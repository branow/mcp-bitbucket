// Package config provides configuration management for the MCP Bitbucket server.
// It loads configuration from environment variables with support for fallback values.
package config

import (
	"github.com/branow/mcp-bitbucket/internal/auth"
	"github.com/branow/mcp-bitbucket/internal/bitbucket/client"
	"github.com/branow/mcp-bitbucket/internal/util"
)

// Global contains the complete configuration for the MCP server.
// It aggregates server, authentication, and Bitbucket client configurations.
type Global struct {
	// Server contains HTTP server configuration
	Server ServerConfig
	// Auth contains authentication configuration (OAuth or Basic)
	Auth auth.AuthConfig
	// Bitbucket contains Bitbucket API client configuration
	Bitbucket client.BitbucketConfig
}

// ServerConfig contains HTTP server configuration.
type ServerConfig struct {
	// Port is the HTTP server port (default: 8080)
	Port int
}

// NewGlobal creates a new Global configuration by loading values from environment variables.
// It reads configuration for the server, Bitbucket client, and authentication from environment.
//
// Environment variables:
//
// Server configuration:
//   - SERVER_PORT: HTTP server port (default: 8080)
//
// Bitbucket configuration:
//   - BITBUCKET_URL: Bitbucket API base URL (default: "https://api.bitbucket.org/2.0")
//   - BITBUCKET_TIMEOUT: HTTP request timeout in seconds (default: 5)
//
// Authentication configuration:
//   - BITBUCKET_AUTH: Authentication type - "basic" or "oauth" (default: "oauth")
//
// For basic authentication:
//   - BITBUCKET_EMAIL: Username/email for basic auth
//   - BITBUCKET_API_TOKEN: Password/API token for basic auth
//
// For OAuth authentication:
//   - SERVER_URL: Base URL of this MCP server
//   - OAUTH_ISSUER: OAuth issuer URL (e.g., "https://auth.example.com")
//   - OAUTH_SCOPES: Required OAuth scopes, semicolon-separated (default: "repository", "pullrequest")
//   - OAUTH_RESOURCE_METADATA_PATH: Path for OAuth metadata endpoint (default: "/.well-known/oauth-protected-resource")
//
// Returns a fully initialized Global configuration with all settings loaded.
func NewGlobal() Global {
	cfg := Global{
		Server: ServerConfig{
			Port: GetInt("SERVER_PORT", 8080),
		},
		Bitbucket: client.BitbucketConfig{
			Url:     GetString("BITBUCKET_URL", "https://api.bitbucket.org/2.0"),
			Timeout: GetInt("BITBUCKET_TIMEOUT", 5),
		},
		Auth: auth.AuthConfig{
			Type: util.AuthType(GetString("BITBUCKET_AUTH", "oauth")),
		},
	}

	switch cfg.Auth.Type {
	case util.BasicAuth:
		cfg.Auth.Basic = auth.BasicConfig{
			Username: GetString("BITBUCKET_EMAIL", ""),
			Password: GetString("BITBUCKET_API_TOKEN", ""),
		}
	case util.OAuth:
		cfg.Auth.OAuth = auth.OAuthConfig{
			ServerUrl:            GetString("SERVER_URL", ""),
			Issuer:               GetString("OAUTH_ISSUER", ""),
			Scopes:               GetList("OAUTH_SCOPES", ";", []string{"repository", "pullrequest"}),
			ResourceMetadataPath: GetString("OAUTH_RESOURCE_METADATA_PATH", "/.well-known/oauth-protected-resource"),
		}
	}

	return cfg
}
