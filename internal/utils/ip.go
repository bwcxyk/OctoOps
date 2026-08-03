package utils

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

// GetCurrentPublicIP 获取本机出口公网IP（用于自动同步安全组白名单等场景）
// 使用多个实测可用的公共接口，按序回退，单个不可用自动切换下一个
func GetCurrentPublicIP() (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	type ipSource struct {
		url    string
		parser func(body string) string // 从响应体提取IP，返回空表示解析失败
	}
	sources := []ipSource{
		// 花生壳：返回 "Current IP Address: x.x.x.x"
		{"https://ddns.oray.com/checkip", func(b string) string {
			fields := strings.Fields(b)
			if len(fields) == 0 {
				return ""
			}
			return fields[len(fields)-1]
		}},
		// ipip.net：返回 "当前 IP：x.x.x.x  来自：..."
		{"https://myip.ipip.net", func(b string) string {
			idx := strings.Index(b, "：")
			if idx < 0 {
				return ""
			}
			rest := b[idx+len("："):]
			if sp := strings.IndexAny(rest, " \t"); sp >= 0 {
				rest = rest[:sp]
			}
			return strings.TrimSpace(rest)
		}},
		// ifconfig.me：返回纯文本 IP
		{"https://ifconfig.me/ip", func(b string) string { return strings.TrimSpace(b) }},
	}
	var lastErr error
	for _, src := range sources {
		resp, err := client.Get(src.url)
		if err != nil {
			log.Printf("[GetPublicIP] 请求 %s 失败: %v", src.url, err)
			lastErr = err
			continue
		}
		func() {
			defer func(Body io.ReadCloser) {
				if e := Body.Close(); e != nil {
					log.Printf("Close Body error: %v", e)
				}
			}(resp.Body)
		}()
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("状态码 %d", resp.StatusCode)
			log.Printf("[GetPublicIP] 请求 %s 返回非200: %d", src.url, resp.StatusCode)
			continue
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}
		ip := strings.TrimSpace(src.parser(string(body)))
		if net.ParseIP(ip) == nil {
			lastErr = fmt.Errorf("返回内容不是合法IP: %q", ip)
			log.Printf("[GetPublicIP] 请求 %s 返回内容非IP: %q", src.url, ip)
			continue
		}
		return ip, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("所有IP查询接口均不可用")
	}
	return "", fmt.Errorf("获取公网IP失败: %v", lastErr)
}
