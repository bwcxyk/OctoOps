package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
	"time"
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

type OctoopsConfig struct {
	Mail MailConfig `yaml:"mail"`
	// 预留字段，后续可扩展
}

type Config struct {
	Octoops   OctoopsConfig   `yaml:"octoops"`
	Postgres  PostgresConfig  `yaml:"postgres"`
	Seatunnel SeatunnelConfig `yaml:"seatunnel"`
}

var (
	SeatunnelBaseURL               string
	SeatunnelJobStatusSyncInterval time.Duration
	PostgresDSN                    string
	mailConfig                     MailConfig
)

func InitConfig() {
	cfg := Config{}

	if data, err := ioutil.ReadFile("config.yaml"); err == nil {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			log.Fatalf("解析 config.yaml 失败: %v", err)
		}
	}

	if v := os.Getenv("SEATUNNEL_BASE_URL"); v != "" {
		cfg.Seatunnel.BaseURL = v
	}
	if v := os.Getenv("POSTGRES_DSN"); v != "" {
		PostgresDSN = v
	} else {
		PostgresDSN = cfg.Postgres.DSN()
	}

	// 校验必填项
	if cfg.Seatunnel.BaseURL == "" {
		log.Fatal("seatunnel.base_url 配置不能为空")
	}
	if cfg.Postgres.Host == "" || cfg.Postgres.User == "" || cfg.Postgres.Password == "" || cfg.Postgres.DBName == "" || cfg.Postgres.Port == 0 || cfg.Postgres.SSLMode == "" || cfg.Postgres.TimeZone == "" {
		log.Fatal("postgres 配置不完整")
	}

	SeatunnelBaseURL = cfg.Seatunnel.BaseURL

	// 邮件配置
	mailConfig = cfg.Octoops.Mail
}

func GetMailConfig() MailConfig {
	return mailConfig
}
