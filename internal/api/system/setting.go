package system

import (
	"encoding/json"
	"net/http"
	"octoops/internal/middleware"
	systemService "octoops/internal/service/system"

	"github.com/gin-gonic/gin"
)

// GetSettings 查询全部系统设置项
func GetSettings(c *gin.Context) {
	data, err := systemService.GetSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取系统设置失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// UpdateSettings 批量更新设置项
func UpdateSettings(c *gin.Context) {
	var req map[string]json.RawMessage
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "没有需要更新的内容"})
		return
	}
	if err := systemService.UpdateSettings(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data, err := systemService.GetSettings()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// RegisterSystemRoutes 注册系统设置相关路由
func RegisterSystemRoutes(r *gin.RouterGroup) {
	r.GET("/system/settings", middleware.AuthMiddleware(), middleware.RequirePermission("system:setting:read"), GetSettings)
	r.PUT("/system/settings", middleware.AuthMiddleware(), middleware.RequirePermission("system:setting:update"), UpdateSettings)
}
