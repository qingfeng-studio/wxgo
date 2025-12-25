package token

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	// WeChatTokenAPI 微信获取 Access Token 的 API 地址
	WeChatTokenAPI = "https://api.weixin.qq.com/cgi-bin/token"

	// defaultLockTTL 分布式锁的默认租约时间（覆盖一次微信请求的耗时）
	defaultLockTTL = 15 * time.Second
)

type cacheKind string

const (
	cacheKindCustom cacheKind = "custom"        // 调用方自定义缓存
	cacheKindRedis  cacheKind = "redis"         // Redis 单点
	cacheKindRC     cacheKind = "redis-cluster" // Redis 集群
	cacheKindMemory cacheKind = "memory"        // 内存缓存
)

// Manager Token 管理器
type Manager struct {
	config     *Config
	cache      Cache
	httpClient *http.Client
	mu         sync.Mutex // 保护并发获取 token（本地）

	distLocker   TokenLocker
	lockStrategy DistLockStrategy
	lockTTL      time.Duration
	selectedKind cacheKind
}

// NewManager 创建 Token 管理器
func NewManager(config *Config) (*Manager, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	cacheImpl, cacheKind := resolveCache(config)
	strategy := config.lockStrategy()
	locker, err := resolveLocker(config, cacheKind, cacheImpl, strategy)
	if err != nil {
		return nil, err
	}

	return &Manager{
		config:       config,
		cache:        cacheImpl,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		distLocker:   locker,
		lockStrategy: strategy,
		lockTTL:      defaultLockTTL,
		selectedKind: cacheKind,
	}, nil
}

// GetAccessToken 获取 Access Token
// 逻辑说明：
// 1) 缓存选择优先级：Cache > RedisClusterClient > RedisClient > 内存；
// 2) 分布式锁策略（DistLockStrategy）：
//   - auto（默认）：若缓存实现 TokenLocker 用它；否则在 Redis/集群上加锁（若可用）；否则只用本地锁
//   - on：强制需要分布式锁，后端不满足则报错
//   - off：只用本地锁
//
// 3) 锁获取失败不静默降级，返回 CodeLock 供上层决策
// 4) 本地 mutex 仍保留，避免同进程重复刷新
func (m *Manager) GetAccessToken(ctx context.Context) (string, Code, error) {
	cacheKey := m.getCacheKey()

	// 先从缓存获取
	token, err := m.cache.Get(ctx, cacheKey)
	if err != nil {
		return "", CodeCacheGet, fmt.Errorf("get token from cache: %w", err)
	}

	// 如果缓存存在且未过期，直接返回
	if token != nil && !token.IsExpired() {
		return token.AccessToken, CodeOK, nil
	}

	// 需要刷新 token，使用 mutex 防止并发请求
	m.mu.Lock()
	defer m.mu.Unlock()

	// 双重检查，可能其他 goroutine 已经刷新了
	token, err = m.cache.Get(ctx, cacheKey)
	if err != nil {
		return "", CodeCacheGet, fmt.Errorf("get token from cache: %w", err)
	}
	if token != nil && !token.IsExpired() {
		return token.AccessToken, CodeOK, nil
	}

	// 如果需要分布式互斥，先取锁
	unlock, err := m.acquireDistLock(ctx)
	if err != nil {
		return "", CodeLock, err
	}
	if unlock != nil {
		defer unlock()
	}

	// 锁内再检查一次，避免其他实例已写入
	token, err = m.cache.Get(ctx, cacheKey)
	if err != nil {
		return "", CodeCacheGet, fmt.Errorf("get token from cache: %w", err)
	}
	if token != nil && !token.IsExpired() {
		return token.AccessToken, CodeOK, nil
	}

	// 从微信 API 获取新 token
	newToken, code, err := m.fetchTokenFromWeChat(ctx)
	if err != nil {
		return "", code, err
	}

	// 保存到缓存
	ttl := time.Duration(newToken.ExpiresIn) * time.Second
	if err := m.cache.Set(ctx, cacheKey, newToken, ttl); err != nil {
		// 返回缓存写入错误，便于上层观测；token 仍返回供调用方兜底使用
		return newToken.AccessToken, CodeCacheSet, fmt.Errorf("set token to cache: %w", err)
	}

	return newToken.AccessToken, CodeOK, nil
}

