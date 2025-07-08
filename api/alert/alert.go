package alert

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"octoops/db"
	alertModel "octoops/model/alert"
	alertService "octoops/service/alert"
)

// 获取所有通知
func ListAlerts(c *gin.Context) {
	var alerts []alertModel.Alert
	db.DB.Order("created_at desc").Find(&alerts)
	c.JSON(http.StatusOK, alerts)
}

// 新增通知
func CreateAlert(c *gin.Context) {
	var alert alertModel.Alert
	if err := c.ShouldBindJSON(&alert); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Create(&alert)
	c.JSON(http.StatusOK, alert)
}

// 更新通知
func UpdateAlert(c *gin.Context) {
	id := c.Param("id")
	var alert alertModel.Alert
	if err := db.DB.First(&alert, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Model(&alert).Updates(req)
	c.JSON(http.StatusOK, alert)
}

// 删除通知
func DeleteAlert(c *gin.Context) {
	id := c.Param("id")
	if err := db.DB.Delete(&alertModel.Alert{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// 测试通知发送
func TestAlert(c *gin.Context) {
	id := c.Param("id")
	var alert alertModel.Alert
	if err := db.DB.First(&alert, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var err error
	switch alert.Type {
	case "email":
		err = alertService.SendTestEmail(&alert)
	case "dingtalk":
		err = alertService.SendTestRobot(&alert)
	case "wechat":
		err = fmt.Errorf("暂未实现企业微信测试发送")
	case "feishu":
		err = fmt.Errorf("暂未实现飞书测试发送")
	default:
		err = fmt.Errorf("未知通知类型: %s", alert.Type)
	}
	if err != nil {
		log.Printf("通知测试发送失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "发送失败", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "测试发送成功"})
}

// 告警组列表
func ListAlertGroups(c *gin.Context) {
	var groups []alertModel.AlertGroup
	db.DB.Order("created_at desc").Find(&groups)
	c.JSON(http.StatusOK, groups)
}

// 新建告警组
func CreateAlertGroup(c *gin.Context) {
	var group alertModel.AlertGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Create(&group)
	c.JSON(http.StatusOK, group)
}

// 更新告警组
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

// 删除告警组
func DeleteAlertGroup(c *gin.Context) {
	id := c.Param("id")
	db.DB.Delete(&alertModel.AlertGroup{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// 获取告警组成员
func ListAlertGroupMembers(c *gin.Context) {
	groupID := c.Param("id")
	var members []alertModel.AlertGroupMember
	db.DB.Where("group_id = ?", groupID).Find(&members)
	c.JSON(http.StatusOK, members)
}

// 添加告警组成员
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

// 删除告警组成员
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

// 路由注册函数
func RegisterAlertRoutes(r *gin.RouterGroup) {
	r.GET("/alerts", ListAlerts)
	r.POST("/alerts", CreateAlert)
	r.PUT("/alerts/:id", UpdateAlert)
	r.DELETE("/alerts/:id", DeleteAlert)
	r.POST("/alerts/:id/test", TestAlert)
}

// 路由注册
func RegisterAlertGroupRoutes(r *gin.RouterGroup) {
	r.GET("/alert-groups", ListAlertGroups)
	r.POST("/alert-groups", CreateAlertGroup)
	r.PUT("/alert-groups/:id", UpdateAlertGroup)
	r.DELETE("/alert-groups/:id", DeleteAlertGroup)
}

// 路由注册扩展
func RegisterAlertGroupMemberRoutes(r *gin.RouterGroup) {
	r.GET("/alert-groups/:id/members", ListAlertGroupMembers)
	r.POST("/alert-groups/:id/members", AddAlertGroupMember)
	r.DELETE("/alert-groups/:id/members/:member_id", DeleteAlertGroupMember)
}

// 告警模板列表
func ListAlertTemplates(c *gin.Context) {
	var templates []alertModel.AlertTemplate
	db.DB.Order("created_at desc").Find(&templates)
	c.JSON(http.StatusOK, templates)
}

// 新建告警模板
func CreateAlertTemplate(c *gin.Context) {
	var tpl alertModel.AlertTemplate
	if err := c.ShouldBindJSON(&tpl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Create(&tpl)
	c.JSON(http.StatusOK, tpl)
}

// 更新告警模板
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

// 删除告警模板
func DeleteAlertTemplate(c *gin.Context) {
	id := c.Param("id")
	db.DB.Delete(&alertModel.AlertTemplate{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// 路由注册
func RegisterAlertTemplateRoutes(r *gin.RouterGroup) {
	r.GET("/alert-templates", ListAlertTemplates)
	r.POST("/alert-templates", CreateAlertTemplate)
	r.PUT("/alert-templates/:id", UpdateAlertTemplate)
	r.DELETE("/alert-templates/:id", DeleteAlertTemplate)
}
