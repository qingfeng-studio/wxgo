package token

import "errors"

// Code 机器可读的错误码，便于上层做国际化或分支处理
type Code string

const (
	// CodeOK 请求成功
	CodeOK Code = "OK"
	// CodeMissingAppID 缺少 AppID
	CodeMissingAppID Code = "E_MISSING_APP_ID"
	// CodeMissingAppSecret 缺少 AppSecret
	CodeMissingAppSecret Code = "E_MISSING_APP_SECRET"
	// CodeCacheGet 缓存读取失败
	CodeCacheGet Code = "E_CACHE_GET"
	// CodeCacheSet 缓存写入失败
	CodeCacheSet Code = "E_CACHE_SET"
	// CodeHTTP 调用 HTTP 失败
	CodeHTTP Code = "E_HTTP"
	// CodeAPIError 微信 API 返回错误
	CodeAPIError Code = "E_WECHAT_API"
	// CodeInvalidResponse 响应解析失败
	CodeInvalidResponse Code = "E_INVALID_RESPONSE"
	// CodeLock 分布式锁获取失败
	CodeLock Code = "E_LOCK"
	// CodeUnknown 未分类错误
	CodeUnknown Code = "E_UNKNOWN"
)

var (
	// ErrMissingAppID AppID 未设置
	ErrMissingAppID = errors.New("wxgo: app_id is required")

	// ErrMissingAppSecret AppSecret 未设置
	ErrMissingAppSecret = errors.New("wxgo: app_secret is required")

	// ErrInvalidResponse 微信返回的响应无效
	ErrInvalidResponse = errors.New("wxgo: invalid response from wechat api")

	// ErrAPIError 微信 API 返回错误
	ErrAPIError = errors.New("wxgo: api error")

	// ErrLockAcquire 分布式锁获取失败
	ErrLockAcquire = errors.New("wxgo: acquire distributed lock failed")

	// ErrLockBackendMissing 需要分布式锁但未配置可用后端
	ErrLockBackendMissing = errors.New("wxgo: distributed lock required but no backend available")
)
