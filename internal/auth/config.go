// Package auth provides authentication configuration and setup for the MCP server.
package auth

import "github.com/branow/mcp-bitbucket/internal/util"

// AuthConfig contains the authentication configuration for the MCP server.
// It supports multiple authentication types (Basic, OAuth) and delegates
// to the appropriate sub-configuration based on the Type field.
type AuthConfig struct {
	// Type specifies which authentication method to use
	Type util.AuthType
	// OAuth contains OAuth 2.0 configuration (used when Type is OAuth)
	OAuth OAuthConfig
	// Basic contains basic authentication configuration (used when Type is BasicAuth)
	Basic BasicConfig
}

// Authorizer creates an Authorizer instance based on the configured authentication type.
// Returns a NoOpAuthorizer if the authentication type is not recognized.
func (c AuthConfig) Authorizer() util.Authorizer {
	switch c.Type {
	case util.BasicAuth:
		return c.Basic.Authorizer()
	case util.OAuth:
		return c.OAuth.Authorizer()
	default:
		return util.NewNoOpAuthorizer()
	}
}

// BasicConfig contains configuration for HTTP basic authentication.
type BasicConfig struct {
	Username string
	Password string
}

// Authorizer creates a BasicAuthorizer using the configured credentials.
func (c BasicConfig) Authorizer() util.Authorizer {
	return util.NewBasicAuthorizer(c.Username, c.Password)
}

// OAuthConfig contains configuration for OAuth 2.0 authentication with Bitbucket.
// Bitbucket uses opaque OAuth 2.0 access tokens (not JWT tokens), which are obtained
// through the OAuth 2.0 authorization flow and validated by Bitbucket's API.
type OAuthConfig struct {
	// ServerUrl is the base URL of this MCP server (e.g., "http://localhost:8080")
	ServerUrl string
	// Issuer is the OAuth authorization server URL (e.g., "https://bitbucket.org")
	// This is used in the OAuth resource metadata for client authentication flows
	Issuer string
	// Scopes are the Bitbucket OAuth scopes required for API access
	// Common scopes: "repository", "pullrequest", "account"
	Scopes []string
	// ResourceMetadataPath is the path where OAuth resource metadata is served
	// (e.g., "/oauth/metadata"). This endpoint provides OAuth discovery information
	ResourceMetadataPath string
}

// Authorizer creates an OAuthAuthorizer using an MCP token extractor.
// The token extractor retrieves opaque access tokens from the MCP context that were
// extracted by the OAuth middleware. These tokens are then forwarded in API requests
// to Bitbucket for validation and authorization.
func (c OAuthConfig) Authorizer() util.Authorizer {
	return util.NewOAuthAuthorizer(util.NewMCPTokenExtractor())
}
