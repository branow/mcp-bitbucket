package auth_test

import (
	"testing"

	"github.com/branow/mcp-bitbucket/internal/auth"
	"github.com/branow/mcp-bitbucket/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestAuthConfig_Authorizer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		authType   util.AuthType
		expectNoOp bool
	}{
		{
			name:       "basic auth",
			authType:   util.BasicAuth,
			expectNoOp: false,
		},
		{
			name:       "oauth",
			authType:   util.OAuth,
			expectNoOp: false,
		},
		{
			name:       "unknown type returns NoOpAuthorizer",
			authType:   util.AuthType("unknown"),
			expectNoOp: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := auth.AuthConfig{
				Type: tt.authType,
				Basic: auth.BasicConfig{
					Username: "user",
					Password: "pass",
				},
				OAuth: auth.OAuthConfig{
					ServerUrl: "http://example.com",
					Issuer:    "https://auth.example.com",
					Scopes:    []string{"repository"},
				},
			}

			authorizer := cfg.Authorizer()
			assert.NotNil(t, authorizer)

			// NoOpAuthorizer is returned for unknown types
			if tt.expectNoOp {
				_, isNoOp := authorizer.(*util.NoOpAuthorizer)
				assert.True(t, isNoOp, "expected NoOpAuthorizer for unknown auth type")
			}
		})
	}
}
