package mcp_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/branow/mcp-bitbucket/cmd/mcp"
	"github.com/stretchr/testify/suite"
)

// E2ETestSuite is the test suite for end-to-end tests
type E2ETestSuite struct {
	suite.Suite
	server  *mcp.McpServer
	baseURL string
	client  *http.Client
}

func TestE2E(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}

func (s *E2ETestSuite) SetupSuite() {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	s.Require().NoError(err, "failed to find available port")

	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	s.server = mcp.NewMcpServer(addr)
	s.baseURL = fmt.Sprintf("http://%s", addr)
	s.client = &http.Client{Timeout: 5 * time.Second}

	go func() {
		if err := s.server.Run(); err != nil && err != http.ErrServerClosed {
			s.T().Logf("server error: %v", err)
		}
	}()

	s.Require().NoError(s.server.WaitUntilReady(5*time.Second), "server failed to start")
}

func (s *E2ETestSuite) TearDownSuite() {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.Require().NoError(s.server.Shutdown(ctx), "failed to shutdown server")
	}
}

func (s *E2ETestSuite) TestHealthEndpoint() {
	resp := s.get("/health")
	s.Assert().Equal(http.StatusOK, resp.StatusCode)
	s.Assert().Equal("application/json", resp.Header.Get("Content-Type"))
	s.Assert().Equal(`{"status":"ok"}`, strings.TrimSpace(s.readBody(resp)))
}

func (s *E2ETestSuite) get(path string) *http.Response {
	return s.do("GET", path, nil)
}

func (s *E2ETestSuite) do(method string, path string, body io.Reader) *http.Response {
	url := s.baseURL + path
	req, err := http.NewRequest(method, url, body)
	s.Require().NoError(err, fmt.Sprintf("failed to create %v:%v request", method, url))
	resp, err := s.client.Do(req)
	s.Require().NoError(err, fmt.Sprintf("failed to make %v:%v request", method, url))
	return resp
}

func (s *E2ETestSuite) readBody(resp *http.Response) string {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	s.Require().NoError(err, "failed to read response body")
	return string(body)
}
