package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

func InitRedis() (*redis.Client, error) {
	redisHost := os.Getenv("REDIS_HOST")         // 從環境變數中獲取 Redis 主機
	redisPassword := os.Getenv("REDIS_PASSWORD") // 獲取 Redis 密碼
	if redisHost == "" {
		redisHost = "localhost"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisHost + ":6379", // 使用主機和預設端口
		Password: redisPassword,       // 使用環境變數中的密碼
		DB:       0,
	})

	// 创建一个5秒超时的上下文，确保不会无限等待
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Printf("hello")
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// 测试 Redis 连接
	err = rdb.Set(ctx, "test_key", "Hello Redis", 0).Err()
	if err != nil {
		log.Fatalf("Could not set key: %v", err)
	}

	val, err := rdb.Get(ctx, "test_key").Result()
	if err != nil {
		log.Fatalf("Could not get key: %v", err)
	}
	log.Printf("Value of 'test_key': %s", val) // 应该打印出 "Hello Redis"

	return rdb, nil
}
