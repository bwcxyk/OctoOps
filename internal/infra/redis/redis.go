package redis

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"octoops/internal/config"

	goredis "github.com/redis/go-redis/v9"
)

var (
	client *goredis.Client
)

func Init(cfg config.RedisConfig) error {
	if cfg.Addr == "" {
		return fmt.Errorf("redis addr is required")
	}

	client = goredis.NewClient(&goredis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return formatRedisConnectError(cfg, err)
	}
	return nil
}

func formatRedisConnectError(cfg config.RedisConfig, err error) error {
	base := fmt.Sprintf("redis 连接失败(addr=%s, db=%d)", cfg.Addr, cfg.DB)

	var opErr *net.OpError
	if !strings.Contains(strings.ToLower(err.Error()), "connectex") && !errors.As(err, &opErr) {
		return fmt.Errorf("%s: %w", base, err)
	}

	return fmt.Errorf(
		"%s: 无法建立连接，请确认 Redis 已启动且地址可达 (原始错误: %w)",
		base, err,
	)
}

func Client() *goredis.Client {
	return client
}
