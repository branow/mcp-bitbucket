// Package auth provides HTTP handlers for OAuth 2.0 authentication.
package auth

import (
	"fmt"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/oauthex"
)

// NewOAuthHandler creates an HTTP handler that serves OAuth 2.0 protected resource metadata.
// This handler responds to requests for OAuth configuration information, telling clients
// where to find Bitbucket's authorization server and what scopes are required.
//
// The handler serves metadata according to RFC 8414 (OAuth 2.0 Authorization Server Metadata)
// and provides information needed by OAuth clients to authenticate with this MCP server
// before accessing Bitbucket resources.
//
// Parameters:
//   - cfg: OAuthConfig containing server URL, issuer (Bitbucket), and required scopes
//
// Returns an http.HandlerFunc that serves the protected resource metadata.
//
// The metadata includes:
//   - Resource URL: The MCP endpoint that requires OAuth authentication
//   - Authorization servers: OAuth token endpoints clients should use to obtain access tokens
//   - Supported scopes: Bitbucket OAuth scopes required (e.g., "repository", "pullrequest")
//
// Note: Bitbucket uses opaque access tokens, not JWT tokens, so the authorization server
// validates tokens through Bitbucket's API rather than local JWT verification.
func NewOAuthHandler(cfg OAuthConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metadata := &oauthex.ProtectedResourceMetadata{
			Resource: fmt.Sprintf("%s/mcp", cfg.ServerUrl),
			AuthorizationServers: []string{
				fmt.Sprintf("%s/site/oauth2/access_token", cfg.Issuer),
			},
			ScopesSupported: cfg.Scopes,
		}
		auth.ProtectedResourceMetadataHandler(metadata).ServeHTTP(w, r)
	}
}
