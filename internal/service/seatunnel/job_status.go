package seatunnel

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"octoops/internal/config"
	"time"
)

// JobStatusResult Seatunnel 作业状态结构体
type JobStatusResult struct {
	JobStatus  string `json:"jobStatus"`
	FinishTime string `json:"finishTime"`
	JobId      string `json:"jobId"`
	JobName    string `json:"jobName"`
}

// QuerySeatunnelJobStatus 查询 seatunnel 作业状态
func QuerySeatunnelJobStatus(jobId string) JobStatusResult {
	url := config.SeatunnelBaseURL + "/job-info/" + jobId
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return JobStatusResult{JobStatus: "UNKNOWN"}
	}
	defer func(body io.ReadCloser) {
		closeErr := body.Close()
		if closeErr != nil {
			fmt.Printf("关闭响应体失败: %v\n", closeErr)
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
