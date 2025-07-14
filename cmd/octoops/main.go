package main

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"octoops/internal/api"
	alertApi "octoops/internal/api/alert"
	aliyunApi "octoops/internal/api/aliyun"
	rbacApi "octoops/internal/api/rbac"
	seatunnelApi "octoops/internal/api/seatunnel"
	"octoops/internal/config"
	"octoops/internal/db"
	"octoops/internal/pkg/jwt"
	"octoops/internal/scheduler"
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
	config.InitConfig()       // 初始化配置
	db.Init()                 // 初始化数据库
	scheduler.InitScheduler() // 初始化定时任务

	// 初始化JWT密钥
	jwt.SetJWTSecret(config.GetJWTSecret())

	// 初始化 Gin 引擎
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// API路由组
	apiGroup := r.Group("/api")

	// 认证相关路由（无需认证）
	rbacApi.RegisterAuthRoutes(apiGroup)

	// 需要认证的路由
	api.RegisterCustomTaskRoutes(apiGroup)
	api.RegisterSchedulerRoutes(apiGroup)
	seatunnelApi.RegisterTaskRoutes(apiGroup)
	aliyunApi.RegisterAliyunRoutes(apiGroup)
	// 告警相关路由
	alertApi.RegisterAlertChannelRoutes(apiGroup)
	alertApi.RegisterAlertGroupRoutes(apiGroup)
	alertApi.RegisterAlertGroupMemberRoutes(apiGroup)
	alertApi.RegisterAlertTemplateRoutes(apiGroup)

	// RBAC管理路由
	rbacApi.RegisterUserRoutes(apiGroup)
	rbacApi.RegisterRoleRoutes(apiGroup)
	rbacApi.RegisterPermissionRoutes(apiGroup)

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
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
