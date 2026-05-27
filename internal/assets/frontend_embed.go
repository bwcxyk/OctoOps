//go:build embed_frontend

package assets

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed all:dist
var distFS embed.FS

var (
	frontendFS http.FileSystem
	assetsFS   http.FileSystem
)

func init() {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic("failed to create sub filesystem: " + err.Error())
	}
	frontendFS = http.FS(sub)
	assetSub, err := fs.Sub(sub, "assets")
	if err != nil {
		panic("failed to create assets sub filesystem: " + err.Error())
	}
	assetsFS = http.FS(assetSub)
}

// SetupFrontend registers embedded frontend routes on the Gin engine.
func SetupFrontend(r *gin.Engine) {
	r.StaticFS("/assets", assetsFS)

	r.GET("/favicon.ico", func(c *gin.Context) {
		c.FileFromFS("/favicon.ico", frontendFS)
	})

	r.GET("/", func(c *gin.Context) {
		c.FileFromFS("/index.html", frontendFS)
	})

	r.NoRoute(func(c *gin.Context) {
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "API not found"})
			return
		}
		c.FileFromFS("/index.html", frontendFS)
	})
}
