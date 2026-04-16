package rbac

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"octoops/internal/config"

	redis "github.com/redis/go-redis/v9"
)

type redisPasswordResetStore struct {
	client *redis.Client
	prefix string
}

func NewRedisPasswordResetStore(addr, password string, db int, prefix string) (PasswordResetStore, error) {
	if addr == "" {
		return nil, fmt.Errorf("redis addr is required")
	}
	if prefix == "" {
		prefix = "octoops:"
	}
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return &redisPasswordResetStore{
		client: client,
		prefix: prefix,
	}, nil
}

func InitPasswordResetStore() error {
	redisCfg := config.GetRedisConfig()
	if !redisCfg.Enable {
		return nil
	}
	store, err := NewRedisPasswordResetStore(redisCfg.Addr, redisCfg.Password, redisCfg.DB, redisCfg.Prefix)
	if err != nil {
		return err
	}
	return SetPasswordResetStore(store)
}

func (s *redisPasswordResetStore) GetCode(email string) (resetCodeEntry, bool, error) {
	key := s.codeKey(email)
	ctx := context.Background()
	val, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return resetCodeEntry{}, false, nil
	}
	if err != nil {
		return resetCodeEntry{}, false, err
	}
	var entry resetCodeEntry
	if err := json.Unmarshal([]byte(val), &entry); err != nil {
		return resetCodeEntry{}, false, err
	}
	return entry, true, nil
}

func (s *redisPasswordResetStore) SetCode(email string, entry resetCodeEntry) error {
	key := s.codeKey(email)
	ctx := context.Background()
	payload, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	ttl := time.Until(entry.ExpiresAt)
	if ttl <= 0 {
		ttl = time.Second
	}
	return s.client.Set(ctx, key, payload, ttl).Err()
}

func (s *redisPasswordResetStore) DeleteCode(email string) error {
	key := s.codeKey(email)
	return s.client.Del(context.Background(), key).Err()
}

func (s *redisPasswordResetStore) GetRate(key string) (resetRateEntry, bool, error) {
	redisKey := s.rateKey(key)
	ctx := context.Background()
	val, err := s.client.Get(ctx, redisKey).Result()
	if err == redis.Nil {
		return resetRateEntry{}, false, nil
	}
	if err != nil {
		return resetRateEntry{}, false, err
	}
	var entry resetRateEntry
	if err := json.Unmarshal([]byte(val), &entry); err != nil {
		return resetRateEntry{}, false, err
	}
	return entry, true, nil
}

func (s *redisPasswordResetStore) SetRate(key string, entry resetRateEntry) error {
	redisKey := s.rateKey(key)
	ctx := context.Background()
	payload, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	ttl := resetCodeSendWindow - time.Since(entry.WindowStart)
	if ttl <= 0 {
		ttl = time.Second
	}
	return s.client.Set(ctx, redisKey, payload, ttl).Err()
}

func (s *redisPasswordResetStore) codeKey(email string) string {
	return s.prefix + "reset:code:" + email
}

func (s *redisPasswordResetStore) rateKey(key string) string {
	return s.prefix + "reset:rate:" + key
}
