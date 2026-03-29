package e2e_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/branow/mcp-bitbucket/internal/config"
	"github.com/branow/mcp-bitbucket/internal/server"
	"github.com/branow/mcp-bitbucket/internal/server/test/e2e"
	"github.com/branow/mcp-bitbucket/internal/util"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/suite"
)

// E2ETestSuite_BasicAuth is the test suite for end-to-end tests
type E2ETestSuite_BasicAuth struct {
	suite.Suite
	baseURL    string
	mcpClient  *mcp.ClientSession
	httpClient *http.Client
	server     *server.McpServer
	bitbucket  *httptest.Server
	cfg        config.Global
	gitBaseDir string
}

func TestE2E_BasicAuth(t *testing.T) {
	suite.Run(t, new(E2ETestSuite_BasicAuth))
}

func (s *E2ETestSuite_BasicAuth) SetupSuite() {
	s.SetupBitbucketServer()
	s.SetupMcpServer()
	s.SetupMcpClient()
}

func (s *E2ETestSuite_BasicAuth) SetupBitbucketServer() {
	auth := e2e.NewBasicAuthMiddleware("test@example.com", "test_token")
	s.bitbucket = e2e.NewBitbucketApiServer(s.T(), auth)
}

func (s *E2ETestSuite_BasicAuth) SetupMcpServer() {
	// Clear config cache to ensure environment variables are re-read for this test
	config.ClearCache()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	s.Require().NoError(err, "failed to find available port")

	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	gitServer, gitBaseDir := e2e.NewBitbucketGitServerWithBaseDir(s.T(), e2e.NewGitTokenMiddleware("test_token"), "test-workspace/test-repository")
	s.gitBaseDir = gitBaseDir

	s.T().Setenv("SERVER_PORT", strconv.Itoa(port))
	s.T().Setenv("BITBUCKET_URL", s.bitbucket.URL)
	s.T().Setenv("BITBUCKET_AUTH", "basic")
	s.T().Setenv("BITBUCKET_EMAIL", "test@example.com")
	s.T().Setenv("BITBUCKET_API_TOKEN", "test_token")
	s.T().Setenv("BITBUCKET_TIMEOUT", "5")
	s.T().Setenv("BITBUCKET_GIT_URL", gitServer.URL)

	s.cfg = config.NewGlobal()
	s.server = server.NewMcpServer(s.cfg)
	s.baseURL = fmt.Sprintf("http://127.0.0.1:%d", port)
	s.httpClient = &http.Client{Timeout: 5 * time.Second}

	go func() {
		if err := s.server.Run(); err != nil && err != http.ErrServerClosed {
			s.T().Logf("server error: %v", err)
		}
	}()

	s.Require().NoError(s.server.WaitUntilReady(5*time.Second), "server failed to start")
}

func (s *E2ETestSuite_BasicAuth) SetupMcpClient() {
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "Test Client",
		Version: "1.0.0",
	}, nil)
	transport := &mcp.StreamableClientTransport{
		Endpoint: fmt.Sprintf("%s/%s", s.baseURL, "mcp"),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	session, err := client.Connect(ctx, transport, nil)
	s.Require().NoError(err, "failed to connect to mcp server")
	s.mcpClient = session
}

func (s *E2ETestSuite_BasicAuth) TearDownSuite() {
	if s.mcpClient != nil {
		s.mcpClient.Close()
	}

	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.Require().NoError(s.server.Shutdown(ctx), "failed to shutdown server")
	}

	if s.bitbucket != nil {
		s.bitbucket.Close()
	}
}

func (s *E2ETestSuite_BasicAuth) TestHealthEndpoint() {
	url := s.baseURL + "/health"
	req, err := http.NewRequest("GET", url, nil)
	s.Require().NoError(err, "failed to create GET /health request")

	resp, err := s.httpClient.Do(req)
	s.Require().NoError(err, "failed to make GET /health request")
	s.Assert().Equal(http.StatusOK, resp.StatusCode)
	s.Assert().Equal("application/json", resp.Header.Get("Content-Type"))
	s.Assert().Equal(`{"status":"ok"}`, strings.TrimSpace(e2e.ReadResponseBody(s.T(), resp)))
}

func (s *E2ETestSuite_BasicAuth) TestMcpInitialize() {
	s.Assert().Equal("Bitbucket MCP", s.mcpClient.InitializeResult().ServerInfo.Title)
	s.Assert().Equal("1.0.0", s.mcpClient.InitializeResult().ServerInfo.Version)
}

