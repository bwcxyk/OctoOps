package postgres

import (
	"fmt"
	alertModel "octoops/internal/model/alert"
	aliyunModel "octoops/internal/model/aliyun"
	rbacModel "octoops/internal/model/rbac"
	seatunnelModel "octoops/internal/model/seatunnel"
	systemModel "octoops/internal/model/system"
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
		&taskModel.TaskLog{},
		&systemModel.Setting{},
		&rbacModel.User{},
		&rbacModel.Role{},
		&rbacModel.Permission{},
		&rbacModel.UserRole{},
		&rbacModel.RolePermission{},
	); err != nil {
		return fmt.Errorf("数据库自动迁移失败: %w", err)
	}
	// 历史任务日志补上类型（新日志写入时会带 task_type）
	if err := DB.Exec(
		`UPDATE "task_logs" SET task_type = 'system' WHERE (task_type IS NULL OR task_type = '') AND task_name IN ('作业状态同步', 'ECS 安全组同步', 'ETL 实时作业状态同步')`,
	).Error; err != nil {
		return fmt.Errorf("回填系统任务日志类型失败: %w", err)
	}
	if err := DB.Exec(`UPDATE "task_logs" SET task_type = 'etl' WHERE task_type IS NULL OR task_type = ''`).Error; err != nil {
		return fmt.Errorf("回填任务日志类型失败: %w", err)
	}
	// 系统任务配置已迁移到 settings 表，删除废弃的 custom_tasks 表
	if err := DB.Exec(`DROP TABLE IF EXISTS "custom_tasks"`).Error; err != nil {
		return fmt.Errorf("删除废弃表 custom_tasks 失败: %w", err)
	}
	return nil
}
