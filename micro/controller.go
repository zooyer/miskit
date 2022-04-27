package micro

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zooyer/miskit/errors"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type Controller struct{}

type Validator interface {
	Valid(ctx *gin.Context) error
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

func (c Controller) Response(ctx *gin.Context, err error, data interface{}) {
	var resp Response
	if err != nil {
		eno, ok := err.(errors.Error)
		if !ok {
			eno = errors.New(errors.UnknownError, err)
		}
		// 记录错误日志，上报metric
		eno.Record(ctx).Metric()
		// 直接返回错误码
		resp.Code = eno.Errno()
		if err = eno.Unwrap(); err != nil {
			resp.Data = err.Error()
		}
	} else {
		resp.Data = data
	}
	resp.Message = errors.Msg(resp.Code)
	ctx.JSON(http.StatusOK, resp)
	ctx.Abort()
}
