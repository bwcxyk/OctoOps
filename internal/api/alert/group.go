package alert

import (
	"net/http"
	"octoops/internal/middleware"
	alertModel "octoops/internal/model/alert"
	alertService "octoops/internal/service/alert"

	"github.com/gin-gonic/gin"
)

// ListAlertGroups 告警组列表
func ListAlertGroups(c *gin.Context) {
	groups, err := alertService.ListAlertGroups()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询告警组失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, groups)
}

// CreateAlertGroup 新建告警组
func CreateAlertGroup(c *gin.Context) {
	var group alertModel.AlertGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := alertService.CreateAlertGroup(&group); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建告警组失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, group)
}

// UpdateAlertGroup 更新告警组
func UpdateAlertGroup(c *gin.Context) {
	id := c.Param("id")
	var req alertModel.AlertGroup
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	group, err := alertService.UpdateAlertGroup(id, req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, group)
}

// DeleteAlertGroup 删除告警组
func DeleteAlertGroup(c *gin.Context) {
	id := c.Param("id")
	if err := alertService.DeleteAlertGroup(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// RegisterAlertGroupRoutes 路由注册
func RegisterAlertGroupRoutes(r *gin.RouterGroup) {
	r.GET("/alert/group", middleware.AuthMiddleware(), middleware.RequirePermission("notify:group:read"), ListAlertGroups)
	r.POST("/alert/group", middleware.AuthMiddleware(), middleware.RequirePermission("notify:group:create"), CreateAlertGroup)
	r.PUT("/alert/group/:id", middleware.AuthMiddleware(), middleware.RequirePermission("notify:group:update"), UpdateAlertGroup)
	r.DELETE("/alert/group/:id", middleware.AuthMiddleware(), middleware.RequirePermission("notify:group:delete"), DeleteAlertGroup)
}
