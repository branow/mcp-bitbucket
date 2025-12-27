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

	"github.com/branow/mcp-bitbucket/internal/bitbucket/client"
	"github.com/branow/mcp-bitbucket/internal/bitbucket/service"
	"github.com/branow/mcp-bitbucket/internal/health"
	"github.com/branow/mcp-bitbucket/internal/mcp"
)

// Config provides configuration for the MCP server and Bitbucket client.
type Config interface {
	ServerPort() int
	client.Config
}

// McpServer represents the HTTP server for the MCP Bitbucket.
type McpServer struct {
	addr      string
	server    *http.Server
	ready     chan struct{}
	bitbucket *service.Service
}

// NewMcpServer creates a new MCP server with the given configuration.
// It initializes the Bitbucket client and service layer.
func NewMcpServer(cfg Config) *McpServer {
	bbClient := client.NewClient(cfg)
	bbService := service.NewService(bbClient)

	return &McpServer{
		addr:      fmt.Sprintf("127.0.0.1:%d", cfg.ServerPort()),
		ready:     make(chan struct{}),
		bitbucket: bbService,
	}
}

// Run starts the HTTP server and begins listening for requests.
// This method blocks until the server is shut down or an error occurs.
//
// The server exposes two endpoints:
//   - /health: Health check endpoint
//   - /mcp: MCP protocol endpoint for Bitbucket integration
//
// Returns an error if the server fails to start or encounters an error while running.
func (s *McpServer) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", health.NewHandler())
	mux.HandleFunc("/mcp", mcp.NewHandler(s.bitbucket))

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
