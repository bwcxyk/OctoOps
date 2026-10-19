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

// Get 读取设置项并解析到 dest。
// found 表示记录是否存在；数据库错误与 JSON 解析失败都通过 error 返回，避免被误当成“未配置”。
func Get(key string, dest any) (found bool, err error) {
	var row systemModel.Setting
	if err := postgres.DB.Where("\"key\" = ?", key).Limit(1).Find(&row).Error; err != nil {
		return false, fmt.Errorf("读取设置项 %s 失败: %w", key, err)
	}
	if row.Key == "" {
		return false, nil
	}
	if err := json.Unmarshal([]byte(row.Value), dest); err != nil {
		return true, fmt.Errorf("解析设置项 %s 失败: %w", key, err)
	}
	return true, nil
}

// Set 写入设置项，已存在则更新 value
func Set(key string, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	var row systemModel.Setting
	if err := postgres.DB.Where("\"key\" = ?", key).Limit(1).Find(&row).Error; err != nil {
		// 查询失败时不走 Create，否则会以唯一键冲突掩盖真实的数据库错误
		return fmt.Errorf("读取设置项 %s 失败: %w", key, err)
	}
	if row.Key != "" {
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
	mailCacheValue systemModel.MailSetting
)

// loadMailCache 命中缓存时返回配置，未缓存返回 hit=false
func loadMailCache() (systemModel.MailSetting, bool) {
	mailCacheMu.RLock()
	defer mailCacheMu.RUnlock()
	if !mailCacheValid {
		return systemModel.MailSetting{}, false
	}
	return mailCacheValue, true
}

func storeMailCache(ms systemModel.MailSetting) {
	mailCacheMu.Lock()
	defer mailCacheMu.Unlock()
	mailCacheValue = ms
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
	if cached, hit := loadMailCache(); hit {
		*dest = cached
		return true, nil
	}
	found, err := Get(systemModel.KeyMailSMTP, dest)
	if err != nil {
		return false, fmt.Errorf("读取邮件配置失败: %w", err)
	}
	if !found {
		// 不缓存“未配置”，避免多实例/后续新增配置后本进程一直读不到；读取错误也不缓存
		return false, nil
	}
	if dest.SMTPPassword == "" {
		storeMailCache(*dest)
		return true, nil
	}
	plain, err := crypto.DecryptAES(dest.SMTPPassword)
	if err != nil {
		// 解密失败不缓存，保留每次重试的机会
		return true, fmt.Errorf("邮件密码解密失败，请确认 octoops.security.aes_key 与加密时一致: %w", err)
	}
	dest.SMTPPassword = plain
	storeMailCache(*dest)
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
