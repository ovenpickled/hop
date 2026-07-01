package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/ovenpickled/hop/config"
)

// StorageService wraps the Redis cache and the Postgres durable store.
// Reads check Redis first; on a cache miss we fall back to Postgres and repopulate the cache.
// Writes go to Postgres first (source of truth) then to Redis.
type StorageService struct {
	redisClient *redis.Client
	pg          *postgresStore
}

var (
	storeService = &StorageService{}
	ctx          = context.Background()
)

// InitializeStore sets up both the Redis and Postgres connections.
// Unlike the per-request store functions, it's still appropriate to fail fast (panic) here: if either dependency is unreachable at boot, the app genuinely cannot serve traffic and shouldn't pretend to start.
func InitializeStore(cfg config.Config) *StorageService {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	if pong, err := redisClient.Ping(ctx).Result(); err != nil {
		panic(fmt.Sprintf("Error init Redis: %v", err))
	} else {
		fmt.Printf("Redis started successfully: pong message = {%s}\n", pong)
	}

	pg, err := newPostgresStore(cfg.PostgresDSN)
	if err != nil {
		panic(fmt.Sprintf("Error init Postgres: %v", err))
	}
	fmt.Println("Postgres started successfully")

	storeService.redisClient = redisClient
	storeService.pg = pg
	return storeService
}

// SaveUrlMapping persists the mapping durably in Postgres, then warms the Redis cache.
// Returns an error instead of panicking so the handler layer can decide how to respond (e.g. 500 vs. a specific message).
func SaveUrlMapping(shortUrl string, originalUrl string, userId string) error {
	if err := storeService.pg.Save(shortUrl, originalUrl, userId); err != nil {
		return fmt.Errorf("saving url mapping to postgres: %w", err)
	}

	// Cache write is best-effort: if Redis is briefly unavailable, the data is still safe in Postgres, so we log and move on rather than failing the whole request.
	if err := storeService.redisClient.Set(ctx, shortUrl, originalUrl, 0).Err(); err != nil {
		fmt.Printf("warning: failed to warm cache for %s: %v\n", shortUrl, err)
	}

	return nil
}

// RetrieveInitialUrl looks up the long url for a given short url. It checks Redis first;
// on a cache miss (or a transient Redis error) it falls back to Postgres and repopulates the cache for next time.
// Returns ErrNotFound if no mapping exists in either store.
func RetrieveInitialUrl(shortUrl string) (string, error) {
	result, err := storeService.redisClient.Get(ctx, shortUrl).Result()
	if err == nil {
		return result, nil
	}
	if !errors.Is(err, redis.Nil) {
		fmt.Printf("warning: redis error on get for %s: %v\n", shortUrl, err)
	}

	longUrl, err := storeService.pg.Get(shortUrl)
	if err != nil {
		return "", err
	}

	// Repopulate the cache so the next read is fast.
	if err := storeService.redisClient.Set(ctx, shortUrl, longUrl, 0).Err(); err != nil {
		fmt.Printf("warning: failed to repopulate cache for %s: %v\n", shortUrl, err)
	}

	return longUrl, nil
}
