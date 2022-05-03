package micro

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/zooyer/miskit/errors"
)

type Response struct {
	Errno   int         `json:"errno"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type Controller struct{}

type Validator interface {
	Valid(ctx *gin.Context) (err error)
}

var (
	errnoCode  = make(map[int]int)
	errnoMutex sync.Mutex
)

// RegisterErrorCode 错误码与http状态码映射
func RegisterErrorCode(codes map[int]int) {
	errnoMutex.Lock()
	defer errnoMutex.Unlock()

	for errno, code := range codes {
		if _, exists := errnoCode[errno]; exists {
			panic(fmt.Sprintf("code: Define called twice for errno %d", errno))
		}

		errnoCode[errno] = code
	}
}

// NewResponse 创建响应结构
func NewResponse(data interface{}, err error) (resp Response) {
	if err != nil {
		errno := errors.New(errors.UnknownError, err)

		resp.Errno = errno.Errno()

		if err = errno.Unwrap(); err != nil {
			resp.Data = err.Error()
		}
	} else {
		resp.Errno = errors.Success
		resp.Data = data
	}

	resp.Message = errors.Msg(resp.Errno)

	return
}

func (c Controller) Bind(ctx *gin.Context, v interface{}) (err error) {
	if err = ctx.Bind(v); err != nil {
		return errors.New(errors.InvalidRequest, err)
	}

	if validator, ok := v.(Validator); ok {
		if err = validator.Valid(ctx); err != nil {
			return errors.New(errors.InvalidRequest, err)
		}
	}

	return
}

func (c Controller) Response(ctx *gin.Context, data interface{}, err error) {
	if err != nil {
		errno := errors.New(errors.UnknownError, err)
		errno.Record(ctx).Metric()
		err = errno
	}

	resp := NewResponse(data, err)

	var code = http.StatusOK
	if err != nil {
		code = http.StatusInternalServerError
	}

	if enoCode, ok := errnoCode[resp.Errno]; ok {
		code = enoCode
	}

	ctx.AbortWithStatusJSON(code, resp)
}
