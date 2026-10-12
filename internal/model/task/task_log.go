package task

import (
	"time"
)

// 任务日志类型
const (
	TaskTypeETL    = "etl"
	TaskTypeSystem = "system"
)

type TaskLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TaskName  string    `gorm:"size:255" json:"task_name"`      // 任务名称
	TaskType  string    `gorm:"size:32;index" json:"task_type"` // 任务类型：etl、system
	Status    string    `gorm:"size:64" json:"status"`          // 状态：success、failed
	Result    string    `gorm:"size:2048" json:"result"`        // 返回内容
	CreatedAt time.Time `json:"created_at"`
}
