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
	"github.com/stretchr/testify/assert"
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
	newBitbucketRepositoriesHandler(s.T(), mux)
	newBitbucketRepositoriesNotFoundHandler(s.T(), mux)
	newBitbucketRepositoryHandler(s.T(), mux)
	newBitbucketRepositoryWithoutReadmeHandler(s.T(), mux)
	newBitbucketRepositoryNotFoundHandler(s.T(), mux)
	newBitbucketRepositorySourceHandler(s.T(), mux)
	newBitbucketRepositorySourceWithoutReadmeHandler(s.T(), mux)
	newBitbucketRepositorySourceNotFoundHandler(s.T(), mux)
	newBitbucketFileSourceReadmeHandler(s.T(), mux)
	newBitbucketPullRequestHandler(s.T(), mux)
	newBitbucketPullRequestNotFoundHandler(s.T(), mux)
	newBitbucketPullRequestCommitsHandler(s.T(), mux)
	newBitbucketPullRequestCommitsNotFoundHandler(s.T(), mux)
	newBitbucketPullRequestDiffHandler(s.T(), mux)
	newBitbucketPullRequestDiffNotFoundHandler(s.T(), mux)
	newBitbucketPullRequestCommentsHandler(s.T(), mux)
	newBitbucketPullRequestCommentsNotFoundHandler(s.T(), mux)
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

func (s *E2ETestSuite_BasicAuth) TestRepositoriesResource() {
	uri := "mcp://bitbucket/test-workspace/repositories?page=1&pageSize=50"
	responses := []string{"repositories.json"}
	testResource(s.T(), s.mcpClient, uri, responses)
}

func (s *E2ETestSuite_BasicAuth) TestRepositoriesResource_NotFound() {
	uri := "mcp://bitbucket/invalid-workspace/repositories?page=1&pageSize=50"
	code := util.CodeResourceNotFoundErr
	err := "You may not have access to this repository or it no longer exists in this workspace. If you think this repository exists and you have access, make sure you are authenticated."
	testResourceError(s.T(), s.mcpClient, uri, code, err)
}

func (s *E2ETestSuite_BasicAuth) TestRepositoryResource() {
	tests := []struct {
		name      string
		uri       string
		responses []string
	}{
		{
			name:      "base",
			uri:       "mcp://bitbucket/test-workspace/repositories/test-repository?src=false",
			responses: []string{"/repository/base.json"},
		},
		{
			name:      "with source",
			uri:       "mcp://bitbucket/test-workspace/repositories/test-repository?readme=invalid&src=true",
			responses: []string{"/repository/with-src.json"},
		},
		{
			name:      "with source without readme",
			uri:       "mcp://bitbucket/test-workspace/repositories/test-repository-without-readme?readme=true&src=true",
			responses: []string{"/repository/with-src-without-readme.json"},
		},
		{
			name:      "with readme",
			uri:       "mcp://bitbucket/test-workspace/repositories/test-repository?readme=true",
			responses: []string{"/repository/with-readme.json"},
		},
		{
			name:      "with source and readme",
			uri:       "mcp://bitbucket/test-workspace/repositories/test-repository?readme=true&src=true",
			responses: []string{"/repository/with-src-and-readme.json"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			testResource(s.T(), s.mcpClient, tt.uri, tt.responses)
		})
	}
}

func (s *E2ETestSuite_BasicAuth) TestRepositoryResource_NotFound() {
	uri := "mcp://bitbucket/test-workspace/repositories/invalid-repository?src=true&readme=true"
	code := util.CodeResourceNotFoundErr
	err := "You may not have access to this repository or it no longer exists in this workspace. If you think this repository exists and you have access, make sure you are authenticated."
	testResourceError(s.T(), s.mcpClient, uri, code, err)
}

func (s *E2ETestSuite_BasicAuth) TestPullRequestResource() {
	tests := []struct {
		name      string
		uri       string
		responses []string
	}{
		{
			name:      "base",
			uri:       "mcp://bitbucket/test-workspace/repositories/test-repository/pullrequests/1",
			responses: []string{"/pullrequest/base.json"},
		},
		{
			name:      "with commits",
			uri:       "mcp://bitbucket/test-workspace/repositories/test-repository/pullrequests/1?commits=true",
			responses: []string{"/pullrequest/with-commits.json"},
		},
		{
			name:      "with diff",
			uri:       "mcp://bitbucket/test-workspace/repositories/test-repository/pullrequests/1?diff=true",
			responses: []string{"/pullrequest/with-diff.json"},
		},
		{
			name:      "with comments",
			uri:       "mcp://bitbucket/test-workspace/repositories/test-repository/pullrequests/1?comments=true",
			responses: []string{"/pullrequest/with-comments.json"},
		},
		{
			name:      "with all",
			uri:       "mcp://bitbucket/test-workspace/repositories/test-repository/pullrequests/1?commits=true&diff=true&comments=true",
			responses: []string{"/pullrequest/with-all.json"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			testResource(s.T(), s.mcpClient, tt.uri, tt.responses)
		})
	}
}

