package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"octoops/internal/db"
	"octoops/internal/model"
	"octoops/internal/scheduler"
	"strconv"
)

func ListCustomTasks(c *gin.Context) {
	var tasks []model.CustomTask
	query := db.DB.Model(&model.CustomTask{})
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
	query.Count(&total)
	query = query.Order("created_at desc").Limit(pageSize).Offset((page-1)*pageSize)
	query.Find(&tasks)
	c.JSON(200, gin.H{
		"data":  tasks,
		"total": total,
	})
}

func CreateCustomTask(c *gin.Context) {
	var task model.CustomTask
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Create(&task)
	scheduler.RegisterCustomTask(
		task.ID,
		task.Name,
		task.CustomType,
		task.CronExpr,
		task.Status,
		scheduler.GetJobFuncByType(task.CustomType),
	)
	c.JSON(http.StatusOK, task)
}

func UpdateCustomTask(c *gin.Context) {
	id := c.Param("id")
	var task model.CustomTask
	if err := db.DB.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Model(&task).Updates(req)
	var uid uint
	if _, err := fmt.Sscanf(id, "%d", &uid); err != nil {
		log.Printf("id 解析失败: %v", err)
		c.JSON(400, gin.H{"error": "无效的ID"})
		return
	}
	scheduler.DisableCustomTask(uid)
	var updatedTask model.CustomTask
	db.DB.First(&updatedTask, id)
	if updatedTask.Status == 1 {
		scheduler.RegisterCustomTask(
			updatedTask.ID,
			updatedTask.Name,
			updatedTask.CustomType,
			updatedTask.CronExpr,
			updatedTask.Status,
			scheduler.GetJobFuncByType(updatedTask.CustomType),
		)
	}
	c.JSON(http.StatusOK, task)
}

func DeleteCustomTask(c *gin.Context) {
	id := c.Param("id")
	var uid uint
	if _, err := fmt.Sscanf(id, "%d", &uid); err != nil {
		log.Printf("id 解析失败: %v", err)
		c.JSON(400, gin.H{"error": "无效的ID"})
		return
	}
	db.DB.Delete(&model.CustomTask{}, id)
	scheduler.DisableCustomTask(uid)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func RegisterCustomTaskRoutes(r *gin.RouterGroup) {
	r.GET("/custom-tasks", ListCustomTasks)
	r.POST("/custom-tasks", CreateCustomTask)
	r.PUT("/custom-tasks/:id", UpdateCustomTask)
	r.DELETE("/custom-tasks/:id", DeleteCustomTask)
}
