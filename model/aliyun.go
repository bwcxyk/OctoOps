package model

import "time"

// 阿里云安全组配置表
// CREATE TABLE aliyun_sg_config (...)
type AliyunSGConfig struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	AccountName       string    `gorm:"column:account_name" json:"account_name"`
	AccessKey         string    `gorm:"column:access_key" json:"access_key"`
	AccessSecret      string    `gorm:"column:access_secret" json:"access_secret"`
	RegionId          string    `gorm:"column:region_id" json:"region_id"`
	SecurityGroupId   string    `gorm:"column:security_group_id" json:"security_group_id"`
	PortList          string    `gorm:"column:port_list" json:"port_list"`
	LastIP            string    `gorm:"column:last_ip" json:"last_ip"` // 最近一次授权的公网IP
	LastIPUpdatedAt *time.Time `gorm:"column:last_ip_updated_at" json:"last_ip_updated_at"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedAt         time.Time `json:"created_at"`
}
