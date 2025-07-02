package model

import "time"

// 通知管理表
// CREATE TABLE alerts (...)
type Alert struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"column:name" json:"name"`           // 通知名称
	Type      string    `gorm:"column:type" json:"type"`           // 通知类型（如 email、webhook、sms 等）
	Target    string    `gorm:"column:target" json:"target"`       // 通知目标（如邮箱、URL、手机号等）
	DingtalkSecret string `gorm:"column:dingtalk_secret" json:"dingtalk_secret"` // 钉钉加签密钥
	Status    int       `json:"status"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	TemplateID uint `json:"template_id"`
} 