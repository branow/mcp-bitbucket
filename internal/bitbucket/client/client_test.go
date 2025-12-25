package client_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/branow/mcp-bitbucket/internal/bitbucket/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testConfig struct {
	url      string
	email    string
	apiToken string
	timeout  int
}

func (c *testConfig) BitbucketUrl() string      { return c.url }
func (c *testConfig) BitbucketEmail() string    { return c.email }
func (c *testConfig) BitbucketApiToken() string { return c.apiToken }
func (c *testConfig) BitbucketTimeout() int     { return c.timeout }

func TestClient_ListRepositories(t *testing.T) {
	t.Parallel()
	workspace, pagelen, page := "test_workspace", 10, 1

	tests := []ClientEndpointTestCase{
		{
			Name:   "Success",
			Status: 200,
			File:   "testdata/repository_list_mock.json",
		},
		{
			Name:   "Bad Request",
			Status: 400,
			File:   "testdata/repository_list_mock_400.json",
			Error:  ClientError("Invalid page"),
		},
		{
			Name:   "Unauthorized",
			Status: 401,
			File:   "testdata/repository_list_mock_401.json",
			Error:  ClientError("401"),
		},
		{
			Name:   "Not Found",
			Status: 404,
			File:   "testdata/repository_list_mock_404.json",
			Error:  ClientError("No workspace with identifier 'test_workspace'."),
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			RunClientTest(t, ClientTestCase[client.ApiResponse[client.Repository]]{
				Status:       tt.Status,
				MockDataFile: tt.File,
				Error:        tt.Error,
				Path:         fmt.Sprintf("/%s/%s", "repositories", workspace),
				Decode:       DecodeJson[client.ApiResponse[client.Repository]],
				CallClient: func(bb *client.Client) (*client.ApiResponse[client.Repository], error) {
					return bb.ListRepositories(workspace, pagelen, page)
				},
			})
		})
	}
}

func TestClient_GetRepository(t *testing.T) {
	t.Parallel()
	workspace, repoSlug := "test_workspace", "test-repo"

	tests := []ClientEndpointTestCase{
		{
			Name:   "Success",
			Status: 200,
			File:   "testdata/repository_mock.json",
		},
		{
			Name:   "Not Found",
			Status: 404,
			File:   "testdata/repository_mock_404.json",
			Error:  ClientError("You may not have access to this repository or it no longer exists in this workspace. If you think this repository exists and you have access, make sure you are authenticated."),
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			RunClientTest(t, ClientTestCase[client.Repository]{
				Status:       tt.Status,
				MockDataFile: tt.File,
				Error:        tt.Error,
				Path:         fmt.Sprintf("/%s/%s/%s", "repositories", workspace, repoSlug),
				Decode:       DecodeJson[client.Repository],
				CallClient: func(bb *client.Client) (*client.Repository, error) {
					return bb.GetRepository(workspace, repoSlug)
				},
			})
		})
	}
}

func TestClient_GetRepositorySource(t *testing.T) {
	t.Parallel()
	workspace, repoSlug := "test_workspace", "test-repo"

	tests := []ClientEndpointTestCase{
		{
			Name:   "Success",
			Status: 200,
			File:   "testdata/repository_src_mock.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			RunClientTest(t, ClientTestCase[client.ApiResponse[client.SourceItem]]{
				Status:       tt.Status,
				MockDataFile: tt.File,
				Error:        tt.Error,
				Path:         fmt.Sprintf("/%s/%s/%s/%s", "repositories", workspace, repoSlug, "src"),
				Decode:       DecodeJson[client.ApiResponse[client.SourceItem]],
				CallClient: func(bb *client.Client) (*client.ApiResponse[client.SourceItem], error) {
					return bb.GetRepositorySource(workspace, repoSlug)
				},
			})
		})
	}
}

