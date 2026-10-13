// Package setting 系统设置项的读写入口：JSON 存取、敏感字段加解密，以及邮件配置的进程内缓存
package setting

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"octoops/internal/infra/postgres"
	systemModel "octoops/internal/model/system"
	"octoops/internal/pkg/crypto"
)

// Get 读取设置项并解析到 dest，不存在或解析失败返回 false
func Get(key string, dest any) bool {
	var row systemModel.Setting
	if err := postgres.DB.Where("\"key\" = ?", key).Limit(1).Find(&row).Error; err != nil || row.Key == "" {
		return false
	}
	return json.Unmarshal([]byte(row.Value), dest) == nil
}

// Set 写入设置项，已存在则更新 value
func Set(key string, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	var row systemModel.Setting
	if err := postgres.DB.Where("\"key\" = ?", key).Limit(1).Find(&row).Error; err == nil && row.Key != "" {
		return postgres.DB.Model(&systemModel.Setting{}).
			Where("\"key\" = ?", key).
			Updates(map[string]interface{}{"value": string(raw), "updated_at": time.Now()}).Error
	}
	return postgres.DB.Create(&systemModel.Setting{Key: key, Value: string(raw)}).Error
}

// 邮件配置缓存：告警等场景下会频繁发信，避免每次都查库并解密
// 缓存常驻内存，只在保存配置（SetMail）时失效，进程重启后重新加载
// 多实例部署时这里需要换成 Redis 等共享缓存，并由变更方广播失效
var (
	mailCacheMu    sync.RWMutex
	mailCacheValid bool
	mailCacheFound bool
	mailCacheValue systemModel.MailSetting
)

// loadMailCache 命中缓存时返回配置，未缓存返回 hit=false
func loadMailCache() (systemModel.MailSetting, bool, bool) {
	mailCacheMu.RLock()
	defer mailCacheMu.RUnlock()
	if !mailCacheValid {
		return systemModel.MailSetting{}, false, false
	}
	return mailCacheValue, mailCacheFound, true
}

func storeMailCache(ms systemModel.MailSetting, found bool) {
	mailCacheMu.Lock()
	defer mailCacheMu.Unlock()
	mailCacheValue = ms
	mailCacheFound = found
	mailCacheValid = true
}

// InvalidateMailCache 保存邮件设置后调用，避免读到旧配置
func InvalidateMailCache() {
	mailCacheMu.Lock()
	defer mailCacheMu.Unlock()
	mailCacheValid = false
	mailCacheValue = systemModel.MailSetting{}
}

// GetMail 读取邮件设置并解密密码，返回该项是否存在
func GetMail(dest *systemModel.MailSetting) (bool, error) {
	if cached, found, hit := loadMailCache(); hit {
		if !found {
			return false, nil
		}
		*dest = cached
		return true, nil
	}
	if !Get(systemModel.KeyMailSMTP, dest) {
		// 未配置也缓存，避免未启用邮件时每次发信都查库
		storeMailCache(systemModel.MailSetting{}, false)
		return false, nil
	}
	if dest.SMTPPassword == "" {
		storeMailCache(*dest, true)
		return true, nil
	}
	plain, err := crypto.DecryptAES(dest.SMTPPassword)
	if err != nil {
		// 解密失败不缓存，保留每次重试的机会
		return true, fmt.Errorf("邮件密码解密失败，请确认 octoops.security.aes_key 与加密时一致: %w", err)
	}
	dest.SMTPPassword = plain
	storeMailCache(*dest, true)
	return true, nil
}

// SetMail 写入邮件设置，密码在落库前加密
func SetMail(ms systemModel.MailSetting) error {
	encrypted, err := crypto.EncryptAES(ms.SMTPPassword)
	if err != nil {
		return err
	}
	ms.SMTPPassword = encrypted
	if err := Set(systemModel.KeyMailSMTP, ms); err != nil {
		return err
	}
	InvalidateMailCache()
	return nil
}
