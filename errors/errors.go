package errors

import (
	"context"
	"fmt"
	"sync"

	"github.com/zooyer/miskit/metric"
)

type Error struct {
	errno   int
	error   error
	message string

	record bool
	metric bool
}

const (
	Success        = 0
	InvalidRequest = 1
	UnknownError   = 3
	ServicePanic   = 4
)

var mutex sync.Mutex

var prefix = "miskit"

var msg = map[int]string{
	Success:        "ok",
	InvalidRequest: "请求无效",
	UnknownError:   "未知错误",
	ServicePanic:   "程序崩溃",
}

func (e Error) String() string {
	if e.error == nil {
		return fmt.Sprintf("%s errno: %d, message:%s", prefix, e.errno, e.message)
	}
	return fmt.Sprintf("%s errno: %d, message:%s, error:%v", prefix, e.errno, e.message, e.error)
}

func (e Error) Errno() int {
	return e.errno
}

func (e Error) Error() string {
	return e.String()
}

func (e Error) Unwrap() error {
	return e.error
}

func (e Error) Metric() {
	if e.metric {
		return
	}
	e.metric = true

	metric.Count("errors::"+prefix, 1, map[string]interface{}{
		"errno":   e.errno,
		"message": e.message,
	})
}

func (e Error) Record(ctx context.Context) Error {
	if e.record {
		return e
	}
	e.record = true
	// TODO LOG
	fmt.Println("LOG::errors:", e)
	return e
}

func Register(name string, errno map[int]string) {
	mutex.Lock()
	defer mutex.Unlock()

	prefix = name

	for errno, message := range errno {
		if _, exists := msg[errno]; exists {
			panic(fmt.Sprintf("errno: Define called twice for errno %d", errno))
		}
		msg[errno] = message
	}
}

func New(errno int, error error) Error {
	if err, ok := error.(Error); ok {
		return err
	}

	return Error{
		errno:   errno,
		error:   error,
		message: Msg(errno),
		record:  false,
		metric:  false,
	}
}

func Msg(errno int) string {
	if msg, exists := msg[errno]; exists {
		return msg
	}
	return "未定义错误码"
}

func Is(err error, errno int) bool {
	if e, ok := err.(Error); ok {
		return e.errno == errno
	}
	return false
}
