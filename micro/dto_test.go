package micro

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type testController struct {
	Controller
}

func (t testController) Valid(ctx *gin.Context) {
	var (
		err   error
		query Query
	)

	defer func() { t.Response(ctx, query, err) }()

	if err = t.Bind(ctx, &query); err != nil {
		return
	}
}

func TestQuery_Valid(t *testing.T) {
	engine := gin.Default()
	var test testController

	engine.GET("/valid", test.Valid)

	var get = func(uri string) *httptest.ResponseRecorder {
		return get(engine, uri)
	}

	var resp *httptest.ResponseRecorder

	resp = get("/valid?size=1&sort=id%20ASC&select=name&name=zhangsan&age=24&omit=age&page=1&where=phone%20%3D%20132")
	assert.Equal(t, http.StatusOK, resp.Code)
	data, _ := ioutil.ReadAll(resp.Body)
	t.Log(string(data))
}

func TestQuery_ByQuery(t *testing.T) {
	type Stu struct {
		gorm.Model
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Level int    `json:"level"`
		Field string `json:"field"`
	}

	var query = Query{
		Page:   2,
		Size:   1,
		Sort:   "id ASC",
		Omit:   []string{"level", "field", "ab"},
		Select: []string{"name", "age", "field"},
		Where:  "level < 10",
	}

	dial := sqlite.Open(":memory:")
	db, err := gorm.Open(dial)
	if err != nil {
		t.Fatal(err)
	}

	db = db.Debug()

	if err = db.AutoMigrate(&Stu{}); err != nil {
		t.Fatal(err)
	}

	db.Create(&Stu{Name: "xiaozhang", Age: 28, Level: 3, Field: "abc"})
	db.Create(&Stu{Name: "xiaoli", Age: 19, Level: 32, Field: "as"})
	db.Create(&Stu{Name: "xiaowang", Age: 22, Level: 5, Field: "bf"})
	db.Create(&Stu{Name: "xiaohong", Age: 18, Level: 1, Field: "sad"})

	var stu []Stu
	if err := db.Scopes(query.ByQuery).Find(&stu).Error; err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 1, len(stu), "len(stu)")
	assert.Equal(t, stu[0].Name, "xiaowang")
	assert.Equal(t, stu[0].Age, 22)
	assert.Equal(t, stu[0].Field, "bf")
	assert.Equal(t, stu[0].Level, 0)
	assert.Equal(t, stu[0].Model, gorm.Model{})

	data, err := json.MarshalIndent(stu, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(data))
}
