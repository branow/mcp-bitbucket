// Package server provides the HTTP server implementation for the MCP Bitbucket service.
//
// This package sets up the HTTP server with health check and MCP endpoints,
// and manages the server lifecycle including graceful shutdown.
package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/branow/mcp-bitbucket/internal/auth"
	"github.com/branow/mcp-bitbucket/internal/bitbucket"
	"github.com/branow/mcp-bitbucket/internal/config"
	"github.com/branow/mcp-bitbucket/internal/health"
	"github.com/branow/mcp-bitbucket/internal/mcp"
	"github.com/branow/mcp-bitbucket/internal/util"
)

// McpServer represents the HTTP server for the MCP Bitbucket service.
// It manages the server lifecycle and integrates authentication, health checks,
// and the MCP protocol endpoint.
type McpServer struct {
	addr      string
	server    *http.Server
	ready     chan struct{}
	bitbucket *bitbucket.Service
	cfg       config.Global
}

// NewMcpServer creates a new MCP server with the given configuration.
// It initializes the Bitbucket client with the configured authentication method
// (basic auth or OAuth) and sets up the service layer.
//
// Parameters:
//   - cfg: Global configuration containing server, authentication, and Bitbucket settings
//
// Returns a fully configured McpServer ready to be started with Run().
func NewMcpServer(cfg config.Global) *McpServer {
	bbClient := bitbucket.NewClient(cfg.Bitbucket, cfg.Auth.Authorizer())
	bbService := bitbucket.NewService(bbClient)

	return &McpServer{
		addr:      fmt.Sprintf("127.0.0.1:%d", cfg.Server.Port),
		ready:     make(chan struct{}),
		bitbucket: bbService,
		cfg:       cfg,
	}
}

// Run starts the HTTP server and begins listening for requests.
// This method blocks until the server is shut down or an error occurs.
//
// The server sets up authentication middleware based on the configured auth type:
//   - OAuth: Validates bearer tokens
//   - Basic: No middleware (authentication handled at API client level)
//
// The server exposes the following endpoints:
//   - /health: Health check endpoint (no authentication required)
//   - /mcp: MCP protocol endpoint for Bitbucket integration (authentication required)
//   - OAuth metadata endpoint: Serves OAuth resource metadata (only when OAuth is enabled)
//
// Returns an error if:
//   - Authentication middleware initialization fails
//   - The server fails to bind to the configured port
//   - The server encounters an error while running
func (s *McpServer) Run() error {
	authorize, err := auth.NewMiddleware(s.cfg.Auth)
	if err != nil {
		return fmt.Errorf("failed to create auth middleware: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", health.NewHandler())
	mux.HandleFunc("/mcp", mcp.NewHandler(s.bitbucket, authorize))
	if s.cfg.Auth.Type == util.OAuth {
		mux.HandleFunc(s.cfg.Auth.OAuth.ResourceMetadataPath, auth.NewOAuthHandler(s.cfg.Auth.OAuth))
	}

	s.server = &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	log.Println("Listening on", s.addr)
	close(s.ready)

	return s.server.Serve(listener)
}

// WaitUntilReady blocks until the server is ready to accept requests or the timeout expires.
// This is useful for testing or coordinating startup with other components.
//
// Parameters:
//   - timeout: Maximum time to wait for the server to become ready
//
// Returns an error if the timeout expires before the server is ready.
func (s *McpServer) WaitUntilReady(timeout time.Duration) error {
	select {
	case <-s.ready:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("server failed to start within %v", timeout)
	}
}

// Shutdown gracefully shuts down the server without interrupting active connections.
//
// Parameters:
//   - ctx: Context to control the shutdown timeout
//
// Returns an error if the shutdown fails or the context is canceled.
func (s *McpServer) Shutdown(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}
