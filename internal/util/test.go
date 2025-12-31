package util

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertJsonRpcError checks if the error is a jsonrpc.Error with the expected code
func AssertJsonRpcError(t *testing.T, err error, expectedCode int64, msgAndArgs ...interface{}) {
	t.Helper()
	var jsonrpcErr *jsonrpc.Error
	require.ErrorAs(t, err, &jsonrpcErr, msgAndArgs...)
	assert.Equal(t, expectedCode, jsonrpcErr.Code, msgAndArgs...)
}

// ObtainAccessToken requests an OAuth access token using client credentials flow
func ObtainAccessToken(clientId, clientSecret, url string) (string, error) {
	if clientId == "" || clientSecret == "" {
		return "", fmt.Errorf("client ID or client secret not configured")
	}

	data := "grant_type=client_credentials"
	req, err := http.NewRequest("POST", url, strings.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientId, clientSecret)

	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to request token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	if tokenResponse.AccessToken == "" {
		return "", fmt.Errorf("access token is empty in response")
	}

	return tokenResponse.AccessToken, nil
}
