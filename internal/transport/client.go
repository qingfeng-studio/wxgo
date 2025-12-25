package transport

import (
	"context"
	"net/http"
	"time"
)

// Client HTTP 传输层客户端封装
type Client struct {
	http      *http.Client
	userAgent string
}

// NewClient 创建 HTTP 客户端
func NewClient() *Client {
	return &Client{
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
		userAgent: "wxgo/1.0.0",
	}
}

// Do 执行 HTTP 请求
// 统一入口，后续可在此添加 retry、backoff、metrics、trace 等功能
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	// 统一设置 User-Agent
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	// 执行请求
	return c.http.Do(req.WithContext(ctx))
}

// SetTimeout 设置请求超时时间
func (c *Client) SetTimeout(timeout time.Duration) {
	c.http.Timeout = timeout
}

