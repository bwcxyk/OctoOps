package seatunnel

import (
	"log"
	"octoops/internal/db"
	alertModel "octoops/internal/model/alert"
	seatunnelModel "octoops/internal/model/seatunnel"
	alertService "octoops/internal/service/alert"
	"strings"
)

// SendTaskAlert 发送任务告警（多渠道分发）
func SendTaskAlert(task seatunnelModel.EtlTask, status string) {
	if task.AlertGroup == "" {
		return
	}
	groupIDs := strings.Split(task.AlertGroup, ",")
	for _, gid := range groupIDs {
		var members []alertModel.AlertGroupMember
		db.DB.Where("group_id = ?", gid).Find(&members)
		for _, m := range members {
			var alert alertModel.AlertChannel
			db.DB.First(&alert, m.ChannelID)
			if alert.Status != 1 {
				continue
			}
			var tpl alertModel.AlertTemplate
			if alert.TemplateID != 0 {
				db.DB.First(&tpl, alert.TemplateID)
			}
			data := map[string]interface{}{
				"JobID":     task.JobID,
				"JobName":   task.Name,
				"Status":    status,
				"StartTime": "",
				"EndTime":   "",
				"TaskType":  task.TaskType,
				"Reason":    "作业状态变为" + status,
			}
			if task.LastRunTime != nil {
				data["StartTime"] = task.LastRunTime.Format("2006-01-02 15:04:05")
			}
			if task.FinishTime != nil {
				data["EndTime"] = task.FinishTime.Format("2006-01-02 15:04:05")
			}
			switch alert.Type {
			case "email":
				if tpl.Content != "" {
					err := alertService.SendEmailWithTemplate(&alert, tpl.Content, data)
					if err != nil {
						log.Printf("[ALERT] 邮件发送失败: %v", err)
					}
				}
			case "dingtalk":
				if tpl.Content != "" {
					err := alertService.SendDingTalkMarkdownWithTemplate(alert.Target, alert.DingtalkSecret, "作业告警", tpl.Content, data)
					if err != nil {
						log.Printf("[ALERT] 钉钉发送失败: %v", err)
					}
				}
			}
		}
	}
}
