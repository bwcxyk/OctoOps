package model

import "time"

type AlertGroup struct {
    ID          uint      `gorm:"primaryKey" json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Status      int       `json:"status"` // 0=禁用, 1=启用
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type AlertGroupMember struct {
    ID          uint      `gorm:"primaryKey" json:"id"`
    GroupID     uint      `json:"group_id"`
    ChannelType string    `json:"channel_type"` // email/robot/other
    ChannelID   uint      `json:"channel_id"`
    CreatedAt   time.Time `json:"created_at"`
} 