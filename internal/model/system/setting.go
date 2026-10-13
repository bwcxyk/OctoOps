package system

import "time"

// Setting 系统设置项，value 为 JSON 字符串，具体结构由各配置项自行定义
type Setting struct {
	Key       string    `gorm:"size:128;primaryKey" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Setting) TableName() string { return "settings" }

// 设置项 key
const (
	KeyTaskLogRetentionDays = "log.retention_days"
	KeyMailSMTP             = "mail.smtp"
)

// MaskedSecret 敏感配置在返回给前端时的占位值
const MaskedSecret = "******"

// 任务日志保留天数默认值与"不清理"标记
const (
	DefaultTaskLogRetentionDays = 30
	TaskLogRetentionDisabled    = -1
)

// TaskSetting 任务类设置项的值结构（key 形如 task.job_status_sync）
type TaskSetting struct {
	Enabled bool   `json:"enabled"`
	Cron    string `json:"cron"`
}

// RetentionSetting 日志保留设置项的值结构
type RetentionSetting struct {
	Days int `json:"retention_days"`
}

// 邮件连接的加密方式
const (
	MailSecurityNone = "none" // 不加密
	MailSecuritySSL  = "ssl"  // 隐式 SSL/TLS，常用端口 465
	MailSecurityTLS  = "tls"  // STARTTLS，常用端口 587
)

// MailSetting 邮件服务设置项的值结构
type MailSetting struct {
	Enable       bool   `json:"enable"`
	SMTPAddress  string `json:"smtp_address"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUser     string `json:"smtp_user"`
	SMTPPassword string `json:"smtp_password"`
	DisplayName  string `json:"display_name"`
	Security     string `json:"security"` // none / ssl / tls
}

// ValidSecurity 判断加密方式取值是否合法
func ValidSecurity(security string) bool {
	switch security {
	case MailSecurityNone, MailSecuritySSL, MailSecurityTLS:
		return true
	}
	return false
}

// DefaultSMTPPort 未指定端口时的默认值
const DefaultSMTPPort = 465
