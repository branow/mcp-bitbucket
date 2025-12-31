package server_test

import (
	"context"
	"encoding/base64"
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
	"github.com/branow/mcp-bitbucket/internal/util"
	"github.com/branow/mcp-bitbucket/internal/util/web"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
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
	mux := http.NewServeMux()
	newBitbucketRepositoryHandler(s.T(), mux)
	auth := newBasicAuthMiddleware("test@example.com", "test_token")
	s.bitbucket = httptest.NewServer(auth(mux))
}

func (s *E2ETestSuite_BasicAuth) SetupMcpServer() {
	// Clear config cache to ensure environment variables are re-read for this test
	config.ClearCache()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	s.Require().NoError(err, "failed to find available port")

	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	s.T().Setenv("SERVER_PORT", strconv.Itoa(port))
	s.T().Setenv("BITBUCKET_URL", s.bitbucket.URL)
	s.T().Setenv("BITBUCKET_AUTH", "basic")
	s.T().Setenv("BITBUCKET_EMAIL", "test@example.com")
	s.T().Setenv("BITBUCKET_API_TOKEN", "test_token")
	s.T().Setenv("BITBUCKET_TIMEOUT", "5")

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
	s.Assert().Equal(`{"status":"ok"}`, strings.TrimSpace(readResponseBody(s.T(), resp)))
}

func (s *E2ETestSuite_BasicAuth) TestMcpInitialize() {
	s.Assert().Equal("Bitbucket MCP", s.mcpClient.InitializeResult().ServerInfo.Title)
	s.Assert().Equal("1.0.0", s.mcpClient.InitializeResult().ServerInfo.Version)
}

func (s *E2ETestSuite_BasicAuth) TestListRepositoriesResource() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	uri := "mcp://bitbucket/test-workspace/repositories?page=1&pageSize=50"
	result, err := s.mcpClient.ReadResource(ctx, &mcp.ReadResourceParams{URI: uri})
	s.Require().NoError(err, "failed to read repositories resource")
	s.Require().NotNil(result)
	s.Require().Len(result.Contents, 1)
	s.Assert().Equal(uri, result.Contents[0].URI)
	s.Assert().Equal("application/json", result.Contents[0].MIMEType)

	expectedData := readMcpServerTestData(s.T(), "list_repositories.json")
	s.Assert().JSONEq(string(expectedData), result.Contents[0].Text)
}

func (s *E2ETestSuite_BasicAuth) TestListRepositoriesResource_NotFound() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	uri := "mcp://bitbucket/invalid-workspace/repositories?page=1&pageSize=50"
	result, err := s.mcpClient.ReadResource(ctx, &mcp.ReadResourceParams{URI: uri})
	s.Require().Error(err, "expected error for invalid namespace")
	s.Assert().Nil(result)

	var jsonrpcErr *jsonrpc.Error
	s.Require().ErrorAs(err, &jsonrpcErr, "error should be a JSON-RPC error")
	s.Assert().Equal(util.CodeResourceNotFoundErr, jsonrpcErr.Code, "error code should be ResourceNotFound")
	s.Assert().Equal("You may not have access to this repository or it no longer exists in this workspace. If you think this repository exists and you have access, make sure you are authenticated.", jsonrpcErr.Message)
}

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
	mux := http.NewServeMux()
	newBitbucketRepositoryHandler(s.T(), mux)
	auth := newOpaqueTokenMiddleware("random-valid-token")
	s.bitbucket = httptest.NewServer(auth(mux))
}

func (s *E2ETestSuite_OAuth) SetupMcpServer() {
	// Clear config cache to ensure environment variables are re-read for this test
	config.ClearCache()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	s.Require().NoError(err, "failed to find available port")

	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	s.T().Setenv("SERVER_PORT", strconv.Itoa(port))
	s.T().Setenv("BITBUCKET_URL", s.bitbucket.URL)
	s.T().Setenv("BITBUCKET_TIMEOUT", "5")
	s.T().Setenv("BITBUCKET_AUTH", "oauth")
	s.T().Setenv("SERVER_URL", fmt.Sprintf("http://127.0.0.1:%d", port))
	s.T().Setenv("OAUTH_ISSUER", "https://bitbucket.org")
	s.T().Setenv("OAUTH_SCOPES", "repository;pullrequest")

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
	s.Assert().Equal(`{"status":"ok"}`, strings.TrimSpace(readResponseBody(s.T(), resp)))
}

func (s *E2ETestSuite_OAuth) TestOAuthMetadataEndpoint() {
	url := s.baseURL + "/.well-known/oauth-protected-resource"
	req, err := http.NewRequest("GET", url, nil)
	s.Require().NoError(err, "failed to create GET metadata request")

	resp, err := s.httpClient.Do(req)
	s.Require().NoError(err, "failed to make GET metadata request")
	s.Assert().Equal(http.StatusOK, resp.StatusCode)
	s.Assert().Equal("application/json", resp.Header.Get("Content-Type"))

	body := readResponseBody(s.T(), resp)

	expected := fmt.Sprintf(`{
		"resource": "%s/mcp",
		"authorization_servers":["https://bitbucket.org/site/oauth2/access_token"],
		"scopes_supported":["repository","pullrequest"]
	}`, s.baseURL)

	s.Assert().JSONEq(expected, body)
}

func (s *E2ETestSuite_OAuth) TestListRepositoriesResource() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	uri := "mcp://bitbucket/test-workspace/repositories?page=1&pageSize=50"
	result, err := s.mcpClient.ReadResource(ctx, &mcp.ReadResourceParams{URI: uri})
	s.Require().NoError(err, "failed to read repositories resource with valid token")
	s.Require().NotNil(result)
	s.Require().Len(result.Contents, 1)
	s.Assert().Equal(uri, result.Contents[0].URI)
	s.Assert().Equal("application/json", result.Contents[0].MIMEType)

	expectedData := readMcpServerTestData(s.T(), "list_repositories.json")
	s.Assert().JSONEq(string(expectedData), result.Contents[0].Text)
}

type Middleware func(http.Handler) http.Handler

func newBasicAuthMiddleware(username, password string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			const prefix = "Basic "
			if !strings.HasPrefix(auth, prefix) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			payload, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(string(payload), ":", 2)
			if len(parts) != 2 || parts[0] != username || parts[1] != password {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func newOpaqueTokenMiddleware(token string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			const prefix = "Bearer "

			if !strings.HasPrefix(auth, prefix) || auth[len(prefix):] != token {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func newBitbucketRepositoryHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) < 2 || pathParts[1] != "test-workspace" {
			w.WriteHeader(http.StatusNotFound)
			w.Write(readBitbucketTestData(t, "not_found_error.json"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(readBitbucketTestData(t, "list_repositories.json"))
	})
}

func readBitbucketTestData(t *testing.T, filename string) []byte {
	t.Helper()
	return readTestData(t, filepath.Join("bitbucket", filename))
}

func readMcpServerTestData(t *testing.T, filename string) []byte {
	t.Helper()
	return readTestData(t, filepath.Join("mcpserver", filename))
}

func readTestData(t *testing.T, filename string) []byte {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("testdata", filename))
	require.NoError(t, err, fmt.Sprintf("failed to read test data file %s", filename))
	return data
}

func readResponseBody(t *testing.T, resp *http.Response) string {
	t.Helper()

	var body string
	err := web.ReadResponseText(resp, &body)
	require.NoError(t, err, "failed to read response body")
	return body
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
