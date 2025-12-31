package util_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/branow/mcp-bitbucket/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicAuthorizer_Authorize(t *testing.T) {
	t.Parallel()

	authorizer := util.NewBasicAuthorizer("testuser", "testpass")
	req := httptest.NewRequest("GET", "http://example.com", nil)
	ctx := context.Background()

	err := authorizer.Authorize(ctx, req)
	require.NoError(t, err)

	username, password, ok := req.BasicAuth()
	assert.True(t, ok)
	assert.Equal(t, "testuser", username)
	assert.Equal(t, "testpass", password)
}

func TestBasicAuthorizer_Authorize_Failure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		username string
		password string
	}{
		{"empty username", "", "testpass"},
		{"empty password", "testuser", ""},
		{"empty credentials", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			authorizer := util.NewBasicAuthorizer(tt.username, tt.password)
			req := httptest.NewRequest("GET", "http://example.com", nil)
			ctx := context.Background()

			err := authorizer.Authorize(ctx, req)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "basic auth credentials not configured")
		})
	}
}

func TestOAuthAuthorizer_Authorize(t *testing.T) {
	t.Parallel()

	extractor := util.NewStaticTokenExtractor("test-token-123")
	authorizer := util.NewOAuthAuthorizer(extractor)
	req := httptest.NewRequest("GET", "http://example.com", nil)
	ctx := context.Background()

	err := authorizer.Authorize(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, "Bearer test-token-123", req.Header.Get("Authorization"))
}

func TestOAuthAuthorizer_Authorize_Failure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		extractor     util.TokenExtractor
		errorContains string
	}{
		{
			name:          "token extraction fails",
			extractor:     &mockTokenExtractor{err: assert.AnError},
			errorContains: "failed to extract OAuth token",
		},
		{
			name:          "token is empty",
			extractor:     &mockTokenExtractor{token: ""},
			errorContains: "OAuth token is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			authorizer := util.NewOAuthAuthorizer(tt.extractor)
			req := httptest.NewRequest("GET", "http://example.com", nil)
			ctx := context.Background()

			err := authorizer.Authorize(ctx, req)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorContains)
		})
	}
}

func TestStaticTokenExtractor_ExtractToken(t *testing.T) {
	t.Parallel()

	extractor := util.NewStaticTokenExtractor("my-static-token")
	ctx := context.Background()

	token, err := extractor.ExtractToken(ctx)
	require.NoError(t, err)
	assert.Equal(t, "my-static-token", token)
}

func TestStaticTokenExtractor_ExtractToken_Failure(t *testing.T) {
	t.Parallel()

	extractor := util.NewStaticTokenExtractor("")
	ctx := context.Background()

	token, err := extractor.ExtractToken(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "static token not configured")
	assert.Empty(t, token)
}

// Note: Success case for MCPTokenExtractor is not tested here because it requires
// the MCP auth middleware to properly set the token info in the context using its
// private tokenInfoKey. The success case will be covered in middleware integration tests.
func TestMCPTokenExtractor_ExtractToken_Failure(t *testing.T) {
	t.Parallel()

	extractor := util.NewMCPTokenExtractor()
	ctx := context.Background()

	token, err := extractor.ExtractToken(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no token info in context")
	assert.Empty(t, token)
}

func TestNoOpAuthorizer_Authorize(t *testing.T) {
	t.Parallel()

	authorizer := util.NewNoOpAuthorizer()
	req := httptest.NewRequest("GET", "http://example.com", nil)
	ctx := context.Background()

	err := authorizer.Authorize(ctx, req)
	require.NoError(t, err)

	// Verify no auth headers were added
	assert.Empty(t, req.Header.Get("Authorization"))
	_, _, ok := req.BasicAuth()
	assert.False(t, ok)
}

// mockTokenExtractor is a test helper that implements TokenExtractor
type mockTokenExtractor struct {
	token string
	err   error
}

func (m *mockTokenExtractor) ExtractToken(ctx context.Context) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.token, nil
}
