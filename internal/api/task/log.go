package task

import (
	"net/http"
	"strconv"

	"octoops/internal/db"
	"octoops/internal/middleware"
	taskModel "octoops/internal/model/task"

	"github.com/gin-gonic/gin"
)

const maxPageSize = 100

// 获取任务日志
func ListTaskLogs(c *gin.Context) {
	var logs []taskModel.TaskLog
	query := db.DB
	if taskName := c.Query("task_name"); taskName != "" {
		query = query.Where("task_name LIKE ?", "%"+taskName+"%")
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if startTime := c.Query("start_time"); startTime != "" {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime := c.Query("end_time"); endTime != "" {
		query = query.Where("created_at <= ?", endTime)
	}

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	var total int64
	if err := query.Model(&taskModel.TaskLog{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "统计任务日志失败",
		})
		return
	}
	query = query.Order("created_at desc").Limit(pageSize).Offset((page - 1) * pageSize)
	if err := query.Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "查询任务日志失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  logs,
		"total": total,
	})
}

func RegisterTaskLogRoutes(r *gin.RouterGroup) {
	r.GET("/task/log", middleware.AuthMiddleware(), middleware.RequirePermission("task:log:read"), ListTaskLogs)
}