func (s *E2ETestSuite_BasicAuth) TestPullRequestResource_NotFound() {
	uri := "mcp://bitbucket/test-workspace/repositories/test-repository/pullrequests/999?commits=true&diff=true&comments=true"
	code := util.CodeResourceNotFoundErr
	err := "Resource not found at"
	testResourceError(s.T(), s.mcpClient, uri, code, err)
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
	newBitbucketRepositoriesHandler(s.T(), mux)
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

	body := readResponseBody(s.T(), resp)

	expected := fmt.Sprintf(`{
		"resource": "%s/mcp",
		"authorization_servers":["https://bitbucket.org/site/oauth2/access_token"],
		"scopes_supported":["repository","pullrequest"]
	}`, s.baseURL)

	s.Assert().JSONEq(expected, body)
}

func (s *E2ETestSuite_OAuth) TestRepositoriesResource() {
	uri := "mcp://bitbucket/test-workspace/repositories?page=1&pageSize=50"
	responses := []string{"repositories.json"}
	testResource(s.T(), s.mcpClient, uri, responses)
}

func testResource(t *testing.T, client *mcp.ClientSession, uri string, responses []string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := client.ReadResource(ctx, &mcp.ReadResourceParams{URI: uri})
	require.NoError(t, err, "failed to read resource")
	require.NotNil(t, result)
	require.Len(t, result.Contents, len(responses))

	for i, resp := range responses {
		assert.Equal(t, uri, result.Contents[i].URI)
		assert.Equal(t, "application/json", result.Contents[i].MIMEType)
		expectedData := readMcpServerTestData(t, resp)
		assert.JSONEq(t, string(expectedData), result.Contents[i].Text)
	}
}

func testResourceError(t *testing.T, client *mcp.ClientSession, uri string, code int64, error string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := client.ReadResource(ctx, &mcp.ReadResourceParams{URI: uri})
	require.Error(t, err)
	assert.Nil(t, result)

	var jsonrpcErr *jsonrpc.Error
	require.ErrorAs(t, err, &jsonrpcErr, "error should be a JSON-RPC error")
	assert.Equal(t, code, jsonrpcErr.Code, "unexpected error code")
	assert.Contains(t, jsonrpcErr.Message, error, "unexpected error message")
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

func newBitbucketRepositoriesHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(readBitbucketTestData(t, "repositories.json"))
	})
}

func newBitbucketRepositoriesNotFoundHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/invalid-workspace", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		w.Write(readBitbucketTestData(t, "not-found.json"))
	})
}

func newBitbucketRepositoryHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/test-repository", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(readBitbucketTestData(t, "repository.json"))
	})
}

func newBitbucketRepositoryWithoutReadmeHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/test-repository-without-readme", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(readBitbucketTestData(t, "repository.json"))
	})
}

func newBitbucketRepositoryNotFoundHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/invalid-repository", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		w.Write(readBitbucketTestData(t, "not-found.json"))
	})
}

func newBitbucketRepositorySourceHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/test-repository/src", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(readBitbucketTestData(t, "repository-source.json"))
	})
}

func newBitbucketRepositorySourceWithoutReadmeHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/test-repository-without-readme/src", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(readBitbucketTestData(t, "repository-source-without-readme.json"))
	})
}

func newBitbucketRepositorySourceNotFoundHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/invalid-repository/src", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		w.Write(readBitbucketTestData(t, "not-found.json"))
	})
}

func newBitbucketFileSourceReadmeHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/test-repository/src/abc123def456/README.md", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain")
		w.Write(readBitbucketTestData(t, "file-source-readme.md"))
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

func newBitbucketPullRequestHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/test-repository/pullrequests/1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(readBitbucketTestData(t, "pull-request.json"))
	})
}

func newBitbucketPullRequestNotFoundHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/test-repository/pullrequests/999", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "text/plain")
		w.Write(readBitbucketTestData(t, "pull-request-not-found.txt"))
	})
}

func newBitbucketPullRequestCommitsHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/test-repository/pullrequests/1/commits", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(readBitbucketTestData(t, "pull-request-commits.json"))
	})
}

func newBitbucketPullRequestCommitsNotFoundHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/test-repository/pullrequests/999/commits", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "text/plain")
		w.Write(readBitbucketTestData(t, "pull-request-not-found.txt"))
	})
}

func newBitbucketPullRequestDiffHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/test-repository/pullrequests/1/diff", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain")
		w.Write(readBitbucketTestData(t, "pull-request-diff.txt"))
	})
}

func newBitbucketPullRequestDiffNotFoundHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/test-repository/pullrequests/999/diff", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "text/plain")
		w.Write(readBitbucketTestData(t, "pull-request-not-found.txt"))
	})
}

func newBitbucketPullRequestCommentsHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/test-repository/pullrequests/1/comments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(readBitbucketTestData(t, "pull-request-comments.json"))
	})
}

func newBitbucketPullRequestCommentsNotFoundHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/test-repository/pullrequests/999/comments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "text/plain")
		w.Write(readBitbucketTestData(t, "pull-request-not-found.txt"))
	})
}
