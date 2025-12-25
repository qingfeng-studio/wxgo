package token

import (
	"context"
	"sync"
	"time"
)

// Cache Token 缓存接口
type Cache interface {
	// Get 获取缓存的 Token
	Get(ctx context.Context, key string) (*TokenInfo, error)

	// Set 设置 Token 到缓存
	Set(ctx context.Context, key string, token *TokenInfo, ttl time.Duration) error

	// Delete 删除 Token（可选）
	Delete(ctx context.Context, key string) error
}

// MemoryCache 内存缓存实现
type MemoryCache struct {
	mu    sync.RWMutex
	store map[string]*TokenInfo
}

// NewMemoryCache 创建内存缓存实例
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		store: make(map[string]*TokenInfo),
	}
}

// Get 从内存获取 Token
func (m *MemoryCache) Get(ctx context.Context, key string) (*TokenInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	token, ok := m.store[key]
	if !ok {
		return nil, nil
	}
	return token, nil
}

// Set 设置 Token 到内存
func (m *MemoryCache) Set(ctx context.Context, key string, token *TokenInfo, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.store[key] = token
	return nil
}

// Delete 删除 Token
func (m *MemoryCache) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.store, key)
	return nil
}

