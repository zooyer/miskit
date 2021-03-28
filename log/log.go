package log

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/zooyer/miskit/trace"
)

type Config struct {
	Filename string // 文件名
	Output   string // stdout/stderr/file
	Level    string
	Align    bool
	Interval time.Duration
}

type Logger struct {
	config    *Config
	tag       []interface{}
	keep      []interface{}
	recorder  Recorder
	recorders map[string]Recorder
}

type Tag struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type Record struct {
	Time    time.Time `json:"time"`
	Level   string    `json:"level"`
	Tag     []Tag     `json:"tag,omitempty"`
	Message string    `json:"message"`
}

var recordPool = sync.Pool{
	New: func() interface{} {
		return new(Record)
	},
}

var levels = map[string]int{
	"TRACE":   0,
	"DEBUG":   1,
	"INFO":    2,
	"WARNING": 3,
	"ERROR":   4,
	"FATAL":   5,
}

func New(config Config, formatter Formatter) (*Logger, error) {
	var logger = new(Logger)
	logger.config = &config

	switch config.Output {
	case "stdout":
		logger.recorder = NewStdoutRecorder(formatter)
	case "stderr":
		logger.recorder = NewStderrRecorder(formatter)
	case "file":
		rotating := NewFileTimeRotating(config.Interval, config.Align)
		recorder, err := NewFileRecorder(config.Filename, formatter, rotating)
		if err != nil {
			return nil, err
		}
		logger.recorder = recorder
	}

	return logger, nil
}

func (l *Logger) SetRecorder(level string, recorder Recorder) {
	if l.recorders == nil {
		l.recorders = make(map[string]Recorder)
	}
	l.recorders[level] = recorder
}

func (l *Logger) SetDefaultRecorder(recorder Recorder) {
	l.recorder = recorder
}

func (l *Logger) New() *Logger {
	var recorders = make(map[string]Recorder)
	for level, recorder := range l.recorders {
		recorders[level] = recorder
	}

	var keep = make([]interface{}, 0, len(l.keep))
	for _, tag := range l.keep {
		keep = append(keep, tag)
	}

	return &Logger{
		config:    l.config,
		tag:       nil,
		keep:      keep,
		recorder:  l.recorder,
		recorders: recorders,
	}
}

func (l *Logger) Tag(keep bool, kv ...interface{}) *Logger {
	if len(kv) == 0 {
		return l
	}

	if keep {
		l.keep = append(l.keep, kv...)
	} else {
		l.tag = append(l.tag, kv...)
	}

	return l
}

func (l *Logger) trace(ctx context.Context, record *Record) {
	if record == nil {
		return
	}

	if t := trace.Get(ctx); t != nil {
		if t.TraceID != "" {
			record.Tag = append(record.Tag, Tag{Key: "trace_id", Value: t.TraceID})
		}
		if t.SpanID != "" {
			record.Tag = append(record.Tag, Tag{Key: "span_id", Value: t.SpanID})
		}
		if t.Lang != "" {
			record.Tag = append(record.Tag, Tag{Key: "lang", Value: t.Lang})
		}
		if t.Tag != "" {
			record.Tag = append(record.Tag, Tag{Key: "tag", Value: t.Tag})
		}
		if len(t.Content) != 0 {
			record.Tag = append(record.Tag, Tag{Key: "content", Value: t.Content})
		}
	}
}

func (l *Logger) format(v ...interface{}) string {
	if len(v) == 0 {
		return ""
	}

	var builder = pool.Get().(*strings.Builder)
	defer pool.Put(builder)
	defer builder.Reset()

	for i, arg := range v {
		if i > 0 {
			builder.WriteByte(' ')
		}
		builder.WriteString(fmt.Sprint(arg))
	}

	return builder.String()
}

func (l *Logger) canRecord(level string) bool {
	if len(l.recorders) == 0 && l.recorder == nil {
		return false
	}

	if levels[level] >= levels[l.config.Level] {
		return true
	}

	return false
}

func (l *Logger) toTag(kv []interface{}) []Tag {
	if len(kv) == 0 {
		return nil
	}

	tag := make([]Tag, 0, len(kv))

	var key string
	for i := 0; i < len(kv)-1; i += 2 {
		switch k := kv[i].(type) {
		case string:
			key = k
		default:
			key = fmt.Sprint(k)
		}
		if val := kv[i+1]; val != nil && key != "" {
			tag = append(tag, Tag{
				Key:   key,
				Value: val,
			})
		}
	}

	return tag
}

func (l *Logger) output(ctx context.Context, level string, v ...interface{}) {
	if !l.canRecord(level) {
		return
	}

	var record = recordPool.Get().(*Record)
	defer recordPool.Put(record)
	defer record.Reset()

	record.Time = time.Now()
	record.Level = level
	record.Message = l.format(v...)
	l.trace(ctx, record)
	record.Tag = append(record.Tag, l.toTag(l.keep)...)
	record.Tag = append(record.Tag, l.toTag(l.tag)...)
	l.tag = l.tag[:0]

	l.record(record)
}

func (l *Logger) record(record *Record) {
	recorder, ok := l.recorders[record.Level]
	if !ok {
		recorder = l.recorder
	}
	if recorder != nil {
		recorder.Record(record)
	}
}

func (l *Logger) Debug(ctx context.Context, v ...interface{}) {
	l.output(ctx, "DEBUG", v...)
}

func (l *Logger) Info(ctx context.Context, v ...interface{}) {
	l.output(ctx, "INFO", v...)
}

func (l *Logger) Warning(ctx context.Context, v ...interface{}) {
	l.output(ctx, "WARNING", v...)
}

func (l *Logger) Error(ctx context.Context, v ...interface{}) {
	l.output(ctx, "ERROR", v...)
}

func (l *Logger) Fatal(ctx context.Context, v ...interface{}) {
	l.output(ctx, "FATAL", v...)
	os.Exit(0)
}
