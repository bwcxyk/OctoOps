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
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("JSON序列化失败: %v", err)
	}
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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("关闭响应体失败: %v\n", err)
		}
	}(resp.Body)

	bodyBytes, _ := io.ReadAll(resp.Body)
	fmt.Printf("钉钉响应状态码: %d, 内容: %s\n", resp.StatusCode, string(bodyBytes))

	if resp.StatusCode != 200 {
		return fmt.Errorf("发送失败，状态码: %d, 响应: %s", resp.StatusCode, string(bodyBytes))
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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("关闭响应体失败: %v\n", err)
		}
	}(resp.Body)
	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("发送失败，状态码: %d, 响应: %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}
