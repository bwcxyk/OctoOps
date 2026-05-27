//go:build embed_frontend

package assets

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// embed dist
//go:embed all:dist
var distFS embed.FS

var (
	sub        fs.FS
	assetsSub  fs.FS
	indexHTML  []byte
)

func init() {
	var err error

	// dist 根目录
	sub, err = fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}

	// dist/assets
	assetsSub, err = fs.Sub(sub, "assets")
	if err != nil {
		panic(err)
	}

	// 预加载 index.html（避免每次 IO）
	indexHTML, err = fs.ReadFile(sub, "index.html")
	if err != nil {
		panic(err)
	}
}

func SetupFrontend(r *gin.Engine) {

	// =========================
	// 1. 静态资源（关键）
	// =========================
	r.StaticFS("/assets", http.FS(assetsSub))

	// =========================
	// 2. favicon
	// =========================
	r.GET("/favicon.ico", func(c *gin.Context) {
		data, err := fs.ReadFile(sub, "favicon.ico")
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		c.Data(http.StatusOK, "image/x-icon", data)
	})

	// =========================
	// 3. SPA fallback
	// =========================
	r.NoRoute(func(c *gin.Context) {

		path := c.Request.URL.Path

		// API 不吞
		if strings.HasPrefix(path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "API not found",
			})
			return
		}

		// 统一返回 index.html（SPA）
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	})
}
