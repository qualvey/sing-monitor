package api

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed web/*
var webFS embed.FS

// serveEmbedded 托管 go:embed 内嵌前端；未命中时回退 index.html（SPA）
func serveEmbedded(r *gin.Engine) {
	sub, err := fs.Sub(webFS, "web")
	if err != nil {
		return
	}
	fileServer := http.FileServer(http.FS(sub))

	r.NoRoute(func(c *gin.Context) {
		p := strings.TrimPrefix(c.Request.URL.Path, "/")
		if p == "" {
			p = "index.html"
		}
		// 静态资源存在则直接返回，否则回退 SPA index
		if _, err := fs.Stat(sub, p); err == nil {
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}
		// SPA fallback：/xxx → index.html
		if data, err := fs.ReadFile(sub, "index.html"); err == nil {
			c.Data(http.StatusOK, "text/html; charset=utf-8", data)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})
}
