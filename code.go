package wxgo

import "github.com/qingfeng-studio/wxgo/internal/token"

// Code 机器可读的错误码，便于调用方做国际化或分支处理
type Code = token.Code

const (
	// CodeOK 请求成功
	CodeOK = token.CodeOK
	// CodeMissingAppID 缺少 AppID
	CodeMissingAppID = token.CodeMissingAppID
	// CodeMissingAppSecret 缺少 AppSecret
	CodeMissingAppSecret = token.CodeMissingAppSecret
	// CodeCacheGet 缓存读取失败
	CodeCacheGet = token.CodeCacheGet
	// CodeCacheSet 缓存写入失败
	CodeCacheSet = token.CodeCacheSet
	// CodeHTTP 调用 HTTP 失败
	CodeHTTP = token.CodeHTTP
	// CodeAPIError 微信 API 返回错误
	CodeAPIError = token.CodeAPIError
	// CodeInvalidResponse 响应解析失败
	CodeInvalidResponse = token.CodeInvalidResponse
	// CodeLock 分布式锁获取失败
	CodeLock = token.CodeLock
	// CodeUnknown 未分类错误
	CodeUnknown = token.CodeUnknown
)

// DistLockStrategy 分布式锁策略
type DistLockStrategy = token.DistLockStrategy

const (
	// DistLockAuto 自动：缓存若带锁优先，用 Redis/集群可回退到 Redis 锁；否则本地锁
	DistLockAuto = token.DistLockAuto
	// DistLockOn 强制分布式锁：后端不支持则报错
	DistLockOn = token.DistLockOn
	// DistLockOff 关闭分布式锁，只用本地互斥
	DistLockOff = token.DistLockOff
)
