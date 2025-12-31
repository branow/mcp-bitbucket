// Package util provides authentication utilities for the Bitbucket MCP server.
package util

import (
	"context"
	"fmt"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/auth"
)

// AuthType represents the authentication method used for Bitbucket API requests.
type AuthType string

const (
	// BasicAuth represents HTTP basic authentication using username and password.
	BasicAuth AuthType = "basic"
	// OAuth represents OAuth 2.0 authentication with bearer tokens.
	OAuth AuthType = "oauth"
)

// Authorizer adds authentication credentials to HTTP requests.
// Implementations handle different authentication methods such as basic auth or OAuth.
type Authorizer interface {
	// Authorize adds authentication information to the provided HTTP request.
	// The context may contain authentication state or token information.
	Authorize(ctx context.Context, req *http.Request) error
}

// BasicAuthorizer implements HTTP basic authentication for API requests.
// It uses a username (email) and password (API token) for authentication.
type BasicAuthorizer struct {
	username string
	password string
}

// NewBasicAuthorizer creates a new BasicAuthorizer with the provided credentials.
//
// Parameters:
//   - username: The username (typically an email address)
//   - password: The password or API token
//
// Returns an Authorizer that adds HTTP basic authentication to requests.
func NewBasicAuthorizer(username, password string) Authorizer {
	return &BasicAuthorizer{
		username: username,
		password: password,
	}
}

// Authorize adds HTTP basic authentication credentials to the request.
// Returns an error if the credentials are not configured.
func (a *BasicAuthorizer) Authorize(ctx context.Context, req *http.Request) error {
	if a.username == "" || a.password == "" {
		return fmt.Errorf("basic auth credentials not configured")
	}
	req.SetBasicAuth(a.username, a.password)
	return nil
}

// TokenExtractor retrieves OAuth access tokens from a context.
// Different implementations can extract tokens from various sources
// (e.g., MCP context, static configuration, or external token services).
type TokenExtractor interface {
	// ExtractToken retrieves an OAuth access token from the given context.
	// Returns an error if the token cannot be retrieved or is invalid.
	ExtractToken(ctx context.Context) (string, error)
}

// OAuthAuthorizer implements OAuth 2.0 bearer token authentication.
// It uses a TokenExtractor to obtain access tokens and adds them to requests.
type OAuthAuthorizer struct {
	extractor TokenExtractor
}

// NewOAuthAuthorizer creates a new OAuthAuthorizer with the provided token extractor.
//
// Parameters:
//   - extractor: TokenExtractor that retrieves OAuth access tokens
//
// Returns an Authorizer that adds bearer token authentication to requests.
func NewOAuthAuthorizer(extractor TokenExtractor) Authorizer {
	return &OAuthAuthorizer{
		extractor: extractor,
	}
}

// Authorize adds an OAuth bearer token to the request's Authorization header.
// The token is retrieved from the context using the configured TokenExtractor.
// Returns an error if token extraction fails or the token is empty.
func (a *OAuthAuthorizer) Authorize(ctx context.Context, req *http.Request) error {
	token, err := a.extractor.ExtractToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to extract OAuth token: %w", err)
	}

	if token == "" {
		return fmt.Errorf("OAuth token is empty")
	}

	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

// MCPTokenExtractor extracts OAuth tokens from MCP (Model Context Protocol) context.
// It retrieves tokens that were validated and stored by the MCP authentication middleware.
type MCPTokenExtractor struct{}

// NewMCPTokenExtractor creates a new MCPTokenExtractor.
// This extractor is used when the MCP server handles OAuth authentication
// and stores validated tokens in the request context.
func NewMCPTokenExtractor() TokenExtractor {
	return &MCPTokenExtractor{}
}

// ExtractToken retrieves the raw OAuth token from the MCP authentication context.
// The token must have been previously validated by MCP's auth middleware and
// stored in the context's TokenInfo.Extra["raw_token"] field.
//
// Returns an error if:
//   - No token info is present in the context
//   - The raw_token field is missing from token info
//   - The raw_token is not a string
//   - The token is empty
func (e *MCPTokenExtractor) ExtractToken(ctx context.Context) (string, error) {
	tokenInfo := auth.TokenInfoFromContext(ctx)
	if tokenInfo == nil {
		return "", fmt.Errorf("no token info in context")
	}

	raw, ok := tokenInfo.Extra["raw_token"]
	if !ok {
		return "", fmt.Errorf("raw_token not found in token info")
	}

	token, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("raw_token is not a string: %T", raw)
	}

	if token == "" {
		return "", fmt.Errorf("token is empty")
	}

	return token, nil
}

// StaticTokenExtractor provides a static OAuth token configured at initialization.
// This is primarily used for testing or when tokens are managed externally.
type StaticTokenExtractor struct {
	token string
}

// NewStaticTokenExtractor creates a new StaticTokenExtractor with a pre-configured token.
//
// Parameters:
//   - token: The static OAuth access token to use for all requests
//
// This is useful for testing or scenarios where the token is managed outside
// the normal OAuth flow.
func NewStaticTokenExtractor(token string) TokenExtractor {
	return &StaticTokenExtractor{token: token}
}

// ExtractToken returns the pre-configured static token.
// Returns an error if no token was configured.
func (e *StaticTokenExtractor) ExtractToken(ctx context.Context) (string, error) {
	if e.token == "" {
		return "", fmt.Errorf("static token not configured")
	}
	return e.token, nil
}

// NoOpAuthorizer is a no-operation authorizer that adds no authentication to requests.
// This is used when authentication is disabled or not required.
type NoOpAuthorizer struct{}

// NewNoOpAuthorizer creates a new NoOpAuthorizer.
// Use this when you need an Authorizer interface implementation
// but don't want to add any authentication to requests.
func NewNoOpAuthorizer() Authorizer {
	return &NoOpAuthorizer{}
}

// Authorize is a no-op that returns without modifying the request.
// Always returns nil (no error).
func (a *NoOpAuthorizer) Authorize(ctx context.Context, req *http.Request) error {
	return nil
}
