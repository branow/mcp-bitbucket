package templates

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	bb "github.com/branow/mcp-bitbucket/internal/bitbucket"
	"github.com/branow/mcp-bitbucket/internal/util"
	sch "github.com/branow/mcp-bitbucket/internal/util/schema"
	"github.com/branow/mcp-bitbucket/internal/util/web"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Repositories implements the TemplateProvider interface
// for listing Bitbucket repositories.
type Repositories struct {
	bitbucket *bb.Service
	template  string
	uriParser *util.UriTemplateParser
}

// NewRepositories creates a new provider for listing repositories.
// The provider supports the URI template:
// mcp://bitbucket/{workspace}/repositories?page={page}&pageSize={pageSize}
//
// Parameters:
//   - bitbucket: The Bitbucket service for making API requests
//
// Returns a configured ListRepositoriesProvider.
func NewRepositories(bitbucket *bb.Service) (*Repositories, error) {
	template := "bitbucket://api/{workspace}/repositories{?page,pageSize}"
	parser, err := util.NewUriTemplateParser(template)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URI template %w", err)
	}

	return &Repositories{
		bitbucket: bitbucket,
		template:  template,
		uriParser: parser,
	}, nil
}

// GetDefinition returns the MCP resource template definition for listing repositories.
// The template includes URI pattern, title, description, and MIME type.
func (p *Repositories) GetDefinition() *mcp.ResourceTemplate {
	return &mcp.ResourceTemplate{
		Name:        "repositories",
		URITemplate: p.template,
		Title:       "List Repositories",
		Description: "List repositories in a Bitbucket workspace.",
		MIMEType:    string(web.MimeApplicationJson),
	}
}

// Handler processes read resource requests for listing repositories.
// It parses and validates the URI parameters, calls the Bitbucket service,
// and returns the repositories as JSON.
//
// URI Parameters:
//   - workspace: The workspace slug or username (required, must not be blank)
//   - page: The page number (optional, defaults to 1, must be positive)
//   - pageSize: The number of items per page (optional, defaults to 50, must be positive)
//
// Returns:
//   - ReadResourceResult containing the list of repositories as JSON
//   - InvalidParamsError if URI parsing or validation fails
//   - ResourceNotFoundError if the workspace doesn't exist
//   - InternalError if internal logic fails
func (p *Repositories) Handler(
	ctx context.Context,
	req *mcp.ReadResourceRequest,
) (*mcp.ReadResourceResult, error) {

	params, err := p.uriParser.Parse(req.Params.URI)
	if err != nil {
		return nil, util.NewInvalidParamsError("uri: " + err.Error())
	}

	workspace := params.Path["workspace"]
	page := sch.Int().Optional(0).Parse(params.Query["page"])
	size := sch.Int().Optional(0).Parse(params.Query["pageSize"])

	res, err := p.bitbucket.ListRepositories(ctx, bb.ListRepositoriesOptions{
		Workspace: workspace,
		Page:      page,
		PageSize:  size,
	})
	if err != nil {
		return nil, err
	}

	bytes, err := json.Marshal(res)
	if err != nil {
		slog.Error("Failed to marshal list repositories response", util.NewLogArgsExtractor().AddPlace("template:repositories").AddError(err).Extract()...,
		)
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