func (s *E2ETestSuite_BasicAuth) TestRepositoriesResource() {
	uri := "bitbucket://api/test-workspace/repositories?page=1&pageSize=50"
	e2e.TestGetResource(s.T(), s.mcpClient, uri, "repositories.json")
}

func (s *E2ETestSuite_BasicAuth) TestRepositoriesResource_NotFound() {
	uri := "bitbucket://api/invalid-workspace/repositories?page=1&pageSize=50"
	code := util.CodeResourceNotFoundErr
	err := "You may not have access to this repository or it no longer exists in this workspace. If you think this repository exists and you have access, make sure you are authenticated."
	e2e.TestGetResourceError(s.T(), s.mcpClient, uri, code, err)
}

func (s *E2ETestSuite_BasicAuth) TestRepositoryResource() {
	tests := []struct {
		name string
		uri  string
		file string
	}{
		{
			name: "base",
			uri:  "bitbucket://api/test-workspace/repositories/test-repository?src=false",
			file: "/repository/base.json",
		},
		{
			name: "with source",
			uri:  "bitbucket://api/test-workspace/repositories/test-repository?readme=invalid&src=true",
			file: "/repository/with-src.json",
		},
		{
			name: "with source without readme",
			uri:  "bitbucket://api/test-workspace/repositories/test-repository-without-readme?readme=true&src=true",
			file: "/repository/with-src-without-readme.json",
		},
		{
			name: "with readme",
			uri:  "bitbucket://api/test-workspace/repositories/test-repository?readme=true",
			file: "/repository/with-readme.json",
		},
		{
			name: "with source and readme",
			uri:  "bitbucket://api/test-workspace/repositories/test-repository?readme=true&src=true",
			file: "/repository/with-src-and-readme.json",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			e2e.TestGetResource(s.T(), s.mcpClient, tt.uri, tt.file)
		})
	}
}

func (s *E2ETestSuite_BasicAuth) TestRepositoryResource_NotFound() {
	uri := "bitbucket://api/test-workspace/repositories/invalid-repository?src=true&readme=true"
	code := util.CodeResourceNotFoundErr
	err := "You may not have access to this repository or it no longer exists in this workspace. If you think this repository exists and you have access, make sure you are authenticated."
	e2e.TestGetResourceError(s.T(), s.mcpClient, uri, code, err)
}

func (s *E2ETestSuite_BasicAuth) TestPullRequestsResource() {
	tests := []struct {
		name string
		uri  string
		file string
	}{
		{
			name: "default state",
			uri:  "bitbucket://api/test-workspace/repositories/test-repository/pullrequests?page=1&pageSize=50",
			file: "/pullrequests/base.json",
		},
		{
			name: "with state MERGED",
			uri:  "bitbucket://api/test-workspace/repositories/test-repository/pullrequests?state=MERGED&page=1&pageSize=50",
			file: "/pullrequests/merged.json",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			e2e.TestGetResource(s.T(), s.mcpClient, tt.uri, tt.file)
		})
	}
}

func (s *E2ETestSuite_BasicAuth) TestPullRequestsResource_NotFound() {
	uri := "bitbucket://api/test-workspace/repositories/invalid-repository/pullrequests?page=1&pageSize=50"
	code := util.CodeResourceNotFoundErr
	err := "You may not have access to this repository or it no longer exists in this workspace. If you think this repository exists and you have access, make sure you are authenticated."
	e2e.TestGetResourceError(s.T(), s.mcpClient, uri, code, err)
}

func (s *E2ETestSuite_BasicAuth) TestFileContentResource() {
	tests := []struct {
		name string
		uri  string
		file string
	}{
		{
			name: "with ref",
			uri:  "bitbucket://api/test-workspace/repositories/test-repository/src/src/main/java/com/example/service/UserService.java?ref=feature-branch",
			file: "/filecontent/with-ref.json",
		},
		{
			name: "without ref",
			uri:  "bitbucket://api/test-workspace/repositories/test-repository/src/src/main/java/com/example/service/UserService.java",
			file: "/filecontent/without-ref.json",
		},
		{
			name: "nested path with ref",
			uri:  "bitbucket://api/test-workspace/repositories/test-repository/src/docs/api/v2/endpoints/users/README.md?ref=abc123def456",
			file: "/filecontent/nested-path.json",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			e2e.TestGetResource(s.T(), s.mcpClient, tt.uri, tt.file)
		})
	}
}

