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
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/suite"
)

// E2ETestSuite_OAuth is the test suite for end-to-end tests with OAuth authentication
type E2ETestSuite_OAuth struct {
	suite.Suite
	baseURL    string
	mcpClient  *mcp.ClientSession
	httpClient *http.Client
	server     *server.McpServer
	bitbucket  *httptest.Server
	cfg        config.Global
}

func TestE2E_OAuth(t *testing.T) {
	suite.Run(t, new(E2ETestSuite_OAuth))
}

func (s *E2ETestSuite_OAuth) SetupSuite() {
	s.SetupBitbucketServer()
	s.SetupMcpServer()
	s.SetupMcpClient()
}

func (s *E2ETestSuite_OAuth) SetupBitbucketServer() {
	auth := e2e.NewOpaqueTokenMiddleware("random-valid-token")
	s.bitbucket = e2e.NewBitbucketApiServer(s.T(), auth)
}

func (s *E2ETestSuite_OAuth) SetupMcpServer() {
	// Clear config cache to ensure environment variables are re-read for this test
	config.ClearCache()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	s.Require().NoError(err, "failed to find available port")

	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	gitBaseURL := e2e.NewGitRepos(s.T(), "test-workspace/test-repository")

	s.T().Setenv("SERVER_PORT", strconv.Itoa(port))
	s.T().Setenv("BITBUCKET_URL", s.bitbucket.URL)
	s.T().Setenv("BITBUCKET_TIMEOUT", "5")
	s.T().Setenv("BITBUCKET_AUTH", "oauth")
	s.T().Setenv("SERVER_URL", fmt.Sprintf("http://127.0.0.1:%d", port))
	s.T().Setenv("OAUTH_ISSUER", "https://bitbucket.org")
	s.T().Setenv("OAUTH_SCOPES", "repository;pullrequest")
	s.T().Setenv("BITBUCKET_GIT_URL", gitBaseURL)

	cfg := config.NewGlobal()
	s.server = server.NewMcpServer(cfg)
	s.baseURL = fmt.Sprintf("http://127.0.0.1:%d", port)
	s.httpClient = &http.Client{Timeout: 5 * time.Second}

	go func() {
		if err := s.server.Run(); err != nil && err != http.ErrServerClosed {
			s.T().Logf("server error: %v", err)
		}
	}()

	s.Require().NoError(s.server.WaitUntilReady(5*time.Second), "server failed to start")
}

func (s *E2ETestSuite_OAuth) SetupMcpClient() {
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "Test Client",
		Version: "1.0.0",
	}, nil)
	transport := &mcp.StreamableClientTransport{
		Endpoint: fmt.Sprintf("%s/%s", s.baseURL, "mcp"),
		HTTPClient: &http.Client{
			Transport: &oauthTransport{
				base:  http.DefaultTransport,
				token: "random-valid-token",
			},
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	session, err := client.Connect(ctx, transport, nil)
	s.Require().NoError(err, "failed to connect to mcp server")
	s.mcpClient = session
}

func (s *E2ETestSuite_OAuth) TearDownSuite() {
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

func (s *E2ETestSuite_OAuth) TestHealthEndpoint() {
	url := s.baseURL + "/health"
	req, err := http.NewRequest("GET", url, nil)
	s.Require().NoError(err, "failed to create GET /health request")

	resp, err := s.httpClient.Do(req)
	s.Require().NoError(err, "failed to make GET /health request")
	s.Assert().Equal(http.StatusOK, resp.StatusCode)
	s.Assert().Equal("application/json", resp.Header.Get("Content-Type"))
	s.Assert().Equal(`{"status":"ok"}`, strings.TrimSpace(e2e.ReadResponseBody(s.T(), resp)))
}

func (s *E2ETestSuite_OAuth) TestMcpInitialize() {
	s.Assert().Equal("Bitbucket MCP", s.mcpClient.InitializeResult().ServerInfo.Title)
	s.Assert().Equal("1.0.0", s.mcpClient.InitializeResult().ServerInfo.Version)
}

func (s *E2ETestSuite_OAuth) TestOAuthMetadataEndpoint() {
	url := s.baseURL + "/.well-known/oauth-protected-resource"
	req, err := http.NewRequest("GET", url, nil)
	s.Require().NoError(err, "failed to create GET metadata request")

	resp, err := s.httpClient.Do(req)
	s.Require().NoError(err, "failed to make GET metadata request")
	s.Assert().Equal(http.StatusOK, resp.StatusCode)
	s.Assert().Equal("application/json", resp.Header.Get("Content-Type"))

	body := e2e.ReadResponseBody(s.T(), resp)

	expected := fmt.Sprintf(`{
		"resource": "%s/mcp",
		"authorization_servers":["https://bitbucket.org/site/oauth2/access_token"],
		"scopes_supported":["repository","pullrequest"]
	}`, s.baseURL)

	s.Assert().JSONEq(expected, body)
}

func (s *E2ETestSuite_OAuth) TestRepositoriesResource() {
	uri := "bitbucket://api/test-workspace/repositories?page=1&pageSize=50"
	e2e.TestGetResource(s.T(), s.mcpClient, uri, "repositories.json")
}

func (s *E2ETestSuite_OAuth) TestCloneRepositoryTool() {
	targetPath := filepath.Join(s.T().TempDir(), "basic")
	args := map[string]any{
		"workspace":   "test-workspace",
		"repository":  "test-repository",
		"target_path": targetPath,
	}
	e2e.TestCallToolDynamic(s.T(), s.mcpClient, "clone_repository", args, map[string]any{
		"path": targetPath,
	})

	_, err := os.Stat(filepath.Join(targetPath, ".git"))
	s.Assert().NoError(err, ".git directory should exist in cloned path")
}

type oauthTransport struct {
	base  http.RoundTripper
	token string
}

func (t *oauthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := req.Clone(req.Context())
	req2.Header.Set("Authorization", "Bearer "+t.token)
	return t.base.RoundTrip(req2)
}
