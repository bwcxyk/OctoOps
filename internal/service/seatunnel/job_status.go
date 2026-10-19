package seatunnel

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"octoops/internal/config"
	"strings"
)

// JobStatusResult Seatunnel 作业状态结构体
type JobStatusResult struct {
	JobStatus  string `json:"jobStatus"`
	FinishTime string `json:"finishTime"`
	JobId      string `json:"jobId"`
	JobName    string `json:"jobName"`
}

// QuerySeatunnelJobStatus 查询 seatunnel 作业状态；请求、状态码、解析或字段缺失都会返回 error
func QuerySeatunnelJobStatus(jobId string) (JobStatusResult, error) {
	if strings.TrimSpace(jobId) == "" {
		return JobStatusResult{}, fmt.Errorf("jobId 为空")
	}
	url := config.SeatunnelBaseURL + "/job-info/" + jobId
	client := &http.Client{Timeout: config.SeatunnelRequestTimeout}
	resp, err := client.Get(url)
	if err != nil {
		return JobStatusResult{}, fmt.Errorf("请求 SeaTunnel 作业状态失败: %w", err)
	}
	defer func(body io.ReadCloser) {
		closeErr := body.Close()
		if closeErr != nil {
			fmt.Printf("关闭响应体失败: %v\n", closeErr)
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return JobStatusResult{}, fmt.Errorf("读取 SeaTunnel 响应失败: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return JobStatusResult{}, fmt.Errorf("SeaTunnel 返回状态码 %d: %s", resp.StatusCode, truncateResponse(body))
	}
	var result JobStatusResult
	if err := json.Unmarshal(body, &result); err != nil {
		return JobStatusResult{}, fmt.Errorf("解析 SeaTunnel 响应失败: %w", err)
	}
	if result.JobStatus == "" {
		return result, fmt.Errorf("SeaTunnel 响应缺少 jobStatus: %s", truncateResponse(body))
	}
	return result, nil
}

// truncateResponse 限制错误信息中的上游响应长度，避免日志/任务日志过长
func truncateResponse(body []byte) string {
	const maxLen = 512
	if len(body) <= maxLen {
		return string(body)
	}
	return strings.ToValidUTF8(string(body[:maxLen]), "") + "..."
}
