package seatunnel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"octoops/internal/config"
	"octoops/internal/db"
	seatunnelModel "octoops/internal/model/seatunnel"
	taskModel "octoops/internal/model/task"
	"strconv"
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

	params := url.Values{}
	params.Set("format", format)
	if isStartWithSavePoint && task.TaskType == "stream" && task.JobID != nil && *task.JobID != "" {
		params.Set("jobId", *task.JobID)
	}
	if task.Name != "" {
		params.Set("jobName", task.Name)
	}
	if isStartWithSavePoint {
		params.Set("isStartWithSavePoint", "true")
	}
	requestURL := config.SeatunnelBaseURL + "/submit-job?" + params.Encode()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(requestURL, "text/plain; charset=utf-8", strings.NewReader(task.Config))
	if err != nil {
		return nil, fmt.Errorf("提交作业失败: %v", err)
	}
	defer func(body io.ReadCloser) {
		closeErr := body.Close()
		if closeErr != nil {
			fmt.Printf("关闭响应体失败: %v\n", closeErr)
		}
	}(resp.Body)

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return respBody, fmt.Errorf("提交作业失败，状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// WriteTaskLog 写入作业日志（提交成功）
func WriteTaskLog(task seatunnelModel.EtlTask, octoopsRespBody []byte) {
	WriteTaskLogWithStatus(task, octoopsRespBody, "success")
}

// UpdateJobIdFromResponse 从响应体中提取 jobId 并更新到数据库
func UpdateJobIdFromResponse(taskID uint, octoopsRespBody []byte) {
	var resultMap map[string]interface{}
	decoder := json.NewDecoder(bytes.NewReader(octoopsRespBody))
	decoder.UseNumber()
	if err := decoder.Decode(&resultMap); err != nil {
		log.Printf("[DEBUG] 解析响应失败: %v", err)
		return
	}

	jobID := ""
	if raw, ok := resultMap["jobId"]; ok {
		jobID = normalizeJobID(raw)
	} else if raw, ok := resultMap["job_id"]; ok {
		jobID = normalizeJobID(raw)
	}

	if jobID != "" {
		if err := db.DB.Model(&seatunnelModel.EtlTask{}).Where("id = ?", taskID).Update("job_id", jobID).Error; err != nil {
			log.Printf("[ERROR] 更新 jobId 失败: taskID=%d, jobId=%s, error=%v", taskID, jobID, err)
		} else {
			log.Printf("[INFO] jobId 更新成功: taskID=%d, jobId=%s", taskID, jobID)
		}
	}
}

func normalizeJobID(raw interface{}) string {
	switch v := raw.(type) {
	case string:
		return strings.TrimSpace(v)
	case json.Number:
		return v.String()
	case float64:
		return strconv.FormatInt(int64(v), 10)
	default:
		return ""
	}
}

// WriteTaskLogWithStatus 写入作业日志（指定状态）
func WriteTaskLogWithStatus(task seatunnelModel.EtlTask, octoopsRespBody []byte, status string) {
	var resultMap map[string]interface{}
	_ = json.Unmarshal(octoopsRespBody, &resultMap)
	taskName := task.Name
	if v, ok := resultMap["jobName"].(string); ok {
		taskName = v
	}
	db.DB.Create(&taskModel.TaskLog{
		TaskName: taskName,
		Result:   string(octoopsRespBody),
		Status:   status,
	})
}