func (s *E2ETestSuite_BasicAuth) TestFileContentResource_NotFound() {
	uri := "bitbucket://api/test-workspace/repositories/test-repository/src/src/main/java/com/example/nonexistent/Missing.java?ref=abc123def456"
	code := util.CodeResourceNotFoundErr
	err := "Resource not found at"
	e2e.TestGetResourceError(s.T(), s.mcpClient, uri, code, err)
}

func (s *E2ETestSuite_BasicAuth) TestPullRequestResource() {
	tests := []struct {
		name string
		uri  string
		file string
	}{
		{
			name: "base",
			uri:  "bitbucket://api/test-workspace/repositories/test-repository/pullrequests/1",
			file: "/pullrequest/base.json",
		},
		{
			name: "with commits",
			uri:  "bitbucket://api/test-workspace/repositories/test-repository/pullrequests/1?commits=true",
			file: "/pullrequest/with-commits.json",
		},
		{
			name: "with diff",
			uri:  "bitbucket://api/test-workspace/repositories/test-repository/pullrequests/1?diff=true",
			file: "/pullrequest/with-diff.json",
		},
		{
			name: "with comments",
			uri:  "bitbucket://api/test-workspace/repositories/test-repository/pullrequests/1?comments=true",
			file: "/pullrequest/with-comments.json",
		},
		{
			name: "with all",
			uri:  "bitbucket://api/test-workspace/repositories/test-repository/pullrequests/1?commits=true&diff=true&comments=true",
			file: "/pullrequest/with-all.json",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			e2e.TestGetResource(s.T(), s.mcpClient, tt.uri, tt.file)
		})
	}
}

func (s *E2ETestSuite_BasicAuth) TestPullRequestResource_NotFound() {
	uri := "bitbucket://api/test-workspace/repositories/test-repository/pullrequests/999?commits=true&diff=true&comments=true"
	code := util.CodeResourceNotFoundErr
	err := "Resource not found at"
	e2e.TestGetResourceError(s.T(), s.mcpClient, uri, code, err)
}

func (s *E2ETestSuite_BasicAuth) TestListRepositoriesTool() {
	tests := []struct {
		name string
		args map[string]any
	}{
		{
			name: "base",
			args: map[string]any{
				"workspace": "test-workspace",
				"page":      1,
				"page_size": 50,
			},
		},
		{
			name: "without page",
			args: map[string]any{
				"workspace": "test-workspace",
			},
		},
		{
			name: "with invalid page",
			args: map[string]any{
				"workspace": "test-workspace",
				"page":      -10000000,
				"page_size": 0,
			},
		},
	}

	tool := "list_repositories"
	for _, tt := range tests {
		s.Run(tt.name, func() {
			e2e.TestCallTool(s.T(), s.mcpClient, tool, tt.args, "repositories.json")
		})
	}
}

func (s *E2ETestSuite_BasicAuth) TestListRepositoriesTool_Failure() {
	tests := []struct {
		name string
		args map[string]any
		code int64
		err  string
	}{
		{
			name: "no workspace",
			args: map[string]any{},
			code: util.CodeInvalidParamsErr,
			err:  "workspace: expected non-blank string, got: ''",
		},
		{
			name: "empty workspace",
			args: map[string]any{
				"workspace": "",
			},
			code: util.CodeInvalidParamsErr,
			err:  "workspace: expected non-blank string, got: ''",
		},
		{
			name: "blank workspace",
			args: map[string]any{
				"workspace": "   ",
			},
			code: util.CodeInvalidParamsErr,
			err:  "workspace: expected non-blank string, got: '   '",
		},
		{
			name: "invalid workspace",
			args: map[string]any{
				"workspace": "invalid-workspace",
			},
			code: util.CodeResourceNotFoundErr,
			err:  "You may not have access to this repository or it no longer exists in this workspace. If you think this repository exists and you have access, make sure you are authenticated.",
		},
	}

	tool := "list_repositories"
	for _, tt := range tests {
		s.Run(tt.name, func() {
			e2e.TestCallToolError(s.T(), s.mcpClient, tool, tt.args, tt.code, tt.err)
		})
	}
}

