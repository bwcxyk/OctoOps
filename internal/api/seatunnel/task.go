package api

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"octoops/internal/config"
	"octoops/internal/db"
	"octoops/internal/model"
	seatunnelModel "octoops/internal/model/seatunnel"
	"octoops/internal/scheduler"
	seatunnel "octoops/internal/service/seatunnel"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// 扩展的任务结构，包含下次执行时间
type TaskWithNextRun struct {
	seatunnelModel.EtlTask
	NextRunTime *time.Time `json:"next_run_time,omitempty"`
}

func ListTasks(c *gin.Context) {
	var tasks []seatunnelModel.EtlTask
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
	if job_id := c.Query("job_id"); job_id != "" {
		query = query.Where("job_id = ?", job_id)
	}
	if job_status := c.Query("job_status"); job_status != "" {
		query = query.Where("job_status = ?", job_status)
	}
	query = query.Order("created_at desc")

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
	query.Model(&seatunnelModel.EtlTask{}).Count(&total)
	query = query.Limit(pageSize).Offset((page - 1) * pageSize)
	query.Find(&tasks)

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

	c.JSON(200, gin.H{
		"data":  tasksWithNextRun,
		"total": total,
	})
}

func CreateTask(c *gin.Context) {
	var task seatunnelModel.EtlTask
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
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
	var req map[string]interface{}
	var dbTask seatunnelModel.EtlTask
	if err := db.DB.First(&dbTask, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 保证ID和JobID不变
	req["id"] = dbTask.ID
	db.DB.Model(&dbTask).Updates(req)
	c.JSON(http.StatusOK, req)
}

func DeleteTask(c *gin.Context) {
	id := c.Param("id")

	// 先获取任务信息
	var task seatunnelModel.EtlTask
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

	// 检查任务状态，stream 类型运行中不允许重复提交
	var task seatunnelModel.EtlTask
	if err := db.DB.First(&task, taskID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	if task.TaskType == "stream" && task.JobStatus == "RUNNING" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "实时数据集成运行中，不允许重复提交作业，请先停止当前作业"})
		return
	}

	// 只有流式任务支持 SavePoint
	var isStartWithSavePoint bool
	if task.TaskType == "stream" {
		isStartWithSavePoint = c.Query("isStartWithSavePoint") == "true"
		// SavePoint 启动时必须有 job_id
		if isStartWithSavePoint && (task.JobID == nil || *task.JobID == "") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "使用 SavePoint 启动时必须有已存在的 job_id，请先正常启动任务"})
			return
		}
	}

	respBody, err := seatunnel.SubmitJobInternal(taskID, isStartWithSavePoint)
	if err != nil {
		// 打印日志
		log.Printf("[ETL] 提交作业失败: taskID=%d, type=%s, error=%v", taskID, task.TaskType, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交作业失败"})
		return
	}

	// 更新最后运行时间
	db.DB.Model(&task).Update("last_run_time", time.Now())

	// 从响应中提取 jobId 并更新到数据库
	seatunnel.UpdateJobIdFromResponse(taskID, respBody)

	log.Printf("[ETL] 提交作业成功: taskID=%d, type=%s, isStartWithSavePoint=%v, result=%s", taskID, task.TaskType, isStartWithSavePoint, string(respBody))

	c.JSON(http.StatusOK, gin.H{"message": "作业提交成功"})
}

// 停止作业
func StopJob(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}
	var task seatunnelModel.EtlTask
	if err := db.DB.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	if task.JobID == nil || *task.JobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "jobId is empty in database"})
		return
	}
	isStopWithSavePoint := c.Query("isStopWithSavePoint")
	body := fmt.Sprintf(`{"jobId": "%s", "isStopWithSavePoint": %s}`, *task.JobID, isStopWithSavePoint)
	url := config.SeatunnelBaseURL + "/stop-job"
	resp, err := http.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		log.Printf("[ETL] 停止作业失败: taskID=%d, jobId=%s, error=%v", task.ID, *task.JobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法连接到 Seatunnel 服务，请检查服务是否已启动且网络正常"})
		return
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("[ETL] 停止作业成功: taskID=%d, jobId=%s", task.ID, *task.JobID)
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

// 更新任务时同步更新调度器
func UpdateTaskWithScheduler(c *gin.Context) {
	id := c.Param("id")
	var dbTask seatunnelModel.EtlTask
	if err := db.DB.First(&dbTask, id).Error; err != nil {
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
	db.DB.Model(&dbTask).Updates(req)

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
			} else {
				if oldStatus == 1 {
					scheduler.RemoveTask(dbTask.ID)
				}
			}
		}
	}

	c.JSON(http.StatusOK, req)
}

// 获取作业日志
func ListTaskLogs(c *gin.Context) {
	var logs []model.TaskLog
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
	query.Model(&model.TaskLog{}).Count(&total)
	query = query.Order("created_at desc").Limit(pageSize).Offset((page - 1) * pageSize)
	query.Find(&logs)

	c.JSON(200, gin.H{
		"data":  logs,
		"total": total,
	})
}

// 手动触发同步所有任务 job_status
func SyncJobStatus(c *gin.Context) {
	log.Printf("[ETL] 触发同步作业状态 /api/sync-job-status")
	seatunnel.SyncAllJobStatus()
	c.JSON(200, gin.H{"message": "同步作业状态已触发"})
}

func RegisterTaskRoutes(r *gin.RouterGroup) {
	r.GET("/tasks", ListTasks)
	r.POST("/tasks", CreateTask)
	r.PUT("/tasks/:id", UpdateTaskWithScheduler)
	r.DELETE("/tasks/:id", DeleteTask)
	r.POST("/submit-job", SubmitJob)
	r.POST("/stop-job", StopJob)
	r.GET("/task-logs", ListTaskLogs)
	r.POST("/sync-job-status", SyncJobStatus)
}
