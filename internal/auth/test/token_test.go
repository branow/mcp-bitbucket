package auth_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/branow/mcp-bitbucket/internal/auth"
	authsdk "github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJwtVerifier_Verify_Success(t *testing.T) {
	token := "token123"
	scopes := []string{"repository", "pullrequest"}

	verifier := auth.NewOpaqueTokenVerifier(scopes)
	info, err := verifier.Verify(context.Background(), token, &http.Request{})

	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, token, info.Extra["raw_token"])
	assert.Equal(t, scopes, info.Scopes)
	assert.NotNil(t, scopes, info.Expiration)
}

func TestJwtVerifier_Verify_Failure(t *testing.T) {
	verifier := auth.NewOpaqueTokenVerifier([]string{"repository", "pullrequest"})
	info, err := verifier.Verify(context.Background(), "", &http.Request{})

	require.EqualError(t, err, authsdk.ErrInvalidToken.Error())
	require.Nil(t, info)
}
