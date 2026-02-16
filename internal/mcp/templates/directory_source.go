package templates

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	bb "github.com/branow/mcp-bitbucket/internal/bitbucket"
	"github.com/branow/mcp-bitbucket/internal/util"
	"github.com/branow/mcp-bitbucket/internal/util/web"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// DirectorySource implements the TemplateProvider interface
// for retrieving directory contents from a Bitbucket repository.
type DirectorySource struct {
	bitbucket *bb.Service
	template  string
	uriParser *util.UriTemplateParser
}

// NewDirectorySource creates a new provider for retrieving directory contents.
// The provider supports the URI template:
// bitbucket://api/{workspace}/repositories/{repository}/src/dir/{+path}?ref={ref}
//
// Parameters:
//   - bitbucket: The Bitbucket service for making API requests
//
// Returns a configured DirectorySourceProvider.
func NewDirectorySource(bitbucket *bb.Service) (*DirectorySource, error) {
	template := "bitbucket://api/{workspace}/repositories/{repository}/src/dir/{+path}{?ref}"
	parser, err := util.NewUriTemplateParser(template)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URI template %w", err)
	}

	return &DirectorySource{
		bitbucket: bitbucket,
		template:  template,
		uriParser: parser,
	}, nil
}

// GetDefinition returns the MCP resource template definition for retrieving directory contents.
// The template includes URI pattern, title, description, and MIME type.
func (p *DirectorySource) GetDefinition() *mcp.ResourceTemplate {
	return &mcp.ResourceTemplate{
		Name:        "directory-source",
		URITemplate: p.template,
		Title:       "Directory Source",
		Description: "Retrieves the contents of a directory in a Bitbucket repository, listing files and subdirectories at the specified path and optional branch/commit ref.",
		MIMEType:    string(web.MimeApplicationJson),
	}
}

// Handler processes read resource requests for retrieving directory contents.
// It parses and validates the URI parameters, calls the Bitbucket service,
// and returns the directory listing as JSON.
//
// URI Parameters:
//   - workspace: The workspace slug or username (required, from path)
//   - repository: The repository name/slug (required, from path)
//   - path: Directory path relative to repository root (required, from path with reserved expansion)
//   - ref: Branch name, tag, or commit hash (optional, from query, defaults to main branch)
//
// Returns:
//   - ReadResourceResult containing the directory listing as JSON
//   - InvalidParamsError if URI parsing or validation fails
//   - ResourceNotFoundError if the directory doesn't exist
//   - InternalError if internal logic fails
func (p *DirectorySource) Handler(
	ctx context.Context,
	req *mcp.ReadResourceRequest,
) (*mcp.ReadResourceResult, error) {

	params, err := p.uriParser.Parse(req.Params.URI)
	if err != nil {
		return nil, util.NewInvalidParamsError("uri: " + err.Error())
	}

	workspace := params.Path["workspace"]
	repository := params.Path["repository"]
	path := params.Path["path"]
	ref := params.Query["ref"]

	res, err := p.bitbucket.GetDirectorySource(ctx, bb.GetDirectorySourceOptions{
		Workspace:  workspace,
		Repository: repository,
		Path:       path,
		Ref:        ref,
	})
	if err != nil {
		return nil, err
	}

	bytes, err := json.Marshal(res)
	if err != nil {
		slog.Error("Failed to marshal get directory source response", util.NewLogArgsExtractor().AddPlace("template:directory-source").AddError(err).Extract()...)
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
