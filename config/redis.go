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

// InitRedis initializes the Redis client connection
func InitRedis() error {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		return err
	}

	// Parse Redis DB value
	redisDB, _ := strconv.Atoi(os.Getenv("REDIS_DB"))

	// Initialize Redis client
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       redisDB,
	})

	// Test the connection
	if err := RedisClient.Ping(Ctx).Err(); err != nil {
		return err
	}

	return nil
} 