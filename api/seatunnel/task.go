package api

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"octoops/config"
	"octoops/db"
	"octoops/model"
	seatunnelModel "octoops/model/seatunnel"
	"octoops/scheduler"
	seatunnelService "octoops/service/seatunnel"
	"strings"
	"time"
)

// 扩展的任务结构，包含下次执行时间
type TaskWithNextRun struct {
	seatunnelModel.Task
	NextRunTime *time.Time `json:"next_run_time,omitempty"`
}

func ListTasks(c *gin.Context) {
	var tasks []seatunnelModel.Task
	query := db.DB
	if taskType := c.Query("task_type"); taskType != "" {
		query = query.Where("task_type = ?", taskType)
	}
	if name := c.Query("name"); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if status := c.Query("status"); status != "" {
		if status == "1" {
			query = query.Where("status = ?", 1)
		} else if status == "0" {
			query = query.Where("status = ?", 0)
		}
	}
	if jobid := c.Query("jobid"); jobid != "" {
		query = query.Where("job_id = ?", jobid)
	}
	if job_status := c.Query("job_status"); job_status != "" {
		query = query.Where("job_status = ?", job_status)
	}
	query = query.Order("created_at desc")
	query.Find(&tasks)

	// 获取所有任务的下次执行时间
	nextRunTimes := scheduler.GetAllTasksNextRunTime()

	// 转换为包含下次执行时间的结构
	var tasksWithNextRun []TaskWithNextRun
	for _, task := range tasks {
		taskWithNextRun := TaskWithNextRun{
			Task: task,
		}
		if nextRun, exists := nextRunTimes[task.ID]; exists {
			taskWithNextRun.NextRunTime = nextRun
		}
		tasksWithNextRun = append(tasksWithNextRun, taskWithNextRun)
	}

	c.JSON(http.StatusOK, tasksWithNextRun)
}

func CreateTask(c *gin.Context) {
	var task seatunnelModel.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if task.JobID == "" {
		// 生成格式为: 20240612153001_1
		now := time.Now().Format("200601021504") // 年月日时分
		var count int64
		// 统计本分钟内的任务数
		start := time.Now().Truncate(time.Minute)
		end := start.Add(time.Minute)
		db.DB.Model(&seatunnelModel.Task{}).Where("created_at >= ? AND created_at < ?", start, end).Count(&count)
		task.JobID = fmt.Sprintf("%s%04d", now, count+1)
	}
	db.DB.Create(&task)

	// 如果是批处理任务且状态为active且有cron表达式，添加到调度器
	if task.TaskType == "batch" && task.Status == 1 && task.CronExpr != "" {
		scheduler.AddTask(task)
	}

	c.JSON(http.StatusOK, task)
}

