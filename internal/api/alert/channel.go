package alert

import (
	"fmt"
	"log"
	"net/http"
	"octoops/internal/middleware"
	alertModel "octoops/internal/model/alert"
	alertService "octoops/internal/service/alert"
	"strings"

	"github.com/gin-gonic/gin"
)

// ListChannels 获取所有渠道
func ListChannels(c *gin.Context) {
	channels, err := alertService.ListChannels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询渠道失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, channels)
}

// CreateChannel 新增渠道
func CreateChannel(c *gin.Context) {
	var channel alertModel.AlertChannel
	if err := c.ShouldBindJSON(&channel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := alertService.CreateChannel(&channel); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建渠道失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, channel)
}

// UpdateChannel 更新渠道
func UpdateChannel(c *gin.Context) {
	id := c.Param("id")
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	channel, err := alertService.UpdateChannel(id, req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, channel)
}

// DeleteChannel 删除渠道
func DeleteChannel(c *gin.Context) {
	id := c.Param("id")
	if err := alertService.DeleteChannel(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// TestChannel 测试渠道发送
func TestChannel(c *gin.Context) {
	id := c.Param("id")
	channel, err := alertService.GetChannelByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var req struct {
		TemplateContent string `json:"template_content"`
	}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	templateContent := strings.TrimSpace(req.TemplateContent)

	sendErr := testChannelMessage(channel, templateContent)
	if sendErr != nil {
		log.Printf("渠道测试发送失败: %v", sendErr)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "发送失败", "error": sendErr.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "测试发送成功"})
}

func testChannelMessage(channel alertModel.AlertChannel, templateContent string) error {
	if templateContent == "" && channel.TemplateID != 0 {
		if tpl, err := alertService.GetAlertTemplateByID(fmt.Sprintf("%d", channel.TemplateID)); err == nil {
			templateContent = strings.TrimSpace(tpl.Content)
		}
	}

	var err error
	switch channel.Type {
	case "email":
		if templateContent != "" {
			err = alertService.SendEmailWithTemplate(&channel, templateContent, nil)
		} else {
			err = alertService.SendTestEmail(&channel)
		}
	case "dingtalk":
		if templateContent != "" {
			err = alertService.SendDingTalkMarkdownWithTemplate(
				channel.Target,
				channel.DingtalkSecret,
				"OctoOps 测试通知",
				templateContent,
				nil,
			)
		} else {
			err = alertService.SendTestRobot(&channel)
		}
	case "wechat":
		err = fmt.Errorf("暂未实现企业微信测试发送")
	case "feishu":
		err = fmt.Errorf("暂未实现飞书测试发送")
	default:
		err = fmt.Errorf("未知渠道类型: %s", channel.Type)
	}
	return err
}

// RegisterAlertChannelRoutes 路由注册函数
func RegisterAlertChannelRoutes(r *gin.RouterGroup) {
	r.GET("/alert/channel", middleware.AuthMiddleware(), middleware.RequirePermission("notify:channel:read"), ListChannels)
	r.POST("/alert/channel", middleware.AuthMiddleware(), middleware.RequirePermission("notify:channel:create"), CreateChannel)
	r.PUT("/alert/channel/:id", middleware.AuthMiddleware(), middleware.RequirePermission("notify:channel:update"), UpdateChannel)
	r.DELETE("/alert/channel/:id", middleware.AuthMiddleware(), middleware.RequirePermission("notify:channel:delete"), DeleteChannel)
	r.POST("/alert/channel/:id/test", middleware.AuthMiddleware(), middleware.RequirePermission("notify:channel:test"), TestChannel)
}
