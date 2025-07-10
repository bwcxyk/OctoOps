package model

import "time"

type CustomTask struct {
    ID           uint      `gorm:"primaryKey" json:"id"`
    Name         string    `gorm:"size:255" json:"name"`
    CustomType   string    `gorm:"size:64" json:"custom_type"`
    CronExpr     string    `gorm:"size:128" json:"cron_expr"`
    Description  string    `gorm:"size:512" json:"description"`
    Status       int       `json:"status"` // 1=启用, 0=禁用
    LastRunTime  *time.Time `json:"last_run_time"`
    LastResult   string    `gorm:"size:1024" json:"last_result"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
} 