// fetchTokenFromWeChat 从微信 API 获取 Token
func (m *Manager) fetchTokenFromWeChat(ctx context.Context) (*TokenInfo, Code, error) {
	params := url.Values{}
	params.Set("grant_type", "client_credential")
	params.Set("appid", m.config.AppID)
	params.Set("secret", m.config.AppSecret)

	reqURL := WeChatTokenAPI + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, CodeHTTP, fmt.Errorf("create request: %w", err)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, CodeHTTP, fmt.Errorf("request wechat api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, CodeHTTP, fmt.Errorf("wechat api status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, CodeHTTP, fmt.Errorf("read response: %w", err)
	}

	var apiResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, CodeInvalidResponse, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}

	// 检查微信 API 错误
	if apiResp.ErrCode != 0 {
		return nil, CodeAPIError, fmt.Errorf("%w: errcode=%d, errmsg=%s", ErrAPIError, apiResp.ErrCode, apiResp.ErrMsg)
	}

	if apiResp.AccessToken == "" {
		return nil, CodeInvalidResponse, ErrInvalidResponse
	}

	// 计算实际过期时间
	tokenInfo := &TokenInfo{
		AccessToken: apiResp.AccessToken,
		ExpiresIn:   apiResp.ExpiresIn,
		ExpiresAt:   time.Now().Add(time.Duration(apiResp.ExpiresIn) * time.Second),
	}

	return tokenInfo, CodeOK, nil
}

// getCacheKey 获取缓存 key
func (m *Manager) getCacheKey() string {
	return fmt.Sprintf("wxgo:token:%s", m.config.AppID)
}

// getLockKey 获取分布式锁 key
func (m *Manager) getLockKey() string {
	return fmt.Sprintf("wxgo:token_lock:%s", m.config.AppID)
}

func (m *Manager) acquireDistLock(ctx context.Context) (func() error, error) {
	if m.distLocker == nil {
		return nil, nil
	}
	unlock, err := m.distLocker.Lock(ctx, m.getLockKey(), m.lockTTL)
	if err != nil {
		return nil, err
	}
	return unlock, nil
}

// resolveCache 根据配置选择缓存实现（优先级：Cache > RedisCluster > Redis > 内存）
func resolveCache(c *Config) (Cache, cacheKind) {
	if c.Cache != nil {
		return c.Cache, cacheKindCustom
	}
	if c.RedisClusterClient != nil {
		return NewRedisClusterCache(c.RedisClusterClient), cacheKindRC
	}
	if c.RedisClient != nil {
		return NewRedisCache(c.RedisClient), cacheKindRedis
	}
	return NewMemoryCache(), cacheKindMemory
}

// resolveLocker 根据策略与缓存类型选择分布式锁
func resolveLocker(c *Config, kind cacheKind, cache Cache, strategy DistLockStrategy) (TokenLocker, error) {
	switch strategy {
	case DistLockOff:
		return nil, nil
	case DistLockAuto, DistLockOn:
		// 1) 缓存自带 TokenLocker 优先
		if locker, ok := cache.(TokenLocker); ok && locker != nil {
			return locker, nil
		}

		// 2) 没有自带锁时，若使用 Redis/集群缓存则复用它做锁（集群优先）
		if kind == cacheKindRC && c.RedisClusterClient != nil {
			return NewRedisLocker(c.RedisClusterClient), nil
		}
		if kind == cacheKindRedis && c.RedisClient != nil {
			return NewRedisLocker(c.RedisClient), nil
		}

		if strategy == DistLockOn {
			return nil, ErrLockBackendMissing
		}
		return nil, nil
	default:
		return nil, nil
	}
}
