// Package config provides environment variable loading and parsing utilities.
// It supports loading from .env files and provides type-safe accessors with fallback values.
package config

import (
	"log/slog"
	"os"

	"github.com/branow/mcp-bitbucket/internal/util/schema"
	"github.com/joho/godotenv"
)

var cfg map[string]any

func init() {
	if err := godotenv.Load(); err != nil {
		slog.Info("Failed to load .env file", "error", err)
	}
	cfg = make(map[string]any)
}

// GetCrit retrieves a critical environment variable using the provided schema.
// If the value is missing or fails validation, it panics.
// The value is cached after the first successful retrieval.
//
// Use this for required configuration values where the application cannot function
// without them (e.g., database connection strings, API keys).
//
// Example:
//
//	username := GetCrit("DB_USERNAME", sch.String().Must(sch.NotBlank()).Critical())
func GetCrit[T any](key string, schema schema.Critical[T]) T {
	if val, ok := cfg[key]; ok {
		return val.(T)
	}
	value := schema.Parse(os.Getenv(key))
	cfg[key] = value
	return value
}

// GetReq retrieves a required environment variable using the provided schema.
// If the value is missing or fails validation, it logs an error and returns the fallback value.
// The value is cached after the first retrieval.
//
// Use this for important configuration values where you want to provide a fallback
// but still log that the proper value is missing.
//
// Example:
//
//	authType := GetReq("AUTH_TYPE", sch.String().Must(sch.In("oauth", "basic")), "oauth")
func GetReq[T any](key string, schema schema.Required[T], fallback T) T {
	if val, ok := cfg[key]; ok {
		return val.(T)
	}

	value, err := schema.Parse(os.Getenv(key))
	if err != nil {
		slog.Error("Missing or invalid Required environment variable, fallback applied",
			"envVar", key,
			"fallback", fallback,
			"error", err,
		)
		value = fallback
	}

	cfg[key] = value
	return value
}

// GetOpt retrieves an optional environment variable using the provided schema.
// If the value is missing or fails validation, it logs an info message and returns the fallback value.
// The value is cached after the first retrieval.
//
// Use this for optional configuration values where a sensible default exists.
//
// Example:
//
//	port := GetOpt("SERVER_PORT", sch.Int().Must(sch.Positive()).Optional(8080))
func GetOpt[T any](key string, schema schema.Optional[T]) T {
	if val, ok := cfg[key]; ok {
		return val.(T)
	}
	value := schema.OnFallback(func(fallback T, err error) {
		slog.Info("Missing or invalid optional environment variable, fallback applied",
			"envVar", key,
			"fallback", fallback,
			"error", err,
		)
	}).Parse(os.Getenv(key))
	cfg[key] = value
	return value
}

// ClearCache clears the internal configuration cache.
// This is useful in tests when environment variables are changed between test cases
// and you need to force re-reading from the environment.
func ClearCache() {
	cfg = make(map[string]any)
}
