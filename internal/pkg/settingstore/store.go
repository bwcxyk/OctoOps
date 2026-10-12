package settingstore

import (
	"encoding/json"
	"fmt"
	"time"

	"octoops/internal/infra/postgres"
	systemModel "octoops/internal/model/system"
	"octoops/internal/pkg/crypto"
)

// Get 读取设置项并解析到 dest，不存在或解析失败返回 false
func Get(key string, dest any) bool {
	var setting systemModel.Setting
	if err := postgres.DB.Where("\"key\" = ?", key).Limit(1).Find(&setting).Error; err != nil || setting.Key == "" {
		return false
	}
	return json.Unmarshal([]byte(setting.Value), dest) == nil
}

// Set 写入设置项，已存在则更新 value
func Set(key string, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	var setting systemModel.Setting
	if err := postgres.DB.Where("\"key\" = ?", key).Limit(1).Find(&setting).Error; err == nil && setting.Key != "" {
		return postgres.DB.Model(&systemModel.Setting{}).
			Where("\"key\" = ?", key).
			Updates(map[string]interface{}{"value": string(raw), "updated_at": time.Now()}).Error
	}
	return postgres.DB.Create(&systemModel.Setting{Key: key, Value: string(raw)}).Error
}

// GetMail 读取邮件设置并解密密码，返回该项是否存在
func GetMail(dest *systemModel.MailSetting) (bool, error) {
	if !Get(systemModel.KeyMailSMTP, dest) {
		return false, nil
	}
	if dest.SMTPPassword == "" {
		return true, nil
	}
	plain, err := crypto.DecryptAES(dest.SMTPPassword)
	if err != nil {
		return true, fmt.Errorf("邮件密码解密失败，请确认 octoops.security.aes_key 与加密时一致: %w", err)
	}
	dest.SMTPPassword = plain
	return true, nil
}

// SetMail 写入邮件设置，密码在落库前加密
func SetMail(ms systemModel.MailSetting) error {
	encrypted, err := crypto.EncryptAES(ms.SMTPPassword)
	if err != nil {
		return err
	}
	ms.SMTPPassword = encrypted
	return Set(systemModel.KeyMailSMTP, ms)
}
