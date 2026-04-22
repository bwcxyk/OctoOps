package seatunnel

import (
	"net/http"
	seatunnelModel "octoops/internal/model/seatunnel"
	"octoops/internal/scheduler"
	seatunnelService "octoops/internal/service/seatunnel"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// 扩展的任务结构，包含下次执行时间
type TaskWithNextRun struct {
	seatunnelModel.EtlTask
	NextRunTime *time.Time `json:"next_run_time,omitempty"`
}

func listTasks(c *gin.Context, fixedTaskType string) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	filter := seatunnelService.TaskListFilter{
		Page:      page,
		PageSize:  pageSize,
		JobID:     c.Query("job_id"),
		JobStatus: c.Query("job_status"),
		Name:      c.Query("name"),
	}
	if fixedTaskType != "" {
		filter.TaskType = fixedTaskType
	} else {
		filter.TaskType = c.Query("task_type")
	}
	if status := c.Query("status"); status == "1" {
		active := 1
		filter.Status = &active
	} else if status == "0" {
		inactive := 0
		filter.Status = &inactive
	}

	tasks, total, err := seatunnelService.ListTasks(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询任务失败: " + err.Error()})
		return
	}

	// 获取所有任务的下次执行时间
	nextRunTimes := scheduler.GetAllTasksNextRunTime()

	// 转换为包含下次执行时间的结构
	var tasksWithNextRun []TaskWithNextRun
	for _, task := range tasks {
		taskWithNextRun := TaskWithNextRun{
			EtlTask: task,
		}
		if nextRun, exists := nextRunTimes[task.ID]; exists {
			taskWithNextRun.NextRunTime = nextRun
		}
		tasksWithNextRun = append(tasksWithNextRun, taskWithNextRun)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  tasksWithNextRun,
		"total": total,
	})
}

func createTask(c *gin.Context, fixedTaskType string) {
	var task seatunnelModel.EtlTask
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if fixedTaskType != "" {
		if task.TaskType != "" && task.TaskType != fixedTaskType {
			c.JSON(http.StatusBadRequest, gin.H{"error": "task_type does not match route type"})
			return
		}
		task.TaskType = fixedTaskType
	}
	if err := seatunnelService.CreateTask(&task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建任务失败: " + err.Error()})
		return
	}

	// 如果是批处理任务且状态为active且有cron表达式，添加到调度器
	if task.TaskType == "batch" && task.Status == 1 && task.CronExpr != "" {
		scheduler.AddTask(task)
	}

	c.JSON(http.StatusOK, task)
}

func deleteTask(c *gin.Context, fixedTaskType string) {
	id := c.Param("id")

	task, err := seatunnelService.GetTaskByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	if fixedTaskType != "" && task.TaskType != fixedTaskType {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// 从调度器中移除任务
	if task.TaskType == "batch" && task.Status == 1 {
		scheduler.RemoveTask(task.ID)
	}

	if err := seatunnelService.DeleteTask(&task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除任务失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Task deleted"})
}

func getTask(c *gin.Context, fixedTaskType string) {
	id := c.Param("id")

	task, err := seatunnelService.GetTaskByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	if fixedTaskType != "" && task.TaskType != fixedTaskType {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// 更新任务时同步更新调度器
func updateTaskWithScheduler(c *gin.Context, fixedTaskType string) {
	id := c.Param("id")
	dbTask, err := seatunnelService.GetTaskByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	if fixedTaskType != "" && dbTask.TaskType != fixedTaskType {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	oldStatus := dbTask.Status

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 保证ID和JobID不变
	req["id"] = dbTask.ID
	req["task_type"] = dbTask.TaskType
	if err := seatunnelService.UpdateTask(&dbTask, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新任务失败: " + err.Error()})
		return
	}

	// 刷新调度器
	if dbTask.TaskType == "batch" && dbTask.CronExpr != "" {
		if v, ok := req["status"]; ok {
			statusInt := 0
			switch vv := v.(type) {
			case float64:
				statusInt = int(vv)
			case int:
				statusInt = vv
			}
			if statusInt == 1 {
				if oldStatus != 1 {
					scheduler.AddTask(dbTask)
				} else {
					scheduler.RemoveTask(dbTask.ID)
					scheduler.AddTask(dbTask)
				}
			} else if oldStatus == 1 {
				scheduler.RemoveTask(dbTask.ID)
			}
		}
	}

	c.JSON(http.StatusOK, req)
}

func ListBatchTasks(c *gin.Context) {
	listTasks(c, "batch")
}

func GetBatchTask(c *gin.Context) {
	getTask(c, "batch")
}

func CreateBatchTask(c *gin.Context) {
	createTask(c, "batch")
}

func UpdateBatchTaskWithScheduler(c *gin.Context) {
	updateTaskWithScheduler(c, "batch")
}

func DeleteBatchTask(c *gin.Context) {
	deleteTask(c, "batch")
}

func ListStreamTasks(c *gin.Context) {
	listTasks(c, "stream")
}

func GetStreamTask(c *gin.Context) {
	getTask(c, "stream")
}

func CreateStreamTask(c *gin.Context) {
	createTask(c, "stream")
}

func UpdateStreamTaskWithScheduler(c *gin.Context) {
	updateTaskWithScheduler(c, "stream")
}

func DeleteStreamTask(c *gin.Context) {
	deleteTask(c, "stream")
}
