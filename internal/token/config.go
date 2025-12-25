package token

import "github.com/go-redis/redis/v8"

// Config Token 管理器配置
type Config struct {
	// AppID 微信公众号/小程序的 AppID
	AppID string

	// AppSecret 微信公众号/小程序的 AppSecret
	AppSecret string

	// Cache 自定义缓存实现（优先级最高）
	Cache Cache

	// RedisClient Redis 单点客户端指针
	RedisClient *redis.Client

	// RedisClusterClient Redis 集群客户端指针
	RedisClusterClient *redis.ClusterClient

	// DistLockStrategy 分布式锁策略：auto/on/off；默认 auto
	DistLockStrategy DistLockStrategy
}

// Validate 验证配置是否有效
func (c *Config) Validate() error {
	if c.AppID == "" {
		return ErrMissingAppID
	}
	if c.AppSecret == "" {
		return ErrMissingAppSecret
	}
	return nil
}

// GetCache 获取缓存实现（按优先级选择）
// 优先级：Cache > RedisClusterClient > RedisClient > 内存
// 即便多种同时传入，也按优先级选定一个，不报错
func (c *Config) GetCache() Cache {
	if c.Cache != nil {
		return c.Cache
	}
	if c.RedisClusterClient != nil {
		return NewRedisClusterCache(c.RedisClusterClient)
	}
	if c.RedisClient != nil {
		return NewRedisCache(c.RedisClient)
	}
	// 默认使用内存缓存
	return NewMemoryCache()
}

// lockStrategy 返回有效的分布式锁策略，默认 auto
func (c *Config) lockStrategy() DistLockStrategy {
	if c.DistLockStrategy == "" {
		return DistLockAuto
	}
	return c.DistLockStrategy
}
