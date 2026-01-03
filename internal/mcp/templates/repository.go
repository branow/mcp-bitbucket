package templates

import (
	"context"
	"encoding/json"

	bitbucket "github.com/branow/mcp-bitbucket/internal/bitbucket/service"
	"github.com/branow/mcp-bitbucket/internal/util"
	sch "github.com/branow/mcp-bitbucket/internal/util/schema"
	"github.com/branow/mcp-bitbucket/internal/util/web"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RepositoryProvider implements the ResourceTemplateProvider interface
// for retrieving a single Bitbucket repository with optional source listing and README.
type RepositoryProvider struct {
	bitbucket *bitbucket.Service
	template  string
	uriParser *util.UriTemplateParser
}

// NewRepositoryProvider creates a new provider for retrieving a single repository.
// The provider supports the URI template:
// mcp://bitbucket/{namespace}/repositories/{repository}?src={src}&readme={readme}
//
// Parameters:
//   - bitbucket: The Bitbucket service for making API requests
//
// Returns a configured RepositoryProvider.
func NewRepositoryProvider(bitbucket *bitbucket.Service) *RepositoryProvider {
	template := "mcp://bitbucket/{namespace}/repositories/{repository}{?src,readme}"
	parser, err := util.NewUriTemplateParser(template)
	if err != nil {
		panic(err)
	}

	return &RepositoryProvider{
		bitbucket: bitbucket,
		template:  template,
		uriParser: parser,
	}
}

// GetDefinition returns the MCP resource template definition for retrieving a repository.
// The template includes URI pattern, title, description, and MIME type.
func (p *RepositoryProvider) GetDefinition() *mcp.ResourceTemplate {
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
//   - namespace: The workspace slug or username (required, must not be blank)
//   - repository: The repository name/slug (required, must not be blank)
//   - src: Include root-level source listing (optional, defaults to false)
//   - readme: Include README file content (optional, defaults to false)
//
// Returns:
//   - ReadResourceResult containing the repository details as JSON
//   - InvalidParamsError if URI parsing or validation fails
//   - ResourceNotFoundError if the repository doesn't exist
//   - InternalError if internal logic fails
func (p *RepositoryProvider) Handler(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	params, err := p.uriParser.Parse(req.Params.URI)
	if err != nil {
		return nil, util.NewInvalidParamsError(err.Error())
	}

	namespace, err := sch.String().Must(sch.NotBlank()).Parse(params.Path["namespace"])
	if err != nil {
		return nil, util.NewInvalidParamsError(err.Error())
	}

	repository, err := sch.String().Must(sch.NotBlank()).Parse(params.Path["repository"])
	if err != nil {
		return nil, util.NewInvalidParamsError(err.Error())
	}

	src := sch.Bool().Optional(false).Parse(params.Query["src"])
	readme := sch.Bool().Optional(false).Parse(params.Query["readme"])

	res, err := p.bitbucket.GetRepository(ctx, namespace, repository, bitbucket.GetRepositoryOptions{IncludeSource: src, IncludeReadme: readme})
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
