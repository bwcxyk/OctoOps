package task

import (
	"strconv"

	"octoops/internal/db"
	taskModel "octoops/internal/model/task"

	"github.com/gin-gonic/gin"
)

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

	var total int64
	query.Model(&taskModel.TaskLog{}).Count(&total)
	query = query.Order("created_at desc").Limit(pageSize).Offset((page - 1) * pageSize)
	query.Find(&logs)

	c.JSON(200, gin.H{
		"data":  logs,
		"total": total,
	})
}

func RegisterTaskLogRoutes(r *gin.RouterGroup) {
	r.GET("/task/log", ListTaskLogs)
}
