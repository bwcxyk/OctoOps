package alert

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"octoops/db"
	alertModel "octoops/model/alert"
	"text/template"
	"time"
)

// 钉钉加签
func dingtalkSign(secret string) (string, string) {
	timestamp := fmt.Sprintf("%d", time.Now().UnixNano()/1e6)
	stringToSign := timestamp + "\n" + secret
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(stringToSign))
	sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return timestamp, url.QueryEscape(sign)
}

// 钉钉机器人测试发送
func SendTestRobot(alert *alertModel.Alert) error {
	msg := map[string]interface{}{
		"msgtype": "text",
		"text":    map[string]string{"content": "这是一条测试机器人通知。"},
	}
	data, _ := json.Marshal(msg)
	webhook := alert.Target
	if alert.Type == "dingtalk" && alert.DingtalkSecret != "" {
		timestamp, sign := dingtalkSign(alert.DingtalkSecret)
		if u, err := url.Parse(webhook); err == nil {
			q := u.Query()
			q.Set("timestamp", timestamp)
			q.Set("sign", sign)
			u.RawQuery = q.Encode()
			webhook = u.String()
		}
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(webhook, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	fmt.Printf("钉钉响应状态码: %d, 内容: %s\n", resp.StatusCode, string(bodyBytes))

	if resp.StatusCode != 200 {
		return fmt.Errorf("发送失败，状态码: %d, 响应: %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}

// 作业失败钉钉告警
func SendDingTalkAlert(taskName string, taskID uint, failTime time.Time, reason string) error {
	content := fmt.Sprintf("作业失败告警：\n任务名称：%s\n任务ID：%d\n失败时间：%s\n失败原因：%s", taskName, taskID, failTime.Format("2006-01-02 15:04:05"), reason)
	var alerts []alertModel.Alert
	db.DB.Where("type = ? AND status = ?", "dingtalk", 1).Find(&alerts)
	msg := map[string]interface{}{
		"msgtype": "text",
		"text":    map[string]string{"content": content},
	}
	data, _ := json.Marshal(msg)
	for _, alert := range alerts {
		webhook := alert.Target
		if alert.DingtalkSecret != "" {
			timestamp, sign := dingtalkSign(alert.DingtalkSecret)
			if u, err := url.Parse(webhook); err == nil {
				q := u.Query()
				q.Set("timestamp", timestamp)
				q.Set("sign", sign)
				u.RawQuery = q.Encode()
				webhook = u.String()
			}
		}
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Post(webhook, "application/json", bytes.NewReader(data))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		fmt.Printf("钉钉告警响应状态码: %d, 内容: %s\n", resp.StatusCode, string(bodyBytes))
	}
	return nil
}

// 支持指定告警组ID列表
func SendDingTalkAlertToGroups(taskName string, taskID uint, failTime time.Time, reason string, groupIDs []string) error {
	content := fmt.Sprintf("作业失败告警：\n任务名称：%s\n任务ID：%d\n失败时间：%s\n失败原因：%s", taskName, taskID, failTime.Format("2006-01-02 15:04:05"), reason)
	var alerts []alertModel.Alert
	if len(groupIDs) > 0 {
		db.DB.Where("type = ? AND status = ? AND id IN ?", "dingtalk", 1, groupIDs).Find(&alerts)
	} else {
		db.DB.Where("type = ? AND status = ?", "dingtalk", 1).Find(&alerts)
	}
	msg := map[string]interface{}{
		"msgtype": "text",
		"text":    map[string]string{"content": content},
	}
	data, _ := json.Marshal(msg)
	for _, alert := range alerts {
		webhook := alert.Target
		if alert.DingtalkSecret != "" {
			timestamp, sign := dingtalkSign(alert.DingtalkSecret)
			if u, err := url.Parse(webhook); err == nil {
				q := u.Query()
				q.Set("timestamp", timestamp)
				q.Set("sign", sign)
				u.RawQuery = q.Encode()
				webhook = u.String()
			}
		}
		fmt.Printf("[DEBUG] 钉钉告警发送: webhook=%s\n[DEBUG] 钉钉告警内容: %s\n", webhook, content)
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Post(webhook, "application/json", bytes.NewReader(data))
		if err != nil {
			fmt.Printf("[ERROR] 钉钉告警发送失败: %v\n", err)
			return err
		}
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		fmt.Printf("[DEBUG] 钉钉告警响应状态码: %d, 内容: %s\n", resp.StatusCode, string(bodyBytes))
	}
	return nil
}

// 发送钉钉 markdown 消息，支持模板渲染
func SendDingTalkMarkdownWithTemplate(webhook, secret, title, tplContent string, data map[string]interface{}) error {
	// 渲染模板
	tpl, err := template.New("msg").Parse(tplContent)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return err
	}
	msg := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title": title,
			"text":  buf.String(),
		},
	}
	jsonData, _ := json.Marshal(msg)
	// 加签
	if secret != "" {
		timestamp, sign := dingtalkSign(secret)
		if u, err := url.Parse(webhook); err == nil {
			q := u.Query()
			q.Set("timestamp", timestamp)
			q.Set("sign", sign)
			u.RawQuery = q.Encode()
			webhook = u.String()
		}
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(webhook, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("发送失败，状态码: %d, 响应: %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}
