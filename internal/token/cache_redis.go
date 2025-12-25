package token

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisCache Redis 缓存实现（单点）
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache 创建 Redis 缓存实例
func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{
		client: client,
	}
}

// Get 从 Redis 获取 Token
func (r *RedisCache) Get(ctx context.Context, key string) (*TokenInfo, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var token TokenInfo
	if err := json.Unmarshal([]byte(val), &token); err != nil {
		return nil, err
	}

	return &token, nil
}

// Set 设置 Token 到 Redis
func (r *RedisCache) Set(ctx context.Context, key string, token *TokenInfo, ttl time.Duration) error {
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}

// Delete 删除 Token
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// RedisClusterCache Redis 集群缓存实现
type RedisClusterCache struct {
	client *redis.ClusterClient
}

// NewRedisClusterCache 创建 Redis 集群缓存实例
func NewRedisClusterCache(client *redis.ClusterClient) *RedisClusterCache {
	return &RedisClusterCache{
		client: client,
	}
}

// Get 从 Redis 集群获取 Token
func (r *RedisClusterCache) Get(ctx context.Context, key string) (*TokenInfo, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var token TokenInfo
	if err := json.Unmarshal([]byte(val), &token); err != nil {
		return nil, err
	}

	return &token, nil
}

// Set 设置 Token 到 Redis 集群
func (r *RedisClusterCache) Set(ctx context.Context, key string, token *TokenInfo, ttl time.Duration) error {
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}

// Delete 删除 Token
func (r *RedisClusterCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

