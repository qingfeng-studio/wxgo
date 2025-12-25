# wxgo

微信 API 接口的 Golang SDK 封装包，提供简洁易用的微信接口调用能力。

## ✨ 特性

- ✅ **Access Token 管理**：自动获取、缓存和刷新 Access Token
- ✅ **灵活的缓存策略**：支持内存、Redis 单点、Redis 集群和自定义缓存
- ✅ **并发安全**：内置并发控制，避免重复请求
- ✅ **自动刷新**：Token 过期前自动刷新（提前 5 分钟）
- ✅ **简单易用**：简洁的 API 设计，快速上手

## 📦 安装

```bash
go get github.com/qingfeng-studio/wxgo
```

## 🚀 快速开始

### 基础用法（内存缓存）

最简单的使用方式，适合单机应用或测试环境：

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/qingfeng-studio/wxgo"
)

func main() {
    client, err := wxgo.NewClient(wxgo.Config{
        AppID:     "your_app_id",
        AppSecret: "your_app_secret",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    token, code, err := client.GetAccessToken(context.Background())
    if err != nil {
        log.Fatalf("get token failed (code=%s): %v", code, err)
    }
    
    fmt.Println("Access Token:", token)
}
```

### Redis 单点缓存

适合生产环境，多实例共享 Token：

```go
import (
    "github.com/go-redis/redis/v8"
    "github.com/qingfeng-studio/wxgo"
)

redisClient := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

client, err := wxgo.NewClient(wxgo.Config{
    AppID:       "your_app_id",
    AppSecret:   "your_app_secret",
    RedisClient: redisClient,
})
if err != nil {
    log.Fatal(err)
}
```

### Redis 集群缓存

适合大规模、高可用场景：

```go
clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
    Addrs: []string{"node1:6379", "node2:6379"},
})

client, err := wxgo.NewClient(wxgo.Config{
    AppID:             "your_app_id",
    AppSecret:         "your_app_secret",
    RedisClusterClient: clusterClient,
})
if err != nil {
    log.Fatal(err)
}
```

### 自定义缓存实现

实现 `token.Cache` 接口即可使用自定义缓存：

```go
import "github.com/qingfeng-studio/wxgo/internal/token"

type CustomCache struct {
    // 你的缓存实现
}

func (c *CustomCache) Get(ctx context.Context, key string) (*token.TokenInfo, error) {
    // 实现获取逻辑
}

func (c *CustomCache) Set(ctx context.Context, key string, token *token.TokenInfo, ttl time.Duration) error {
    // 实现设置逻辑
}

func (c *CustomCache) Delete(ctx context.Context, key string) error {
    // 实现删除逻辑
}

// 使用自定义缓存
client, err := wxgo.NewClient(wxgo.Config{
    AppID:     "your_app_id",
    AppSecret: "your_app_secret",
    Cache:     &CustomCache{},
})
```

## 📖 API 文档

### Config

客户端配置结构：

```go
type Config struct {
    AppID              string              // 微信公众号/小程序的 AppID（必填）
    AppSecret          string              // 微信公众号/小程序的 AppSecret（必填）
    Cache              token.Cache         // 自定义缓存实现（优先级最高）
    RedisClient        *redis.Client       // Redis 单点客户端
    RedisClusterClient *redis.ClusterClient // Redis 集群客户端
}
```

**缓存优先级**：`Cache` > `RedisClient` > `RedisClusterClient` > 内存缓存（默认）

### Client

微信 API 客户端：

```go
// NewClient 创建微信客户端
func NewClient(cfg Config) (*Client, error)

// GetAccessToken 获取 Access Token（自动处理缓存和刷新）
func (c *Client) GetAccessToken(ctx context.Context) (string, error)
```

## 💡 使用示例

项目提供了完整的使用示例，展示三种不同的缓存方式：

```bash
cd examples
go run main.go
```

示例会依次演示：
1. **内存缓存**：最简单的方式，适合单机或测试
2. **Redis 单点**：适合生产环境，多实例共享 Token
3. **Redis 集群**：适合大规模、高可用场景

**注意**：运行前请将代码中的 `your_app_id` 和 `your_app_secret` 替换为实际的微信 AppID 和 AppSecret。

如果 Redis 未运行，示例会自动跳过 Redis 相关的测试，只运行内存缓存示例。

## 🔧 工作原理

1. **Token 获取**：首次调用 `GetAccessToken()` 时，从微信 API 获取 Token
2. **缓存存储**：Token 自动存储到配置的缓存中
3. **自动刷新**：Token 过期前 5 分钟自动刷新，避免边界情况
4. **并发控制**：使用互斥锁防止并发请求导致重复刷新

## 📋 功能规划

### 当前版本（v0.1.0）
- ✅ Access Token 管理（获取、缓存、自动刷新）
- ✅ 灵活的存储方式（内存、Redis 单点、Redis 集群、自定义）

### 后续计划
- 🔲 二维码生成
- 🔲 公众号消息处理
- 🔲 小程序接口
- 🔲 支付接口
- 🔲 企业微信接口

## 📄 License

MIT
