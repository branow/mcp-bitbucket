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

	const workspace = "test_workspace"
	const pagelen = 10
	const page = 1

	RunClientTest(t, ClientTestCase[bitbucket.BitbucketApiResponse[bitbucket.BitbucketRepository]]{
		MockDataFile: "testdata/repository_list_mock.json",
		Path:         fmt.Sprintf("/%s/%s", "repositories", workspace),
		Decode:       DecodeJson[bitbucket.BitbucketApiResponse[bitbucket.BitbucketRepository]],
		CallClient: func(client *bitbucket.Client) (*bitbucket.BitbucketApiResponse[bitbucket.BitbucketRepository], error) {
			return client.ListRepositories(workspace, pagelen, page)
		},
	})
}

func TestClient_GetRepository(t *testing.T) {
	t.Parallel()

	const workspace = "test_workspace"
	const repoSlug = "test-repo"

	RunClientTest(t, ClientTestCase[bitbucket.BitbucketRepository]{
		MockDataFile: "testdata/repository_mock.json",
		Path:         fmt.Sprintf("/%s/%s/%s", "repositories", workspace, repoSlug),
		Decode:       DecodeJson[bitbucket.BitbucketRepository],
		CallClient: func(client *bitbucket.Client) (*bitbucket.BitbucketRepository, error) {
			return client.GetRepository(workspace, repoSlug)
		},
	})
}

func TestClient_GetRepositorySource(t *testing.T) {
	t.Parallel()

	const workspace = "test_workspace"
	const repoSlug = "test-repo"

	RunClientTest(t, ClientTestCase[bitbucket.BitbucketApiResponse[bitbucket.BitbucketSourceItem]]{
		MockDataFile: "testdata/repository_src_mock.json",
		Path:         fmt.Sprintf("/%s/%s/%s/%s", "repositories", workspace, repoSlug, "src"),
		Decode:       DecodeJson[bitbucket.BitbucketApiResponse[bitbucket.BitbucketSourceItem]],
		CallClient: func(client *bitbucket.Client) (*bitbucket.BitbucketApiResponse[bitbucket.BitbucketSourceItem], error) {
			return client.GetRepositorySource(workspace, repoSlug)
		},
	})
}

func TestClient_ListPullRequests(t *testing.T) {
	t.Parallel()

	const workspace = "test_workspace"
	const repoSlug = "test-repo"
	const pagelen = 10
	const page = 1

	RunClientTest(t, ClientTestCase[bitbucket.BitbucketApiResponse[bitbucket.BitbucketPullRequest]]{
		MockDataFile: "testdata/pull_requests_mock.json",
		Path:         fmt.Sprintf("/%s/%s/%s/%s", "repositories", workspace, repoSlug, "pullrequests"),
		Decode:       DecodeJson[bitbucket.BitbucketApiResponse[bitbucket.BitbucketPullRequest]],
		CallClient: func(client *bitbucket.Client) (*bitbucket.BitbucketApiResponse[bitbucket.BitbucketPullRequest], error) {
			return client.ListPullRequests(workspace, repoSlug, pagelen, page, nil)
		},
	})
}

func TestClient_GetPullRequest(t *testing.T) {
	t.Parallel()

	const workspace = "test_workspace"
	const repoSlug = "test-repo"
	const pullRequestId = 1

	RunClientTest(t, ClientTestCase[bitbucket.BitbucketPullRequest]{
		MockDataFile: "testdata/pull_request_mock.json",
		Path:         fmt.Sprintf("/%s/%s/%s/%s/%d", "repositories", workspace, repoSlug, "pullrequests", pullRequestId),
		Decode:       DecodeJson[bitbucket.BitbucketPullRequest],
		CallClient: func(client *bitbucket.Client) (*bitbucket.BitbucketPullRequest, error) {
			return client.GetPullRequest(workspace, repoSlug, pullRequestId)
		},
	})
}

