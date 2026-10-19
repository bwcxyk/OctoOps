package seatunnel

import (
	"fmt"
	"log"
	"octoops/internal/infra/postgres"
	seatunnelModel "octoops/internal/model/seatunnel"
)

// SyncAllJobStatus 同步全部实时作业状态，返回执行结果与错误。
// 并非所有作业都已提交到 SeaTunnel，查不到（接口返回 200 但没有 jobStatus）属正常情况，只计数不计成败；
// 只有通信失败、接口返回异常等流程问题才算同步任务失败。
func SyncAllJobStatus() (string, error) {
	log.Printf("[Scheduler] 开始同步作业状态")
	var tasks []seatunnelModel.EtlTask
	if err := postgres.DB.Where("task_type = ?", "stream").Find(&tasks).Error; err != nil {
		return "", fmt.Errorf("查询实时作业失败: %w", err)
	}
	synced := 0
	notFound := 0
	failed := 0
	for _, task := range tasks {
		if task.JobID == nil || *task.JobID == "" {
			continue
		}
		oldStatus := task.JobStatus
		result, found, err := QuerySeatunnelJobStatus(*task.JobID)
		if err != nil {
			log.Printf("[Scheduler] 查询作业状态失败: taskID=%d, jobId=%s, error=%v", task.ID, *task.JobID, err)
			failed++
			continue
		}
		if !found {
			notFound++
			continue
		}
		status := result.JobStatus
		if err := postgres.DB.Model(&task).Update("job_status", status).Error; err != nil {
			log.Printf("[Scheduler] 更新作业状态失败: taskID=%d, jobId=%s, error=%v", task.ID, *task.JobID, err)
			failed++
			continue
		}
		if result.FinishTime != "" {
			if err := postgres.DB.Model(&task).Update("finish_time", result.FinishTime).Error; err != nil {
				log.Printf("[Scheduler] 更新作业完成时间失败: taskID=%d, jobId=%s, error=%v", task.ID, *task.JobID, err)
			}
		}
		// 状态变为FAILED，alert_group不为空则通知
		if oldStatus != "FAILED" && status == "FAILED" {
			SendTaskAlert(task, status)
		}
		synced++
	}
	log.Printf("[Scheduler] 完成同步作业状态，成功 %d 个，未找到 %d 个，失败 %d 个", synced, notFound, failed)
	if failed > 0 {
		return "", fmt.Errorf("作业状态同步失败：%d 个作业查询失败", failed)
	}
	return fmt.Sprintf("作业状态同步完成：成功 %d 个，未找到 %d 个", synced, notFound), nil
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
	result, found, err := QuerySeatunnelJobStatus(*task.JobID)
	if err != nil {
		return "", err
	}
	if !found {
		return "", fmt.Errorf("SeaTunnel 上未找到作业 %s", *task.JobID)
	}
	status := result.JobStatus

	updates := map[string]interface{}{
		"job_status": status,
	}
	if result.FinishTime != "" {
		updates["finish_time"] = result.FinishTime
	}
	if err := postgres.DB.Model(&task).Updates(updates).Error; err != nil {
		return "", fmt.Errorf("更新任务状态失败: %w", err)
	}

	if oldStatus != "FAILED" && status == "FAILED" {
		SendTaskAlert(task, status)
	}
	return status, nil
}
