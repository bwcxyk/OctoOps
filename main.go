package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"octoops/api"
	alertApi "octoops/api/alert"
	aliyunApi "octoops/api/aliyun"
	seatunnelApi "octoops/api/seatunnel"
	"octoops/config"
	"octoops/db"
	"octoops/scheduler"
	alertService "octoops/service/alert"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 主函数入口
	// 设置Gin框架为生产模式
	gin.SetMode(gin.ReleaseMode)
	// 初始化应用配置、数据库连接、定时任务和邮件服务
	config.InitConfig()                                            // 初始化配置
	db.Init()                                                      // 初始化数据库
	scheduler.InitScheduler()                                      // 初始化定时任务
	alertService.InitEmailConfigFromStruct(config.GetMailConfig()) // 初始化邮件配置

	// 初始化 Gin 引擎
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// API路由组
	apiGroup := r.Group("/api")
	seatunnelApi.RegisterTaskRoutes(apiGroup)
	alertApi.RegisterAlertRoutes(apiGroup)
	aliyunApi.RegisterAliyunRoutes(apiGroup)
	api.RegisterCustomTaskRoutes(apiGroup)
	api.RegisterSchedulerRoutes(apiGroup)
	alertApi.RegisterAlertGroupRoutes(apiGroup)
	alertApi.RegisterAlertGroupMemberRoutes(apiGroup)
	alertApi.RegisterAlertTemplateRoutes(apiGroup)

	// 静态资源托管
	r.Static("/assets", "./public/assets")
	r.StaticFile("/favicon.ico", "./public/favicon.ico")

	// 首页路由 - 直接返回index.html
	r.GET("/", func(c *gin.Context) {
		c.File("./public/index.html")
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
		c.File("./public/index.html")
	})

	// 优雅启动和关闭
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		log.Println("服务启动于 :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("收到关闭信号，正在优雅关闭服务...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("服务优雅关闭失败: %v", err)
	}
	log.Println("服务已优雅退出")
}
