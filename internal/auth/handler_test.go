package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/branow/mcp-bitbucket/internal/auth"
	"github.com/stretchr/testify/assert"
)

func TestNewOAuthHandler(t *testing.T) {
	t.Parallel()

	cfg := auth.OAuthConfig{
		ServerUrl:            "http://example.com",
		Issuer:               "https://auth.example.com",
		Scopes:               []string{"repository", "pullrequest"},
		ResourceMetadataPath: "/oauth/metadata",
	}

	handler := auth.NewOAuthHandler(cfg)
	assert.NotNil(t, handler)

	req := httptest.NewRequest("GET", "http://example.com/oauth/metadata", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	expected := `{
    "resource": "http://example.com/mcp",
    "authorization_servers": ["https://auth.example.com/site/oauth2/access_token"],
    "scopes_supported": ["repository", "pullrequest"]
  }`

	assert.JSONEq(t, expected, rec.Body.String())
}
