// Package templates provides MCP resource template providers and dispatchers.
//
// This package defines the interface for resource templates and manages
// registering them with the MCP server.
package templates

import (
	"context"

	bitbucket "github.com/branow/mcp-bitbucket/internal/bitbucket/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ResourceTemplateProvider defines the interface for MCP resource template providers.
// Implementations must provide both the template definition and a handler for reading resources.
type ResourceTemplateProvider interface {
	GetDefinition() *mcp.ResourceTemplate
	Handler(context.Context, *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error)
}

// ResourceTemplateDispatcher manages multiple resource template providers
// and registers them with an MCP server.
type ResourceTemplateDispatcher[T ResourceTemplateProvider] struct {
	providers []ResourceTemplateProvider
}

// NewResourceTemplateDispatcher creates a new dispatcher with all available resource template providers.
// Currently includes repositories, repository, and pull request providers.
//
// Parameters:
//   - bitbucket: The Bitbucket service used by resource providers
//
// Returns a dispatcher ready to register resource templates with an MCP server.
func NewResourceTemplateDispatcher(bitbucket *bitbucket.Service) *ResourceTemplateDispatcher[ResourceTemplateProvider] {
	return &ResourceTemplateDispatcher[ResourceTemplateProvider]{
		providers: []ResourceTemplateProvider{
			NewRepositoriesProvider(bitbucket),
			NewRepositoryProvider(bitbucket),
			NewPullRequestProvider(bitbucket),
		},
	}
}

// Dispatch registers all resource template providers with the given MCP server.
// Each provider's template definition and handler are added to the server.
func (d *ResourceTemplateDispatcher[T]) Dispatch(server *mcp.Server) {
	for _, provider := range d.providers {
		server.AddResourceTemplate(provider.GetDefinition(), provider.Handler)
	}
}
