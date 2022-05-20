package webfs

import (
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"
	"time"
)

type memFile struct {
	name   string
	data   []byte
	offset int64
}

type webMFS struct {
	files map[string][]byte
}

func (f *memFile) Readdir(count int) ([]fs.FileInfo, error) {
	return nil, nil
}

func (f *memFile) Stat() (fs.FileInfo, error) {
	return f, nil
}

func (f *memFile) Read(buf []byte) (int, error) {
	if f.offset == f.Size() {
		return 0, io.EOF
	}

	var n = copy(buf, f.data[f.offset:])
	f.offset += int64(n)

	return n, nil
}

func (f *memFile) Seek(offset int64, whence int) (int64, error) {
	var size = f.Size()

	var startOffset int64

	switch whence {
	case io.SeekStart:
		startOffset = 0
	case io.SeekCurrent:
		startOffset = f.offset
	case io.SeekEnd:
		startOffset = size
	}

	if startOffset+offset < 0 {
		return 0, errors.New("an attempt was made to move the file pointer before the beginning of the file")
	}

	if startOffset+offset < size {
		f.offset = offset
	} else {
		f.offset = size
	}

	return f.offset, nil
}

func (f *memFile) Close() error {
	return nil
}

func (f *memFile) Name() string {
	return f.name
}

func (f *memFile) Size() int64 {
	return int64(len(f.data))
}

func (f *memFile) Mode() fs.FileMode {
	return 0666
}

func (f *memFile) ModTime() time.Time {
	return time.Now()
}

func (f *memFile) IsDir() bool {
	return false
}

func (f *memFile) Sys() interface{} {
	return f
}

// WebMFS 文件系统目录，嵌入式文件系统
func WebMFS(files map[string][]byte) http.FileSystem {
	return &webMFS{
		files: files,
	}
}

func (w *webMFS) Open(name string) (http.File, error) {
	data, exists := w.files[name]
	if !exists {
		return nil, os.ErrNotExist
	}

	return &memFile{name: name, data: data}, nil
}
