package token

import (
	"context"
	"time"
)

// DistLockStrategy 分布式锁策略
type DistLockStrategy string

const (
	// DistLockAuto 自动判断：优先使用缓存自带的 TokenLocker，否则在 Redis/RedisCluster 上启用分布式锁；否则退回本地锁
	DistLockAuto DistLockStrategy = "auto"
	// DistLockOn 强制需要分布式锁；若当前后端不支持，则返回配置错误
	DistLockOn DistLockStrategy = "on"
	// DistLockOff 关闭分布式锁，只用本地互斥
	DistLockOff DistLockStrategy = "off"
)

// TokenLocker 可选接口：缓存若实现它，可提供分布式锁
type TokenLocker interface {
	// Lock 获取锁，返回解锁函数
	Lock(ctx context.Context, key string, ttl time.Duration) (func() error, error)
}