func TestClient_ListPullRequestCommits(t *testing.T) {
	t.Parallel()

	const workspace = "test_workspace"
	const repoSlug = "test-repo"
	const pullRequestId = 1

	RunClientTest(t, ClientTestCase[bitbucket.BitbucketApiResponse[bitbucket.BitbucketCommit]]{
		MockDataFile: "testdata/pull_request_commits_mock.json",
		Path:         fmt.Sprintf("/%s/%s/%s/%s/%d/%s", "repositories", workspace, repoSlug, "pullrequests", pullRequestId, "commits"),
		Decode:       DecodeJson[bitbucket.BitbucketApiResponse[bitbucket.BitbucketCommit]],
		CallClient: func(client *bitbucket.Client) (*bitbucket.BitbucketApiResponse[bitbucket.BitbucketCommit], error) {
			return client.ListPullRequestCommits(workspace, repoSlug, pullRequestId)
		},
	})
}

func TestClient_ListPullRequestComments(t *testing.T) {
	t.Parallel()

	const workspace = "test_workspace"
	const repoSlug = "test-repo"
	const pullRequestId = 1
	const pagelen = 10
	const page = 1

	RunClientTest(t, ClientTestCase[bitbucket.BitbucketApiResponse[bitbucket.BitbucketPullRequestComment]]{
		MockDataFile: "testdata/pull_request_comments_mock.json",
		Path:         fmt.Sprintf("/%s/%s/%s/%s/%d/%s", "repositories", workspace, repoSlug, "pullrequests", pullRequestId, "comments"),
		Decode:       DecodeJson[bitbucket.BitbucketApiResponse[bitbucket.BitbucketPullRequestComment]],
		CallClient: func(client *bitbucket.Client) (*bitbucket.BitbucketApiResponse[bitbucket.BitbucketPullRequestComment], error) {
			return client.ListPullRequestComments(workspace, repoSlug, pullRequestId, pagelen, page)
		},
	})
}

func TestClient_GetPullRequestDiff(t *testing.T) {
	t.Parallel()

	const workspace = "test_workspace"
	const repoSlug = "test-repo"
	const pullRequestId = 1

	RunClientTest(t, ClientTestCase[string]{
		MockDataFile: "testdata/pull_request_diff_mock.txt",
		Path:         fmt.Sprintf("/%s/%s/%s/%s/%d/%s", "repositories", workspace, repoSlug, "pullrequests", pullRequestId, "diff"),
		Decode:       DecodeText,
		CallClient: func(client *bitbucket.Client) (*string, error) {
			return client.GetPullRequestDiff(workspace, repoSlug, pullRequestId)
		},
	})
}

func TestClient_GetFileSource(t *testing.T) {
	t.Parallel()

	const workspace = "test_workspace"
	const repoSlug = "test-repo"
	const commit = "54ad501s"
	const path = "src/test-path/test-file.ext"

	RunClientTest(t, ClientTestCase[string]{
		MockDataFile: "testdata/file_source_mock.txt",
		Path:         fmt.Sprintf("/%s/%s/%s/%s/%s/%s", "repositories", workspace, repoSlug, "src", commit, path),
		Decode:       DecodeText,
		CallClient: func(client *bitbucket.Client) (*string, error) {
			return client.GetFileSource(workspace, repoSlug, commit, path)
		},
	})
}

func DecodeJson[T any](data []byte, res *T) error {
	return json.Unmarshal(data, res)
}

func DecodeText(data []byte, res *string) error {
	*res = string(data)
	return nil
}

type ClientTestCase[T any] struct {
	MockDataFile string
	Path         string
	CallClient   func(*bitbucket.Client) (*T, error)
	Decode       func(data []byte, res *T) error
}

func RunClientTest[T any](t *testing.T, tc ClientTestCase[T]) {
	t.Helper()

	mockData, err := os.ReadFile(tc.MockDataFile)
	require.NoError(t, err, "failed to read mock data file")

	var expectedResponse T
	err = tc.Decode(mockData, &expectedResponse)
	require.NoError(t, err, "failed to decode expected response")

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
