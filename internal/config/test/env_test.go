package config_test

import (
	"fmt"
	"math/rand/v2"
	"os"
	"testing"

	"github.com/branow/mcp-bitbucket/internal/config"
	sch "github.com/branow/mcp-bitbucket/internal/util/schema"
	"github.com/stretchr/testify/assert"
)

func TestGetCrit(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		expectPanic bool
	}{
		{"Missing env var", "", true},
		{"Blank env var", "  ", true},
		{"Valid env var", "value", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := fmt.Sprintf("key-%d", rand.Int())
			old := os.Getenv(key)
			defer os.Setenv(key, old)
			defer config.ClearCache()

			os.Setenv(key, tt.value)
			if tt.expectPanic {
				assert.Panics(t, func() {
					config.GetCrit(key, sch.String().Must(sch.NotBlank()).Critical())
				})
			} else {
				actual := config.GetCrit(key, sch.String().Must(sch.NotBlank()).Critical())
				assert.Equal(t, tt.value, actual)
			}
		})
	}
}

func TestGetReq(t *testing.T) {
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
			defer config.ClearCache()

			os.Setenv(key, tt.value)
			actual := config.GetReq(key, sch.String().Must(sch.NotBlank()), tt.fallback)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestGetOpt(t *testing.T) {
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
			defer config.ClearCache()

			os.Setenv(key, tt.value)
			actual := config.GetOpt(key, sch.String().Must(sch.NotBlank()).Optional(tt.fallback))
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestClearCache(t *testing.T) {
	key := fmt.Sprintf("key-%d", rand.Int())
	old := os.Getenv(key)
	defer os.Setenv(key, old)

	os.Setenv(key, "initial")
	first := config.GetOpt(key, sch.String().Optional("fallback"))
	assert.Equal(t, "initial", first)

	os.Setenv(key, "changed")
	second := config.GetOpt(key, sch.String().Optional("fallback"))
	assert.Equal(t, "initial", second, "Should return cached value")

	config.ClearCache()
	third := config.GetOpt(key, sch.String().Optional("fallback"))
	assert.Equal(t, "changed", third, "Should return new value after cache clear")
}