func (s *E2ETestSuite_BasicAuth) TestGetRepositoryTool() {
	tests := []struct {
		name string
		args map[string]any
		file string
	}{
		{
			name: "base",
			args: map[string]any{
				"workspace":  "test-workspace",
				"repository": "test-repository",
			},
			file: "/repository/base.json",
		},
		{
			name: "with source",
			args: map[string]any{
				"workspace":      "test-workspace",
				"repository":     "test-repository",
				"include_source": true,
			},
			file: "/repository/with-src.json",
		},
		{
			name: "with source without readme",
			args: map[string]any{
				"workspace":      "test-workspace",
				"repository":     "test-repository-without-readme",
				"include_source": true,
				"include_readme": false,
			},
			file: "/repository/with-src-without-readme.json",
		},
		{
			name: "with readme",
			args: map[string]any{
				"workspace":      "test-workspace",
				"repository":     "test-repository",
				"include_readme": true,
			},
			file: "/repository/with-readme.json",
		},
		{
			name: "with source and readme",
			args: map[string]any{
				"workspace":      "test-workspace",
				"repository":     "test-repository",
				"include_source": true,
				"include_readme": true,
			},
			file: "/repository/with-src-and-readme.json",
		},
	}

	tool := "get_repository"
	for _, tt := range tests {
		s.Run(tt.name, func() {
			e2e.TestCallTool(s.T(), s.mcpClient, tool, tt.args, tt.file)
		})
	}
}

func (s *E2ETestSuite_BasicAuth) TestGetRepositoryTool_Failure() {
	tests := []struct {
		name string
		args map[string]any
		code int64
		err  string
	}{
		{
			name: "no workspace",
			args: map[string]any{
				"repository": "test-repository",
			},
			code: util.CodeInvalidParamsErr,
			err:  "workspace: expected non-blank string, got: ''",
		},
		{
			name: "empty workspace",
			args: map[string]any{
				"workspace":  "",
				"repository": "test-repository",
			},
			code: util.CodeInvalidParamsErr,
			err:  "workspace: expected non-blank string, got: ''",
		},
		{
			name: "blank workspace",
			args: map[string]any{
				"workspace":  "   ",
				"repository": "test-repository",
			},
			code: util.CodeInvalidParamsErr,
			err:  "workspace: expected non-blank string, got: '   '",
		},
		{
			name: "no repository",
			args: map[string]any{
				"workspace": "test-workspace",
			},
			code: util.CodeInvalidParamsErr,
			err:  "repository: expected non-blank string, got: ''",
		},
		{
			name: "empty repository",
			args: map[string]any{
				"workspace":  "test-workspace",
				"repository": "",
			},
			code: util.CodeInvalidParamsErr,
			err:  "repository: expected non-blank string, got: ''",
		},
		{
			name: "blank repository",
			args: map[string]any{
				"workspace":  "test-workspace",
				"repository": "   ",
			},
			code: util.CodeInvalidParamsErr,
			err:  "repository: expected non-blank string, got: '   '",
		},
		{
			name: "invalid repository",
			args: map[string]any{
				"workspace":  "test-workspace",
				"repository": "invalid-repository",
			},
			code: util.CodeResourceNotFoundErr,
			err:  "You may not have access to this repository or it no longer exists in this workspace. If you think this repository exists and you have access, make sure you are authenticated.",
		},
	}

	tool := "get_repository"
	for _, tt := range tests {
		s.Run(tt.name, func() {
			e2e.TestCallToolError(s.T(), s.mcpClient, tool, tt.args, tt.code, tt.err)
		})
	}
}

