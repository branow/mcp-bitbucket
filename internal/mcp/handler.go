package mcp

import (
	"fmt"
	"net/http"

	"github.com/branow/mcp-bitbucket/internal/auth"
	"github.com/branow/mcp-bitbucket/internal/bitbucket"
	"github.com/branow/mcp-bitbucket/internal/mcp/templates"
	"github.com/branow/mcp-bitbucket/internal/mcp/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Dispatcher is an interface for components that register themselves with
// an MCP server.
type Dispatcher[T any] interface {
	Dispatch(*mcp.Server)
}

// NewHandler creates a new HTTP handler for the MCP server.
// It initializes the MCP server with Bitbucket integration and tools.
//
// Parameters:
//   - bitbucket: The Bitbucket service for making API requests
//
// Returns an HTTP handler function that can be used with an HTTP server.
func NewHandler(
	bitbucket *bitbucket.Service,
	authorize auth.Middleware,
) (http.HandlerFunc, error) {

	server := mcp.NewServer(&mcp.Implementation{
		Title:   "Bitbucket MCP",
		Version: "1.0.0",
	}, nil)

	template, err := templates.NewDispatcher(bitbucket)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize templates dispatcher: %w", err)
	}
	tool, err := tools.NewDispatcher(bitbucket)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tools dispatcher: %w", err)
	}

	template.Dispatch(server)
	tool.Dispatch(server)

	mcpHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return server
	}, nil)

	return authorize(mcpHandler).ServeHTTP, nil
}
