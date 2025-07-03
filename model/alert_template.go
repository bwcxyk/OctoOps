package model

import "time"

type AlertTemplate struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    Name      string    `gorm:"size:255" json:"name"`
    Type      string    `gorm:"size:64" json:"type"`     // dd/wx/fs/email
    Content   string    `gorm:"size:2048" json:"content"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
} 