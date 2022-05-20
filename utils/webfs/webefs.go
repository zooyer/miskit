package webfs

import (
	"embed"
	"fmt"
	"net/http"
	"path"
	"path/filepath"
	"strings"
)

// webEFS web嵌入文件系统
type webEFS struct {
	Path  string
	Embed http.FileSystem
}

// WebEFS 文件系统目录，嵌入式文件系统
func WebEFS(path string, fs embed.FS) http.FileSystem {
	return &webEFS{
		Path:  path,
		Embed: http.FS(fs),
	}
}

// Open 代码与http.Dir.Open函数相似
func (w *webEFS) Open(filename string) (http.File, error) {
	if filepath.Separator != '/' && strings.ContainsRune(filename, filepath.Separator) {
		return nil, fmt.Errorf("http: invalid character in file path %v", filename)
	}
	dir := w.Path
	if dir == "" {
		dir = "."
	}
	fullName := filepath.Join(dir, filepath.FromSlash(path.Clean("/"+filename)))
	// embed暂不支持windows separator, 统一转为"/"处理
	fullName = strings.ReplaceAll(fullName, "\\", "/")

	return w.Embed.Open(fullName)
}
