package mail

import (
	"log"

	systemModel "octoops/internal/model/system"
	"octoops/internal/pkg/settingstore"

	"gopkg.in/gomail.v2"
)

type MailOptions struct {
	To      string   // 收件人
	Cc      []string // 抄送，可选
	Subject string   // 主题
	Body    string   // 正文（支持HTML）
}

// SendMail 发送邮件，支持抄送、SSL、配置化
func SendMail(opt MailOptions) error {
	cfg := CurrentMailConfig()
	if !cfg.Enable {
		return nil // 未开启邮件
	}
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
	d.SSL = cfg.SSL
	return d.DialAndSend(m)
}

// CurrentMailConfig 邮件配置，来源为系统设置
func CurrentMailConfig() systemModel.MailSetting {
	var ms systemModel.MailSetting
	found, err := settingstore.GetMail(&ms)
	if err != nil {
		log.Printf("[Mail] 读取邮件配置失败: %v", err)
		return systemModel.MailSetting{SMTPPort: systemModel.DefaultSMTPPort, SSL: true}
	}
	if !found {
		return systemModel.MailSetting{SMTPPort: systemModel.DefaultSMTPPort, SSL: true}
	}
	if ms.SMTPPort == 0 {
		ms.SMTPPort = systemModel.DefaultSMTPPort
	}
	return ms
}
