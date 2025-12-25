package wxgo

import (
	"context"
	"fmt"

	"github.com/qingfeng-studio/wxgo/internal/token"
	"github.com/qingfeng-studio/wxgo/internal/transport"
)

// Client 微信 API 客户端
type Client struct {
	cfg   Config
	http  *transport.Client
	token *token.Manager
}

// NewClient 创建微信客户端
func NewClient(cfg Config) (*Client, error) {
	// 构建 token 配置
	tokenConfig := &token.Config{
		AppID:              cfg.AppID,
		AppSecret:          cfg.AppSecret,
		Cache:              cfg.Cache,
		RedisClient:        cfg.RedisClient,
		RedisClusterClient: cfg.RedisClusterClient,
		DistLockStrategy:   cfg.DistLockStrategy,
	}

	// 初始化 token manager
	tokenMgr, err := token.NewManager(tokenConfig)
	if err != nil {
		return nil, fmt.Errorf("create token manager: %w", err)
	}

	// 初始化 transport client
	httpClient := transport.NewClient()
	if cfg.HTTPTimeout > 0 {
		httpClient.SetTimeout(cfg.HTTPTimeout)
	}

	return &Client{
		cfg:   cfg,
		http:  httpClient,
		token: tokenMgr,
	}, nil
}

// GetAccessToken 获取 Access Token，返回值：(token, code, err)
// code 便于上层做国际化/分支处理；不需要时可用空标识符忽略
func (c *Client) GetAccessToken(ctx context.Context) (string, token.Code, error) {
	return c.token.GetAccessToken(ctx)
}

// authHeader 获取鉴权 Header（供内部 Service 使用）
func (c *Client) authHeader(ctx context.Context) (string, error) {
	tk, _, err := c.token.GetAccessToken(ctx)
	if err != nil {
		return "", err
	}
	return "Bearer " + tk, nil
}
