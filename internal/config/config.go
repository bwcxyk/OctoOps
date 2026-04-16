package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type PostgresConfig struct {
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	Port     int    `yaml:"port"`
	SSLMode  string `yaml:"sslmode"`
	TimeZone string `yaml:"timezone"`
}

func (p PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		p.Host, p.User, p.Password, p.DBName, p.Port, p.SSLMode, p.TimeZone,
	)
}

type MailConfig struct {
	SMTPAddress  string `yaml:"smtp_address"`
	SMTPPort     int    `yaml:"smtp_port"`
	SMTPUser     string `yaml:"smtp_user"`
	SMTPPassword string `yaml:"smtp_password"`
	DisplayName  string `yaml:"display_name"`
	Enable       bool   `yaml:"enable"`
	SSL          bool   `yaml:"ssl"`
}

type SeatunnelConfig struct {
	BaseURL string `yaml:"base_url"`
}

type AliyunConfig struct {
	AesKey string `yaml:"aes_key"`
}

type AuthConfig struct {
	JWTSecret string `yaml:"jwt_secret"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type RedisConfig struct {
	Enable   bool   `yaml:"enable"`
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	Prefix   string `yaml:"prefix"`
}

type OctoopsConfig struct {
	Mail   MailConfig   `yaml:"mail"`
	Aliyun AliyunConfig `yaml:"aliyun"`
	Auth   AuthConfig   `yaml:"auth"`
	Server ServerConfig `yaml:"server"`
	Redis  RedisConfig  `yaml:"redis"`
	// 预留字段，后续可扩展
}

type Config struct {
	Octoops   OctoopsConfig   `yaml:"octoops"`
	Postgres  PostgresConfig  `yaml:"postgres"`
	Seatunnel SeatunnelConfig `yaml:"seatunnel"`
}

var (
	SeatunnelBaseURL string
	PostgresDSN      string
	mailConfig       MailConfig
	aliyunAesKey     string
	jwtSecret        string
	serverPort       int
	redisConfig      RedisConfig
)

func overrideStringField(envVar string, field *string) {
	if v := os.Getenv(envVar); v != "" {
		*field = v
	}
}

func overrideIntField(envVar string, field *int) {
	if v := os.Getenv(envVar); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			*field = n
		}
	}
}

func overrideBoolField(envVar string, field *bool) {
	if v := os.Getenv(envVar); v != "" {
		if v == "true" || v == "1" {
			*field = true
		} else if v == "false" || v == "0" {
			*field = false
		}
	}
}

func InitConfig() error {
	cfg := Config{}

	if data, err := os.ReadFile("config.yaml"); err == nil {
		log.Println("读取 config.yaml 配置文件")
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("解析 config.yaml 失败: %w", err)
		}
	} else {
		return fmt.Errorf("读取 config.yaml 失败: %w", err)
	}

	// Postgres
	overrideStringField("POSTGRES_HOST", &cfg.Postgres.Host)
	overrideStringField("POSTGRES_USER", &cfg.Postgres.User)
	overrideStringField("POSTGRES_PASSWORD", &cfg.Postgres.Password)
	overrideStringField("POSTGRES_DBNAME", &cfg.Postgres.DBName)
	overrideIntField("POSTGRES_PORT", &cfg.Postgres.Port)
	overrideStringField("POSTGRES_SSLMODE", &cfg.Postgres.SSLMode)
	overrideStringField("POSTGRES_TIMEZONE", &cfg.Postgres.TimeZone)

	// Seatunnel
	overrideStringField("SEATUNNEL_BASE_URL", &cfg.Seatunnel.BaseURL)

	// Octoops.Mail
	overrideStringField("OCTOOPS_MAIL_SMTP_ADDRESS", &cfg.Octoops.Mail.SMTPAddress)
	overrideIntField("OCTOOPS_MAIL_SMTP_PORT", &cfg.Octoops.Mail.SMTPPort)
	overrideStringField("OCTOOPS_MAIL_SMTP_USER", &cfg.Octoops.Mail.SMTPUser)
	overrideStringField("OCTOOPS_MAIL_SMTP_PASSWORD", &cfg.Octoops.Mail.SMTPPassword)
	overrideStringField("OCTOOPS_MAIL_DISPLAY_NAME", &cfg.Octoops.Mail.DisplayName)
	overrideBoolField("OCTOOPS_MAIL_ENABLE", &cfg.Octoops.Mail.Enable)
	overrideBoolField("OCTOOPS_MAIL_SSL", &cfg.Octoops.Mail.SSL)

	// Octoops.Aliyun
	overrideStringField("OCTOOPS_ALIYUN_AES_KEY", &cfg.Octoops.Aliyun.AesKey)

	// Octoops.Auth
	overrideStringField("OCTOOPS_AUTH_JWT_SECRET", &cfg.Octoops.Auth.JWTSecret)
	// Octoops.Server
	overrideIntField("OCTOOPS_SERVER_PORT", &cfg.Octoops.Server.Port)
	// Octoops.Redis
	overrideBoolField("OCTOOPS_REDIS_ENABLE", &cfg.Octoops.Redis.Enable)
	overrideStringField("OCTOOPS_REDIS_ADDR", &cfg.Octoops.Redis.Addr)
	overrideStringField("OCTOOPS_REDIS_PASSWORD", &cfg.Octoops.Redis.Password)
	overrideIntField("OCTOOPS_REDIS_DB", &cfg.Octoops.Redis.DB)
	overrideStringField("OCTOOPS_REDIS_PREFIX", &cfg.Octoops.Redis.Prefix)

	// 校验必填项
	if cfg.Seatunnel.BaseURL == "" {
		return errors.New("seatunnel.base_url 配置不能为空")
	}
	if cfg.Postgres.Host == "" || cfg.Postgres.User == "" || cfg.Postgres.Password == "" || cfg.Postgres.DBName == "" || cfg.Postgres.Port == 0 || cfg.Postgres.SSLMode == "" || cfg.Postgres.TimeZone == "" {
		return errors.New("postgres 配置不完整")
	}

	SeatunnelBaseURL = cfg.Seatunnel.BaseURL
	PostgresDSN = cfg.Postgres.DSN()
	mailConfig = cfg.Octoops.Mail
	aliyunAesKey = cfg.Octoops.Aliyun.AesKey
	jwtSecret = cfg.Octoops.Auth.JWTSecret
	serverPort = cfg.Octoops.Server.Port
	redisConfig = cfg.Octoops.Redis
	if redisConfig.Prefix == "" {
		redisConfig.Prefix = "octoops:"
	}
	if serverPort == 0 {
		serverPort = 8080
	}

	// 设置JWT密钥
	// Enforce a non-empty JWT secret.
	if jwtSecret == "" {
		return errors.New("octoops.auth.jwt_secret is required")
	}
	return nil
}

func GetMailConfig() MailConfig {
	return mailConfig
}

func GetAliyunAesKey() string {
	return aliyunAesKey
}

func GetJWTSecret() string {
	return jwtSecret
}

func GetServerPort() int {
	return serverPort
}

func GetRedisConfig() RedisConfig {
	return redisConfig
}
