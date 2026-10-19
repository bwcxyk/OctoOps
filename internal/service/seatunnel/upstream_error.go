package seatunnel

import "fmt"

// UpstreamError SeaTunnel 接口调用失败，保留上游状态码与原始响应，便于接口层结构化返回
type UpstreamError struct {
	Action     string // 操作名，如「提交作业」「停止作业」
	StatusCode int    // 上游 HTTP 状态码，请求未到达上游时为 0
	Body       []byte // 上游原始响应体
	Err        error  // 请求未到达上游时的错误（如连接失败）
}

// NewUpstreamError 上游返回非 2xx 时的错误
func NewUpstreamError(action string, statusCode int, body []byte) *UpstreamError {
	return &UpstreamError{Action: action, StatusCode: statusCode, Body: body}
}

// NewUpstreamRequestError 请求未到达上游时的错误
func NewUpstreamRequestError(action string, err error) *UpstreamError {
	return &UpstreamError{Action: action, Err: err}
}

func (e *UpstreamError) Error() string {
	// 用于日志与任务日志 result，保留状态码便于排查
	if e.Err != nil {
		return fmt.Sprintf("%s失败: %v", e.Action, e.Err)
	}
	return fmt.Sprintf("%s失败，状态码: %d, 响应: %s", e.Action, e.StatusCode, string(e.Body))
}

func (e *UpstreamError) Unwrap() error {
	return e.Err
}
