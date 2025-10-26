package bitbucket_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/branow/mcp-bitbucket/internal/bitbucket"
	"github.com/branow/mcp-bitbucket/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_ListRepositories(t *testing.T) {
	t.Parallel()

	const expectedResponse = "list of repositories"

	config := NewTestConfig()
	path := util.JoinUrlPath("repositories", config.Namespace)
	config.BaseUrl = NewTestServer(t, path, func(resp http.ResponseWriter, req *http.Request) {
		actualUsername, actualPassword, ok := req.BasicAuth()
		require.True(t, ok, "expected basic auth")
		require.Equal(t, config.Username, actualUsername)
		require.Equal(t, config.Password, actualPassword)
		resp.Write([]byte(expectedResponse))
	})

	client := bitbucket.NewClient(config)
	actualResponse, err := client.ListRepositories()
	require.NoError(t, err)
	assert.Equal(t, expectedResponse, actualResponse)
}

func NewTestConfig() bitbucket.Config {
	return bitbucket.Config{
		Username:  "test_user",
		Password:  "test_password",
		Namespace: "test_namespace",
		Timeout:   1,
	}
}

func NewTestServer(t *testing.T, attern string, handle func(http.ResponseWriter, *http.Request)) string {
	t.Helper()
	handler := http.NewServeMux()
	handler.HandleFunc("/", handle)
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server.URL
}
