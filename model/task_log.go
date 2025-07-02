package model

import (
    "time"
)

type TaskLog struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    TaskID    uint           `json:"task_id"`      // 关联任务ID
    JobID     string         `json:"job_id"`       // octoops返回的jobId
    JobName   string         `json:"job_name"`     // octoops返回的jobName
    Result    string         `json:"result"`       // octoops原始返回内容（json字符串）
    TaskType  string         `json:"task_type"`
    CreatedAt time.Time      `json:"created_at"`
} 