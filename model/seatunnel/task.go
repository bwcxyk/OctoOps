package model

import (
	"time"
	"gorm.io/gorm"
)

type Task struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Name            string         `gorm:"size:255" json:"name"`
	Description     string         `gorm:"size:512" json:"description"`
	TaskType        string         `gorm:"size:64" json:"task_type"`
	CronExpr        string         `gorm:"size:128" json:"cron_expr"`
	Config          string         `json:"config"`
	ConfigFormat    string         `gorm:"size:32" json:"config_format"`
	JobID           string         `gorm:"size:128;uniqueIndex" json:"jobid"`
	JobStatus       string         `gorm:"size:64" json:"job_status"`
	AlertGroup      string         `gorm:"size:255" json:"alert_group"`
	Status          int            `json:"status"`
	LastRunTime     *time.Time     `json:"last_run_time"`
	FinishTime      *time.Time     `json:"finish_time"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
} 