func TestClient_ListPullRequests(t *testing.T) {
	t.Parallel()
	workspace, repoSlug, pagelen, page := "test_workspace", "test-repo", 10, 1

	tests := []ClientEndpointTestCase{
		{
			Name:   "Success",
			Status: 200,
			File:   "testdata/pull_requests_mock.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			RunClientTest(t, ClientTestCase[client.ApiResponse[client.PullRequest]]{
				Status:       tt.Status,
				MockDataFile: tt.File,
				Error:        tt.Error,
				Path:         fmt.Sprintf("/%s/%s/%s/%s", "repositories", workspace, repoSlug, "pullrequests"),
				Decode:       DecodeJson[client.ApiResponse[client.PullRequest]],
				CallClient: func(bb *client.Client) (*client.ApiResponse[client.PullRequest], error) {
					return bb.ListPullRequests(workspace, repoSlug, pagelen, page, nil)
				},
			})
		})
	}
}

func TestClient_GetPullRequest(t *testing.T) {
	t.Parallel()
	workspace, repoSlug, pullRequestId := "test_workspace", "test-repo", 1

	tests := []ClientEndpointTestCase{
		{
			Name:   "Success",
			Status: 200,
			File:   "testdata/pull_request_mock.json",
		},
		{
			Name:   "Not Found",
			Status: 404,
			File:   "testdata/pull_request_mock_404.txt",
			Error:  ClientError("404"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			RunClientTest(t, ClientTestCase[client.PullRequest]{
				Status:       tt.Status,
				MockDataFile: tt.File,
				Error:        tt.Error,
				Path:         fmt.Sprintf("/%s/%s/%s/%s/%d", "repositories", workspace, repoSlug, "pullrequests", pullRequestId),
				Decode:       DecodeJson[client.PullRequest],
				CallClient: func(bb *client.Client) (*client.PullRequest, error) {
					return bb.GetPullRequest(workspace, repoSlug, pullRequestId)
				},
			})
		})
	}
}

func TestClient_ListPullRequestCommits(t *testing.T) {
	t.Parallel()
	workspace, repoSlug, pullRequestId := "test_workspace", "test-repo", 1

	tests := []ClientEndpointTestCase{
		{
			Name:   "Success",
			Status: 200,
			File:   "testdata/pull_request_commits_mock.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			RunClientTest(t, ClientTestCase[client.ApiResponse[client.Commit]]{
				Status:       tt.Status,
				MockDataFile: tt.File,
				Error:        tt.Error,
				Path:         fmt.Sprintf("/%s/%s/%s/%s/%d/%s", "repositories", workspace, repoSlug, "pullrequests", pullRequestId, "commits"),
				Decode:       DecodeJson[client.ApiResponse[client.Commit]],
				CallClient: func(bb *client.Client) (*client.ApiResponse[client.Commit], error) {
					return bb.ListPullRequestCommits(workspace, repoSlug, pullRequestId)
				},
			})
		})
	}
}

func TestClient_ListPullRequestComments(t *testing.T) {
	t.Parallel()
	workspace, repoSlug, pullRequestId, pagelen, page := "test_workspace", "test-repo", 1, 10, 1

	tests := []ClientEndpointTestCase{
		{
			Name:   "Success",
			Status: 200,
			File:   "testdata/pull_request_comments_mock.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			RunClientTest(t, ClientTestCase[client.ApiResponse[client.PullRequestComment]]{
				Status:       tt.Status,
				MockDataFile: tt.File,
				Error:        tt.Error,
				Path:         fmt.Sprintf("/%s/%s/%s/%s/%d/%s", "repositories", workspace, repoSlug, "pullrequests", pullRequestId, "comments"),
				Decode:       DecodeJson[client.ApiResponse[client.PullRequestComment]],
				CallClient: func(bb *client.Client) (*client.ApiResponse[client.PullRequestComment], error) {
					return bb.ListPullRequestComments(workspace, repoSlug, pullRequestId, pagelen, page)
				},
			})
		})
	}
}

func TestClient_GetPullRequestDiff(t *testing.T) {
	t.Parallel()
	workspace, repoSlug, pullRequestId := "test_workspace", "test-repo", 1

	tests := []ClientEndpointTestCase{
		{
			Name:   "Success",
			Status: 200,
			File:   "testdata/pull_request_diff_mock.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			RunClientTest(t, ClientTestCase[string]{
				Status:       tt.Status,
				MockDataFile: tt.File,
				Error:        tt.Error,
				Path:         fmt.Sprintf("/%s/%s/%s/%s/%d/%s", "repositories", workspace, repoSlug, "pullrequests", pullRequestId, "diff"),
				Decode:       DecodeText,
				CallClient: func(bb *client.Client) (*string, error) {
					return bb.GetPullRequestDiff(workspace, repoSlug, pullRequestId)
				},
			})
		})
	}
}

