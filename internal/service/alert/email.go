package alert

import (
	"bytes"
	"strings"
	"text/template"

	"octoops/internal/model/alert"
	"octoops/internal/utils"

	"github.com/yuin/goldmark"
)

// 邮件测试发送
func SendTestEmail(alert *alert.AlertChannel) error {
	// 直接调用 utils.SendMail
	return utils.SendMail(utils.MailOptions{
		To:      alert.Target,
		Subject: "OctoOps 测试通知",
		Body:    "这是一条测试邮件通知。",
	})
}

// 邮件模板发送
func SendEmailWithTemplate(alert *alert.AlertChannel, tplContent string, data map[string]interface{}) error {
	content := tplContent
	if data != nil {
		// 先渲染 Go 模板变量
		tpl, err := template.New("mail").Parse(tplContent)
		if err != nil {
			return err
		}
		var buf bytes.Buffer
		if err := tpl.Execute(&buf, data); err != nil {
			return err
		}
		content = buf.String()
	}

	// 统一模板模式：邮件渠道始终按 Markdown 转 HTML 发送
	body, err := markdownToHTML(content)
	if err != nil {
		return err
	}

	return utils.SendMail(utils.MailOptions{
		To:      alert.Target,
		Subject: "OctoOps 告警通知",
		Body:    body,
	})
}

func markdownToHTML(markdown string) (string, error) {
	var htmlBuf bytes.Buffer
	md := goldmark.New()
	if err := md.Convert([]byte(strings.ReplaceAll(markdown, "\r\n", "\n")), &htmlBuf); err != nil {
		return "", err
	}
	return htmlBuf.String(), nil
}
