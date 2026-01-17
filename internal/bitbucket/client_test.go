package bitbucket_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/branow/mcp-bitbucket/internal/bitbucket"
	"github.com/branow/mcp-bitbucket/internal/util"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			Name:      "Bad Request",
			Status:    400,
			File:      "testdata/repository_list_mock_400.json",
			ErrorCode: util.CodeInvalidParamsErr,
		},
		{
			Name:      "Unauthorized",
			Status:    401,
			File:      "testdata/repository_list_mock_401.json",
			ErrorCode: util.CodeInvalidParamsErr,
		},
		{
			Name:      "Not Found",
			Status:    404,
			File:      "testdata/repository_list_mock_404.json",
			ErrorCode: util.CodeResourceNotFoundErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			RunClientTest(t, ClientTestCase[bitbucket.ApiResponse[bitbucket.ApiRepository]]{
				Status:       tt.Status,
				MockDataFile: tt.File,
				ErrorCode:    tt.ErrorCode,
				Path:         fmt.Sprintf("/%s/%s", "repositories", workspace),
				Decode:       DecodeJson[bitbucket.ApiResponse[bitbucket.ApiRepository]],
				CallClient: func(bb *bitbucket.Client) (*bitbucket.ApiResponse[bitbucket.ApiRepository], error) {
					return bb.ListRepositories(context.Background(), workspace, pagelen, page)
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
			Name:      "Not Found",
			Status:    404,
			File:      "testdata/repository_mock_404.json",
			ErrorCode: util.CodeResourceNotFoundErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			RunClientTest(t, ClientTestCase[bitbucket.ApiRepository]{
				Status:       tt.Status,
				MockDataFile: tt.File,
				ErrorCode:    tt.ErrorCode,
				Path:         fmt.Sprintf("/%s/%s/%s", "repositories", workspace, repoSlug),
				Decode:       DecodeJson[bitbucket.ApiRepository],
				CallClient: func(bb *bitbucket.Client) (*bitbucket.ApiRepository, error) {
					return bb.GetRepository(context.Background(), workspace, repoSlug)
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
			RunClientTest(t, ClientTestCase[bitbucket.ApiResponse[bitbucket.ApiSourceItem]]{
				Status:       tt.Status,
				MockDataFile: tt.File,
				ErrorCode:    tt.ErrorCode,
				Path:         fmt.Sprintf("/%s/%s/%s/%s", "repositories", workspace, repoSlug, "src"),
				Decode:       DecodeJson[bitbucket.ApiResponse[bitbucket.ApiSourceItem]],
				CallClient: func(bb *bitbucket.Client) (*bitbucket.ApiResponse[bitbucket.ApiSourceItem], error) {
					return bb.GetRepositorySource(context.Background(), workspace, repoSlug)
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
			RunClientTest(t, ClientTestCase[bitbucket.ApiResponse[bitbucket.ApiPullRequest]]{
				Status:       tt.Status,
				MockDataFile: tt.File,
				ErrorCode:    tt.ErrorCode,
				Path:         fmt.Sprintf("/%s/%s/%s/%s", "repositories", workspace, repoSlug, "pullrequests"),
				Decode:       DecodeJson[bitbucket.ApiResponse[bitbucket.ApiPullRequest]],
				CallClient: func(bb *bitbucket.Client) (*bitbucket.ApiResponse[bitbucket.ApiPullRequest], error) {
					return bb.ListPullRequests(context.Background(), workspace, repoSlug, pagelen, page, nil)
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
			Name:      "Not Found",
			Status:    404,
			File:      "testdata/pull_request_mock_404.txt",
			ErrorCode: util.CodeResourceNotFoundErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			RunClientTest(t, ClientTestCase[bitbucket.ApiPullRequest]{
				Status:       tt.Status,
				MockDataFile: tt.File,
				ErrorCode:    tt.ErrorCode,
				Path:         fmt.Sprintf("/%s/%s/%s/%s/%d", "repositories", workspace, repoSlug, "pullrequests", pullRequestId),
				Decode:       DecodeJson[bitbucket.ApiPullRequest],
				CallClient: func(bb *bitbucket.Client) (*bitbucket.ApiPullRequest, error) {
					return bb.GetPullRequest(context.Background(), workspace, repoSlug, pullRequestId)
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
			RunClientTest(t, ClientTestCase[bitbucket.ApiResponse[bitbucket.ApiCommit]]{
				Status:       tt.Status,
				MockDataFile: tt.File,
				ErrorCode:    tt.ErrorCode,
				Path:         fmt.Sprintf("/%s/%s/%s/%s/%d/%s", "repositories", workspace, repoSlug, "pullrequests", pullRequestId, "commits"),
				Decode:       DecodeJson[bitbucket.ApiResponse[bitbucket.ApiCommit]],
				CallClient: func(bb *bitbucket.Client) (*bitbucket.ApiResponse[bitbucket.ApiCommit], error) {
					return bb.ListPullRequestCommits(context.Background(), workspace, repoSlug, pullRequestId)
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
			RunClientTest(t, ClientTestCase[bitbucket.ApiResponse[bitbucket.ApiPullRequestComment]]{
				Status:       tt.Status,
				MockDataFile: tt.File,
				ErrorCode:    tt.ErrorCode,
				Path:         fmt.Sprintf("/%s/%s/%s/%s/%d/%s", "repositories", workspace, repoSlug, "pullrequests", pullRequestId, "comments"),
				Decode:       DecodeJson[bitbucket.ApiResponse[bitbucket.ApiPullRequestComment]],
				CallClient: func(bb *bitbucket.Client) (*bitbucket.ApiResponse[bitbucket.ApiPullRequestComment], error) {
					return bb.ListPullRequestComments(context.Background(), workspace, repoSlug, pullRequestId, pagelen, page)
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
				ErrorCode:    tt.ErrorCode,
				Path:         fmt.Sprintf("/%s/%s/%s/%s/%d/%s", "repositories", workspace, repoSlug, "pullrequests", pullRequestId, "diff"),
				Decode:       DecodeText,
				CallClient: func(bb *bitbucket.Client) (*string, error) {
					return bb.GetPullRequestDiff(context.Background(), workspace, repoSlug, pullRequestId)
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
				ErrorCode:    tt.ErrorCode,
				Path:         fmt.Sprintf("/%s/%s/%s/%s/%s/%s", "repositories", workspace, repoSlug, "src", commit, path),
				Decode:       DecodeText,
				CallClient: func(bb *bitbucket.Client) (*string, error) {
					return bb.GetFileSource(context.Background(), workspace, repoSlug, commit, path)
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
			Name:      "Not Found",
			Status:    404,
			File:      "testdata/repository_src_mock_404.json",
			ErrorCode: util.CodeResourceNotFoundErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			RunClientTest(t, ClientTestCase[bitbucket.ApiResponse[bitbucket.ApiSourceItem]]{
				Status:       tt.Status,
				MockDataFile: tt.File,
				ErrorCode:    tt.ErrorCode,
				Path:         fmt.Sprintf("/%s/%s/%s/%s/%s", "repositories", workspace, repoSlug, "src", commit),
				Decode:       DecodeJson[bitbucket.ApiResponse[bitbucket.ApiSourceItem]],
				CallClient: func(bb *bitbucket.Client) (*bitbucket.ApiResponse[bitbucket.ApiSourceItem], error) {
					return bb.GetDirectorySource(context.Background(), workspace, repoSlug, commit, path)
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
	Name      string
	Status    int
	ErrorCode int64
	File      string
}

type ClientTestCase[T any] struct {
	MockDataFile string
	Status       int
	Path         string
	CallClient   func(*bitbucket.Client) (*T, error)
	Decode       func(data []byte, res *T) error
	ErrorCode    int64
}

func RunClientTest[T any](t *testing.T, tc ClientTestCase[T]) {
	t.Helper()

	mockData, err := os.ReadFile(tc.MockDataFile)
	require.NoError(t, err, "failed to read mock data file")

	var expectedResponse T

	if tc.ErrorCode == 0 {
		err = tc.Decode(mockData, &expectedResponse)
		require.NoError(t, err, "failed to decode expected response")
	}

	testUsername := "test_user"
	testPassword := "test_password"

	serverURL := NewTestServer(t, tc.Path, func(resp http.ResponseWriter, req *http.Request) {
		actualUsername, actualPassword, ok := req.BasicAuth()
		require.True(t, ok, "expected basic auth")
		require.Equal(t, testUsername, actualUsername)
		require.Equal(t, testPassword, actualPassword)
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(tc.Status)
		resp.Write(mockData)
	})

	config := bitbucket.Config{
		Url:     serverURL,
		Timeout: 1,
	}

	authorizer := util.NewBasicAuthorizer(testUsername, testPassword)
	bb := bitbucket.NewClient(config, authorizer)
	actualResponse, err := tc.CallClient(bb)

	if tc.ErrorCode == 0 {
		require.NoError(t, err)
		assert.Equal(t, &expectedResponse, actualResponse)
	} else {
		require.Error(t, err)
		var jsonrpcErr *jsonrpc.Error
		require.ErrorAs(t, err, &jsonrpcErr, "Error should be a jsonrpc.Error")
		assert.Equal(t, tc.ErrorCode, jsonrpcErr.Code, "Error code should match")
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
