package config

import (
	"context"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

var RedisClient *redis.Client
var Ctx = context.Background()

func InitRedis() error {
	if err := godotenv.Load(); err != nil {
		return err
	}

	redisDB, _ := strconv.Atoi(os.Getenv("REDIS_DB"))

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       redisDB,
	})

	if err := RedisClient.Ping(Ctx).Err(); err != nil {
		return err
	}

	return nil
} 