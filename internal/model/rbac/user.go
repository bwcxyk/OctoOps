package model

import (
	"time"
)

// User 用户模型
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"uniqueIndex;not null"`
	Password  string    `json:"-" gorm:"not null"` // 密码不返回给前端
	Email     string    `json:"email" gorm:"uniqueIndex"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	Status    int       `json:"status" gorm:"default:1"` // 1:正常 0:禁用
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// 关联关系
	Roles []Role `json:"roles" gorm:"many2many:user_roles;"`
}

// UserRole 用户角色关联表
type UserRole struct {
	UserID uint `json:"user_id" gorm:"primaryKey"`
	RoleID uint `json:"role_id" gorm:"primaryKey"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

func (UserRole) TableName() string {
	return "user_roles"
} 