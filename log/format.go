package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type Formatter func(record *Record) string

const timeLayout = "2006-01-02 15:04:05.000"

// {"time":"2020-10-25T16:36:04.4143659+08:00","level":"DEBUG","tag":[{"key":"module","value":"apollo"},{"key":"trace","value":"10001241235"},{"key":"type","value":"rpc"},{"key":"rpc","value":"http"},{"key":"latency","value":"10"},{"key":"other","value":{"json":"value"," ":"   "}}],"message":"hello world"}
func JSONFormatter(indent bool) Formatter {
	return func(record *Record) string {
		data, _ := json.Marshal(record)
		if indent {
			var buf bytes.Buffer
			if err := json.Indent(&buf, data, "", "  "); err == nil {
				data = buf.Bytes()
			}
		}
		return string(data)
	}
}

//    date        time     level      tag                                           message
// 2019-10-27 00:04:36.028 DEBUG "trace"="10001241235" "type"="rpc" "rpc"="http" "hello world"
func TextFormatter(color bool) Formatter {
	return func(record *Record) string {
		var builder = pool.Get().(*strings.Builder)
		defer pool.Put(builder)
		defer builder.Reset()

		// time
		var str = record.Time.Format(timeLayout)
		if color {
			str = yellow(str)
		}
		builder.WriteString(str)
		builder.WriteString(" ")

		// level
		if str = record.Level; color {
			switch record.Level {
			case "DEBUG":
				str = green(str)
			case "INFO":
				str = blue(str)
			case "WARNING":
				str = orange(str)
			case "ERROR":
				str = red(str)
			case "FATAL":
				str = red(str)
			default:
				str = red(str)
			}
		}
		builder.WriteString(str)
		builder.WriteString(" ")

		// tag
		if tag := record.Tag; len(tag) > 0 {
			var write bool
			for i := range tag {
				if tag[i].Value == nil {
					continue
				}

				key := tag[i].Key
				var val string
				switch v := tag[i].Value.(type) {
				case string:
					val = v
				case json.RawMessage:
					if data, err := json.Marshal(v); err == nil {
						val = string(data)
					} else {
						val = string(v)
					}
				default:
					val = fmt.Sprint(tag[i].Value)
				}
				if key == "" || val == "" {
					continue
				}
				if color {
					key = cyan(key)
					val = cyan(val)
				}
				if i != 0 {
					builder.WriteString(" ")
				}
				builder.WriteString(key)
				builder.WriteString("=")
				builder.WriteString(val)
				if !write {
					write = true
				}
			}
			if write {
				builder.WriteString(" ")
			}
		}

		// message
		if str = strconv.Quote(record.Message); len(str)-2 == len(record.Message) {
			str = record.Message
		}
		builder.WriteString(str)

		return builder.String()
	}
}

// colors render
func green(s string) string {
	return "\033[32m" + s + "\033[0m"
}

func red(s string) string {
	return "\033[31m" + s + "\033[0m"
}

func blue(s string) string {
	return "\033[34m" + s + "\033[0m"
}

func orange(s string) string {
	return "\033[35m" + s + "\033[0m"
}

func yellow(s string) string {
	return "\033[33m" + s + "\033[0m"
}

func cyan(s string) string {
	return "\033[36m" + s + "\033[0m"
}
