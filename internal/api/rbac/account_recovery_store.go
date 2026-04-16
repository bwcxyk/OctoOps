package rbac

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"octoops/internal/config"
	infraRedis "octoops/internal/infra/redis"

	redis "github.com/redis/go-redis/v9"
)

// RecoveryStore is a storage abstraction for account-recovery code and rate data.
type RecoveryStore interface {
	GetCode(email string) (resetCodeEntry, bool, error)
	SetCode(email string, entry resetCodeEntry) error
	DeleteCode(email string) error

	GetRate(key string) (resetRateEntry, bool, error)
	SetRate(key string, entry resetRateEntry) error
}

func GetRecoveryStore() (RecoveryStore, error) {
	return NewRedisRecoveryStore(infraRedis.Client(), config.GetRedisConfig().Prefix)
}

type redisRecoveryStore struct {
	client *redis.Client
	prefix string
}

func NewRedisRecoveryStore(client *redis.Client, prefix string) (RecoveryStore, error) {
	if client == nil {
		return nil, fmt.Errorf("redis client is nil")
	}
	if prefix == "" {
		prefix = "octoops:"
	}

	return &redisRecoveryStore{
		client: client,
		prefix: prefix,
	}, nil
}

func (s *redisRecoveryStore) GetCode(email string) (resetCodeEntry, bool, error) {
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

func (s *redisRecoveryStore) SetCode(email string, entry resetCodeEntry) error {
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

func (s *redisRecoveryStore) DeleteCode(email string) error {
	key := s.codeKey(email)
	return s.client.Del(context.Background(), key).Err()
}

func (s *redisRecoveryStore) GetRate(key string) (resetRateEntry, bool, error) {
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

func (s *redisRecoveryStore) SetRate(key string, entry resetRateEntry) error {
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

func (s *redisRecoveryStore) codeKey(email string) string {
	return s.prefix + "reset:code:" + email
}

func (s *redisRecoveryStore) rateKey(key string) string {
	return s.prefix + "reset:rate:" + key
}