func TestClient_GetFileSource(t *testing.T) {
	t.Parallel()
	workspace, repoSlug, commit, path := "test_workspace", "test-repo", "54ad501s", "src/test-path/test-file.ext"

	tests := []ClientEndpointTestCase{
		{
			Name:   "Success",
			Status: 200,
			File:   "testdata/file_source_mock.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			RunClientTest(t, ClientTestCase[string]{
				Status:       tt.Status,
				MockDataFile: tt.File,
				Error:        tt.Error,
				Path:         fmt.Sprintf("/%s/%s/%s/%s/%s/%s", "repositories", workspace, repoSlug, "src", commit, path),
				Decode:       DecodeText,
				CallClient: func(bb *client.Client) (*string, error) {
					return bb.GetFileSource(workspace, repoSlug, commit, path)
				},
			})
		})
	}
}

func TestClient_GetDirectorySource(t *testing.T) {
	t.Parallel()
	workspace, repoSlug, commit, path := "test_workspace", "test-repo", "abc123def456", ""

	tests := []ClientEndpointTestCase{
		{
			Name:   "Success",
			Status: 200,
			File:   "testdata/repository_src_mock.json",
		},
		{
			Name:   "Not Found",
			Status: 404,
			File:   "testdata/repository_src_mock_404.json",
			Error:  ClientError("No such file or directory: d"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			RunClientTest(t, ClientTestCase[client.ApiResponse[client.SourceItem]]{
				Status:       tt.Status,
				MockDataFile: tt.File,
				Error:        tt.Error,
				Path:         fmt.Sprintf("/%s/%s/%s/%s/%s", "repositories", workspace, repoSlug, "src", commit),
				Decode:       DecodeJson[client.ApiResponse[client.SourceItem]],
				CallClient: func(bb *client.Client) (*client.ApiResponse[client.SourceItem], error) {
					return bb.GetDirectorySource(workspace, repoSlug, commit, path)
				},
			})
		})
	}
}

func DecodeJson[T any](data []byte, res *T) error {
	return json.Unmarshal(data, res)
}

func DecodeText(data []byte, res *string) error {
	*res = string(data)
	return nil
}

type ClientEndpointTestCase struct {
	Name   string
	Status int
	Error  string
	File   string
}

type ClientTestCase[T any] struct {
	MockDataFile string
	Status       int
	Path         string
	CallClient   func(*client.Client) (*T, error)
	Decode       func(data []byte, res *T) error
	Error        string
}

func RunClientTest[T any](t *testing.T, tc ClientTestCase[T]) {
	t.Helper()

	mockData, err := os.ReadFile(tc.MockDataFile)
	require.NoError(t, err, "failed to read mock data file")

	var expectedResponse T

	if tc.Error == "" {
		err = tc.Decode(mockData, &expectedResponse)
		require.NoError(t, err, "failed to decode expected response")
	}

	config := &testConfig{
		email:    "test_user",
		apiToken: "test_password",
		timeout:  1,
	}

	config.url = NewTestServer(t, tc.Path, func(resp http.ResponseWriter, req *http.Request) {
		actualUsername, actualPassword, ok := req.BasicAuth()
		require.True(t, ok, "expected basic auth")
		require.Equal(t, config.BitbucketEmail(), actualUsername)
		require.Equal(t, config.BitbucketApiToken(), actualPassword)
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(tc.Status)
		resp.Write(mockData)
	})

	bb := client.NewClient(config)
	actualResponse, err := tc.CallClient(bb)

	if tc.Error == "" {
		require.NoError(t, err)
		assert.Equal(t, &expectedResponse, actualResponse)
	} else {
		require.EqualError(t, err, tc.Error)
		assert.Nil(t, actualResponse)
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

func ClientError(message string) string {
	return fmt.Sprintf("%s: %s", client.ErrClientBitbucket.Error(), message)
}
