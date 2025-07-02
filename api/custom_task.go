package api

import (
    "net/http"
    "octoops/db"
    "octoops/model"
    "octoops/scheduler"
    "github.com/gin-gonic/gin"
    "fmt"
)

func ListCustomTasks(c *gin.Context) {
    var tasks []model.CustomTask
    db.DB.Order("created_at desc").Find(&tasks)
    c.JSON(http.StatusOK, tasks)
}

func CreateCustomTask(c *gin.Context) {
    var task model.CustomTask
    if err := c.ShouldBindJSON(&task); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    db.DB.Create(&task)
    scheduler.RegisterCustomTask(
        fmt.Sprintf("custom_%d", task.ID),
        task.Name,
        task.CustomType,
        task.CronExpr,
        task.Status == 1,
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
    var req model.CustomTask
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    db.DB.Model(&task).Updates(req)
    scheduler.DisableCustomTask(fmt.Sprintf("custom_%s", id))
    scheduler.RegisterCustomTask(
        fmt.Sprintf("custom_%s", id),
        req.Name,
        req.CustomType,
        req.CronExpr,
        req.Status == 1,
        scheduler.GetJobFuncByType(req.CustomType),
    )
    c.JSON(http.StatusOK, task)
}

func DeleteCustomTask(c *gin.Context) {
    id := c.Param("id")
    db.DB.Delete(&model.CustomTask{}, id)
    scheduler.DisableCustomTask(fmt.Sprintf("custom_%s", id))
    c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func EnableCustomTask(c *gin.Context) {
    id := c.Param("id")
    scheduler.EnableCustomTask(fmt.Sprintf("custom_%s", id))
    c.JSON(http.StatusOK, gin.H{"message": "已启用"})
}

func DisableCustomTask(c *gin.Context) {
    id := c.Param("id")
    scheduler.DisableCustomTask(fmt.Sprintf("custom_%s", id))
    c.JSON(http.StatusOK, gin.H{"message": "已禁用"})
}

func RegisterCustomTaskRoutes(r *gin.RouterGroup) {
    r.GET("/custom-tasks", ListCustomTasks)
    r.POST("/custom-tasks", CreateCustomTask)
    r.PUT("/custom-tasks/:id", UpdateCustomTask)
    r.DELETE("/custom-tasks/:id", DeleteCustomTask)
    r.POST("/custom-tasks/:id/enable", EnableCustomTask)
    r.POST("/custom-tasks/:id/disable", DisableCustomTask)
} 