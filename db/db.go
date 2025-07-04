package db

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"octoops/config"
	"octoops/model"
	aliyunModel "octoops/model/aliyun"
	seatunnelModel "octoops/model/seatunnel"
	alertModel "octoops/model/alert"
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
		&seatunnelModel.EtlTask{},
		&model.TaskLog{},
		&aliyunModel.AliyunSGConfig{},
		&alertModel.Alert{},
		&alertModel.AlertGroup{},
		&alertModel.AlertGroupMember{},
		&alertModel.AlertTemplate{},
		&model.CustomTask{},
	)
}
