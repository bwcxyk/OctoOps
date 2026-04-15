package alert

import (
	"errors"
	"net/http"
	"octoops/internal/middleware"
	alertModel "octoops/internal/model/alert"
	alertService "octoops/internal/service/alert"

	"github.com/gin-gonic/gin"
)

// ListAlertGroupMembers 获取告警组成员
func ListAlertGroupMembers(c *gin.Context) {
	groupID := c.Param("id")
	members, err := alertService.ListAlertGroupMembers(groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询成员失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, members)
}

// AddAlertGroupMember 添加告警组成员
func AddAlertGroupMember(c *gin.Context) {
	groupID := c.Param("id")
	var member alertModel.AlertGroupMember
	if err := c.ShouldBindJSON(&member); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	groupIDUint, err := alertService.ParseUint(groupID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的组ID: " + err.Error()})
		return
	}
	if err := alertService.CreateAlertGroupMember(groupIDUint, &member); err != nil {
		if errors.Is(err, alertService.ErrAlertGroupMemberExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "成员已存在，不能重复添加"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建成员失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, member)
}

// DeleteAlertGroupMember 删除告警组成员
func DeleteAlertGroupMember(c *gin.Context) {
	memberID := c.Param("member_id")
	if err := alertService.DeleteAlertGroupMember(memberID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// RegisterAlertGroupMemberRoutes 路由注册扩展
func RegisterAlertGroupMemberRoutes(r *gin.RouterGroup) {
	r.GET("/alert/group/:id/members", middleware.AuthMiddleware(), middleware.RequirePermission("notify:group:member:read"), ListAlertGroupMembers)
	r.POST("/alert/group/:id/members", middleware.AuthMiddleware(), middleware.RequirePermission("notify:group:member:create"), AddAlertGroupMember)
	r.DELETE("/alert/group/:id/members/:member_id", middleware.AuthMiddleware(), middleware.RequirePermission("notify:group:member:delete"), DeleteAlertGroupMember)
}
