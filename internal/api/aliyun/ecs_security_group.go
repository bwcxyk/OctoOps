package aliyun

import (
	"errors"
	"net/http"
	"octoops/internal/middleware"
	aliyunModel "octoops/internal/model/aliyun"
	aliyunService "octoops/internal/service/aliyun"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 获取所有安全组配置
func ListAliyunSGConfigs(c *gin.Context) {
	configs, err := aliyunService.ListEcsSecurityGroupConfigs(c.Query("status"), c.Query("access_key"), c.Query("name"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询安全组配置失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, configs)
}

// 新增安全组配置
func CreateAliyunSGConfig(c *gin.Context) {
	var cfg aliyunModel.SGConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := aliyunService.CreateEcsSecurityGroupConfig(&cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

// 更新安全组配置
func UpdateAliyunSGConfig(c *gin.Context) {
	id := c.Param("id")
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cfg, err := aliyunService.UpdateEcsSecurityGroupConfig(id, req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

// 删除安全组配置
func DeleteAliyunSGConfig(c *gin.Context) {
	id := c.Param("id")
	if err := aliyunService.DeleteEcsSecurityGroupConfig(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// 单条同步安全组端口到阿里云
func SyncAliyunSGConfig(c *gin.Context) {
	id := c.Param("id")
	if err := aliyunService.SyncEcsSecurityGroupConfigByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "同步成功"})
}

// 路由注册函数
func RegisterAliyunRoutes(r *gin.RouterGroup) {
	r.GET("/aliyun/ecs-sg", middleware.AuthMiddleware(), middleware.RequirePermission("aliyun:ecs_sg:read"), ListAliyunSGConfigs)
	r.POST("/aliyun/ecs-sg", middleware.AuthMiddleware(), middleware.RequirePermission("aliyun:ecs_sg:create"), CreateAliyunSGConfig)
	r.PUT("/aliyun/ecs-sg/:id", middleware.AuthMiddleware(), middleware.RequirePermission("aliyun:ecs_sg:update"), UpdateAliyunSGConfig)
	r.DELETE("/aliyun/ecs-sg/:id", middleware.AuthMiddleware(), middleware.RequirePermission("aliyun:ecs_sg:delete"), DeleteAliyunSGConfig)
	r.POST("/aliyun/ecs-sg/:id/sync", middleware.AuthMiddleware(), middleware.RequirePermission("aliyun:ecs_sg:sync"), SyncAliyunSGConfig)
}
