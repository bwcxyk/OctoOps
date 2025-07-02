package model

import "time"

type CustomTask struct {
    ID           uint      `gorm:"primaryKey" json:"id"`
    Name         string    `json:"name"`
    CronExpr     string    `json:"cron_expr"`
    Status       int       `json:"status"` // 1=启用, 0=禁用
    CustomType   string    `json:"custom_type"`
    Description  string    `json:"description"`
    LastRunTime  *time.Time `json:"last_run_time"`
    LastResult   string    `json:"last_result"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
} 