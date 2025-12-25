package main

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/qingfeng-studio/wxgo"
)

func main() {
	appID := "your_app_id"
	appSecret := "your_app_secret"
	ctx := context.Background()

	// ==========================================
	// 用法1：内存缓存（默认，最简单）
	// ==========================================
	fmt.Println("=== 用法1：内存缓存 ===")
	client1, err := wxgo.NewClient(wxgo.Config{
		AppID:     appID,
		AppSecret: appSecret,
		// 不传 Cache、RedisClient、RedisClusterClient
		// 自动使用内存缓存
	})
	if err != nil {
		log.Fatal("创建客户端失败:", err)
	}

	token1, code1, err := client1.GetAccessToken(context.Background())
	if err != nil {
		log.Printf("获取 Token 失败(code=%s): %v\n", code1, err)
	} else {
		fmt.Printf("✅ Token: %s\n\n", token1)
	}

	// ==========================================
	// 用法2：Redis 单点
	// ==========================================
	fmt.Println("=== 用法2：Redis 单点 ===")
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 测试 Redis 连接（可选）
	if err := redisClient.Ping(ctx).Err(); err != nil {
		fmt.Printf("⚠️  Redis 未连接，跳过此示例: %v\n\n", err)
	} else {
		client2, err := wxgo.NewClient(wxgo.Config{
			AppID:            appID,
			AppSecret:        appSecret,
			RedisClient:      redisClient,       // 传入 Redis 单点客户端
			DistLockStrategy: wxgo.DistLockAuto, // 默认 auto：多实例时自动用 Redis 分布式锁
			// 如明确是单实例但仍想用 Redis 缓存，可改为 DistLockOff 关闭分布式锁
		})
		if err != nil {
			log.Fatal("创建客户端失败:", err)
		}

		token2, code2, err := client2.GetAccessToken(ctx)
		if err != nil {
			log.Printf("获取 Token 失败(code=%s): %v\n", code2, err)
		} else {
			fmt.Printf("✅ Token: %s\n\n", token2)
		}
	}

	// ==========================================
	// 用法3：Redis 集群
	// ==========================================
	fmt.Println("=== 用法3：Redis 集群 ===")
	clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{"node1:6379", "node2:6379"},
	})

	// 测试 Redis 集群连接（可选）
	if err := clusterClient.Ping(ctx).Err(); err != nil {
		fmt.Printf("⚠️  Redis 集群未连接，跳过此示例: %v\n\n", err)
	} else {
		client3, err := wxgo.NewClient(wxgo.Config{
			AppID:              appID,
			AppSecret:          appSecret,
			RedisClusterClient: clusterClient, // 传入 Redis 集群客户端
			DistLockStrategy:   wxgo.DistLockAuto,
		})
		if err != nil {
			log.Fatal("创建客户端失败:", err)
		}

		token3, code3, err := client3.GetAccessToken(ctx)
		if err != nil {
			log.Printf("获取 Token 失败(code=%s): %v\n", code3, err)
		} else {
			fmt.Printf("✅ Token: %s\n\n", token3)
		}
	}

	fmt.Println("💡 提示：请将 'your_app_id' 和 'your_app_secret' 替换为实际的微信 AppID 和 AppSecret")
}
