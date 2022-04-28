package micro

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/zooyer/jsons"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func get(engine *gin.Engine, uri string, header ...http.Header) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", uri, nil)
	for _, header := range header {
		for key, header := range header {
			req.Header[key] = header
		}
	}
	res := httptest.NewRecorder()
	engine.ServeHTTP(res, req)
	return res
}

func postJson(engine *gin.Engine, uri string, data interface{}) *httptest.ResponseRecorder {
	body, _ := json.Marshal(data)
	req := httptest.NewRequest("POST", uri, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	engine.ServeHTTP(res, req)
	return res
}

type Student struct {
	Name  string `json:"name" binding:"required"`
	Age   int    `json:"age"`
	Phone string `json:"phone"`
}

func (s *Student) Valid(ctx *gin.Context) (err error) {
	if s.Age == 0 {
		s.Age = 25
	}

	if s.Phone == "" {
		return errors.New("invalid phone")
	}

	return
}

type studentController struct {
	Controller
}

func (s studentController) Valid(ctx *gin.Context) {
	var (
		err error
		req Student
	)

	defer func() { s.Response(ctx, req, err) }()

	if err = s.Bind(ctx, &req); err != nil {
		return
	}
}

func TestController_Bind(t *testing.T) {
	engine := gin.Default()

	stu := new(studentController)

	engine.POST("/valid", stu.Valid)

	var tests = []struct {
		Input Student
		Equal bool
	}{
		{
			Input: Student{
				Name:  "张三",
				Age:   0,
				Phone: "",
			},
			Equal: false,
		},
		{
			Input: Student{
				Name:  "李四",
				Age:   16,
				Phone: "",
			},
			Equal: false,
		},
		{
			Input: Student{
				Name:  "王五",
				Age:   0,
				Phone: "999",
			},
			Equal: true,
		},
		{
			Input: Student{
				Name:  "赵六",
				Age:   15,
				Phone: "111",
			},
			Equal: true,
		},
	}

	var resp *httptest.ResponseRecorder
	var data []byte

	for _, test := range tests {
		resp = postJson(engine, "/valid", test.Input)
		data, _ = ioutil.ReadAll(resp.Body)
		assert.Equal(t, http.StatusOK, resp.Code)
		if test.Equal {
			assert.Equal(t, "0", jsons.Raw(data).Number("code").JSONString())
		} else {
			assert.NotEqual(t, "0", jsons.Raw(data).Number("code").JSONString())
		}
		t.Log(string(data))
	}
}
