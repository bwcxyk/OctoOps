package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"octoops/config"
	"octoops/db"
	"octoops/model"
	"strings"
)

// SubmitJobInternal 内部提交作业方法
func SubmitJobInternal(taskID uint, isStartWithSavePoint bool) ([]byte, error) {
	var task model.Task
	if err := db.DB.First(&task, taskID).Error; err != nil {
		return nil, fmt.Errorf("任务不存在: %v", err)
	}

	if task.Config == "" {
		return nil, fmt.Errorf("任务配置为空")
	}

	format := task.ConfigFormat
	if format == "" {
		format = "json"
	}

	// 构建URL
	url := config.SeatunnelBaseURL + "/submit-job?format=" + format
	if task.JobID != "" {
		url += "&jobId=" + task.JobID
	}
	if task.Name != "" {
		url += "&jobName=" + task.Name
	}
	if isStartWithSavePoint {
		url += "&isStartWithSavePoint=true"
	}

	// 发送请求
	resp, err := http.Post(url, "text/plain; charset=utf-8", strings.NewReader(task.Config))
	if err != nil {
		return nil, fmt.Errorf("提交作业失败: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return respBody, fmt.Errorf("提交作业失败，状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// 写入作业日志的公共方法
func WriteTaskLog(task model.Task, octoopsRespBody []byte) {
	var resultMap map[string]interface{}
	_ = json.Unmarshal(octoopsRespBody, &resultMap)
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
	db.DB.Create(&model.TaskLog{
		TaskID:   task.ID,
		JobID:    jobId,
		JobName:  jobName,
		Result:   string(octoopsRespBody),
		TaskType: task.TaskType,
	})
}

// Seatunnel作业状态结构体
type JobStatusResult struct {
	JobStatus    string `json:"jobStatus"`
	FinishTime   string `json:"finishTime"`
}

// 查询 seatunnel 作业状态
func QuerySeatunnelJobStatus(jobId string) JobStatusResult {
	url := config.SeatunnelBaseURL + "/job-info/" + jobId
	resp, err := http.Get(url)
	if err != nil {
		return JobStatusResult{JobStatus: "UNKNOWN"}
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result JobStatusResult
	err = json.Unmarshal(body, &result)
	log.Printf("[DEBUG] 解析后 result: jobStatus=%s, finishTime=%s", result.JobStatus, result.FinishTime)
	if err != nil {
		return JobStatusResult{JobStatus: "UNKNOWN"}
	}
	return result
}

// 发送任务告警（多渠道分发）
func SendTaskAlert(task model.Task, status string) {
	if task.AlertGroup == "" {
		return
	}
	var groupIDs []string = strings.Split(task.AlertGroup, ",")
	for _, gid := range groupIDs {
		var members []model.AlertGroupMember
		db.DB.Where("group_id = ?", gid).Find(&members)
		for _, m := range members {
			var alert model.Alert
			db.DB.First(&alert, m.ChannelID)
			if alert.Status != 1 {
				continue
			}
			// 查找模板内容
			var tpl model.AlertTemplate
			if alert.TemplateID != 0 {
				db.DB.First(&tpl, alert.TemplateID)
			}
			// 组装变量
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
					err := SendEmailWithTemplate(&alert, tpl.Content, data)
					if err != nil {
						log.Printf("[ALERT] 邮件发送失败: %v", err)
					}
				}
			case "dingtalk":
				if tpl.Content != "" {
					err := SendDingTalkMarkdownWithTemplate(alert.Target, alert.DingtalkSecret, "作业告警", tpl.Content, data)
					if err != nil {
						log.Printf("[ALERT] 钉钉发送失败: %v", err)
					}
				}
			}
		}
	}
}
