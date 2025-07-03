package model

import (
    "time"
)

type TaskLog struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    TaskID    uint           `json:"task_id"`      // 关联任务ID
    JobID     string         `gorm:"size:128" json:"job_id"`       // octoops返回的jobId
    JobName   string         `gorm:"size:255" json:"job_name"`     // octoops返回的jobName
    TaskType  string         `gorm:"size:64" json:"task_type"`
    Result    string         `gorm:"size:2048" json:"result"`       // octoops原始返回内容（json字符串）
    CreatedAt time.Time      `json:"created_at"`
} 