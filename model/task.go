package model

import (
	"time"
	"gorm.io/gorm"
)

type Task struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Name            string         `json:"name"`
	Description     string         `json:"description"`
	CronExpr        string         `json:"cron_expr"`
	Status          int            `json:"status"`
	TaskType        string         `json:"task_type"`
	Config          string         `json:"config"`
	ConfigFormat    string         `json:"config_format"`
	JobID           string         `json:"jobid" gorm:"uniqueIndex"`
	LastRunTime     *time.Time     `json:"last_run_time"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
	JobStatus       string         `json:"job_status"`
	AlertGroup       string         `json:"alert_group"`   // 告警组，存通知组ID列表或逗号分隔
	FinishTime      *time.Time     `json:"finish_time"`
} 

