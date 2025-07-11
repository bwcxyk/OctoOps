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
    fmt.Sscanf(id, "%d", &uid)
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
    db.DB.Delete(&model.CustomTask{}, id)
    var uid uint
    fmt.Sscanf(id, "%d", &uid)
    scheduler.DisableCustomTask(uid)
    c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func RegisterCustomTaskRoutes(r *gin.RouterGroup) {
    r.GET("/custom-tasks", ListCustomTasks)
    r.POST("/custom-tasks", CreateCustomTask)
    r.PUT("/custom-tasks/:id", UpdateCustomTask)
    r.DELETE("/custom-tasks/:id", DeleteCustomTask)
} 