func UpdateTask(c *gin.Context) {
	id := c.Param("id")
	var req seatunnelModel.Task
	var dbTask seatunnelModel.Task
	if err := db.DB.First(&dbTask, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.JobID == "" {
		req.JobID = dbTask.JobID
		if req.JobID == "" {
			now := time.Now().Format("200601021504")
			var count int64
			start := time.Now().Truncate(time.Minute)
			end := start.Add(time.Minute)
			db.DB.Model(&seatunnelModel.Task{}).Where("created_at >= ? AND created_at < ?", start, end).Count(&count)
			req.JobID = fmt.Sprintf("%s%04d", now, count+1)
		}
	}
	req.ID = dbTask.ID // 保证ID不变
	db.DB.Model(&dbTask).Updates(req)
	c.JSON(http.StatusOK, req)
}

func DeleteTask(c *gin.Context) {
	id := c.Param("id")

	// 先获取任务信息
	var task seatunnelModel.Task
	if err := db.DB.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// 从调度器中移除任务
	if task.TaskType == "batch" && task.Status == 1 {
		scheduler.RemoveTask(task.ID)
	}

	// 删除任务
	db.DB.Delete(&task)
	c.JSON(http.StatusOK, gin.H{"message": "Task deleted"})
}

// 提交作业
func SubmitJob(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}
	var taskID uint
	if _, err := fmt.Sscanf(id, "%d", &taskID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}
	isStartWithSavePoint := c.Query("isStartWithSavePoint") == "true"

	respBody, err := seatunnelService.SubmitJobInternal(taskID, isStartWithSavePoint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 解析 seatunnel 返回内容并保存日志
	var resultMap map[string]interface{}
	_ = json.Unmarshal(respBody, &resultMap)
	jobId := ""
	jobName := ""
	if v, ok := resultMap["jobId"].(string); ok {
		jobId = v
	} else if v, ok := resultMap["jobId"].(float64); ok {
		jobId = fmt.Sprintf("%.0f", v)
	}
	if v, ok := resultMap["jobName"].(string); ok {
		jobName = v
	}
	fmt.Println("写入日志：", jobId, jobName, string(respBody))
	var task seatunnelModel.Task
	db.DB.First(&task, taskID)
	// 更新最后运行时间
	db.DB.Model(&task).Update("last_run_time", time.Now())
	seatunnelService.WriteTaskLog(task, respBody)

	c.JSON(http.StatusOK, gin.H{"message": "作业提交成功"})
}

// 停止作业
func StopJob(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}
	var task seatunnelModel.Task
	if err := db.DB.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	if task.JobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "jobId is empty in database"})
		return
	}
	isStopWithSavePoint := c.Query("isStopWithSavePoint")
	body := fmt.Sprintf(`{"jobId": "%s", "isStopWithSavePoint": %s}`, task.JobID, isStopWithSavePoint)
	url := config.SeatunnelBaseURL + "/stop-job"
	resp, err := http.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
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

// 更新任务时同步更新调度器
func UpdateTaskWithScheduler(c *gin.Context) {
	var task seatunnelModel.Task
	id := c.Param("id")
	if err := db.DB.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// 保存原始状态用于比较
	oldStatus := task.Status

	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 保存任务
	db.DB.Save(&task)

	// 如果是批处理任务且有cron表达式，更新调度器
	if task.TaskType == "batch" && task.CronExpr != "" {
		if task.Status == 1 {
			// 如果状态从inactive变为active，添加任务到调度器
			if oldStatus != 1 {
				scheduler.AddTask(task)
			} else {
				// 如果状态已经是active，重新加载所有任务（cron表达式可能改变了）
				scheduler.ReloadTasks()
			}
		} else {
			// 如果状态变为inactive，从调度器中移除任务
			if oldStatus == 1 {
				scheduler.RemoveTask(task.ID)
			}
		}
	}

	c.JSON(http.StatusOK, task)
}

// 获取作业日志
func ListTaskLogs(c *gin.Context) {
	var logs []model.TaskLog
	query := db.DB
	if jobID := c.Query("job_id"); jobID != "" {
		query = query.Where("job_id = ?", jobID)
	}
	if taskType := c.Query("task_type"); taskType != "" {
		query = query.Where("task_type = ?", taskType)
	}
	query.Order("created_at desc").Find(&logs)
	c.JSON(http.StatusOK, logs)
}

// 手动触发同步所有任务 job_status
func SyncJobStatus(c *gin.Context) {
	log.Printf("[API] 手动触发同步作业状态 /api/sync-job-status")
	scheduler.SyncAllJobStatus()
	c.JSON(200, gin.H{"message": "同步作业状态已触发"})
}

func RegisterTaskRoutes(r *gin.RouterGroup) {
	r.GET("/tasks", ListTasks)
	r.POST("/tasks", CreateTask)
	r.PUT("/tasks/:id", UpdateTaskWithScheduler)
	r.DELETE("/tasks/:id", DeleteTask)
	r.POST("/submit-job", SubmitJob)
	r.POST("/stop-job", StopJob)
	r.GET("/scheduler/status", GetSchedulerStatus)
	r.POST("/scheduler/reload", ReloadScheduler)
	r.GET("/task-logs", ListTaskLogs)
	r.POST("/sync-job-status", SyncJobStatus)
} 