package postgres

import (
	"fmt"
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

func Init(dsn string) error {
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return nil
}

func Migrate() error {
	if DB == nil {
		return fmt.Errorf("database is not initialized")
	}
	if err := DB.AutoMigrate(
		&seatunnelModel.EtlTask{},
		&aliyunModel.SGConfig{},
		&alertModel.AlertChannel{},
		&alertModel.AlertGroup{},
		&alertModel.AlertGroupMember{},
		&alertModel.AlertTemplate{},
		&taskModel.CustomTask{},
		&taskModel.TaskLog{},
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
