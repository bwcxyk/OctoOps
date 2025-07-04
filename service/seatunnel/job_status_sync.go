package seatunnel

import (
	"log"
	"octoops/db"
	seatunnelModel "octoops/model/seatunnel"
)

func SyncAllJobStatus() {
	log.Printf("[Scheduler] 开始同步作业状态")
	var tasks []seatunnelModel.EtlTask
	db.DB.Where("task_type = ?", "stream").Find(&tasks)
	for _, task := range tasks {
		if task.JobID != "" {
			oldStatus := task.JobStatus
			result := QuerySeatunnelJobStatus(task.JobID)
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
 