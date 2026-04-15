package aliyun

import (
	"encoding/base64"
	"fmt"
	"octoops/internal/db"
	aliyunModel "octoops/internal/model/aliyun"
	"octoops/internal/utils"
	"strings"

	"gorm.io/gorm"
)

func ListEcsSecurityGroupConfigs(status, accessKey, name string) ([]aliyunModel.SGConfig, error) {
	var configs []aliyunModel.SGConfig
	query := db.DB.Order("created_at desc")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if accessKey != "" {
		query = query.Where("access_key = ?", accessKey)
	}
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	err := query.Find(&configs).Error
	return configs, err
}

func CreateEcsSecurityGroupConfig(cfg *aliyunModel.SGConfig) error {
	sk, err := utils.EncryptAES(cfg.AccessSecret)
	if err != nil {
		return fmt.Errorf("SK加密失败: %v", err)
	}
	cfg.AccessSecret = sk
	return db.DB.Create(cfg).Error
}

func GetEcsSecurityGroupConfigByID(id string) (aliyunModel.SGConfig, error) {
	var cfg aliyunModel.SGConfig
	err := db.DB.First(&cfg, id).Error
	return cfg, err
}

func UpdateEcsSecurityGroupConfig(id string, req map[string]interface{}) (aliyunModel.SGConfig, error) {
	cfg, err := GetEcsSecurityGroupConfigByID(id)
	if err != nil {
		return aliyunModel.SGConfig{}, err
	}
	if sk, ok := req["access_secret"].(string); ok && sk != "" {
		_, decodeErr := base64.StdEncoding.DecodeString(sk)
		if decodeErr != nil || len(sk) < 32 {
			encrypted, encErr := utils.EncryptAES(sk)
			if encErr != nil {
				return aliyunModel.SGConfig{}, fmt.Errorf("SK加密失败: %v", encErr)
			}
			req["access_secret"] = encrypted
		}
	}
	if err := db.DB.Model(&cfg).Updates(req).Error; err != nil {
		return aliyunModel.SGConfig{}, err
	}
	return cfg, nil
}

func DeleteEcsSecurityGroupConfig(id string) error {
	return db.DB.Delete(&aliyunModel.SGConfig{}, id).Error
}

func SyncEcsSecurityGroupConfigByID(id string) error {
	cfg, err := GetEcsSecurityGroupConfigByID(id)
	if err != nil {
		return err
	}
	dbIns := db.DB.Session(&gorm.Session{})
	dbIns = dbIns.Model(&aliyunModel.SGConfig{}).Where("id = ?", cfg.ID)
	err = UpdateSecurityGroupIfIPChanged(dbIns)
	if err != nil {
		if strings.Contains(err.Error(), "InvalidSecurityGroupId.NotFound") {
			return fmt.Errorf("找不到安全组，请检查安全组ID、Region和AK/SK配置是否正确。")
		}
		return err
	}
	return nil
}
