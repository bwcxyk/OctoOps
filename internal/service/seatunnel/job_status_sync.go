package seatunnel

import (
	"fmt"
	"log"
	"octoops/internal/db"
	seatunnelModel "octoops/internal/model/seatunnel"
)

func SyncAllJobStatus() {
	log.Printf("[Scheduler] 开始同步作业状态")
	var tasks []seatunnelModel.EtlTask
	db.DB.Where("task_type = ?", "stream").Find(&tasks)
	for _, task := range tasks {
		if task.JobID != nil && *task.JobID != "" {
			oldStatus := task.JobStatus
			result := QuerySeatunnelJobStatus(*task.JobID)
			status := result.JobStatus
			db.DB.Model(&task).Update("job_status", status)
			if result.FinishTime != "" {
				db.DB.Model(&task).Update("finish_time", result.FinishTime)
			}
			// 状态变为FAILED，alert_group不为空则通知
			if oldStatus != "FAILED" && status == "FAILED" {
				SendTaskAlert(task, status)
			}
		}
	}
	log.Printf("[Scheduler] 完成同步作业状态")
}

// SyncJobStatusByTaskID 同步单个任务作业状态并回写数据库
func SyncJobStatusByTaskID(taskID uint) (string, error) {
	var task seatunnelModel.EtlTask
	if err := db.DB.First(&task, taskID).Error; err != nil {
		return "", fmt.Errorf("任务不存在: %v", err)
	}
	if task.JobID == nil || *task.JobID == "" {
		return "", fmt.Errorf("任务 jobId 为空")
	}

	oldStatus := task.JobStatus
	result := QuerySeatunnelJobStatus(*task.JobID)
	status := result.JobStatus
	if status == "" {
		status = "UNKNOWN"
	}

	updates := map[string]interface{}{
		"job_status": status,
	}
	if result.FinishTime != "" {
		updates["finish_time"] = result.FinishTime
	}
	if err := db.DB.Model(&task).Updates(updates).Error; err != nil {
		return "", fmt.Errorf("更新任务状态失败: %v", err)
	}

	if oldStatus != "FAILED" && status == "FAILED" {
		SendTaskAlert(task, status)
	}
	return status, nil
}
