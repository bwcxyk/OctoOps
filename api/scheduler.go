package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"octoops/scheduler"
)

// 启动调度器
func StartSchedulerHandler(c *gin.Context) {
	scheduler.StartScheduler()
	c.JSON(http.StatusOK, gin.H{"message": "Scheduler started"})
}

// 停止调度器
func StopSchedulerHandler(c *gin.Context) {
	scheduler.StopScheduler()
	c.JSON(http.StatusOK, gin.H{"message": "Scheduler stopped"})
}

// 获取调度器状态
func GetSchedulerStatus(c *gin.Context) {
	status := scheduler.GetSchedulerStatus()
	c.JSON(http.StatusOK, status)
}

// 重新加载调度器
func ReloadScheduler(c *gin.Context) {
	scheduler.ReloadTasks()
	c.JSON(http.StatusOK, gin.H{"message": "调度器重新加载成功"})
}

// 注册调度器相关路由
func RegisterSchedulerRoutes(r *gin.RouterGroup) {
	r.POST("/scheduler/start", StartSchedulerHandler)
	r.POST("/scheduler/stop", StopSchedulerHandler)
	r.GET("/scheduler/status", GetSchedulerStatus)
	r.POST("/scheduler/reload", ReloadScheduler)
} 