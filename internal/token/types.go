package token

import "time"

// TokenInfo Access Token 信息
type TokenInfo struct {
	AccessToken string    `json:"access_token"`
	ExpiresIn   int       `json:"expires_in"` // 过期时间（秒）
	ExpiresAt   time.Time // 实际过期时间
}

// IsExpired 检查 Token 是否已过期
// 提前 5 分钟刷新，避免边界情况
func (t *TokenInfo) IsExpired() bool {
	return time.Now().Add(5 * time.Minute).After(t.ExpiresAt)
}

