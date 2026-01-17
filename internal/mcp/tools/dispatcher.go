// Package tools provides MCP tool providers and dispatchers.
//
// This package defines the interface for tools and manages
// registering them with the MCP server.
package tools

import (
	"context"
	"fmt"

	bb "github.com/branow/mcp-bitbucket/internal/bitbucket"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ToolProvider defines the interface for MCP tool providers.
// Implementations must provide both the tool definition and a handler for
// calling tools.
type ToolProvider interface {
	GetDefinition() *mcp.Tool
	Handler(context.Context, *mcp.CallToolRequest) (*mcp.CallToolResult, error)
}

// ToolDispatcher manages multiple tool providers
// and registers them with an MCP server.
type ToolDispatcher struct {
	providers []ToolProvider
}

// NewDispatcher creates a new dispatcher with all available tool providers.
// Currently includes list_repositories, get_repository, and get_pull_request tools.
//
// Parameters:
//   - bitbucket: The Bitbucket service used by tool providers
//
// Returns a dispatcher ready to register tools with an MCP server.
func NewDispatcher(bitbucket *bb.Service) (*ToolDispatcher, error) {
	type constructor func() (ToolProvider, error)

	configs := []struct {
		nameFn string
		make   constructor
	}{
		{"repositories", func() (ToolProvider, error) { return NewListRepositories(bitbucket) }},
		{"repository", func() (ToolProvider, error) { return NewGetRepository(bitbucket) }},
		{"pull request", func() (ToolProvider, error) { return NewGetPullRequest(bitbucket) }},
	}

	var providers []ToolProvider
	for _, cfg := range configs {
		p, err := cfg.make()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize %s tool provider: %w", cfg.nameFn, err)
		}
		providers = append(providers, p)
	}

	return &ToolDispatcher{providers: providers}, nil
}

// Dispatch registers all tool providers with the given MCP server.
// Each provider's tool definition and handler are added to the server.
func (d *ToolDispatcher) Dispatch(server *mcp.Server) {
	for _, provider := range d.providers {
		server.AddTool(provider.GetDefinition(), provider.Handler)
	}
}
