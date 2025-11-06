package bitbucket_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/branow/mcp-bitbucket/internal/bitbucket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_ListRepositories(t *testing.T) {
	t.Parallel()

	const namespace = "test_workspace"
	const pagelen = 10
	const page = 1

	mockData, err := os.ReadFile("testdata/repository_list_mock.json")
	require.NoError(t, err, "failed to read mock data file")

	var expectedResponse bitbucket.BitBucketResponse[bitbucket.Repository]
	err = json.Unmarshal(mockData, &expectedResponse)
	require.NoError(t, err, "failed to unmarshal expected response")

	config := NewTestConfig()
	path := fmt.Sprintf("/%s/%s", "repositories", namespace)
	config.BaseUrl = NewTestServer(t, path, func(resp http.ResponseWriter, req *http.Request) {
		actualUsername, actualPassword, ok := req.BasicAuth()
		require.True(t, ok, "expected basic auth")
		require.Equal(t, config.Username, actualUsername)
		require.Equal(t, config.Password, actualPassword)
		resp.Header().Set("Content-Type", "application/json")
		resp.Write(mockData)
	})

	client := bitbucket.NewClient(config)
	actualResponse, err := client.ListRepositories(namespace, pagelen, page)
	require.NoError(t, err)
	assert.Equal(t, &expectedResponse, actualResponse)
}

func NewTestConfig() bitbucket.Config {
	return bitbucket.Config{
		Username: "test_user",
		Password: "test_password",
		Timeout:  1,
	}
}

func NewTestServer(t *testing.T, pattern string, handle func(http.ResponseWriter, *http.Request)) string {
	t.Helper()
	handler := http.NewServeMux()
	handler.HandleFunc(pattern, handle)
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server.URL
}
