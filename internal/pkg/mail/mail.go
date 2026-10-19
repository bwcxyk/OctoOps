package mail

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	netmail "net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"

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
	cfg, err := CurrentMailConfig()
	if err != nil {
		return err
	}
	if !cfg.Enable {
		return nil // 未开启邮件
	}
	return sendWithConfig(cfg, opt)
}

// SendTestMail 发送测试邮件验证邮件配置是否可用，不受邮件通知总开关限制
func SendTestMail(to string) error {
	cfg, err := CurrentMailConfig()
	if err != nil {
		return err
	}
	if strings.TrimSpace(cfg.SMTPAddress) == "" || strings.TrimSpace(cfg.SMTPUser) == "" {
		return errors.New("请先填写 SMTP 服务器与邮箱账号并保存")
	}
	return sendWithConfig(cfg, MailOptions{
		To:      to,
		Subject: "OctoOps 邮件配置测试",
		Body:    "<p>这是一封测试邮件，收到即说明当前邮件配置可用。</p>",
	})
}

// CurrentMailConfig 邮件配置，来源为系统设置；读取/解密失败会返回 error，避免静默按“未启用”处理
func CurrentMailConfig() (systemModel.MailSetting, error) {
	var ms systemModel.MailSetting
	found, err := syssetting.GetMail(&ms)
	if err != nil {
		log.Printf("[Mail] 读取邮件配置失败: %v", err)
		return systemModel.MailSetting{}, err
	}
	if !found {
		return defaultMailSetting(), nil
	}
	if !systemModel.ValidSecurity(ms.Security) {
		ms.Security = systemModel.MailSecuritySSL
	}
	return ms, nil
}

// defaultMailSetting 未配置时的默认值
func defaultMailSetting() systemModel.MailSetting {
	return systemModel.MailSetting{}
}

const (
	smtpDialTimeout = 10 * time.Second
	smtpDeadline    = 30 * time.Second
)

// sendWithConfig 按 security 选择连接方式：
//   - ssl：直接建立 TLS 连接
//   - tls：明文连接后强制 STARTTLS，服务器不支持则拒绝发送
//   - none：不主动升级 TLS；服务器要求认证时凭据以明文传输，仅适用于内网可信 SMTP
func sendWithConfig(cfg systemModel.MailSetting, opt MailOptions) error {
	if cfg.SMTPPort < 1 || cfg.SMTPPort > 65535 {
		return fmt.Errorf("SMTP 端口不合法: %d", cfg.SMTPPort)
	}
	if strings.TrimSpace(cfg.SMTPAddress) == "" {
		return errors.New("SMTP 服务器地址为空")
	}
	if strings.TrimSpace(cfg.SMTPUser) == "" {
		return errors.New("SMTP 邮箱账号为空")
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

	from, err := netmail.ParseAddress(cfg.SMTPUser)
	if err != nil {
		return fmt.Errorf("发件人地址不合法: %w", err)
	}
	recipients, err := envelopeRecipients(opt)
	if err != nil {
		return err
	}
	return sendSMTP(cfg, from.Address, recipients, m)
}

// envelopeRecipients 从 To/Cc 头解析出 SMTP 信封收件人
func envelopeRecipients(opt MailOptions) ([]string, error) {
	headers := make([]string, 0, 1+len(opt.Cc))
	headers = append(headers, opt.To)
	headers = append(headers, opt.Cc...)

	recipients := make([]string, 0, len(headers))
	for _, header := range headers {
		if strings.TrimSpace(header) == "" {
			continue
		}
		addrs, err := netmail.ParseAddressList(header)
		if err != nil {
			return nil, fmt.Errorf("收件人地址不合法: %w", err)
		}
		for _, addr := range addrs {
			recipients = append(recipients, addr.Address)
		}
	}
	if len(recipients) == 0 {
		return nil, errors.New("收件人为空")
	}
	return recipients, nil
}

func sendSMTP(cfg systemModel.MailSetting, from string, to []string, msg *gomail.Message) error {
	host := strings.TrimSpace(cfg.SMTPAddress)
	addr := net.JoinHostPort(host, strconv.Itoa(cfg.SMTPPort))
	tlsConfig := &tls.Config{ServerName: host}
	dialer := &net.Dialer{Timeout: smtpDialTimeout}

	var conn net.Conn
	var err error
	switch cfg.Security {
	case systemModel.MailSecuritySSL:
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
	case systemModel.MailSecurityTLS, systemModel.MailSecurityNone:
		conn, err = dialer.Dial("tcp", addr)
	default:
		return fmt.Errorf("不支持的邮件加密方式: %s", cfg.Security)
	}
	if err != nil {
		return fmt.Errorf("连接 SMTP 服务器失败: %w", err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(smtpDeadline))

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("初始化 SMTP 连接失败: %w", err)
	}

	if cfg.Security == systemModel.MailSecurityTLS {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return errors.New("SMTP 服务器未提供 STARTTLS，已拒绝降级为明文连接")
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("STARTTLS 失败: %w", err)
		}
	}
	// none 明确不调用 StartTLS；ssl 在建立连接时已经是 TLS。

	if err := authenticate(client, cfg, host); err != nil {
		return err
	}
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("设置发件人失败: %w", err)
	}
	for _, rcpt := range to {
		if err := client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("设置收件人 %s 失败: %w", rcpt, err)
		}
	}
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("开始发送邮件失败: %w", err)
	}
	if _, err := msg.WriteTo(writer); err != nil {
		_ = writer.Close()
		return fmt.Errorf("写入邮件内容失败: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("提交邮件内容失败: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("结束 SMTP 会话失败: %w", err)
	}
	return nil
}

// authenticate 按服务器声明的机制选择认证方式；host 必须与建立连接时的 serverName 一致，
// 否则标准库的 PlainAuth 与 loginAuth 会因名称不匹配拒绝发送
func authenticate(client *smtp.Client, cfg systemModel.MailSetting, host string) error {
	if strings.TrimSpace(cfg.SMTPUser) == "" {
		return nil
	}
	ok, mechanisms := client.Extension("AUTH")
	if !ok {
		return nil // 服务器未声明 AUTH，按匿名方式投递
	}
	var auth smtp.Auth
	switch {
	case strings.Contains(mechanisms, "CRAM-MD5"):
		auth = smtp.CRAMMD5Auth(cfg.SMTPUser, cfg.SMTPPassword)
	case strings.Contains(mechanisms, "LOGIN") && !strings.Contains(mechanisms, "PLAIN"):
		auth = &loginAuth{
			username: cfg.SMTPUser,
			password: cfg.SMTPPassword,
			host:     host,
		}
	default:
		auth = smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPassword, host)
	}
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP 认证失败: %w", err)
	}
	return nil
}

// loginAuth 兼容只声明 LOGIN 的服务器；与标准库的 PlainAuth 一样要求 TLS
type loginAuth struct {
	username string
	password string
	host     string
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	if !server.TLS {
		return "", nil, errors.New("SMTP LOGIN 认证要求 TLS 连接")
	}
	if !strings.EqualFold(server.Name, a.host) {
		return "", nil, fmt.Errorf("SMTP 服务器名称不匹配: %s", server.Name)
	}
	return "LOGIN", nil, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if !more {
		return nil, nil
	}
	switch strings.ToLower(strings.TrimSpace(string(fromServer))) {
	case "username:":
		return []byte(a.username), nil
	case "password:":
		return []byte(a.password), nil
	default:
		return nil, fmt.Errorf("未知的 SMTP LOGIN 挑战: %s", fromServer)
	}
}
