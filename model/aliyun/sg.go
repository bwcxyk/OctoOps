package aliyun

import "time"

// 阿里云安全组配置表
type SGConfig struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	Name            string     `gorm:"column:name;size:255" json:"name"`
	AccessKey       string     `gorm:"column:access_key;size:128" json:"access_key"`
	AccessSecret    string     `gorm:"column:access_secret;size:256" json:"access_secret"`
	RegionId        string     `gorm:"column:region_id;size:64" json:"region_id"`
	SecurityGroupId string     `gorm:"column:security_group_id;size:64" json:"security_group_id"`
	PortList        string     `gorm:"column:port_list;size:128" json:"port_list"`
	Status          int        `gorm:"column:status" json:"status"`
	LastIP          string     `gorm:"column:last_ip;size:64" json:"last_ip"` // 最近一次授权的公网IP
	LastIPUpdatedAt *time.Time `gorm:"column:last_ip_updated_at" json:"last_ip_updated_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// 指定表名
func (SGConfig) TableName() string {
	return "aliyun_sg_configs"
}
