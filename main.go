package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"octoops/api"
	"octoops/config"
	"octoops/db"
	"octoops/scheduler"
	"octoops/service"
)

func main() {
	config.InitConfig()                                       // 初始化配置
	db.Init()                                                 // 初始化数据库
	scheduler.InitScheduler()                                 // 初始化定时任务
	service.InitEmailConfigFromStruct(config.GetMailConfig()) // 初始化邮件配置

	r := gin.Default()

	// API路由组
	apiGroup := r.Group("/api")
	api.RegisterTaskRoutes(apiGroup)
	api.RegisterAlertRoutes(apiGroup)
	api.RegisterAliyunRoutes(apiGroup)
	api.RegisterCustomTaskRoutes(apiGroup)
	api.RegisterAlertGroupRoutes(apiGroup)
	api.RegisterAlertGroupMemberRoutes(apiGroup)
	api.RegisterAlertTemplateRoutes(apiGroup)

	// 静态资源托管
	r.Static("/assets", "./web/public/assets")
	r.StaticFile("/favicon.ico", "./web/public/favicon.ico")

	// 首页路由 - 直接返回index.html
	r.GET("/", func(c *gin.Context) {
		c.File("./web/public/index.html")
	})

	// 其他所有前端路由 - 返回index.html
	r.NoRoute(func(c *gin.Context) {
		// 如果是API路径但未匹配,返回404
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "API not found",
			})
			return
		}
		// 非API路径都返回前端页面
		c.File("./web/public/index.html")
	})

	r.Run(":8080")
}
