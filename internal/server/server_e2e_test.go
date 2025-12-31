package server_test

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
	"github.com/branow/mcp-bitbucket/internal/util"
	"github.com/branow/mcp-bitbucket/internal/util/web"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/suite"
)

// E2ETestSuite is the test suite for end-to-end tests
type E2ETestSuite struct {
	suite.Suite
	baseURL    string
	mcpClient  *mcp.ClientSession
	httpClient *http.Client
	server     *server.McpServer
	bitbucket  *httptest.Server
}

func TestE2E(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}

func (s *E2ETestSuite) SetupSuite() {
	s.SetupBitbucketServer()
	s.SetupMcpServer()
	s.SetupMcpClient()
}

func (s *E2ETestSuite) SetupBitbucketServer() {
	s.bitbucket = httptest.NewServer(s.createBitbucketHandler())
}

func (s *E2ETestSuite) SetupMcpServer() {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	s.Require().NoError(err, "failed to find available port")

	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	s.T().Setenv("SERVER_PORT", strconv.Itoa(port))
	s.T().Setenv("BITBUCKET_URL", s.bitbucket.URL)
	s.T().Setenv("BITBUCKET_EMAIL", "test@example.com")
	s.T().Setenv("BITBUCKET_API_TOKEN", "test_token")
	s.T().Setenv("BITBUCKET_TIMEOUT", "5")

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

func (s *E2ETestSuite) SetupMcpClient() {
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

func (s *E2ETestSuite) TearDownSuite() {
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

func (s *E2ETestSuite) TestHealthEndpoint() {
	url := s.baseURL + "/health"
	req, err := http.NewRequest("GET", url, nil)
	s.Require().NoError(err, "failed to create GET /health request")

	resp, err := s.httpClient.Do(req)
	s.Require().NoError(err, "failed to make GET /health request")
	s.Assert().Equal(http.StatusOK, resp.StatusCode)
	s.Assert().Equal("application/json", resp.Header.Get("Content-Type"))
	s.Assert().Equal(`{"status":"ok"}`, strings.TrimSpace(s.readResponseBody(resp)))
}

func (s *E2ETestSuite) TestMcpInitialize() {
	s.Assert().Equal("Bitbucket MCP", s.mcpClient.InitializeResult().ServerInfo.Title)
	s.Assert().Equal("1.0.0", s.mcpClient.InitializeResult().ServerInfo.Version)
}

func (s *E2ETestSuite) TestListRepositoriesResource() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	uri := "mcp://bitbucket/test-workspace/repositories?page=1&pageSize=50"
	result, err := s.mcpClient.ReadResource(ctx, &mcp.ReadResourceParams{URI: uri})
	s.Require().NoError(err, "failed to read repositories resource")
	s.Require().NotNil(result)
	s.Require().Len(result.Contents, 1)
	s.Assert().Equal(uri, result.Contents[0].URI)
	s.Assert().Equal("application/json", result.Contents[0].MIMEType)

	expectedData := s.readMcpServerTestData("list_repositories.json")
	s.Assert().JSONEq(string(expectedData), result.Contents[0].Text)
}

func (s *E2ETestSuite) TestListRepositoriesResource_NotFound() {
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

func (s *E2ETestSuite) createBitbucketHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/repositories/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) < 2 || pathParts[1] != "test-workspace" {
			w.WriteHeader(http.StatusNotFound)
			w.Write(s.readBitbucketTestData("not_found_error.json"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(s.readBitbucketTestData("list_repositories.json"))
	})

	return mux
}

func (s *E2ETestSuite) readBitbucketTestData(filename string) []byte {
	s.T().Helper()
	return s.readTestData(filepath.Join("bitbucket", filename))
}

func (s *E2ETestSuite) readMcpServerTestData(filename string) []byte {
	s.T().Helper()
	return s.readTestData(filepath.Join("mcpserver", filename))
}

func (s *E2ETestSuite) readTestData(filename string) []byte {
	s.T().Helper()

	data, err := os.ReadFile(filepath.Join("testdata", filename))
	s.Require().NoError(err, fmt.Sprintf("failed to read test data file %s", filename))
	return data
}

func (s *E2ETestSuite) readResponseBody(resp *http.Response) string {
	s.T().Helper()

	var body string
	err := web.ReadResponseText(resp, &body)
	s.Require().NoError(err, "failed to read response body")
	return body
}
