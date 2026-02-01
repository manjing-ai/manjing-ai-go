package redis

import (
	"context"
	"time"

	"manjing-ai-go/config"

	"github.com/redis/go-redis/v9"
)

// Client Redis 客户端
type Client struct {
	RDB *redis.Client
}

// New 创建 Redis 客户端
func New(cfg config.RedisConfig) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	return &Client{RDB: rdb}
}

// Close 关闭连接
func (c *Client) Close() error {
	return c.RDB.Close()
}

// Ping 健康检查
func (c *Client) Ping(ctx context.Context) error {
	return c.RDB.Ping(ctx).Err()
}

// SetTokenBlacklisted 拉黑 token
func (c *Client) SetTokenBlacklisted(ctx context.Context, token string, ttl time.Duration) error {
	return c.RDB.Set(ctx, tokenBlacklistKey(token), 1, ttl).Err()
}

// IsTokenBlacklisted 判断 token 是否被拉黑
func (c *Client) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	val, err := c.RDB.Exists(ctx, tokenBlacklistKey(token)).Result()
	if err != nil {
		return false, err
	}
	return val == 1, nil
}

func tokenBlacklistKey(token string) string {
	return "auth:blacklist:" + token
}
