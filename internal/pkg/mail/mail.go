package mail

import (
	"errors"
	"log"
	"strings"

	systemModel "octoops/internal/model/system"
	syssetting "octoops/internal/pkg/setting"

	"gopkg.in/gomail.v2"
)

type MailOptions struct {
	To      string   // 收件人
	Cc      []string // 抄送，可选
	Subject string   // 主题
	Body    string   // 正文（支持HTML）
}

// SendMail 发送邮件，支持抄送、加密方式（无/SSL/STARTTLS）、配置化
func SendMail(opt MailOptions) error {
	cfg := CurrentMailConfig()
	if !cfg.Enable {
		return nil // 未开启邮件
	}
	return sendWithConfig(cfg, opt)
}

// SendTestMail 发送测试邮件验证邮件配置是否可用，不受邮件通知总开关限制
func SendTestMail(to string) error {
	cfg := CurrentMailConfig()
	if strings.TrimSpace(cfg.SMTPAddress) == "" || strings.TrimSpace(cfg.SMTPUser) == "" {
		return errors.New("请先填写 SMTP 服务器与邮箱账号并保存")
	}
	return sendWithConfig(cfg, MailOptions{
		To:      to,
		Subject: "OctoOps 邮件配置测试",
		Body:    "<p>这是一封测试邮件，收到即说明当前邮件配置可用。</p>",
	})
}

func sendWithConfig(cfg systemModel.MailSetting, opt MailOptions) error {
	m := gomail.NewMessage()
	if cfg.DisplayName != "" {
		m.SetHeader("From", m.FormatAddress(cfg.SMTPUser, cfg.DisplayName))
	} else {
		m.SetHeader("From", cfg.SMTPUser)
	}
	m.SetHeader("To", opt.To)
	if len(opt.Cc) > 0 {
		m.SetHeader("Cc", opt.Cc...)
	}
	m.SetHeader("Subject", opt.Subject)
	m.SetBody("text/html", opt.Body)

	port := cfg.SMTPPort
	if port == 0 {
		port = systemModel.DefaultSMTPPort
	}
	d := gomail.NewDialer(cfg.SMTPAddress, port, cfg.SMTPUser, cfg.SMTPPassword)
	// ssl：直接建立 TLS 连接（465）；tls：明文连接后用 STARTTLS 升级（587）；none：不主动加密
	d.SSL = cfg.Security == systemModel.MailSecuritySSL
	return d.DialAndSend(m)
}

// CurrentMailConfig 邮件配置，来源为系统设置
func CurrentMailConfig() systemModel.MailSetting {
	var ms systemModel.MailSetting
	found, err := syssetting.GetMail(&ms)
	if err != nil {
		log.Printf("[Mail] 读取邮件配置失败: %v", err)
		return defaultMailSetting()
	}
	if !found {
		return defaultMailSetting()
	}
	if ms.SMTPPort == 0 {
		ms.SMTPPort = systemModel.DefaultSMTPPort
	}
	if !systemModel.ValidSecurity(ms.Security) {
		ms.Security = systemModel.MailSecuritySSL
	}
	return ms
}

// defaultMailSetting 未配置或读取失败时的兜底配置
func defaultMailSetting() systemModel.MailSetting {
	return systemModel.MailSetting{SMTPPort: systemModel.DefaultSMTPPort, Security: systemModel.MailSecuritySSL}
}
