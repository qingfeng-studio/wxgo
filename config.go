package wxgo

import (
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/qingfeng-studio/wxgo/internal/token"
)

// Config 客户端配置
type Config struct {
	// AppID 微信公众号/小程序的 AppID
	AppID string

	// AppSecret 微信公众号/小程序的 AppSecret
	AppSecret string

	// Cache 自定义缓存实现（优先级最高）
	Cache token.Cache

	// RedisClient Redis 单点客户端指针
	RedisClient *redis.Client

	// RedisClusterClient Redis 集群客户端指针
	RedisClusterClient *redis.ClusterClient

	// DistLockStrategy 分布式锁策略：auto/on/off；默认 auto
	DistLockStrategy token.DistLockStrategy

	// HTTPTimeout 调用微信接口的超时时间；默认 10s
	HTTPTimeout time.Duration
}
