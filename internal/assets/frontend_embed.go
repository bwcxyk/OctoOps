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
	distRoot  fs.FS
	assetsFS  fs.FS
	indexHTML []byte
	favicon   []byte
)

func init() {
	var err error

	// 1. 获取 dist 根目录
	distRoot, err = fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}

	// 2. 切出 assets 子目录，精准映射静态资源
	assetsFS, err = fs.Sub(distRoot, "assets")
	if err != nil {
		panic(err)
	}

	// 3. 预加载 index.html
	indexHTML, err = fs.ReadFile(distRoot, "index.html")
	if err != nil {
		panic(err)
	}

	// 4. 预加载 favicon.ico（如果存在）
	favicon, err = fs.ReadFile(distRoot, "favicon.ico")
	if err != nil {
		// 如果前端打包没生成 favicon，保留为 nil 即可，避免程序直接崩溃
		favicon = nil
	}
}

// SetupFrontend 注册前端托管路由
func SetupFrontend(r *gin.Engine) {

	// 静态资源路由，直接映射到 dist/assets 目录
	r.StaticFS("/assets", http.FS(assetsFS))

	// 处理 favicon.ico 请求
	r.GET("/favicon.ico", func(c *gin.Context) {
		if favicon == nil {
			c.Status(http.StatusNotFound)
			return
		}
		c.Data(http.StatusOK, "image/x-icon", favicon)
	})

	// 其他所有路由都返回 index.html，由前端路由接管
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// API 路由未命中，直接返回标准 404 JSON
		if strings.HasPrefix(path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "API not found",
			})
			return
		}

		// 其余所有页面路由统一返回内存中的 index.html，交由前端路由接管
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	})
}
