package aliyun

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"

	"log"
	"octoops/internal/db"
	aliyunModel "octoops/internal/model/aliyun"
	"octoops/internal/utils"

	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	ecs "github.com/alibabacloud-go/ecs-20140526/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/aliyun/credentials-go/credentials"
)

// 获取当前公网IP
func GetCurrentPublicIP() (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://www.ipplus360.com/getIP")
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Close Body error: %v", err)
		}
	}(resp.Body)
	var result struct {
		Data string `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Data, nil
}

// 获取数据库中最新的安全组配置
func GetAliyunSGConfig(db *gorm.DB) (*aliyunModel.SGConfig, error) {
	var cfg aliyunModel.SGConfig
	err := db.Order("updated_at desc").First(&cfg).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// 初始化ECS客户端（无AK方式，推荐）
func Initialization(cfg *aliyunModel.SGConfig) (*ecs.Client, error) {
	ak := cfg.AccessKey
	sk, err := utils.DecryptAES(cfg.AccessSecret)
	if err != nil {
		if strings.Contains(err.Error(), "invalid padding") {
			return nil, fmt.Errorf("ECS客户端初始化失败: 请检查AES密钥与加密时使用是否一致")
		}
		return nil, fmt.Errorf("ECS客户端初始化失败: %v", err)
	}
	config := new(credential.Config).
		SetType("access_key").
		SetAccessKeyId(ak).
		SetAccessKeySecret(sk)
	cred, err := credential.NewCredential(config)
	if err != nil {
		return nil, err
	}
	openCfg := &openapi.Config{
		Credential: cred,
		RegionId:   tea.String(cfg.RegionId),
	}
	return ecs.NewClient(openCfg)
}

// 授权安全组（官方示例风格）
func AuthorizeSecurityGroup(client *ecs.Client, cfg *aliyunModel.SGConfig, port int, sourceCidrIp string) error {
	log.Printf("[Authorize] IP: %s, Port: %d, SecurityGroup: %s", sourceCidrIp, port, cfg.SecurityGroupId)
	req := &ecs.AuthorizeSecurityGroupRequest{}
	req.RegionId = tea.String(cfg.RegionId)
	req.SecurityGroupId = tea.String(cfg.SecurityGroupId)
	req.IpProtocol = tea.String("tcp")
	req.PortRange = tea.String(fmt.Sprintf("%d/%d", port, port))
	req.NicType = tea.String("internet")
	req.Policy = tea.String("accept")
	req.Priority = tea.String("1")
	// 统一加/32
	if !strings.Contains(sourceCidrIp, "/") {
		sourceCidrIp = sourceCidrIp + "/32"
	}
	req.SourceCidrIp = tea.String(sourceCidrIp)
	_, err := client.AuthorizeSecurityGroup(req)
	return err
}

// 撤销安全组授权，支持端口段
func RevokeSecurityGroup(client *ecs.Client, cfg *aliyunModel.SGConfig, portRange string, sourceCidrIp string) error {
	if !strings.Contains(sourceCidrIp, "/") {
		sourceCidrIp = sourceCidrIp + "/32"
	}
	log.Printf("[Revoke] IP: %s, PortRange: %s, SecurityGroup: %s", sourceCidrIp, portRange, cfg.SecurityGroupId)
	req := &ecs.RevokeSecurityGroupRequest{}
	req.RegionId = tea.String(cfg.RegionId)
	req.SecurityGroupId = tea.String(cfg.SecurityGroupId)
	req.IpProtocol = tea.String("tcp")
	req.PortRange = tea.String(portRange)
	req.SourceCidrIp = tea.String(sourceCidrIp)
	_, err := client.RevokeSecurityGroup(req)
	return err
}

// 查询安全组详情（官方示例风格）
func DescribeSecurityGroupAttribute(client *ecs.Client, cfg *aliyunModel.SGConfig) (*ecs.DescribeSecurityGroupAttributeResponse, error) {
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
	var portList []int
	for _, p := range strings.Split(cfg.PortList, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		var port int
		_, scanErr := fmt.Sscanf(p, "%d", &port)
		if scanErr != nil {
			log.Printf("解析端口失败: %s, 错误: %v", p, scanErr)
			continue
		}
		if port > 0 {
			portList = append(portList, port)
		}
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

	// 1. 撤销oldIP下所有tcp规则
	if oldIP != "" {
		resp, err := DescribeSecurityGroupAttribute(client, cfg)
		if err != nil {
			if strings.Contains(err.Error(), "StatusCode: 403") || strings.Contains(err.Error(), "Forbidden.RAM") {
				return fmt.Errorf("没有权限查询安全组规则，请联系管理员为该账号授权安全组相关权限（ecs:DescribeSecurityGroupAttribute 等）")
			}
			return fmt.Errorf("查询安全组规则失败: %v", err)
		}
		for _, perm := range resp.Body.Permissions.Permission {
			if tea.StringValue(perm.SourceCidrIp) == fmt.Sprintf("%s/32", oldIP) && strings.ToLower(tea.StringValue(perm.IpProtocol)) == "tcp" {
				portRange := tea.StringValue(perm.PortRange)
				if err := RevokeSecurityGroup(client, cfg, portRange, oldIP); err != nil {
					if strings.Contains(err.Error(), "InvalidSecurityGroupRule.RuleNotExist") {
						continue
					}
					return fmt.Errorf("端口范围%s撤销旧授权失败: %v", portRange, err)
				}
			}
		}
	}

	// 2. 授权当前IP所有端口
	for _, port := range portList {
		if err := AuthorizeSecurityGroup(client, cfg, port, newIP); err != nil {
			return fmt.Errorf("端口%d授权失败: %v", port, err)
		}
	}

	// 3. 更新last_ip和last_ip_updated_at
	db.Model(cfg).Select("last_ip", "last_ip_updated_at").Updates(map[string]interface{}{
		"last_ip":            newIP,
		"last_ip_updated_at": time.Now(),
	})
	return nil
}

/**
// 查询安全组规则详情并返回字符串（可用于日志或前端展示）
func GetSecurityGroupDetailString(client *ecs.Client, cfg *aliyunModel.SGConfig) (string, error) {
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
**/

// 批量同步所有ECS安全组配置
func SyncAllECSSecurityGroups() error {
	var configs []aliyunModel.SGConfig
	dbIns := db.DB
	dbIns.Where("status != 0").Find(&configs)
	var failed []string
	for _, cfg := range configs {
		ins := dbIns.Session(&gorm.Session{}).Model(&aliyunModel.SGConfig{}).Where("id = ?", cfg.ID)
		err := UpdateSecurityGroupIfIPChanged(ins)
		if err != nil {
			log.Printf("[ECS SG Sync] 配置ID=%d 同步失败: %v", cfg.ID, err)
			failed = append(failed, fmt.Sprintf("ID=%d: %v", cfg.ID, err))
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("部分安全组同步失败: %v", failed)
	}
	return nil
}

// 封装统一同步函数
func SyncECSSecurityGroups() string {
	log.Printf("[Scheduler] 开始同步ECS安全组")
	err := SyncAllECSSecurityGroups()
	if err != nil {
		log.Printf("[Scheduler] ECS安全组同步失败: %v", err)
		return "ECS安全组同步失败: " + err.Error()
	}
	log.Printf("[Scheduler] ECS安全组同步完成")
	return "ECS安全组同步完成"
}
