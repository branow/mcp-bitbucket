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

	RunClientTest(t, ClientTestCase[bitbucket.BitbucketApiResponse[bitbucket.Repository]]{
		MockDataFile: "testdata/repository_list_mock.json",
		Path:         fmt.Sprintf("/%s/%s", "repositories", namespace),
		CallClient: func(client *bitbucket.Client) (*bitbucket.BitbucketApiResponse[bitbucket.Repository], error) {
			return client.ListRepositories(namespace, pagelen, page)
		},
	})
}

func TestClient_GetRepository(t *testing.T) {
	t.Parallel()

	const namespace = "test_workspace"
	const repoSlug = "test-repo"

	RunClientTest(t, ClientTestCase[bitbucket.Repository]{
		MockDataFile: "testdata/repository_mock.json",
		Path:         fmt.Sprintf("/%s/%s/%s", "repositories", namespace, repoSlug),
		CallClient: func(client *bitbucket.Client) (*bitbucket.Repository, error) {
			return client.GetRepository(namespace, repoSlug)
		},
	})
}

func TestClient_GetRepositorySource(t *testing.T) {
	t.Parallel()

	const namespace = "test_workspace"
	const repoSlug = "test-repo"

	RunClientTest(t, ClientTestCase[bitbucket.BitbucketApiResponse[bitbucket.SourceItem]]{
		MockDataFile: "testdata/repository_src_mock.json",
		Path:         fmt.Sprintf("/%s/%s/%s/%s", "repositories", namespace, repoSlug, "src"),
		CallClient: func(client *bitbucket.Client) (*bitbucket.BitbucketApiResponse[bitbucket.SourceItem], error) {
			return client.GetRepositorySource(namespace, repoSlug)
		},
	})
}

type ClientTestCase[T any] struct {
	MockDataFile string
	Path         string
	CallClient   func(*bitbucket.Client) (*T, error)
}

func RunClientTest[T any](t *testing.T, tc ClientTestCase[T]) {
	t.Helper()

	mockData, err := os.ReadFile(tc.MockDataFile)
	require.NoError(t, err, "failed to read mock data file")

	var expectedResponse T
	err = json.Unmarshal(mockData, &expectedResponse)
	require.NoError(t, err, "failed to unmarshal expected response")

	config := bitbucket.Config{
		Username: "test_user",
		Password: "test_password",
		Timeout:  1,
	}
	config.BaseUrl = NewTestServer(t, tc.Path, func(resp http.ResponseWriter, req *http.Request) {
		actualUsername, actualPassword, ok := req.BasicAuth()
		require.True(t, ok, "expected basic auth")
		require.Equal(t, config.Username, actualUsername)
		require.Equal(t, config.Password, actualPassword)
		resp.Header().Set("Content-Type", "application/json")
		resp.Write(mockData)
	})

	client := bitbucket.NewClient(config)
	actualResponse, err := tc.CallClient(client)
	require.NoError(t, err)
	assert.Equal(t, &expectedResponse, actualResponse)
}

func NewTestServer(t *testing.T, pattern string, handle func(http.ResponseWriter, *http.Request)) string {
	t.Helper()
	handler := http.NewServeMux()
	handler.HandleFunc(pattern, handle)
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server.URL
}
