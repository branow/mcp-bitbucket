// Package auth provides opaque token verification for OAuth 2.0 authentication with Bitbucket.
package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/auth"
)

// OpaqueTokenVerifier validates opaque OAuth 2.0 access tokens.
// Unlike JWT tokens, opaque tokens do not contain embedded claims or signatures
// that can be verified locally. This verifier performs basic validation (checking
// that the token is not empty) and stores the token for forwarding to Bitbucket.
//
// Bitbucket uses opaque access tokens rather than JWTs, so actual token validation
// is performed by Bitbucket's API when the token is used. This verifier's role is
// to extract the token from requests and make it available to the API client.
type OpaqueTokenVerifier struct {
	// scopes are the required OAuth scopes configured for this resource.
	// These are stored in the TokenInfo for scope checking by the middleware.
	scopes []string
}

// NewOpaqueTokenVerifier creates a new opaque token verifier.
//
// Parameters:
//   - scopes: The OAuth scopes required for accessing this resource
//
// Returns a verifier that performs basic token validation and stores tokens
// for forwarding to the Bitbucket API.
func NewOpaqueTokenVerifier(scopes []string) *OpaqueTokenVerifier {
	return &OpaqueTokenVerifier{scopes: scopes}
}

// Verify performs basic validation on an opaque access token.
//
// Since opaque tokens cannot be validated without calling the token introspection
// endpoint (which Bitbucket does not expose), this method only checks that the
// token is present and non-empty. The actual authorization is performed when
// the token is used to call Bitbucket APIs.
//
// Parameters:
//   - ctx: Request context (currently unused but required by interface)
//   - tokenStr: The opaque access token string to verify
//   - req: The HTTP request (currently unused but required by interface)
//
// Returns TokenInfo containing:
//   - A placeholder expiration time (5 minutes from now)
//   - The configured scopes (for middleware scope checking)
//   - The raw token stored in Extra["raw_token"] for forwarding to Bitbucket
//
// Returns auth.ErrInvalidToken if the token is empty or missing.
func (v *OpaqueTokenVerifier) Verify(ctx context.Context, tokenStr string, _ *http.Request) (*auth.TokenInfo, error) {
	if tokenStr == "" || tokenStr == "Bearer" {
		return nil, auth.ErrInvalidToken
	}

	return &auth.TokenInfo{
		Expiration: time.Now().Add(time.Minute * 5), // Placeholder expiration
		Scopes:     v.scopes,                        // Required scopes
		Extra:      map[string]any{"raw_token": tokenStr},
	}, nil
}
