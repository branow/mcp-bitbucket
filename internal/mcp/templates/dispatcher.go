package templates

import (
	"context"
	"fmt"

	bb "github.com/branow/mcp-bitbucket/internal/bitbucket"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TemplateProvider defines the interface for MCP resource template providers.
// Implementations must provide both the template definition and a handler
// for reading resources.
type TemplateProvider interface {
	GetDefinition() *mcp.ResourceTemplate
	Handler(context.Context, *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error)
}

// TemplateDispatcher manages multiple resource template providers
// and registers them with an MCP server.
type TemplateDispatcher struct {
	providers []TemplateProvider
}

// NewDispatcher creates a new dispatcher with all available template providers.
// Currently includes repositories, repository, and pull request providers.
//
// Parameters:
//   - bitbucket: The Bitbucket service used by resource providers
//
// Returns a dispatcher ready to register resource templates with an MCP server.
func NewDispatcher(bitbucket *bb.Service) (*TemplateDispatcher, error) {
	type constructor func() (TemplateProvider, error)

	configs := []struct {
		nameFn string
		make   constructor
	}{
		{"repositories", func() (TemplateProvider, error) { return NewRepositories(bitbucket) }},
		{"repository", func() (TemplateProvider, error) { return NewRepository(bitbucket) }},
		{"pull request", func() (TemplateProvider, error) { return NewPullRequest(bitbucket) }},
		{"file content", func() (TemplateProvider, error) { return NewFileContent(bitbucket) }},
		{"directory source", func() (TemplateProvider, error) { return NewDirectorySource(bitbucket) }},
	}

	var providers []TemplateProvider
	for _, cfg := range configs {
		p, err := cfg.make()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize %s template provider: %w", cfg.nameFn, err)
		}
		providers = append(providers, p)
	}

	return &TemplateDispatcher{providers: providers}, nil
}

// Dispatch registers all resource template providers with the given MCP server.
// Each provider's template definition and handler are added to the server.
func (d *TemplateDispatcher) Dispatch(server *mcp.Server) {
	for _, provider := range d.providers {
		server.AddResourceTemplate(provider.GetDefinition(), provider.Handler)
	}
}