func (s *E2ETestSuite_BasicAuth) TestGetPullRequestTool() {
	tests := []struct {
		name string
		args map[string]any
		file string
	}{
		{
			name: "base",
			args: map[string]any{
				"workspace":  "test-workspace",
				"repository": "test-repository",
				"id":         1,
			},
			file: "/pullrequest/base.json",
		},
		{
			name: "with commits",
			args: map[string]any{
				"workspace":       "test-workspace",
				"repository":      "test-repository",
				"id":              1,
				"include_commits": true,
			},
			file: "/pullrequest/with-commits.json",
		},
		{
			name: "with diff",
			args: map[string]any{
				"workspace":    "test-workspace",
				"repository":   "test-repository",
				"id":           1,
				"include_diff": true,
			},
			file: "/pullrequest/with-diff.json",
		},
		{
			name: "with comments",
			args: map[string]any{
				"workspace":        "test-workspace",
				"repository":       "test-repository",
				"id":               1,
				"include_comments": true,
			},
			file: "/pullrequest/with-comments.json",
		},
		{
			name: "with all",
			args: map[string]any{
				"workspace":        "test-workspace",
				"repository":       "test-repository",
				"id":               1,
				"include_commits":  true,
				"include_diff":     true,
				"include_comments": true,
			},
			file: "/pullrequest/with-all.json",
		},
	}

	tool := "get_pull_request"
	for _, tt := range tests {
		s.Run(tt.name, func() {
			e2e.TestCallTool(s.T(), s.mcpClient, tool, tt.args, tt.file)
		})
	}
}

func (s *E2ETestSuite_BasicAuth) TestGetPullRequestTool_Failure() {
	tests := []struct {
		name string
		args map[string]any
		code int64
		err  string
	}{
		{
			name: "blank workspace",
			args: map[string]any{
				"workspace":  "   ",
				"repository": "test-repository",
				"id":         1,
			},
			code: util.CodeInvalidParamsErr,
			err:  "workspace: expected non-blank string, got: '   '",
		},
		{
			name: "blank repository",
			args: map[string]any{
				"workspace":  "test-workspace",
				"repository": "   ",
				"id":         1,
			},
			code: util.CodeInvalidParamsErr,
			err:  "repository: expected non-blank string, got: '   '",
		},
		{
			name: "no id",
			args: map[string]any{
				"workspace":  "test-workspace",
				"repository": "test-repository",
			},
			code: util.CodeInvalidParamsErr,
			err:  "id: expected positive integer (> 0), got: 0",
		},
		{
			name: "invalid id",
			args: map[string]any{
				"workspace":  "test-workspace",
				"repository": "test-repository",
				"id":         -123,
			},
			code: util.CodeInvalidParamsErr,
			err:  "id: expected positive integer (> 0), got: -123",
		},
		{
			name: "not existing id",
			args: map[string]any{
				"workspace":  "test-workspace",
				"repository": "test-repository",
				"id":         999,
			},
			code: util.CodeResourceNotFoundErr,
			err:  "Resource not found at",
		},
	}

	tool := "get_pull_request"
	for _, tt := range tests {
		s.Run(tt.name, func() {
			e2e.TestCallToolError(s.T(), s.mcpClient, tool, tt.args, tt.code, tt.err)
		})
	}
}

func (s *E2ETestSuite_BasicAuth) TestListPullRequestsTool() {
	tests := []struct {
		name string
		args map[string]any
		file string
	}{
		{
			name: "base",
			args: map[string]any{
				"workspace":  "test-workspace",
				"repository": "test-repository",
				"page":       1,
				"page_size":  50,
			},
			file: "/pullrequests/base.json",
		},
		{
			name: "without page",
			args: map[string]any{
				"workspace":  "test-workspace",
				"repository": "test-repository",
			},
			file: "/pullrequests/base.json",
		},
		{
			name: "with invalid page",
			args: map[string]any{
				"workspace":  "test-workspace",
				"repository": "test-repository",
				"page":       -10000000,
				"page_size":  0,
			},
			file: "/pullrequests/base.json",
		},
		{
			name: "with state MERGED",
			args: map[string]any{
				"workspace":  "test-workspace",
				"repository": "test-repository",
				"state":      []string{"MERGED"},
			},
			file: "/pullrequests/merged.json",
		},
	}

	tool := "list_pull_requests"
	for _, tt := range tests {
		s.Run(tt.name, func() {
			e2e.TestCallTool(s.T(), s.mcpClient, tool, tt.args, tt.file)
		})
	}
}

