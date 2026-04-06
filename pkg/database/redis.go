package database

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func InitRedis(addr, password string, db int) error {
	RDB = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	_, err := RDB.Ping(context.Background()).Result()
	if err != nil {
		return err
	}
	fmt.Println("Redis 连接成功")
	return nil
}
