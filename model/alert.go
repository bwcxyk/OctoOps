package model

import "time"

// 通知管理表
// CREATE TABLE alerts (...)
type Alert struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"column:name;size:255" json:"name"`           // 通知名称
	Type      string    `gorm:"column:type;size:64" json:"type"`           // 通知类型
	Target    string    `gorm:"column:target;size:255" json:"target"`       // 通知目标
	DingtalkSecret string `gorm:"column:dingtalk_secret;size:255" json:"dingtalk_secret"` // 钉钉加签密钥
	Status    int       `json:"status"`
	TemplateID uint     `json:"template_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
} 