func (s *E2ETestSuite_BasicAuth) TestListPullRequestsTool_Failure() {
	tests := []struct {
		name string
		args map[string]any
		code int64
		err  string
	}{
		{
			name: "no workspace",
			args: map[string]any{
				"repository": "test-repository",
			},
			code: util.CodeInvalidParamsErr,
			err:  "workspace: expected non-blank string, got: ''",
		},
		{
			name: "empty workspace",
			args: map[string]any{
				"workspace":  "",
				"repository": "test-repository",
			},
			code: util.CodeInvalidParamsErr,
			err:  "workspace: expected non-blank string, got: ''",
		},
		{
			name: "blank workspace",
			args: map[string]any{
				"workspace":  "   ",
				"repository": "test-repository",
			},
			code: util.CodeInvalidParamsErr,
			err:  "workspace: expected non-blank string, got: '   '",
		},
		{
			name: "no repository",
			args: map[string]any{
				"workspace": "test-workspace",
			},
			code: util.CodeInvalidParamsErr,
			err:  "repository: expected non-blank string, got: ''",
		},
		{
			name: "empty repository",
			args: map[string]any{
				"workspace":  "test-workspace",
				"repository": "",
			},
			code: util.CodeInvalidParamsErr,
			err:  "repository: expected non-blank string, got: ''",
		},
		{
			name: "blank repository",
			args: map[string]any{
				"workspace":  "test-workspace",
				"repository": "   ",
			},
			code: util.CodeInvalidParamsErr,
			err:  "repository: expected non-blank string, got: '   '",
		},
		{
			name: "invalid repository",
			args: map[string]any{
				"workspace":  "test-workspace",
				"repository": "invalid-repository",
			},
			code: util.CodeResourceNotFoundErr,
			err:  "You may not have access to this repository or it no longer exists in this workspace. If you think this repository exists and you have access, make sure you are authenticated.",
		},
	}

	tool := "list_pull_requests"
	for _, tt := range tests {
		s.Run(tt.name, func() {
			e2e.TestCallToolError(s.T(), s.mcpClient, tool, tt.args, tt.code, tt.err)
		})
	}
}

func (s *E2ETestSuite_BasicAuth) TestDirectorySourceResource() {
	tests := []struct {
		name string
		uri  string
		file string
	}{
		{
			name: "with ref",
			uri:  "bitbucket://api/test-workspace/repositories/test-repository/src/dir/src/main/java/com/example/service?ref=feature-branch",
			file: "/directorysource/with-ref.json",
		},
		{
			name: "without ref",
			uri:  "bitbucket://api/test-workspace/repositories/test-repository/src/dir/src/main/java/com/example/service",
			file: "/directorysource/without-ref.json",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			e2e.TestGetResource(s.T(), s.mcpClient, tt.uri, tt.file)
		})
	}
}

func (s *E2ETestSuite_BasicAuth) TestDirectorySourceResource_NotFound() {
	uri := "bitbucket://api/test-workspace/repositories/test-repository/src/dir/src/main/java/com/example/nonexistent?ref=abc123def456"
	code := util.CodeResourceNotFoundErr
	err := "No such file or directory"
	e2e.TestGetResourceError(s.T(), s.mcpClient, uri, code, err)
}

func (s *E2ETestSuite_BasicAuth) TestGetDirectorySourceTool() {
	tests := []struct {
		name string
		args map[string]any
		file string
	}{
		{
			name: "with ref",
			args: map[string]any{
				"workspace":  "test-workspace",
				"repository": "test-repository",
				"path":       "src/main/java/com/example/service",
				"ref":        "feature-branch",
			},
			file: "/directorysource/with-ref.json",
		},
		{
			name: "without ref",
			args: map[string]any{
				"workspace":  "test-workspace",
				"repository": "test-repository",
				"path":       "src/main/java/com/example/service",
			},
			file: "/directorysource/without-ref.json",
		},
	}

	tool := "get_directory_source"
	for _, tt := range tests {
		s.Run(tt.name, func() {
			e2e.TestCallTool(s.T(), s.mcpClient, tool, tt.args, tt.file)
		})
	}
}

func (s *E2ETestSuite_BasicAuth) TestGetDirectorySourceTool_Failure() {
	tests := []struct {
		name string
		args map[string]any
		code int64
		err  string
	}{
		{
			name: "blank workspace",
			args: map[string]any{
				"workspace":  "   ",
				"repository": "test-repository",
				"path":       "src/main/java/com/example/service",
			},
			code: util.CodeInvalidParamsErr,
			err:  "workspace: expected non-blank string, got: '   '",
		},
		{
			name: "blank repository",
			args: map[string]any{
				"workspace":  "test-workspace",
				"repository": "   ",
				"path":       "src/main/java/com/example/service",
			},
			code: util.CodeInvalidParamsErr,
			err:  "repository: expected non-blank string, got: '   '",
		},
		{
			name: "blank path",
			args: map[string]any{
				"workspace":  "test-workspace",
				"repository": "test-repository",
				"path":       "   ",
			},
			code: util.CodeInvalidParamsErr,
			err:  "path: expected non-blank string, got: '   '",
		},
		{
			name: "not found directory",
			args: map[string]any{
				"workspace":  "test-workspace",
				"repository": "test-repository",
				"path":       "src/main/java/com/example/nonexistent",
				"ref":        "abc123def456",
			},
			code: util.CodeResourceNotFoundErr,
			err:  "No such file or directory",
		},
	}

	tool := "get_directory_source"
	for _, tt := range tests {
		s.Run(tt.name, func() {
			e2e.TestCallToolError(s.T(), s.mcpClient, tool, tt.args, tt.code, tt.err)
		})
	}
}

