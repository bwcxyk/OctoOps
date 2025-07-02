package service

import (
	"errors"
	"gopkg.in/gomail.v2"
	"log"
	"octoops/config"
	"octoops/model"
	"strconv"
	"bytes"
	"text/template"
	"github.com/russross/blackfriday/v2"
)

var (
	emailFrom     string
	emailPassword string
	smtpHost      string
	smtpPort      string
	displayName   string
	enableMail    bool
	enableSSL     bool
)

func InitEmailConfigFromStruct(cfg config.MailConfig) {
	emailFrom = cfg.SMTPUser
	emailPassword = cfg.SMTPPassword
	smtpHost = cfg.SMTPAddress
	smtpPort = strconv.Itoa(cfg.SMTPPort)
	displayName = cfg.DisplayName
	enableMail = cfg.Enable
	enableSSL = cfg.SSL
}

// 邮件测试发送
func SendTestEmail(alert *model.Alert) error {
	if !enableMail {
		return errors.New("未开启邮件通知")
	}
	m := gomail.NewMessage()
	if displayName != "" {
		m.SetHeader("From", m.FormatAddress(emailFrom, displayName))
	} else {
		m.SetHeader("From", emailFrom)
	}
	m.SetHeader("To", alert.Target)
	m.SetHeader("Subject", "OctoOps 测试通知")
	m.SetBody("text/plain", "这是一条测试邮件通知。")

	portInt, err := strconv.Atoi(smtpPort)
	if err != nil {
		log.Printf("邮件端口转换失败: %v", err)
		return err
	}
	d := gomail.NewDialer(smtpHost, portInt, emailFrom, emailPassword)
	d.SSL = enableSSL
	err = d.DialAndSend(m)
	if err != nil {
		log.Printf("邮件发送失败: %v", err)
	}
	return err
}

// 邮件模板发送
func SendEmailWithTemplate(alert *model.Alert, tplContent string, data map[string]interface{}) error {
	if !enableMail {
		return errors.New("未开启邮件通知")
	}
	// 渲染模板
	tpl, err := template.New("mail").Parse(tplContent)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return err
	}
	// markdown 转 html
	htmlBody := string(blackfriday.Run(buf.Bytes()))
	m := gomail.NewMessage()
	if displayName != "" {
		m.SetHeader("From", m.FormatAddress(emailFrom, displayName))
	} else {
		m.SetHeader("From", emailFrom)
	}
	m.SetHeader("To", alert.Target)
	m.SetHeader("Subject", "OctoOps 告警通知")
	m.SetBody("text/html", htmlBody)

	portInt, err := strconv.Atoi(smtpPort)
	if err != nil {
		log.Printf("邮件端口转换失败: %v", err)
		return err
	}
	d := gomail.NewDialer(smtpHost, portInt, emailFrom, emailPassword)
	d.SSL = enableSSL
	err = d.DialAndSend(m)
	if err != nil {
		log.Printf("邮件发送失败: %v", err)
	}
	return err
}
