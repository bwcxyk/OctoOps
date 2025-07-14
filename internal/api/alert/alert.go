package alert

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"octoops/internal/db"
	alertModel "octoops/internal/model/alert"
	alertService "octoops/internal/service/alert"
)

// Channel相关接口

// ListChannels 获取所有渠道
func ListChannels(c *gin.Context) {
	var channels []alertModel.Channel
	db.DB.Order("created_at desc").Find(&channels)
	c.JSON(http.StatusOK, channels)
}

// CreateChannel 新增渠道
func CreateChannel(c *gin.Context) {
	var channel alertModel.Channel
	if err := c.ShouldBindJSON(&channel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Create(&channel)
	c.JSON(http.StatusOK, channel)
}

// UpdateChannel 更新渠道
func UpdateChannel(c *gin.Context) {
	id := c.Param("id")
	var channel alertModel.Channel
	if err := db.DB.First(&channel, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Model(&channel).Updates(req)
	c.JSON(http.StatusOK, channel)
}

// DeleteChannel 删除渠道
func DeleteChannel(c *gin.Context) {
	id := c.Param("id")
	if err := db.DB.Delete(&alertModel.Channel{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// TestChannel 测试渠道发送
func TestChannel(c *gin.Context) {
	id := c.Param("id")
	var channel alertModel.Channel
	if err := db.DB.First(&channel, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var err error
	switch channel.Type {
	case "email":
		err = alertService.SendTestEmail(&channel)
	case "dingtalk":
		err = alertService.SendTestRobot(&channel)
	case "wechat":
		err = fmt.Errorf("暂未实现企业微信测试发送")
	case "feishu":
		err = fmt.Errorf("暂未实现飞书测试发送")
	default:
		err = fmt.Errorf("未知渠道类型: %s", channel.Type)
	}
	if err != nil {
		log.Printf("渠道测试发送失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "发送失败", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "测试发送成功"})
}

// RegisterChannelRoutes 路由注册函数
func RegisterAlertChannelRoutes(r *gin.RouterGroup) {
	r.GET("/channels", ListChannels)
	r.POST("/channels", CreateChannel)
	r.PUT("/channels/:id", UpdateChannel)
	r.DELETE("/channels/:id", DeleteChannel)
	r.POST("/channels/:id/test", TestChannel)
}

// ListAlertGroups 告警组列表
func ListAlertGroups(c *gin.Context) {
	var groups []alertModel.AlertGroup
	db.DB.Order("created_at desc").Find(&groups)
	c.JSON(http.StatusOK, groups)
}

// CreateAlertGroup 新建告警组
func CreateAlertGroup(c *gin.Context) {
	var group alertModel.AlertGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Create(&group)
	c.JSON(http.StatusOK, group)
}

// UpdateAlertGroup 更新告警组
func UpdateAlertGroup(c *gin.Context) {
	id := c.Param("id")
	var group alertModel.AlertGroup
	if err := db.DB.First(&group, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var req alertModel.AlertGroup
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Model(&group).Updates(map[string]interface{}{
		"name":        req.Name,
		"description": req.Description,
		"status":      req.Status,
	})
	c.JSON(http.StatusOK, group)
}

// DeleteAlertGroup 删除告警组
func DeleteAlertGroup(c *gin.Context) {
	id := c.Param("id")
	db.DB.Delete(&alertModel.AlertGroup{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ListAlertGroupMembers 获取告警组成员
func ListAlertGroupMembers(c *gin.Context) {
	groupID := c.Param("id")
	var members []alertModel.AlertGroupMember
	db.DB.Where("group_id = ?", groupID).Find(&members)
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
	groupIDUint, err := parseUint(groupID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的组ID: " + err.Error()})
		return
	}
	member.GroupID = groupIDUint
	if err := db.DB.Create(&member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建成员失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, member)
}

// DeleteAlertGroupMember 删除告警组成员
func DeleteAlertGroupMember(c *gin.Context) {
	memberID := c.Param("member_id")
	db.DB.Delete(&alertModel.AlertGroupMember{}, memberID)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// 工具函数
func parseUint(s string) (uint, error) {
	var i uint
	_, err := fmt.Sscanf(s, "%d", &i)
	if err != nil {
		return 0, fmt.Errorf("无效的数字格式: %s", s)
	}
	return i, nil
}

// RegisterAlertGroupRoutes 路由注册
func RegisterAlertGroupRoutes(r *gin.RouterGroup) {
	r.GET("/alert-groups", ListAlertGroups)
	r.POST("/alert-groups", CreateAlertGroup)
	r.PUT("/alert-groups/:id", UpdateAlertGroup)
	r.DELETE("/alert-groups/:id", DeleteAlertGroup)
}

// RegisterAlertGroupMemberRoutes 路由注册扩展
func RegisterAlertGroupMemberRoutes(r *gin.RouterGroup) {
	r.GET("/alert-groups/:id/members", ListAlertGroupMembers)
	r.POST("/alert-groups/:id/members", AddAlertGroupMember)
	r.DELETE("/alert-groups/:id/members/:member_id", DeleteAlertGroupMember)
}

// ListAlertTemplates 告警模板列表
func ListAlertTemplates(c *gin.Context) {
	var templates []alertModel.AlertTemplate
	db.DB.Order("created_at desc").Find(&templates)
	c.JSON(http.StatusOK, templates)
}

// CreateAlertTemplate 新建告警模板
func CreateAlertTemplate(c *gin.Context) {
	var tpl alertModel.AlertTemplate
	if err := c.ShouldBindJSON(&tpl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Create(&tpl)
	c.JSON(http.StatusOK, tpl)
}

// UpdateAlertTemplate 更新告警模板
func UpdateAlertTemplate(c *gin.Context) {
	id := c.Param("id")
	var tpl alertModel.AlertTemplate
	if err := db.DB.First(&tpl, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var req alertModel.AlertTemplate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Model(&tpl).Updates(map[string]interface{}{
		"name":    req.Name,
		"type":    req.Type,
		"content": req.Content,
	})
	c.JSON(http.StatusOK, tpl)
}

// DeleteAlertTemplate 删除告警模板
func DeleteAlertTemplate(c *gin.Context) {
	id := c.Param("id")
	db.DB.Delete(&alertModel.AlertTemplate{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// RegisterAlertTemplateRoutes 路由注册
func RegisterAlertTemplateRoutes(r *gin.RouterGroup) {
	r.GET("/alert-templates", ListAlertTemplates)
	r.POST("/alert-templates", CreateAlertTemplate)
	r.PUT("/alert-templates/:id", UpdateAlertTemplate)
	r.DELETE("/alert-templates/:id", DeleteAlertTemplate)
}
