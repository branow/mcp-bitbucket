package auth_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/branow/mcp-bitbucket/internal/auth"
	"github.com/stretchr/testify/require"
)

func TestNewOAuthMiddleware_Success(t *testing.T) {
	handler := newHandler(auth.OAuthConfig{
		ServerUrl:            "http://localhost:8080",
		Issuer:               "http://issuer.example.com",
		Scopes:               []string{"repository", "pullrequest"},
		ResourceMetadataPath: "/.well-known/oauth-resource-metadata",
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", "token123"))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, "OK", rr.Body.String())
}

func TestNewOAuthMiddleware_Failure(t *testing.T) {
	url := "http://localhost:8080"
	metadata := "/.well-known/oauth-resource-metadata"

	handler := newHandler(auth.OAuthConfig{
		ServerUrl:            url,
		Issuer:               "http://issuer.example.com",
		Scopes:               []string{"repository", "pullrequest"},
		ResourceMetadataPath: metadata,
	})

	tests := []struct {
		name     string
		token    string
		header   bool
		expError string
	}{
		{
			name:     "missing authorization header",
			token:    "",
			header:   false,
			expError: "no bearer token",
		},
		{
			name:     "empty token",
			token:    "Bearer ",
			header:   true,
			expError: "invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.header {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, http.StatusUnauthorized, rr.Code)

			wwwAuth := rr.Header().Get("WWW-Authenticate")
			require.NotEmpty(t, wwwAuth, "WWW-Authenticate header should be present")
			require.Contains(t, wwwAuth, "Bearer", "WWW-Authenticate header should contain Bearer scheme")
			require.Contains(t, wwwAuth, "resource_metadata="+url+metadata)
			require.Equal(t, tt.expError, strings.TrimSpace(rr.Body.String()))
		})
	}
}

func newHandler(cfg auth.OAuthConfig) http.Handler {
	middleware := auth.NewOAuthMiddleware(cfg)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	return middleware(testHandler)
}
