package store

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)

const CacheDuration = 6 * time.Hour

// struct wrapper around the raw Redis client
type StorageService struct {
	redisClient *redis.Client
}

// Top level declarations for the storeService and Redis context
var (
	storeService = &StorageService{}
	ctx = context.Background()
)

// Initializing the store service and return a store pointer
func InitializeStore() *StorageService {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		Password: "",
		DB: 0,
	})

	pong, err := redisClient.Ping(ctx).Result()
	if err != nil {
		panic(fmt.Sprintf("Error init Redis: %v", err))
	}

	fmt.Printf("\nRedis started successfully: pong message = {%s}", pong)
	storeService.redisClient = redisClient
	return storeService
}

// Save the mapping between the originalUrl and the generated shortUrl
func SaveUrlMapping(shortUrl string, originalUrl string, userId string) {
	err := storeService.redisClient.Set(ctx, shortUrl, originalUrl, CacheDuration).Err()
	if err != nil {
		panic(fmt.Sprintf("Failed saving key url | Error: %v - shortUrl: %s - originalUrl: %s\n", err, shortUrl, originalUrl))
	}
}

// Retrieve the initial long URL once the short is provided. Here we retrieve the long url and then redirect
func RetrieveInitialUrl(shortUrl string) string {
	result, err := storeService.redisClient.Get(ctx, shortUrl).Result()
	if err != nil {
		panic(fmt.Sprintf("Failed RetrieveInitialUrl url | Error: %v - shortUrl: %s\n", err, shortUrl))
	}
	return result
}
