package token

import (
	"context"
	crand "crypto/rand"
	"encoding/hex"
	randv2 "math/rand/v2"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	// redisLockRetryBaseInterval 基准退避间隔
	redisLockRetryBaseInterval = 250 * time.Millisecond
	// redisLockRetryMaxInterval 最大退避间隔
	redisLockRetryMaxInterval = 900 * time.Millisecond
	// redisLockMaxRetry 获取锁的最大重试次数
	redisLockMaxRetry = 3
	// redisLockJitterPercent 抖动百分比（0.2 表示 ±20%）
	redisLockJitterPercent = 0.2
)

func jitterInterval(base time.Duration) time.Duration {
	if base <= 0 || redisLockJitterPercent <= 0 {
		return base
	}
	p := redisLockJitterPercent
	min := float64(base) * (1 - p)
	max := float64(base) * (1 + p)
	return time.Duration(min + randv2.Float64()*(max-min))
}

func backoffInterval(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	interval := redisLockRetryBaseInterval * time.Duration(1<<attempt)
	if interval > redisLockRetryMaxInterval {
		interval = redisLockRetryMaxInterval
	}
	return jitterInterval(interval)
}

var unlockScript = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
  return redis.call("del", KEYS[1])
end
return 0
`)

// RedisLocker 基于 Redis/Redis 集群的分布式锁
type RedisLocker struct {
	client redis.Cmdable
}

// NewRedisLocker 创建基于 Redis 的锁实现；cmd 可为 *redis.Client 或 *redis.ClusterClient
func NewRedisLocker(cmd redis.Cmdable) *RedisLocker {
	if cmd == nil {
		return nil
	}
	return &RedisLocker{client: cmd}
}

// Lock 获取锁，带有限次数重试
func (r *RedisLocker) Lock(ctx context.Context, key string, ttl time.Duration) (func() error, error) {
	lockVal := randomLockValue()

	for i := 0; i < redisLockMaxRetry; i++ {
		ok, err := r.client.SetNX(ctx, key, lockVal, ttl).Result()
		if err != nil {
			return nil, err
		}
		if ok {
			unlock := func() error {
				return unlockScript.Run(ctx, r.client, []string{key}, lockVal).Err()
			}
			return unlock, nil
		}
		if i == redisLockMaxRetry-1 {
			break
		}
		// 尊重调用方上下文，避免无意义等待；指数退避 + 抖动
		wait := backoffInterval(i)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(wait):
		}
	}

	return nil, ErrLockAcquire
}

func randomLockValue() string {
	b := make([]byte, 16)
	if _, err := crand.Read(b); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(b)
}
