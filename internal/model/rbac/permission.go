package model

import (
	"time"
)

// Permission 权限模型
type Permission struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null"`
	Code        string    `json:"code" gorm:"uniqueIndex;not null"` // 权限代码，如: user:read
	Description string    `json:"description"`
	Type        string    `json:"type"` // menu:菜单权限 api:接口权限
	Path        string    `json:"path"` // 菜单路径或API路径
	Method      string    `json:"method"` // HTTP方法，如: GET, POST, PUT, DELETE
	Status      int       `json:"status" gorm:"default:1"` // 1:正常 0:禁用
	ParentID    uint      `json:"parent_id" gorm:"default:0;index"` // 新增
	Children    []Permission `json:"children" gorm:"-"`
	OrderNum    int       `json:"order_num" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	
	// 关联关系
	Roles []Role `json:"roles" gorm:"many2many:role_permissions;"`
}

// TableName 指定表名
func (Permission) TableName() string {
	return "permissions"
} 