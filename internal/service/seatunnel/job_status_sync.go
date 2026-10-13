package seatunnel

import (
	"fmt"
	"log"
	"octoops/internal/infra/postgres"
	seatunnelModel "octoops/internal/model/seatunnel"
)

// SyncAllJobStatus 同步全部实时作业状态，返回执行结果与错误
func SyncAllJobStatus() (string, error) {
	log.Printf("[Scheduler] 开始同步作业状态")
	var tasks []seatunnelModel.EtlTask
	if err := postgres.DB.Where("task_type = ?", "stream").Find(&tasks).Error; err != nil {
		return "", fmt.Errorf("查询实时作业失败: %w", err)
	}
	synced := 0
	for _, task := range tasks {
		if task.JobID == nil || *task.JobID == "" {
			continue
		}
		oldStatus := task.JobStatus
		result := QuerySeatunnelJobStatus(*task.JobID)
		status := result.JobStatus
		postgres.DB.Model(&task).Update("job_status", status)
		if result.FinishTime != "" {
			postgres.DB.Model(&task).Update("finish_time", result.FinishTime)
		}
		// 状态变为FAILED，alert_group不为空则通知
		if oldStatus != "FAILED" && status == "FAILED" {
			SendTaskAlert(task, status)
		}
		synced++
	}
	log.Printf("[Scheduler] 完成同步作业状态，共 %d 个作业", synced)
	return fmt.Sprintf("作业状态同步完成，共 %d 个作业", synced), nil
}

// SyncJobStatusByTaskID 同步单个任务作业状态并回写数据库
func SyncJobStatusByTaskID(taskID uint) (string, error) {
	var task seatunnelModel.EtlTask
	if err := postgres.DB.First(&task, taskID).Error; err != nil {
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
	if err := postgres.DB.Model(&task).Updates(updates).Error; err != nil {
		return "", fmt.Errorf("更新任务状态失败: %v", err)
	}

	if oldStatus != "FAILED" && status == "FAILED" {
		SendTaskAlert(task, status)
	}
	return status, nil
}
