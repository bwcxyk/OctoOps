package db

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"octoops/config"
	"octoops/model"
)

var DB *gorm.DB

func Init() {
	// 支持通过环境变量或 config 包配置数据库连接
	dsn := config.PostgresDSN
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect database: %v", err))
	}
	DB.AutoMigrate(
		&model.Task{},
		&model.TaskLog{},
		&model.AliyunSGConfig{},
		&model.Alert{},
		&model.CustomTask{},
		&model.AlertGroup{},
		&model.AlertGroupMember{},
		&model.AlertTemplate{},
	)
}
