package alert

import "time"

type AlertGroup struct {
    ID          uint      `gorm:"primaryKey" json:"id"`
    Name        string    `gorm:"size:255" json:"name"`
    Description string    `gorm:"size:512" json:"description"`
    Status      int       `json:"status"` // 0=禁用, 1=启用
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type AlertGroupMember struct {
    ID          uint      `gorm:"primaryKey" json:"id"`
    GroupID     uint      `json:"group_id"`
    ChannelType string    `gorm:"size:64" json:"channel_type"` // email/robot/other
    ChannelID   uint      `json:"channel_id"`
    CreatedAt   time.Time `json:"created_at"`
} 
