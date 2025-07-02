package model

import "time"

type AlertTemplate struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    Name      string    `json:"name"`
    Type      string    `json:"type"`     // dd/wx/fs/email
    Content   string    `json:"content"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
} 