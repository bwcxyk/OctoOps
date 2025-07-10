package utils

import (
	"gopkg.in/gomail.v2"
	"octoops/internal/config"
)

type MailOptions struct {
	To      string   // 收件人
	Cc      []string // 抄送，可选
	Subject string   // 主题
	Body    string   // 正文（支持HTML）
}

// SendMail 发送邮件，支持抄送、SSL、配置化
func SendMail(opt MailOptions) error {
	cfg := config.GetMailConfig()
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

	d := gomail.NewDialer(cfg.SMTPAddress, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPassword)

	d.SSL = cfg.SSL
	return d.DialAndSend(m)
} 