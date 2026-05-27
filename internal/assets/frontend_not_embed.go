//go:build !embed_frontend

package assets

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupFrontend registers a catch-all route that informs the user
// the frontend is not bundled. Use 'npm run dev' in web/ for local development.
func SetupFrontend(r *gin.Engine) {
	r.NoRoute(func(c *gin.Context) {
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "API not found"})
			return
		}
		c.String(http.StatusOK, "Frontend is not bundled. Run 'npm run dev' in web/ for local development, or build with '-tags embed_frontend' to embed frontend.")
	})
}
