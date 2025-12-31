// Package auth provides HTTP middleware for authentication.
package auth

import (
	"fmt"
	"net/http"

	"github.com/branow/mcp-bitbucket/internal/util"
	"github.com/modelcontextprotocol/go-sdk/auth"
)

// Middleware is a function that wraps an HTTP handler to add cross-cutting concerns.
// In this package, it's used to add authentication checks to HTTP endpoints.
type Middleware func(http.Handler) http.Handler

// NewMiddleware creates an authentication middleware based on the provided configuration.
// It selects the appropriate middleware implementation based on the authentication type:
//   - OAuth: Returns middleware that validates bearer tokens
//   - BasicAuth: Returns empty middleware (basic auth is handled at the API client level)
//   - Other types: Returns an error
//
// Parameters:
//   - cfg: AuthConfig specifying the authentication type and related configuration
//
// Returns a Middleware function or an error if the authentication type is not supported.
func NewMiddleware(cfg AuthConfig) (Middleware, error) {
	switch cfg.Type {
	case util.OAuth:
		return NewOAuthMiddleware(cfg.OAuth), nil
	case util.BasicAuth:
		return NewEmptyMiddleware(), nil
	default:
		return nil, fmt.Errorf("invalid authentication type: %v", cfg.Type)
	}
}

// NewEmptyMiddleware creates a no-op middleware that passes requests through unchanged.
// This is used when authentication is handled elsewhere (e.g., at the API client level)
// or when authentication is not required.
func NewEmptyMiddleware() Middleware {
	return func(h http.Handler) http.Handler { return h }
}

// NewOAuthMiddleware creates middleware that validates OAuth 2.0 bearer tokens.
// This middleware is designed for use with Bitbucket's opaque access tokens.
//
// Unlike JWT-based OAuth implementations, Bitbucket uses opaque tokens that cannot
// be validated locally. The middleware performs basic checks (token presence) and
// stores the token for forwarding to Bitbucket. Actual authorization is verified
// when the token is used to call Bitbucket APIs.
//
// Parameters:
//   - cfg: OAuthConfig containing server URL, issuer, required scopes, and metadata path
//
// Returns a Middleware that:
//   - Extracts bearer tokens from the Authorization header
//   - Performs basic validation (checks token is not empty)
//   - Stores the token info in the request context
//   - Returns 401 Unauthorized for missing/empty tokens
//   - Provides OAuth resource metadata for client authentication flows
//
// The middleware always succeeds for non-empty tokens. Token validity is determined
// by the Bitbucket API when requests are made.
func NewOAuthMiddleware(cfg OAuthConfig) Middleware {
	resourceMetadataUrl := fmt.Sprintf("%s%s", cfg.ServerUrl, cfg.ResourceMetadataPath)
	return auth.RequireBearerToken(NewOpaqueTokenVerifier(cfg.Scopes).Verify, &auth.RequireBearerTokenOptions{
		Scopes:              cfg.Scopes,
		ResourceMetadataURL: resourceMetadataUrl,
	})
}
