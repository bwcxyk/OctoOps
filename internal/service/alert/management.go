package alert

import (
	"errors"
	"fmt"
	"octoops/internal/db"
	alertModel "octoops/internal/model/alert"

	"gorm.io/gorm"
)

var ErrAlertGroupMemberExists = errors.New("告警组成员已存在")

func ListChannels() ([]alertModel.AlertChannel, error) {
	var channels []alertModel.AlertChannel
	err := db.DB.Order("created_at desc").Find(&channels).Error
	return channels, err
}

func CreateChannel(channel *alertModel.AlertChannel) error {
	return db.DB.Create(channel).Error
}

func GetChannelByID(id string) (alertModel.AlertChannel, error) {
	var channel alertModel.AlertChannel
	err := db.DB.First(&channel, id).Error
	return channel, err
}

func UpdateChannel(id string, updates map[string]interface{}) (alertModel.AlertChannel, error) {
	channel, err := GetChannelByID(id)
	if err != nil {
		return alertModel.AlertChannel{}, err
	}
	if err := db.DB.Model(&channel).Updates(updates).Error; err != nil {
		return alertModel.AlertChannel{}, err
	}
	return channel, nil
}

func DeleteChannel(id string) error {
	return db.DB.Delete(&alertModel.AlertChannel{}, id).Error
}

func ListAlertGroups() ([]alertModel.AlertGroup, error) {
	var groups []alertModel.AlertGroup
	err := db.DB.Order("created_at desc").Find(&groups).Error
	return groups, err
}

func CreateAlertGroup(group *alertModel.AlertGroup) error {
	return db.DB.Create(group).Error
}

func GetAlertGroupByID(id string) (alertModel.AlertGroup, error) {
	var group alertModel.AlertGroup
	err := db.DB.First(&group, id).Error
	return group, err
}

func UpdateAlertGroup(id string, req alertModel.AlertGroup) (alertModel.AlertGroup, error) {
	group, err := GetAlertGroupByID(id)
	if err != nil {
		return alertModel.AlertGroup{}, err
	}
	updates := map[string]interface{}{
		"name":        req.Name,
		"description": req.Description,
		"status":      req.Status,
	}
	if err := db.DB.Model(&group).Updates(updates).Error; err != nil {
		return alertModel.AlertGroup{}, err
	}
	return group, nil
}

func DeleteAlertGroup(id string) error {
	return db.DB.Delete(&alertModel.AlertGroup{}, id).Error
}

func ListAlertGroupMembers(groupID string) ([]alertModel.AlertGroupMember, error) {
	var members []alertModel.AlertGroupMember
	err := db.DB.Where("group_id = ?", groupID).Find(&members).Error
	return members, err
}

func CreateAlertGroupMember(groupID uint, member *alertModel.AlertGroupMember) error {
	member.GroupID = groupID
	var existing alertModel.AlertGroupMember
	if err := db.DB.Where(
		"group_id = ? AND channel_type = ? AND channel_id = ?",
		member.GroupID,
		member.ChannelType,
		member.ChannelID,
	).First(&existing).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	} else {
		return ErrAlertGroupMemberExists
	}
	if err := db.DB.Create(member).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrAlertGroupMemberExists
		}
		return err
	}
	return nil
}

func DeleteAlertGroupMember(memberID string) error {
	return db.DB.Delete(&alertModel.AlertGroupMember{}, memberID).Error
}

func ListAlertTemplates() ([]alertModel.AlertTemplate, error) {
	var templates []alertModel.AlertTemplate
	err := db.DB.Order("created_at desc").Find(&templates).Error
	return templates, err
}

func CreateAlertTemplate(tpl *alertModel.AlertTemplate) error {
	return db.DB.Create(tpl).Error
}

func GetAlertTemplateByID(id string) (alertModel.AlertTemplate, error) {
	var tpl alertModel.AlertTemplate
	err := db.DB.First(&tpl, id).Error
	return tpl, err
}

func UpdateAlertTemplate(id string, req alertModel.AlertTemplate) (alertModel.AlertTemplate, error) {
	tpl, err := GetAlertTemplateByID(id)
	if err != nil {
		return alertModel.AlertTemplate{}, err
	}
	updates := map[string]interface{}{
		"name":    req.Name,
		"content": req.Content,
	}
	if err := db.DB.Model(&tpl).Updates(updates).Error; err != nil {
		return alertModel.AlertTemplate{}, err
	}
	return tpl, nil
}

func DeleteAlertTemplate(id string) error {
	return db.DB.Delete(&alertModel.AlertTemplate{}, id).Error
}

func ParseUint(s string) (uint, error) {
	var i uint
	_, err := fmt.Sscanf(s, "%d", &i)
	if err != nil {
		return 0, fmt.Errorf("无效的数字格式: %s", s)
	}
	return i, nil
}
