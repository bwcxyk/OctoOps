package seatunnel

import (
	"octoops/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterBatchTaskRoutes(r *gin.RouterGroup) {
	r.GET("/seatunnel/batch", middleware.AuthMiddleware(), middleware.RequirePermission("etl:batch:read"), ListBatchTasks)
	r.POST("/seatunnel/batch", middleware.AuthMiddleware(), middleware.RequirePermission("etl:batch:create"), CreateBatchTask)
	r.PUT("/seatunnel/batch/:id", middleware.AuthMiddleware(), middleware.RequirePermission("etl:batch:update"), UpdateBatchTaskWithScheduler)
	r.DELETE("/seatunnel/batch/:id", middleware.AuthMiddleware(), middleware.RequirePermission("etl:batch:delete"), DeleteBatchTask)
}
