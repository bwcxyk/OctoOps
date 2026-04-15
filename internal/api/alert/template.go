package alert

import (
	"net/http"
	"octoops/internal/middleware"
	alertModel "octoops/internal/model/alert"
	alertService "octoops/internal/service/alert"

	"github.com/gin-gonic/gin"
)

// ListAlertTemplates 告警模板列表
func ListAlertTemplates(c *gin.Context) {
	templates, err := alertService.ListAlertTemplates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询模板失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, templates)
}

// CreateAlertTemplate 新建告警模板
func CreateAlertTemplate(c *gin.Context) {
	var tpl alertModel.AlertTemplate
	if err := c.ShouldBindJSON(&tpl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := alertService.CreateAlertTemplate(&tpl); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建模板失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, tpl)
}

// UpdateAlertTemplate 更新告警模板
func UpdateAlertTemplate(c *gin.Context) {
	id := c.Param("id")
	var req alertModel.AlertTemplate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tpl, err := alertService.UpdateAlertTemplate(id, req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, tpl)
}

// DeleteAlertTemplate 删除告警模板
func DeleteAlertTemplate(c *gin.Context) {
	id := c.Param("id")
	if err := alertService.DeleteAlertTemplate(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// RegisterAlertTemplateRoutes 路由注册
func RegisterAlertTemplateRoutes(r *gin.RouterGroup) {
	r.GET("/alert/template", middleware.AuthMiddleware(), middleware.RequirePermission("notify:template:read"), ListAlertTemplates)
	r.POST("/alert/template", middleware.AuthMiddleware(), middleware.RequirePermission("notify:template:create"), CreateAlertTemplate)
	r.PUT("/alert/template/:id", middleware.AuthMiddleware(), middleware.RequirePermission("notify:template:update"), UpdateAlertTemplate)
	r.DELETE("/alert/template/:id", middleware.AuthMiddleware(), middleware.RequirePermission("notify:template:delete"), DeleteAlertTemplate)
}