func (s *E2ETestSuite_BasicAuth) TestGetFileContentTool() {
	tests := []struct {
		name string
		args map[string]any
		file string
	}{
		{
			name: "with ref",
			args: map[string]any{
				"workspace":  "test-workspace",
				"repository": "test-repository",
				"path":       "src/main/java/com/example/service/UserService.java",
				"ref":        "feature-branch",
			},
			file: "/filecontent/with-ref.json",
		},
		{
			name: "without ref",
			args: map[string]any{
				"workspace":  "test-workspace",
				"repository": "test-repository",
				"path":       "src/main/java/com/example/service/UserService.java",
			},
			file: "/filecontent/without-ref.json",
		},
		{
			name: "nested path with ref",
			args: map[string]any{
				"workspace":  "test-workspace",
				"repository": "test-repository",
				"path":       "docs/api/v2/endpoints/users/README.md",
				"ref":        "abc123def456",
			},
			file: "/filecontent/nested-path.json",
		},
	}

	tool := "get_file_content"
	for _, tt := range tests {
		s.Run(tt.name, func() {
			e2e.TestCallTool(s.T(), s.mcpClient, tool, tt.args, tt.file)
		})
	}
}

func (s *E2ETestSuite_BasicAuth) TestGetFileContentTool_Failure() {
	tests := []struct {
		name string
		args map[string]any
		code int64
		err  string
	}{
		{
			name: "blank workspace",
			args: map[string]any{
				"workspace":  "   ",
				"repository": "test-repository",
				"path":       "src/main/java/com/example/service/UserService.java",
			},
			code: util.CodeInvalidParamsErr,
			err:  "workspace: expected non-blank string, got: '   '",
		},
		{
			name: "blank repository",
			args: map[string]any{
				"workspace":  "test-workspace",
				"repository": "   ",
				"path":       "src/main/java/com/example/service/UserService.java",
			},
			code: util.CodeInvalidParamsErr,
			err:  "repository: expected non-blank string, got: '   '",
		},
		{
			name: "blank path",
			args: map[string]any{
				"workspace":  "test-workspace",
				"repository": "test-repository",
				"path":       "   ",
			},
			code: util.CodeInvalidParamsErr,
			err:  "path: expected non-blank string, got: '   '",
		},
		{
			name: "not found file",
			args: map[string]any{
				"workspace":  "test-workspace",
				"repository": "test-repository",
				"path":       "src/main/java/com/example/nonexistent/Missing.java",
				"ref":        "abc123def456",
			},
			code: util.CodeResourceNotFoundErr,
			err:  "Resource not found at",
		},
	}

	tool := "get_file_content"
	for _, tt := range tests {
		s.Run(tt.name, func() {
			e2e.TestCallToolError(s.T(), s.mcpClient, tool, tt.args, tt.code, tt.err)
		})
	}
}

func (s *E2ETestSuite_BasicAuth) TestCloneRepositoryTool() {
	tests := []struct {
		name string
		args map[string]any
	}{
		{
			name: "basic clone",
			args: map[string]any{
				"workspace":   "test-workspace",
				"repository":  "test-repository",
				"target_path": filepath.Join(s.T().TempDir(), "basic"),
			},
		},
		{
			name: "shallow clone",
			args: map[string]any{
				"workspace":   "test-workspace",
				"repository":  "test-repository",
				"target_path": filepath.Join(s.T().TempDir(), "shallow"),
				"depth":       1,
			},
		},
		{
			name: "clone with ref",
			args: map[string]any{
				"workspace":   "test-workspace",
				"repository":  "test-repository",
				"target_path": filepath.Join(s.T().TempDir(), "ref"),
				"ref":         "main",
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			targetPath := tt.args["target_path"].(string)
			e2e.TestCallToolDynamic(s.T(), s.mcpClient, "clone_repository", tt.args, map[string]any{
				"path": targetPath,
			})

			_, err := os.Stat(filepath.Join(targetPath, ".git"))
			s.Assert().NoError(err, ".git directory should exist in cloned path")
		})
	}
}

