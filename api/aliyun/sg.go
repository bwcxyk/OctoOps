package aliyun

import (
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"octoops/db"
	aliyunModel "octoops/model/aliyun"
	aliyunService "octoops/service/aliyun"
	"octoops/util"
	"strings"
)

// 获取所有安全组配置
func ListAliyunSGConfigs(c *gin.Context) {
	var configs []aliyunModel.AliyunSGConfig
	status := c.Query("status")
	accessKey := c.Query("access_key")
	name := c.Query("name")
	query := db.DB.Order("created_at desc")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if accessKey != "" {
		query = query.Where("access_key = ?", accessKey)
	}
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	query.Find(&configs)
	c.JSON(http.StatusOK, configs)
}

// 新增安全组配置
func CreateAliyunSGConfig(c *gin.Context) {
	var cfg aliyunModel.AliyunSGConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 只加密 SK
	sk, err := util.EncryptAES(cfg.AccessSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SK加密失败: " + err.Error()})
		return
	}
	cfg.AccessSecret = sk
	db.DB.Create(&cfg)
	c.JSON(http.StatusOK, cfg)
}

// 更新安全组配置
func UpdateAliyunSGConfig(c *gin.Context) {
	id := c.Param("id")
	var cfg aliyunModel.AliyunSGConfig
	if err := db.DB.First(&cfg, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if sk, ok := req["access_secret"].(string); ok && sk != "" {
		_, decodeErr := base64.StdEncoding.DecodeString(sk)
		if decodeErr != nil || len(sk) < 32 {
			encrypted, err := util.EncryptAES(sk)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "SK加密失败: " + err.Error()})
				return
			}
			req["access_secret"] = encrypted
		}
	}
	db.DB.Model(&cfg).Updates(req)
	c.JSON(http.StatusOK, cfg)
}

// 删除安全组配置
func DeleteAliyunSGConfig(c *gin.Context) {
	id := c.Param("id")
	db.DB.Delete(&aliyunModel.AliyunSGConfig{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// 单条同步安全组端口到阿里云
func SyncAliyunSGConfig(c *gin.Context) {
	id := c.Param("id")
	var cfg aliyunModel.AliyunSGConfig
	if err := db.DB.First(&cfg, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	dbIns := db.DB.Session(&gorm.Session{})
	dbIns = dbIns.Model(&aliyunModel.AliyunSGConfig{}).Where("id = ?", cfg.ID)
	err := aliyunService.UpdateSecurityGroupIfIPChanged(dbIns)
	if err != nil {
		if strings.Contains(err.Error(), "InvalidSecurityGroupId.NotFound") {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "找不到安全组，请检查安全组ID、Region和AK/SK配置是否正确。"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "同步成功"})
}

// 路由注册函数
func RegisterAliyunRoutes(r *gin.RouterGroup) {
	r.GET("/aliyun-sg-configs", ListAliyunSGConfigs)
	r.POST("/aliyun-sg-configs", CreateAliyunSGConfig)
	r.PUT("/aliyun-sg-configs/:id", UpdateAliyunSGConfig)
	r.DELETE("/aliyun-sg-configs/:id", DeleteAliyunSGConfig)
	r.POST("/aliyun-sg-configs/:id/sync", SyncAliyunSGConfig)
}
