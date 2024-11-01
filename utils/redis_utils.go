package utils

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

func SetKey(r *redis.Client, ctx context.Context, key string, value string, expiration time.Duration) error {
	return r.Set(ctx, key, value, expiration).Err()
}

func GetKey(r *redis.Client, ctx context.Context, key string) (string, error) {
	return r.Get(ctx, key).Result()
}

func KeyExists(r *redis.Client, ctx context.Context, key string) (bool, error) {
	res, err := r.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return res == 1, nil
}

func DeleteKey(r *redis.Client, ctx context.Context, key string) error {
	return r.Del(ctx, key).Err()
}

func ExpireKey(r *redis.Client, ctx context.Context, key string, expiration time.Duration) error {
	return r.Expire(ctx, key, expiration).Err()
}

func PrintRedisKeys(ctx context.Context, r *redis.Client) {
	keys, err := r.Keys(ctx, "*").Result() // 獲取所有鍵
	if err != nil {
		log.Printf("Error fetching keys: %v", err)
		return
	}

	for _, key := range keys {
		value, err := r.Get(ctx, key).Result() // 獲取鍵的值
		if err != nil {
			log.Printf("Error fetching value for key %s: %v", key, err)
			continue
		}
		log.Printf("Key: %s, Value: %s", key, value)
	}
}

// 使用 Redis 存储和获取在线用户
func UpdateUserOnlineStatus(redisClient *redis.Client, ctx context.Context, username string, online bool) error {
	if online {
		// 用户上线，设置键值对并设置过期时间为 1 小时
		return SetKey(redisClient, ctx, username, "online", time.Hour)
	} else {
		// 用户下线，删除键
		return DeleteKey(redisClient, ctx, username)
	}
}
