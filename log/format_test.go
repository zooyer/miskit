package log

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestFormat(t *testing.T) {
	var record = new(Record)
	record.Time = time.Now()
	record.Level = "DEBUG"
	record.Tag = []Tag{
		{"module", "apollo"},
		{"trace", "10001241235"},
		{"type", "rpc"},
		{"rpc", "http"},
		{"latency", "10"},
		{"other", json.RawMessage(`{"json":"value", " ": "   "}`)},
	}
	record.Message = "hello world"

	t.Log(JSONFormatter(false)(record))
	t.Log(TextFormatter(false)(record))
	t.Log(TextFormatter(true)(record))
	record.Level = "INFO"
	t.Log(TextFormatter(true)(record))
	record.Level = "WARNING"
	t.Log(TextFormatter(true)(record))
	record.Level = "ERROR"
	t.Log(TextFormatter(true)(record))
}

func TestFormatText(t *testing.T) {
	for i := 0; i < 100; i++ {
		fmt.Println(TextFormatter(true)(&Record{
			Level:   "DEBUG",
			Time:    time.Now(),
			Message: "null",
			Tag: []Tag{
				{Key: "1", Value: ""},
				{Key: "2", Value: ""},
				{Key: "3", Value: ""},
				{Key: "4", Value: ""},
				{Key: "5", Value: ""},
				{Key: "6", Value: ""},
				{Key: "7", Value: ""},
				{Key: "8", Value: ""},
			},
		}))
	}
}
