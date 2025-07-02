package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"

	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	ecs "github.com/alibabacloud-go/ecs-20140526/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	"octoops/model"
	"octoops/db"
)

// 获取当前公网IP
func GetCurrentPublicIP() (string, error) {
	resp, err := http.Get("https://www.ipplus360.com/getIP")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var result struct {
		Data string `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Data, nil
}

// 获取数据库中最新的安全组配置
func GetAliyunSGConfig(db *gorm.DB) (*model.AliyunSGConfig, error) {
	var cfg model.AliyunSGConfig
	err := db.Order("updated_at desc").First(&cfg).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// 初始化ECS客户端（官方示例风格）
func Initialization(cfg *model.AliyunSGConfig) (*ecs.Client, error) {
	openCfg := &openapi.Config{}
	openCfg.AccessKeyId = tea.String(cfg.AccessKey)
	openCfg.AccessKeySecret = tea.String(cfg.AccessSecret)
	openCfg.RegionId = tea.String(cfg.RegionId)
	return ecs.NewClient(openCfg)
}

// 授权安全组（官方示例风格）
func AuthorizeSecurityGroup(client *ecs.Client, cfg *model.AliyunSGConfig, port int, sourceCidrIp string) error {
	req := &ecs.AuthorizeSecurityGroupRequest{}
	req.RegionId = tea.String(cfg.RegionId)
	req.SecurityGroupId = tea.String(cfg.SecurityGroupId)
	req.IpProtocol = tea.String("tcp")
	req.PortRange = tea.String(fmt.Sprintf("%d/%d", port, port))
	req.NicType = tea.String("internet")
	req.Policy = tea.String("accept")
	req.Priority = tea.String("1")
	req.SourceCidrIp = tea.String(fmt.Sprintf("%s/32", sourceCidrIp))
	_, err := client.AuthorizeSecurityGroup(req)
	return err
}

// 撤销安全组授权（官方示例风格）
func RevokeSecurityGroup(client *ecs.Client, cfg *model.AliyunSGConfig, port int, sourceCidrIp string) error {
	req := &ecs.RevokeSecurityGroupRequest{}
	req.RegionId = tea.String(cfg.RegionId)
	req.SecurityGroupId = tea.String(cfg.SecurityGroupId)
	req.IpProtocol = tea.String("tcp")
	req.PortRange = tea.String(fmt.Sprintf("%d/%d", port, port))
	req.SourceCidrIp = tea.String(fmt.Sprintf("%s/32", sourceCidrIp))
	_, err := client.RevokeSecurityGroup(req)
	return err
}

// 查询安全组详情（官方示例风格）
func DescribeSecurityGroupAttribute(client *ecs.Client, cfg *model.AliyunSGConfig) (*ecs.DescribeSecurityGroupAttributeResponse, error) {
	req := &ecs.DescribeSecurityGroupAttributeRequest{}
	req.RegionId = tea.String(cfg.RegionId)
	req.SecurityGroupId = tea.String(cfg.SecurityGroupId)
	req.Direction = tea.String("all")
	return client.DescribeSecurityGroupAttribute(req)
}

// 主流程：如IP变化则更新安全组规则（官方示例风格+数据库参数）
func UpdateSecurityGroupIfIPChanged(db *gorm.DB) error {
	cfg, err := GetAliyunSGConfig(db)
	if err != nil {
		return fmt.Errorf("获取安全组配置失败: %v", err)
	}
	portList := []int{}
	for _, p := range strings.Split(cfg.PortList, ",") {
		p = strings.TrimSpace(p)
		if p == "" { continue }
		var port int
		fmt.Sscanf(p, "%d", &port)
		if port > 0 { portList = append(portList, port) }
	}

	client, err := Initialization(cfg)
	if err != nil {
		return fmt.Errorf("ECS客户端初始化失败: %v", err)
	}

	oldIP := cfg.LastIP
	newIP, err := GetCurrentPublicIP()
	if err != nil {
		return fmt.Errorf("获取公网IP失败: %v", err)
	}
	if newIP == oldIP {
		return nil // 无需更新
	}

	// 授权新IP
	for _, port := range portList {
		if err := AuthorizeSecurityGroup(client, cfg, port, newIP); err != nil {
			return fmt.Errorf("端口%d授权失败: %v", port, err)
		}
	}
	// 撤销旧IP
	if oldIP != "" {
		for _, port := range portList {
			if err := RevokeSecurityGroup(client, cfg, port, oldIP); err != nil {
				if strings.Contains(err.Error(), "InvalidSecurityGroupRule.RuleNotExist") {
					// 规则不存在，跳过
					continue
				}
				return fmt.Errorf("端口%d撤销旧授权失败: %v", port, err)
			}
		}
	}
	// 更新last_ip和last_ip_updated_at
	db.Model(cfg).Updates(map[string]interface{}{
		"last_ip": newIP,
		"last_ip_updated_at": time.Now(),
	})
	return nil
}

// 查询安全组规则详情并返回字符串（可用于日志或前端展示）
func GetSecurityGroupDetailString(client *ecs.Client, cfg *model.AliyunSGConfig) (string, error) {
	resp, err := DescribeSecurityGroupAttribute(client, cfg)
	if err != nil {
		return "", err
	}
	detail := resp.Body
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("安全组 %s(%s) 规则如下：\n", tea.StringValue(detail.SecurityGroupName), tea.StringValue(detail.SecurityGroupId)))
	for _, permission := range detail.Permissions.Permission {
		sb.WriteString(fmt.Sprintf("  规则描述：%s; 方向：%s; 端口范围：%s; 源端IPv4 CIDR地址段：%s; 网卡类型：%s; 访问权限：%s; 规则优先级：%s; 创建时间：%s;\n",
			tea.StringValue(permission.Description),
			tea.StringValue(permission.Direction),
			tea.StringValue(permission.SourcePortRange),
			tea.StringValue(permission.SourceCidrIp),
			tea.StringValue(permission.NicType),
			tea.StringValue(permission.Policy),
			tea.StringValue(permission.Priority),
			tea.StringValue(permission.CreateTime),
		))
	}
	return sb.String(), nil
}

// 批量同步所有ECS安全组配置
func SyncAllECSSecurityGroups() error {
	var configs []model.AliyunSGConfig
	dbIns := db.DB
	dbIns.Find(&configs)
	for _, cfg := range configs {
		ins := dbIns.Session(&gorm.Session{})
		ins = ins.Model(&model.AliyunSGConfig{}).Where("id = ?", cfg.ID)
		err := UpdateSecurityGroupIfIPChanged(ins)
		if err != nil {
			return err
		}
	}
	return nil
}
