package seatunnel

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"octoops/internal/config"
	"octoops/internal/infra/postgres"
	"octoops/internal/middleware"
	seatunnelModel "octoops/internal/model/seatunnel"
	seatunnel "octoops/internal/service/seatunnel"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func waitForTaskStatus(taskID uint, attempts int, interval time.Duration, shouldStop func(string) bool) string {
	jobStatus := "UNKNOWN"
	for i := 0; i < attempts; i++ {
		status, syncErr := seatunnel.SyncJobStatusByTaskID(taskID)
		if syncErr != nil {
			log.Printf("[ETL] 同步作业状态失败: taskID=%d, error=%v", taskID, syncErr)
		} else {
			jobStatus = status
			if shouldStop(status) {
				return jobStatus
			}
		}
		if i < attempts-1 {
			time.Sleep(interval)
		}
	}
	return jobStatus
}

func requireTaskActionPermission(c *gin.Context, taskType, action string) bool {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return false
	}
	permissionCode := fmt.Sprintf("etl:%s:%s", taskType, action)
	if !middleware.HasPermission(user, permissionCode) {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return false
	}
	return true
}

// writeUpstreamError 输出上游错误：error 为给前端展示的短文案，detail 为上游状态码与原始响应
func writeUpstreamError(c *gin.Context, message string, err error) {
	body := gin.H{"error": message}

	var upstreamErr *seatunnel.UpstreamError
	if errors.As(err, &upstreamErr) {
		detail := gin.H{}
		if upstreamErr.StatusCode > 0 {
			detail["status_code"] = upstreamErr.StatusCode
		}
		if upstreamErr.Err != nil {
			detail["reason"] = upstreamErr.Err.Error()
		}
		if len(upstreamErr.Body) > 0 {
			// 上游返回 JSON 时解析成对象，避免再套一层转义字符串
			var parsed interface{}
			if json.Unmarshal(upstreamErr.Body, &parsed) == nil {
				detail["response"] = parsed
			} else {
				detail["response"] = string(upstreamErr.Body)
			}
		}
		body["detail"] = detail
	} else {
		body["detail"] = err.Error()
	}

	c.JSON(http.StatusInternalServerError, body)
}

// 提交作业
func SubmitJob(c *gin.Context) {
	id := c.Param("id")
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
	if err := postgres.DB.First(&task, taskID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	if !requireTaskActionPermission(c, task.TaskType, "submit") {
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
		log.Printf("[ETL] 提交作业失败: taskID=%d, type=%s, error=%v", taskID, task.TaskType, err)
		writeUpstreamError(c, "提交作业失败", err)
		return
	}

	// 更新最后运行时间
	postgres.DB.Model(&task).Update("last_run_time", time.Now())

	// 从响应中提取 jobId 并更新到数据库
	seatunnel.UpdateJobIdFromResponse(taskID, respBody)

	// 提交成功后等待作业进入明确状态，前端只需刷新一次列表
	jobStatus := waitForTaskStatus(taskID, 15, time.Second, func(status string) bool {
		return status == "RUNNING" || status == "FAILED" || status == "FINISHED" || status == "CANCEL"
	})

	log.Printf("[ETL] 提交作业成功: taskID=%d, type=%s, isStartWithSavePoint=%v, result=%s", taskID, task.TaskType, isStartWithSavePoint, string(respBody))

	c.JSON(http.StatusOK, gin.H{
		"message":    "作业提交成功",
		"job_status": jobStatus,
	})
}

// 停止作业
func StopJob(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}
	var task seatunnelModel.EtlTask
	if err := postgres.DB.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	if !requireTaskActionPermission(c, task.TaskType, "stop") {
		return
	}
	if task.JobID == nil || *task.JobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "jobId is empty in database"})
		return
	}
	isStopWithSavePoint := c.DefaultQuery("isStopWithSavePoint", "false")
	if isStopWithSavePoint != "true" && isStopWithSavePoint != "false" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "isStopWithSavePoint must be true or false"})
		return
	}
	body := fmt.Sprintf(`{"jobId": "%s", "isStopWithSavePoint": %s}`, *task.JobID, isStopWithSavePoint)
	url := config.SeatunnelBaseURL + "/stop-job"
	client := &http.Client{Timeout: config.SeatunnelRequestTimeout}
	resp, err := client.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		log.Printf("[ETL] 停止作业失败: taskID=%d, jobId=%s, error=%v", task.ID, *task.JobID, err)
		writeUpstreamError(c, "停止作业失败", seatunnel.NewUpstreamRequestError("停止作业", err))
		return
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.Printf("[ETL] 停止作业失败: taskID=%d, jobId=%s, statusCode=%d, response=%s", task.ID, *task.JobID, resp.StatusCode, string(respBody))
		writeUpstreamError(c, "停止作业失败", seatunnel.NewUpstreamError("停止作业", resp.StatusCode, respBody))
		return
	}

	// 停止成功后等待状态退出 RUNNING，前端只需刷新一次列表
	jobStatus := waitForTaskStatus(task.ID, 15, time.Second, func(status string) bool {
		return status != "" && status != "RUNNING" && status != "UNKNOWN"
	})

	log.Printf("[ETL] 停止作业成功: taskID=%d, jobId=%s, jobStatus=%s", task.ID, *task.JobID, jobStatus)
	c.JSON(http.StatusOK, gin.H{
		"message":    "作业停止成功",
		"job_status": jobStatus,
		"result":     string(respBody),
	})
}

// 手动触发同步所有任务 job_status
func SyncJobStatus(c *gin.Context) {
	log.Printf("[ETL] 触发同步作业状态 /api/seatunnel/tasks/sync-status")
	result, err := seatunnel.SyncAllJobStatus()
	if err != nil {
		log.Printf("[ETL] 同步作业状态失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": result})
}

func RegisterStreamTaskRoutes(r *gin.RouterGroup) {
	r.GET("/seatunnel/stream", middleware.AuthMiddleware(), middleware.RequirePermission("etl:stream:read"), ListStreamTasks)
	r.GET("/seatunnel/stream/:id", middleware.AuthMiddleware(), middleware.RequirePermission("etl:stream:read"), GetStreamTask)
	r.POST("/seatunnel/stream", middleware.AuthMiddleware(), middleware.RequirePermission("etl:stream:create"), CreateStreamTask)
	r.PUT("/seatunnel/stream/:id", middleware.AuthMiddleware(), middleware.RequirePermission("etl:stream:update"), UpdateStreamTaskWithScheduler)
	r.DELETE("/seatunnel/stream/:id", middleware.AuthMiddleware(), middleware.RequirePermission("etl:stream:delete"), DeleteStreamTask)

	// 作业控制（stream/batch 手动触发都使用这组接口）
	r.POST("/seatunnel/tasks/:id/start", middleware.AuthMiddleware(), middleware.RequireAnyPermission("etl:stream:submit", "etl:batch:submit"), SubmitJob)
	r.POST("/seatunnel/tasks/:id/stop", middleware.AuthMiddleware(), middleware.RequirePermission("etl:stream:stop"), StopJob)
	r.POST("/seatunnel/tasks/sync-status", middleware.AuthMiddleware(), middleware.RequirePermission("etl:stream:sync_status"), SyncJobStatus)
}
