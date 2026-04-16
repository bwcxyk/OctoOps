package db

import (
	"fmt"
	"octoops/internal/config"
	alertModel "octoops/internal/model/alert"
	aliyunModel "octoops/internal/model/aliyun"
	rbacModel "octoops/internal/model/rbac"
	seatunnelModel "octoops/internal/model/seatunnel"
	taskModel "octoops/internal/model/task"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() error {
	// 支持通过环境变量或 config 包配置数据库连接
	dsn := config.PostgresDSN
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	// Set connection pool parameters
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	// Auto migrate models
	if err := DB.AutoMigrate(
		&seatunnelModel.EtlTask{},
		&aliyunModel.SGConfig{},
		&alertModel.AlertChannel{},
		&alertModel.AlertGroup{},
		&alertModel.AlertGroupMember{},
		&alertModel.AlertTemplate{},
		&taskModel.CustomTask{},
		&taskModel.TaskLog{},
		// RBAC相关模型
		&rbacModel.User{},
		&rbacModel.Role{},
		&rbacModel.Permission{},
		&rbacModel.UserRole{},
		&rbacModel.RolePermission{},
	); err != nil {
		return fmt.Errorf("数据库自动迁移失败: %w", err)
	}
	return nil
}
