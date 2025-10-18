package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

var cfg map[string]any

func init() {
	if err := godotenv.Load(); err != nil {
		slog.Error("Failed to load .env file", "error", err)
	}
	cfg = make(map[string]any)
}

// GetInt retrieves an integer environment variable by key.
// If the environment variable is missing or invalid, it returns the provided fallback.
func GetInt(key string, fallback int) int {
	return get(key, fallback, func(value string) (int, error) {
		return strconv.Atoi(value)
	})
}

// GetString retrieves an string environment variable by key.
// If the environment variable is missing, it returns the provided fallback.
func GetString(key string, fallback string) string {
	return get(key, fallback, func(value string) (string, error) {
		return value, nil
	})
}

func get[T any](key string, fallback T, transform func(string) (T, error)) T {
	if val, ok := cfg[key]; ok {
		return val.(T)
	}

	raw, err := getEnvVar(key)
	if err != nil {
		slog.Warn("Environment variable not set, fallback applied",
			"envVar", key,
			"fallback", fallback,
			"error", err,
		)
		cfg[key] = fallback
		return fallback
	}

	value, err := transform(raw)
	if err != nil {
		slog.Warn("Environment variable invalid, fallback applied",
			"envVar", key,
			"value", raw,
			"fallback", fallback,
			"error", err,
		)
		value = fallback
	}

	cfg[key] = value
	return value
}

func getEnvVar(key string) (string, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return "", MissingEnvVarError{Key: key}
	}
	return value, nil
}
