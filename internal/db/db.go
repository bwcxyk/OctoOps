package db

import (
	"fmt"
	"log"
	"octoops/internal/config"
	"octoops/internal/model"
	alertModel "octoops/internal/model/alert"
	aliyunModel "octoops/internal/model/aliyun"
	rbacModel "octoops/internal/model/rbac"
	seatunnelModel "octoops/internal/model/seatunnel"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

	// Set connection pool parameters
	sqlDB, err := DB.DB()
	if err != nil {
		panic(fmt.Sprintf("failed to get database connection: %v", err))
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	if err := DB.AutoMigrate(
		&seatunnelModel.EtlTask{},
		&model.TaskLog{},
		&aliyunModel.SGConfig{},
		&alertModel.Alert{},
		&alertModel.AlertGroup{},
		&alertModel.AlertGroupMember{},
		&alertModel.AlertTemplate{},
		&model.CustomTask{},
		// RBAC相关模型
		&rbacModel.User{},
		&rbacModel.Role{},
		&rbacModel.Permission{},
		&rbacModel.UserRole{},
		&rbacModel.RolePermission{},
	); err != nil {
		log.Fatalf("数据库自动迁移失败: %v", err)
	}
}
