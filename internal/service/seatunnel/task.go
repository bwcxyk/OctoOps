package seatunnel

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"octoops/internal/config"
	"octoops/internal/db"
	"octoops/internal/model"
	alertModel "octoops/internal/model/alert"
	seatunnelModel "octoops/internal/model/seatunnel"
	alertService "octoops/internal/service/alert"
	"strings"
	"time"
)

// SubmitJobInternal 内部提交作业方法
func SubmitJobInternal(taskID uint, isStartWithSavePoint bool) ([]byte, error) {
	var task seatunnelModel.EtlTask
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
	// 仅实时任务（stream）传递 jobId，离线任务（batch）不传
	if task.TaskType == "stream" && task.JobID != nil && *task.JobID != "" {
		url += "&jobId=" + *task.JobID
	}
	if task.Name != "" {
		url += "&jobName=" + task.Name
	}
	if isStartWithSavePoint {
		url += "&isStartWithSavePoint=true"
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "text/plain; charset=utf-8", strings.NewReader(task.Config))
	if err != nil {
		return nil, fmt.Errorf("提交作业失败: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("关闭响应体失败: %v\n", err)
		}
	}(resp.Body)

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return respBody, fmt.Errorf("提交作业失败，状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// 写入作业日志的公共方法
func WriteTaskLog(task seatunnelModel.EtlTask, octoopsRespBody []byte) {
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
	JobStatus  string `json:"jobStatus"`
	FinishTime string `json:"finishTime"`
	JobId      string `json:"jobId"`
	JobName    string `json:"jobName"`
}

// 查询 seatunnel 作业状态
func QuerySeatunnelJobStatus(jobId string) JobStatusResult {
	url := config.SeatunnelBaseURL + "/job-info/" + jobId
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return JobStatusResult{JobStatus: "UNKNOWN"}
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("关闭响应体失败: %v\n", err)
		}
	}(resp.Body)
	body, _ := io.ReadAll(resp.Body)
	var result JobStatusResult
	err = json.Unmarshal(body, &result)
	log.Printf("[DEBUG] 解析后 result: jobId=%s, jobName=%s, jobStatus=%s, finishTime=%s", result.JobId, result.JobName, result.JobStatus, result.FinishTime)
	if err != nil {
		return JobStatusResult{JobStatus: "UNKNOWN"}
	}
	return result
}

// 发送任务告警（多渠道分发）
func SendTaskAlert(task seatunnelModel.EtlTask, status string) {
	if task.AlertGroup == "" {
		return
	}
	var groupIDs = strings.Split(task.AlertGroup, ",")
	for _, gid := range groupIDs {
		var members []alertModel.AlertGroupMember
		db.DB.Where("group_id = ?", gid).Find(&members)
		for _, m := range members {
			var alert alertModel.Channel
			db.DB.First(&alert, m.ChannelID)
			if alert.Status != 1 {
				continue
			}
			// 查找模板内容
			var tpl alertModel.AlertTemplate
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
