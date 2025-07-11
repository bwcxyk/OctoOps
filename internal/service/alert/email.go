package alert

import (
	"bytes"
	"github.com/russross/blackfriday/v2"
	"octoops/internal/model/alert"
	"octoops/internal/utils"
	"text/template"
)

// 邮件测试发送
func SendTestEmail(alert *alert.Channel) error {
	// 直接调用 utils.SendMail
	return utils.SendMail(utils.MailOptions{
		To:      alert.Target,
		Subject: "OctoOps 测试通知",
		Body:    "这是一条测试邮件通知。",
	})
}

// 邮件模板发送
func SendEmailWithTemplate(alert *alert.Channel, tplContent string, data map[string]interface{}) error {
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
	return utils.SendMail(utils.MailOptions{
		To:      alert.Target,
		Subject: "OctoOps 告警通知",
		Body:    htmlBody,
	})
}
