package antd

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type antFS struct {
	http.FileSystem
	route map[string]bool
}

func AntFS(fs http.FileSystem, route ...string) *antFS {
	var routes = make(map[string]bool)
	for _, r := range route {
		routes[r] = true
	}
	return &antFS{
		FileSystem: fs,
		route:      routes,
	}
}

func (a *antFS) Open(filename string) (http.File, error) {
	if a.route[filename] {
		return a.FileSystem.Open("index.html")
	}
	return a.FileSystem.Open(filename)
}

func (a *antFS) Router(route string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.FileFromFS(route, a)
	}
}

func (a *antFS) Mount(relativePath string, route gin.IRouter) {
	for r := range a.route {
		route.HEAD(r, a.Router(r))
		route.GET(r, a.Router(r))
	}
	route.StaticFS(relativePath, a)
}

func (a *antFS) Redirect(src, dest string, router gin.IRouter) {
	router.GET(src, func(ctx *gin.Context) {
		ctx.Redirect(http.StatusFound, dest)
	})
}
