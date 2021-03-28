package log

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Recorder
type Recorder interface {
	Record(record ...*Record)
	Close()
}

// FileRotating
type FileRotating interface {
	Rotating(filename string, file *os.File) *os.File
}

// RecordBuffer
type RecordBuffer struct {
	off int
	buf []*Record
	rcd Recorder
}

// StdioRecorder
type StdioRecorder struct {
	formatter Formatter
	closed    bool
	stdio     *os.File
}

// FileRecorder
type FileRecorder struct {
	file      *os.File
	filename  string
	formatter Formatter
	rotating  FileRotating
}

// FileCountRotating
type FileCountRotating struct {
	maxBytes int
	bakCount int
}

// FileTimeRotating
type FileTimeRotating struct {
	align    bool
	interval time.Duration
}

var pool = sync.Pool{
	New: func() interface{} {
		return new(strings.Builder)
	},
}

func (r *Record) Reset() {
	r.Tag = r.Tag[:0]
	r.Level = ""
	r.Message = ""
	r.Time = time.Time{}
}

func NewRecordBuffer(recorder Recorder) *RecordBuffer {
	return NewRecordBufferSize(recorder, 128)
}

func NewRecordBufferSize(recorder Recorder, size int) *RecordBuffer {
	r, ok := recorder.(*RecordBuffer)
	if ok {
		if len(r.buf) != size && size > 0 {
			r.Flush()
			r.buf = make([]*Record, size)
		}
		return r
	}

	if size <= 0 {
		size = 128
	}

	return &RecordBuffer{
		buf: make([]*Record, size),
		rcd: recorder,
	}
}

func (r *RecordBuffer) Record(record ...*Record) {
	for len(record) > r.Available() {
		var n int
		if r.Buffered() == 0 {
			r.rcd.Record(record...)
			n = len(record)
		} else {
			n = copy(r.buf[r.off:], record)
			r.off += n
			r.Flush()
		}
		record = record[n:]
	}

	r.off += copy(r.buf[r.off:], record)
}

func (r *RecordBuffer) Reset(recorder Recorder) {
	r.off = 0
	r.rcd = recorder
}

func (r *RecordBuffer) Size() int {
	return len(r.buf)
}

func (r *RecordBuffer) Buffered() int {
	return r.off
}

func (r *RecordBuffer) Available() int {
	return len(r.buf) - r.off
}

func (r *RecordBuffer) Flush() {
	if r.off <= 0 {
		return
	}
	r.rcd.Record(r.buf[:r.off]...)
	r.off = 0
}

func (r *RecordBuffer) Close() {
	r.Flush()
}

func NewStdoutRecorder(formatter Formatter) *StdioRecorder {
	var s = new(StdioRecorder)
	s.formatter = formatter
	s.stdio = os.Stdout

	return s
}

func NewStderrRecorder(formatter Formatter) *StdioRecorder {
	var s = new(StdioRecorder)
	s.formatter = formatter
	s.stdio = os.Stderr

	return s
}

func (s *StdioRecorder) Record(record ...*Record) {
	if s.closed {
		return
	}
	var builder = pool.Get().(*strings.Builder)
	defer pool.Put(builder)
	defer builder.Reset()

	for _, r := range record {
		_, _ = fmt.Fprintln(builder, s.formatter(r))
	}

	if len(record) > 0 {
		_, _ = fmt.Fprint(s.stdio, builder.String())
	}
}

func (s *StdioRecorder) Close() {
	s.closed = true
	return
}

func NewFileRecorder(filename string, formatter Formatter, rotating FileRotating) (*FileRecorder, error) {
	_ = os.MkdirAll(filepath.Dir(filename), 0755)

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return &FileRecorder{
		file:      file,
		filename:  filename,
		rotating:  rotating,
		formatter: formatter,
	}, nil
}

func (f *FileRecorder) Record(record ...*Record) {
	if f.file == nil {
		return
	}
	if f.file != nil && f.rotating != nil {
		f.file = f.rotating.Rotating(f.filename, f.file)
	}

	var builder = pool.Get().(*strings.Builder)
	defer pool.Put(builder)
	defer builder.Reset()

	for _, r := range record {
		_, _ = fmt.Fprintln(builder, f.formatter(r))
	}

	if len(record) > 0 {
		_, _ = fmt.Fprint(f.file, builder.String())
	}
}

func (f *FileRecorder) Close() {
	if f.file == nil {
		return
	}
	_ = f.file.Close()
	f.file = nil
}

func NewFileCountRotating(maxBytes, bakCount int) *FileCountRotating {
	return &FileCountRotating{
		maxBytes: maxBytes,
		bakCount: bakCount,
	}
}

func (f *FileCountRotating) Rotating(filename string, file *os.File) *os.File {
	// 文件备份数量为0, 则不备份
	if f.bakCount == 0 {
		_ = file.Close()
		return nil
	}

	info, err := file.Stat()
	if err != nil {
		return file
	}

	// 不限制文件大小或不满足rotate条件, 则不rotate
	if f.maxBytes <= 0 || int(info.Size()) < f.maxBytes {
		return file
	}

	if err = file.Close(); err != nil {
		return file
	}

	for i := f.bakCount - 2; i >= 0; i-- {
		oldName := fmt.Sprintf("%s.%04d", filename, i)
		if i == 0 {
			oldName = filename
		}
		newName := fmt.Sprintf("%s.%04d", filename, i+1)
		_ = os.Rename(oldName, newName)
	}

	empty, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return file
	}

	return empty
}

func NewFileTimeRotating(interval time.Duration, align bool) *FileTimeRotating {
	return &FileTimeRotating{
		align:    align,
		interval: interval,
	}
}

func (f *FileTimeRotating) Rotating(filename string, file *os.File) *os.File {
	if f.interval == 0 {
		_ = file.Close()
		return nil
	}

	s, err := file.Stat()
	if err != nil {
		return file
	}
	info := fileInfo(s)

	if info.CreateTime().Add(f.interval).Unix() > time.Now().Unix() {
		return file
	}

	if err = file.Close(); err != nil {
		return file
	}

	name := fmt.Sprintf("%s.%s", filename, info.CreateTime().Format("20060102150405"))

	if f.align {
		switch {
		case f.interval >= time.Hour*24:
			name = fmt.Sprintf("%s.%s", filename, info.CreateTime().Format("20060102000000"))
		case f.interval >= time.Hour:
			name = fmt.Sprintf("%s.%s", filename, info.CreateTime().Format("20060102150000"))
		case f.interval >= time.Minute:
			name = fmt.Sprintf("%s.%s", filename, info.CreateTime().Format("20060102150400"))
		}
	}

	_ = os.Rename(filename, name)

	empty, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return file
	}

	return empty
}
