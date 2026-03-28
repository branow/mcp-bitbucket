package config_test

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/branow/mcp-bitbucket/internal/config"
	"github.com/branow/mcp-bitbucket/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGlobal_BasicAuth(t *testing.T) {
	t.Cleanup(config.ClearCache)

	emailKey := randKey("BITBUCKET_EMAIL")
	tokenKey := randKey("BITBUCKET_API_TOKEN")
	authKey := randKey("BITBUCKET_AUTH")

	t.Setenv(emailKey, "user@example.com")
	t.Setenv(tokenKey, "my-api-token")
	t.Setenv(authKey, "basic")

	// Temporarily map the well-known env keys to our unique test keys.
	// Since NewGlobal reads fixed env var names, we use t.Setenv directly.
	t.Setenv("BITBUCKET_AUTH", "basic")
	t.Setenv("BITBUCKET_EMAIL", "user@example.com")
	t.Setenv("BITBUCKET_API_TOKEN", "my-api-token")
	config.ClearCache()

	cfg := config.NewGlobal()

	assert.Equal(t, util.BasicAuth, cfg.Auth.Type)
	assert.Equal(t, "user@example.com", cfg.Auth.Basic.Username)
	assert.Equal(t, "my-api-token", cfg.Auth.Basic.Password)
	_ = emailKey
	_ = tokenKey
	_ = authKey
}

func TestNewGlobal_OAuth(t *testing.T) {
	t.Cleanup(config.ClearCache)

	t.Setenv("BITBUCKET_AUTH", "oauth")
	t.Setenv("SERVER_URL", "http://localhost:8080")
	t.Setenv("OAUTH_ISSUER", "https://bitbucket.org")
	t.Setenv("OAUTH_SCOPES", "repository;pullrequest")
	t.Setenv("OAUTH_RESOURCE_METADATA_PATH", "/.well-known/oauth-protected-resource")
	config.ClearCache()

	cfg := config.NewGlobal()

	assert.Equal(t, util.OAuth, cfg.Auth.Type)
	assert.Equal(t, "http://localhost:8080", cfg.Auth.OAuth.ServerUrl)
	assert.Equal(t, "https://bitbucket.org", cfg.Auth.OAuth.Issuer)
	assert.Equal(t, []string{"repository", "pullrequest"}, cfg.Auth.OAuth.Scopes)
	assert.Equal(t, "/.well-known/oauth-protected-resource", cfg.Auth.OAuth.ResourceMetadataPath)
}

func TestNewGlobal_DefaultValues(t *testing.T) {
	t.Cleanup(config.ClearCache)

	t.Setenv("BITBUCKET_AUTH", "oauth")
	t.Setenv("SERVER_URL", "http://localhost:9090")
	// Unset optional env vars so defaults are used
	t.Setenv("SERVER_PORT", "")
	t.Setenv("BITBUCKET_URL", "")
	t.Setenv("BITBUCKET_TIMEOUT", "")
	t.Setenv("BITBUCKET_GIT_URL", "")
	t.Setenv("OAUTH_ISSUER", "")
	t.Setenv("OAUTH_SCOPES", "")
	t.Setenv("OAUTH_RESOURCE_METADATA_PATH", "")
	config.ClearCache()

	cfg := config.NewGlobal()

	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "https://api.bitbucket.org/2.0", cfg.BitbucketApi.Url)
	assert.Equal(t, 5, cfg.BitbucketApi.Timeout)
	assert.Equal(t, "https://bitbucket.org", cfg.BitbucketGit.BaseURL)
	assert.Equal(t, "https://bitbucket.org", cfg.Auth.OAuth.Issuer)
	assert.Equal(t, []string{"repository", "pullrequest"}, cfg.Auth.OAuth.Scopes)
	assert.Equal(t, "/.well-known/oauth-protected-resource", cfg.Auth.OAuth.ResourceMetadataPath)
}

func TestNewGlobal_CustomServerPort(t *testing.T) {
	t.Cleanup(config.ClearCache)

	t.Setenv("BITBUCKET_AUTH", "oauth")
	t.Setenv("SERVER_URL", "http://localhost:9090")
	t.Setenv("SERVER_PORT", "9090")
	config.ClearCache()

	cfg := config.NewGlobal()

	assert.Equal(t, 9090, cfg.Server.Port)
}

func TestNewGlobal_CustomBitbucketConfig(t *testing.T) {
	t.Cleanup(config.ClearCache)

	t.Setenv("BITBUCKET_AUTH", "oauth")
	t.Setenv("SERVER_URL", "http://localhost:9090")
	t.Setenv("BITBUCKET_URL", "https://api.bitbucket.example.com/2.0")
	t.Setenv("BITBUCKET_GIT_URL", "https://bitbucket.example.com")
	t.Setenv("BITBUCKET_TIMEOUT", "10")
	config.ClearCache()

	cfg := config.NewGlobal()

	assert.Equal(t, "https://api.bitbucket.example.com/2.0", cfg.BitbucketApi.Url)
	assert.Equal(t, "https://bitbucket.example.com", cfg.BitbucketGit.BaseURL)
	assert.Equal(t, 10, cfg.BitbucketApi.Timeout)
}

func TestNewGlobal_MissingBasicAuthCredentials_Panics(t *testing.T) {
	t.Cleanup(config.ClearCache)

	t.Setenv("BITBUCKET_AUTH", "basic")
	t.Setenv("BITBUCKET_EMAIL", "")
	t.Setenv("BITBUCKET_API_TOKEN", "")
	config.ClearCache()

	require.Panics(t, func() {
		config.NewGlobal()
	})
}

func TestNewGlobal_MissingOAuthServerURL_Panics(t *testing.T) {
	t.Cleanup(config.ClearCache)

	t.Setenv("BITBUCKET_AUTH", "oauth")
	t.Setenv("SERVER_URL", "")
	config.ClearCache()

	require.Panics(t, func() {
		config.NewGlobal()
	})
}

// setEnv sets a unique env key to the given value and registers cleanup.
// It returns the unique key used.
func randKey(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, rand.Int())
}
