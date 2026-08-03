package utils

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

var publicIPClient = &http.Client{
	Timeout: 10 * time.Second,
}

type ipSource struct {
	url    string
	parser func(body string) string
}

// GetCurrentPublicIP 获取本机出口公网IPv4
// 用于自动同步安全组白名单等场景
// 使用多个公共接口，按序回退，单个不可用自动切换下一个
func GetCurrentPublicIP() (string, error) {

	sources := []ipSource{
		{
			// 花生壳：Current IP Address: x.x.x.x
			url: "https://ddns.oray.com/checkip",
			parser: func(b string) string {
				fields := strings.Fields(b)

				for _, field := range fields {
					if net.ParseIP(field) != nil {
						return field
					}
				}

				return ""
			},
		},
		{
			// ipip.net：当前 IP：x.x.x.x
			url: "https://myip.ipip.net",
			parser: func(b string) string {
				idx := strings.IndexAny(b, ":：")
				if idx < 0 {
					return ""
				}

				rest := strings.TrimSpace(b[idx+1:])
				fields := strings.Fields(rest)

				if len(fields) == 0 {
					return ""
				}

				return fields[0]
			},
		},
		{
			// ifconfig.me：纯文本IP
			url: "https://ifconfig.me/ip",
			parser: func(b string) string {
				return strings.TrimSpace(b)
			},
		},
	}

	var errs []string

	for _, src := range sources {

		req, err := http.NewRequest(
			http.MethodGet,
			src.url,
			nil,
		)
		if err != nil {
			errs = append(
				errs,
				fmt.Sprintf("%s: %v", src.url, err),
			)
			continue
		}

		req.Header.Set(
			"User-Agent",
			"Mozilla/5.0",
		)

		resp, err := publicIPClient.Do(req)
		if err != nil {
			errs = append(
				errs,
				fmt.Sprintf("%s: %v", src.url, err),
			)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()

			errs = append(
				errs,
				fmt.Sprintf(
					"%s: status=%d",
					src.url,
					resp.StatusCode,
				),
			)
			continue
		}

		body, err := io.ReadAll(
			io.LimitReader(resp.Body, 1024),
		)

		resp.Body.Close()

		if err != nil {
			errs = append(
				errs,
				fmt.Sprintf("%s: %v", src.url, err),
			)
			continue
		}

		ip := strings.TrimSpace(
			src.parser(string(body)),
		)

		parsedIP := net.ParseIP(ip)

		// 安全组白名单场景，只接受IPv4
		if parsedIP == nil || parsedIP.To4() == nil {
			errs = append(
				errs,
				fmt.Sprintf(
					"%s: invalid ipv4=%q",
					src.url,
					ip,
				),
			)
			continue
		}

		return ip, nil
	}

	if len(errs) == 0 {
		return "", fmt.Errorf("所有公网IP查询接口均不可用")
	}

	return "", fmt.Errorf(
		"获取公网IP失败: %s",
		strings.Join(errs, "; "),
	)
}