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

// Repository implements the TemplateProvider interface
// for retrieving a single Bitbucket repository with optional source listing and README.
type Repository struct {
	bitbucket *bb.Service
	template  string
	uriParser *util.UriTemplateParser
}

// NewRepository creates a new provider for retrieving a single repository.
// The provider supports the URI template:
// mcp://bitbucket/{workspace}/repositories/{repository}?src={src}&readme={readme}
//
// Parameters:
//   - bitbucket: The Bitbucket service for making API requests
//
// Returns a configured RepositoryProvider.
func NewRepository(bitbucket *bb.Service) (*Repository, error) {
	template := "bitbucket://api/{workspace}/repositories/{repository}{?src,readme}"
	parser, err := util.NewUriTemplateParser(template)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URI template %w", err)
	}

	return &Repository{
		bitbucket: bitbucket,
		template:  template,
		uriParser: parser,
	}, nil
}

// GetDefinition returns the MCP resource template definition for retrieving a repository.
// The template includes URI pattern, title, description, and MIME type.
func (p *Repository) GetDefinition() *mcp.ResourceTemplate {
	return &mcp.ResourceTemplate{
		Name:        "repository",
		URITemplate: p.template,
		Title:       "Repository",
		Description: "Retrieves a repository from the configured Bitbucket workspace, including metadata such as repository name, slug, and visibility. Optionally includes root-level source listing (src=true) and README file content (readme=true).",
		MIMEType:    string(web.MimeApplicationJson),
	}
}

// Handler processes read resource requests for retrieving a single repository.
// It parses and validates the URI parameters, calls the Bitbucket service,
// and returns the repository details as JSON.
//
// URI Parameters:
//   - workspace: The workspace slug or username (required, must not be blank)
//   - repository: The repository name/slug (required, must not be blank)
//   - src: Include root-level source listing (optional, defaults to false)
//   - readme: Include README file content (optional, defaults to false)
//
// Returns:
//   - ReadResourceResult containing the repository details as JSON
//   - InvalidParamsError if URI parsing or validation fails
//   - ResourceNotFoundError if the repository doesn't exist
//   - InternalError if internal logic fails
func (p *Repository) Handler(
	ctx context.Context,
	req *mcp.ReadResourceRequest,
) (*mcp.ReadResourceResult, error) {

	params, err := p.uriParser.Parse(req.Params.URI)
	if err != nil {
		return nil, util.NewInvalidParamsError("uri: " + err.Error())
	}

	workspace := params.Path["workspace"]
	repository := params.Path["repository"]
	src := sch.Bool().Optional(false).Parse(params.Query["src"])
	readme := sch.Bool().Optional(false).Parse(params.Query["readme"])

	res, err := p.bitbucket.GetRepository(ctx, bb.GetRepositoryOptions{
		Workspace:     workspace,
		Repository:    repository,
		IncludeSource: src,
		IncludeReadme: readme,
	})
	if err != nil {
		return nil, err
	}

	bytes, err := json.Marshal(res)
	if err != nil {
		slog.Error("Failed to marshal get repository response", util.NewLogArgsExtractor().AddPlace("template:repository").AddError(err).Extract()...)
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
