package system

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"octoops/internal/infra/postgres"
	systemModel "octoops/internal/model/system"
	syssetting "octoops/internal/pkg/setting"
	"octoops/internal/scheduler"

	"github.com/robfig/cron/v3"
)

// cronParser 与调度器保持一致：6 段表达式（秒 分 时 日 月 周）
var cronParser = cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

func taskSettingKey(taskType string) string { return "task." + taskType }

// GetSettings 返回全部设置项，value 已还原为 JSON 对象
func GetSettings() (map[string]json.RawMessage, error) {
	var settings []systemModel.Setting
	if err := postgres.DB.Find(&settings).Error; err != nil {
		return nil, err
	}
	result := make(map[string]json.RawMessage, len(settings))
	for _, s := range settings {
		result[s.Key] = json.RawMessage(s.Value)
	}
	// 邮件密码不返回明文（库中为密文），前端原样回传 MaskedSecret 表示不修改
	var ms systemModel.MailSetting
	found, err := syssetting.GetMail(&ms)
	if err != nil {
		return nil, err
	}
	if found {
		if ms.SMTPPassword != "" {
			ms.SMTPPassword = systemModel.MaskedSecret
		}
		if masked, err := json.Marshal(ms); err == nil {
			result[systemModel.KeyMailSMTP] = masked
		}
	}
	return result, nil
}

// UpdateSettings 校验并保存设置项，保存成功后热更新调度器
func UpdateSettings(values map[string]json.RawMessage) error {
	taskUpdates := make(map[string]systemModel.TaskSetting, len(values))
	retentionDays := 0
	hasRetention := false
	var mailSetting systemModel.MailSetting
	hasMail := false

	for key, raw := range values {
		switch {
		case strings.HasPrefix(key, "task."):
			taskType := strings.TrimPrefix(key, "task.")
			if _, ok := scheduler.SystemTaskDefByType(taskType); !ok {
				return fmt.Errorf("未知的设置项: %s", key)
			}
			var ts systemModel.TaskSetting
			if err := json.Unmarshal(raw, &ts); err != nil {
				return fmt.Errorf("设置项 %s 格式错误: %w", key, err)
			}
			if strings.TrimSpace(ts.Cron) == "" {
				return fmt.Errorf("设置项 %s 缺少调度周期", key)
			}
			if _, err := cronParser.Parse(strings.TrimSpace(ts.Cron)); err != nil {
				return fmt.Errorf("设置项 %s 的调度周期不合法: %w", key, err)
			}
			taskUpdates[taskType] = ts
		case key == systemModel.KeyTaskLogRetentionDays:
			var rs systemModel.RetentionSetting
			if err := json.Unmarshal(raw, &rs); err != nil {
				return fmt.Errorf("设置项 %s 格式错误: %w", key, err)
			}
			if rs.Days != systemModel.TaskLogRetentionDisabled && rs.Days < 1 {
				return fmt.Errorf("日志保留天数需不小于 1，-1 表示不清理")
			}
			retentionDays = rs.Days
			hasRetention = true
		case key == systemModel.KeyMailSMTP:
			if err := json.Unmarshal(raw, &mailSetting); err != nil {
				return fmt.Errorf("设置项 %s 格式错误: %w", key, err)
			}
			if mailSetting.Enable {
				if strings.TrimSpace(mailSetting.SMTPAddress) == "" {
					return fmt.Errorf("启用邮件时必须填写 SMTP 服务器地址")
				}
				if strings.TrimSpace(mailSetting.SMTPUser) == "" {
					return fmt.Errorf("启用邮件时必须填写邮箱账号")
				}
			}
			if mailSetting.SMTPPort == 0 {
				mailSetting.SMTPPort = systemModel.DefaultSMTPPort
			}
			if !systemModel.ValidSecurity(mailSetting.Security) {
				return fmt.Errorf("邮件加密方式只能是 none（无）、ssl 或 tls")
			}
			hasMail = true
		default:
			return fmt.Errorf("未知的设置项: %s", key)
		}
	}

	for taskType, ts := range taskUpdates {
		if err := saveTaskSetting(taskSettingKey(taskType), ts); err != nil {
			return fmt.Errorf("保存设置项 %s 失败: %w", taskType, err)
		}
	}
	if hasRetention {
		if err := saveRetentionSetting(retentionDays); err != nil {
			return fmt.Errorf("保存设置项 %s 失败: %w", systemModel.KeyTaskLogRetentionDays, err)
		}
	}
	if hasMail {
		if err := saveMailSetting(mailSetting); err != nil {
			return fmt.Errorf("保存设置项 %s 失败: %w", systemModel.KeyMailSMTP, err)
		}
	}
	// 保存成功后再应用到调度器
	for taskType, ts := range taskUpdates {
		scheduler.ApplySystemTask(taskType, ts.Enabled, ts.Cron)
	}
	return nil
}

// GetTaskLogRetentionDays 读取日志保留天数，-1 表示不清理
func GetTaskLogRetentionDays() int {
	setting, ok := loadSetting(systemModel.KeyTaskLogRetentionDays)
	if !ok {
		return systemModel.DefaultTaskLogRetentionDays
	}
	var rs systemModel.RetentionSetting
	if err := json.Unmarshal([]byte(setting.Value), &rs); err != nil {
		log.Printf("[System] 设置项 %s 解析失败，使用默认值: %v", systemModel.KeyTaskLogRetentionDays, err)
		return systemModel.DefaultTaskLogRetentionDays
	}
	if rs.Days != systemModel.TaskLogRetentionDisabled && rs.Days < 1 {
		return systemModel.DefaultTaskLogRetentionDays
	}
	return rs.Days
}

// loadSetting 读取设置项，不存在时返回 false（用 Find 避免 gorm 打印 record not found）
func loadSetting(key string) (systemModel.Setting, bool) {
	var setting systemModel.Setting
	if err := postgres.DB.Where("\"key\" = ?", key).Limit(1).Find(&setting).Error; err != nil || setting.Key == "" {
		return systemModel.Setting{}, false
	}
	return setting, true
}

func saveRetentionSetting(days int) error {
	raw, err := json.Marshal(systemModel.RetentionSetting{Days: days})
	if err != nil {
		return err
	}
	if setting, ok := loadSetting(systemModel.KeyTaskLogRetentionDays); ok {
		setting.Value = string(raw)
		setting.UpdatedAt = time.Now()
		return postgres.DB.Save(&setting).Error
	}
	return postgres.DB.Create(&systemModel.Setting{Key: systemModel.KeyTaskLogRetentionDays, Value: string(raw)}).Error
}

// saveMailSetting 保存邮件设置；密码为脱敏占位值时保持原密码不变，明文密码落库前加密
func saveMailSetting(ms systemModel.MailSetting) error {
	if ms.SMTPPassword == systemModel.MaskedSecret || strings.TrimSpace(ms.SMTPPassword) == "" {
		var old systemModel.MailSetting
		found, err := syssetting.GetMail(&old)
		if err != nil {
			return err
		}
		if found {
			ms.SMTPPassword = old.SMTPPassword
		}
	}
	return syssetting.SetMail(ms)
}

func saveTaskSetting(key string, ts systemModel.TaskSetting) error {
	raw, err := json.Marshal(ts)
	if err != nil {
		return err
	}
	if setting, ok := loadSetting(key); ok {
		setting.Value = string(raw)
		setting.UpdatedAt = time.Now()
		return postgres.DB.Save(&setting).Error
	}
	return postgres.DB.Create(&systemModel.Setting{Key: key, Value: string(raw)}).Error
}
