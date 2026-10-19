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

// QuerySeatunnelJobStatus 查询 SeaTunnel 作业状态。
// 作业未提交到 SeaTunnel 时接口仍返回 200，只是响应里没有 jobStatus（形如 {"jobId":"xxx"}），
// 此时 found=false 且 err 为 nil；只有通信失败、非 2xx 响应或响应无法解析才返回 error。
func QuerySeatunnelJobStatus(jobId string) (JobStatusResult, bool, error) {
	if strings.TrimSpace(jobId) == "" {
		return JobStatusResult{}, false, fmt.Errorf("jobId 为空")
	}
	url := config.SeatunnelBaseURL + "/job-info/" + jobId
	client := &http.Client{Timeout: config.SeatunnelRequestTimeout}
	resp, err := client.Get(url)
	if err != nil {
		return JobStatusResult{}, false, fmt.Errorf("请求 SeaTunnel 作业状态失败: %w", err)
	}
	defer func(body io.ReadCloser) {
		closeErr := body.Close()
		if closeErr != nil {
			fmt.Printf("关闭响应体失败: %v\n", closeErr)
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return JobStatusResult{}, false, fmt.Errorf("读取 SeaTunnel 响应失败: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return JobStatusResult{}, false, fmt.Errorf("SeaTunnel 返回状态码 %d: %s", resp.StatusCode, truncateResponse(body))
	}
	var result JobStatusResult
	if err := json.Unmarshal(body, &result); err != nil {
		return JobStatusResult{}, false, fmt.Errorf("解析 SeaTunnel 响应失败: %w", err)
	}
	if result.JobStatus == "" {
		return result, false, nil
	}
	return result, true, nil
}

// truncateResponse 限制错误信息中的上游响应长度，避免日志/任务日志过长
func truncateResponse(body []byte) string {
	const maxLen = 512
	if len(body) <= maxLen {
		return string(body)
	}
	return strings.ToValidUTF8(string(body[:maxLen]), "") + "..."
}
