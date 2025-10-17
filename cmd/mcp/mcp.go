package mcp

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/branow/mcp-bitbucket/internal/health"
)

type McpServer struct {
	addr   string
	server *http.Server
	ready  chan struct{}
}

func NewMcpServer(addr string) *McpServer {
	return &McpServer{
		addr:  addr,
		ready: make(chan struct{}),
	}
}

func (s *McpServer) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", health.Handler)

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

func (s *McpServer) WaitUntilReady(timeout time.Duration) error {
	select {
	case <-s.ready:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("server failed to start within %v", timeout)
	}
}

func (s *McpServer) Shutdown(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}
