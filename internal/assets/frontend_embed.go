//go:build embed_frontend

package assets

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed all:dist
var distFS embed.FS

var (
	sub       fs.FS
	indexHTML []byte // 缓存 index.html 提高并发性能
)

func init() {
	var err error
	// 剥离出 dist 目录
	sub, err = fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}

	// 提前将 index.html 读入内存
	indexHTML, err = fs.ReadFile(sub, "index.html")
	if err != nil {
		panic(err)
	}
}

// SetupFrontend registers embedded frontend routes on the Gin engine.
func SetupFrontend(r *gin.Engine) {
	// 1. 直接将 sub 映射到 /assets，Gin 会自动在 sub (即 dist 目录) 下寻找 "assets" 文件夹
	r.StaticFS("/assets", http.FS(sub))

	// 2. Favicon 图标路由
	r.GET("/favicon.ico", func(c *gin.Context) {
		data, err := fs.ReadFile(sub, "favicon.ico")
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		c.Data(http.StatusOK, "image/x-icon", data)
	})

	// 3. SPA 兜底路由 (HTML5 History Mode)
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// 如果是没找到的 API 请求，直接返回 404 JSON，不要返回 index.html
		if strings.HasPrefix(path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"error": "API not found"})
			return
		}

		// 其它所有前端路由（如 /dashboard, /login），都返回 index.html 让前端路由去处理
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	})
}
