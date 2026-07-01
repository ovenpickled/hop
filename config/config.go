package config

import (
	"fmt"
	"os"
)

// Config holds all environment-driven settings for the service.
type Config struct {
	ServerPort string
	BaseURL string

	RedisAddr string
	RedisPassword string
	RedisDB int

	PostgresDSN string
}

// Load reads configuration from environment variables, falling back to sane local-development defaults when a variable isn't set.
func Load() Config {
	return Config{
		ServerPort: getEnv("PORT", "9808"),
		BaseURL:    getEnv("BASE_URL", "http://localhost:9808/"),

		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       0,

		PostgresDSN: getEnv(
			"POSTGRES_DSN",
			"postgres://hop:hop@localhost:5432/hop?sslmode=disable",
		),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// String is handy for debug logging at startup without leaking secrets.
func (c Config) String() string {
	return fmt.Sprintf(
		"ServerPort=%s BaseURL=%s RedisAddr=%s PostgresDSN=%s",
		c.ServerPort, c.BaseURL, c.RedisAddr, maskDSN(c.PostgresDSN),
	)
}

func maskDSN(dsn string) string {
	// Doesn't print credentials in logs
	return "postgres://****@****"
}
