package templates

import (
	"context"
	"encoding/json"

	bitbucket "github.com/branow/mcp-bitbucket/internal/bitbucket/service"
	"github.com/branow/mcp-bitbucket/internal/util"
	"github.com/branow/mcp-bitbucket/internal/util/schema"
	"github.com/branow/mcp-bitbucket/internal/util/web"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListRepositoriesProvider implements the ResourceTemplateProvider interface
// for listing Bitbucket repositories.
type ListRepositoriesProvider struct {
	bitbucket *bitbucket.Service
	template  string
	uriParser *util.UriTemplateParser
}

// NewListRepositoriesProvider creates a new provider for listing repositories.
// The provider supports the URI template:
// mcp://bitbucket/{namespace}/repositories?page={page}&pageSize={pageSize}
//
// Parameters:
//   - bitbucket: The Bitbucket service for making API requests
//
// Returns a configured ListRepositoriesProvider.
func NewListRepositoriesProvider(bitbucket *bitbucket.Service) *ListRepositoriesProvider {
	template := "mcp://bitbucket/{namespace}/repositories?page={page}&pageSize={pageSize}"
	parser, err := util.NewUriTemplateParser(template)
	if err != nil {
		panic(err)
	}

	return &ListRepositoriesProvider{
		bitbucket: bitbucket,
		template:  template,
		uriParser: parser,
	}
}

// GetDefinition returns the MCP resource template definition for listing repositories.
// The template includes URI pattern, title, description, and MIME type.
func (p *ListRepositoriesProvider) GetDefinition() *mcp.ResourceTemplate {
	return &mcp.ResourceTemplate{
		Name:        "repositories",
		URITemplate: p.template,
		Title:       "List Repositories",
		Description: "Retrieves a list of repositories from the configured Bitbucket workspace, including metadata such as repository name, slug, and visibility.",
		MIMEType:    string(web.MimeApplicationJson),
	}
}

// Handler processes read resource requests for listing repositories.
// It parses and validates the URI parameters, calls the Bitbucket service,
// and returns the repositories as JSON.
//
// URI Parameters:
//   - namespace: The workspace slug or username (required, must not be blank)
//   - page: The page number (optional, defaults to 1, must be positive)
//   - pageSize: The number of items per page (optional, defaults to 50, must be positive)
//
// Returns:
//   - ReadResourceResult containing the list of repositories as JSON
//   - InvalidParamsError if URI parsing or validation fails
//   - ResourceNotFoundError if the namespace doesn't exist
//   - InternalError if internal logic fails
func (p *ListRepositoriesProvider) Handler(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	params, err := p.uriParser.Parse(req.Params.URI)
	if err != nil {
		return nil, util.NewInvalidParamsError(err.Error())
	}

	namespace, err := schema.String().Must(schema.NotBlank()).Parse(params.Path["namespace"])
	if err != nil {
		return nil, util.NewInvalidParamsError(err.Error())
	}

	page := schema.Int().Must(schema.Positive()).Optional(1).Parse(params.Query["page"])
	size := schema.Int().Must(schema.Positive()).Optional(50).Parse(params.Query["pageSize"])

	res, err := p.bitbucket.ListRepositories(ctx, namespace, page, size)
	if err != nil {
		return nil, err
	}

	bytes, err := json.Marshal(res)
	if err != nil {
		return nil, util.NewInternalError()
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      req.Params.URI,
				MIMEType: string(web.MimeApplicationJson),
				Text:     string(bytes),
			},
		},
	}, nil
}
