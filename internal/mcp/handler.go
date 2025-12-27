// Package mcp provides the MCP (Model Context Protocol) server implementation for Bitbucket.
//
// This package sets up the MCP server with resource templates and handlers
// for interacting with Bitbucket repositories through the MCP protocol.
package mcp

import (
	"net/http"

	"github.com/branow/mcp-bitbucket/internal/bitbucket/service"
	"github.com/branow/mcp-bitbucket/internal/mcp/templates"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Dispatcher is an interface for components that register themselves with an MCP server.
type Dispatcher[T any] interface {
	Dispatch(*mcp.Server)
}

var handler *mcp.StreamableHTTPHandler

// NewHandler creates a new HTTP handler for the MCP server.
// It initializes the MCP server with Bitbucket integration and resource templates.
//
// Parameters:
//   - bitbucket: The Bitbucket service for making API requests
//
// Returns an HTTP handler function that can be used with an HTTP server.
func NewHandler(bitbucket *service.Service) http.HandlerFunc {
	server := mcp.NewServer(&mcp.Implementation{
		Title:   "Bitbucket MCP",
		Version: "1.0.0",
	}, nil)

	templates.NewResourceTemplateDispatcher(bitbucket).Dispatch(server)

	handler = mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return server
	}, nil)

	return func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	}
}