func (s *E2ETestSuite_BasicAuth) TestCloneRepositoryTool_PullIfExists() {
	s.Run("already up to date", func() {
		targetPath := filepath.Join(s.T().TempDir(), "repo")
		args := map[string]any{
			"workspace":   "test-workspace",
			"repository":  "test-repository",
			"target_path": targetPath,
		}
		e2e.TestCallToolDynamic(s.T(), s.mcpClient, "clone_repository", args, map[string]any{"path": targetPath})

		// Pull with pull_if_exists=true — repo already cloned, nothing new to pull.
		pullArgs := map[string]any{
			"workspace":      "test-workspace",
			"repository":     "test-repository",
			"target_path":    targetPath,
			"pull_if_exists": true,
		}
		e2e.TestCallToolDynamic(s.T(), s.mcpClient, "clone_repository", pullArgs, map[string]any{"path": targetPath})

		_, err := os.Stat(filepath.Join(targetPath, ".git"))
		s.Assert().NoError(err, ".git directory should still exist after pull")
	})

	s.Run("pull new commit", func() {
		targetPath := filepath.Join(s.T().TempDir(), "repo")
		args := map[string]any{
			"workspace":   "test-workspace",
			"repository":  "test-repository",
			"target_path": targetPath,
		}
		e2e.TestCallToolDynamic(s.T(), s.mcpClient, "clone_repository", args, map[string]any{"path": targetPath})

		e2e.AddCommitToBareRepo(s.T(), s.gitBaseDir, "test-workspace/test-repository", "PULLED.md", "pulled")

		pullArgs := map[string]any{
			"workspace":      "test-workspace",
			"repository":     "test-repository",
			"target_path":    targetPath,
			"pull_if_exists": true,
		}
		e2e.TestCallToolDynamic(s.T(), s.mcpClient, "clone_repository", pullArgs, map[string]any{"path": targetPath})

		_, err := os.Stat(filepath.Join(targetPath, "PULLED.md"))
		s.Assert().NoError(err, "PULLED.md should exist after pull")
	})
}

func (s *E2ETestSuite_BasicAuth) TestCloneRepositoryTool_Failure() {
	tests := []struct {
		name string
		args map[string]any
		code int64
		err  string
	}{
		{
			name: "blank workspace",
			args: map[string]any{
				"workspace":   "",
				"repository":  "test-repository",
				"target_path": s.T().TempDir(),
			},
			code: util.CodeInvalidParamsErr,
			err:  "workspace: expected non-blank string",
		},
		{
			name: "blank repository",
			args: map[string]any{
				"workspace":   "test-workspace",
				"repository":  "",
				"target_path": s.T().TempDir(),
			},
			code: util.CodeInvalidParamsErr,
			err:  "repository: expected non-blank string",
		},
		{
			name: "blank target_path",
			args: map[string]any{
				"workspace":   "test-workspace",
				"repository":  "test-repository",
				"target_path": "",
			},
			code: util.CodeInvalidParamsErr,
			err:  "target_path: expected non-blank string",
		},
		{
			name: "negative depth",
			args: map[string]any{
				"workspace":   "test-workspace",
				"repository":  "test-repository",
				"target_path": s.T().TempDir(),
				"depth":       -1,
			},
			code: util.CodeInvalidParamsErr,
			err:  "depth:",
		},
		{
			name: "non-existent repository",
			args: map[string]any{
				"workspace":   "test-workspace",
				"repository":  "nonexistent-repository",
				"target_path": filepath.Join(s.T().TempDir(), "repo"),
			},
			code: util.CodeInvalidParamsErr,
			err:  "git clone failed:",
		},
	}

	tool := "clone_repository"
	for _, tt := range tests {
		s.Run(tt.name, func() {
			e2e.TestCallToolError(s.T(), s.mcpClient, tool, tt.args, tt.code, tt.err)
		})
	}
}
