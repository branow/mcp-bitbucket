package config_test

import (
	"fmt"
	"math/rand/v2"
	"os"
	"testing"

	"github.com/branow/mcp-bitbucket/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestGetInt(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		fallback int
		expected int
	}{
		{"Missing env var", "", 10, 10},
		{"Blank env var", "  ", 10, 10},
		{"Invalid env var", "value", 10, 10},
		{"Valid env var", "20", 10, 20},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := fmt.Sprintf("key-%d", rand.Int())
			old := os.Getenv(key)
			defer os.Setenv(key, old)

			os.Setenv(key, tt.value)
			actual := config.GetInt(key, tt.fallback)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestGetString(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		fallback string
		expected string
	}{
		{"Missing env var", "", "fallback", "fallback"},
		{"Blank env var", "  ", "fallback", "fallback"},
		{"Valid env var", "value", "fallback", "value"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := fmt.Sprintf("key-%d", rand.Int())
			old := os.Getenv(key)
			defer os.Setenv(key, old)

			os.Setenv(key, tt.value)
			actual := config.GetString(key, tt.fallback)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
