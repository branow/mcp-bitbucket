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

// FileContent implements the TemplateProvider interface
// for retrieving file content from a Bitbucket repository.
type FileContent struct {
	bitbucket *bb.Service
	template  string
	uriParser *util.UriTemplateParser
}

// NewFileContent creates a new provider for retrieving file content.
// The provider supports the URI template:
// bitbucket://api/{workspace}/repositories/{repository}/src/{+path}?ref={ref}
//
// Parameters:
//   - bitbucket: The Bitbucket service for making API requests
//
// Returns a configured FileContentProvider.
func NewFileContent(bitbucket *bb.Service) (*FileContent, error) {
	template := "bitbucket://api/{workspace}/repositories/{repository}/src/{+path}{?ref}"
	parser, err := util.NewUriTemplateParser(template)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URI template %w", err)
	}

	return &FileContent{
		bitbucket: bitbucket,
		template:  template,
		uriParser: parser,
	}, nil
}

// GetDefinition returns the MCP resource template definition for retrieving file content.
// The template includes URI pattern, title, description, and MIME type.
func (p *FileContent) GetDefinition() *mcp.ResourceTemplate {
	return &mcp.ResourceTemplate{
		Name:        "file-content",
		URITemplate: p.template,
		Title:       "File Content",
		Description: "Get file content from a repository by path and optional branch/commit ref",
		MIMEType:    string(web.MimeApplicationJson),
	}
}

// Handler processes read resource requests for retrieving file content.
// It parses and validates the URI parameters, calls the Bitbucket service,
// and returns the file content as JSON.
//
// URI Parameters:
//   - workspace: The workspace slug or username (required, from path)
//   - repository: The repository name/slug (required, from path)
//   - path: File path relative to repository root (required, from path with reserved expansion)
//   - ref: Branch name, tag, or commit hash (optional, from query, defaults to main branch)
//
// Returns:
//   - ReadResourceResult containing the file content as JSON
//   - InvalidParamsError if URI parsing or validation fails
//   - ResourceNotFoundError if the file doesn't exist
//   - InternalError if internal logic fails
func (p *FileContent) Handler(
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

	res, err := p.bitbucket.GetFileContent(ctx, bb.GetFileContentOptions{
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
		slog.Error("Failed to marshal get file content response", util.NewLogArgsExtractor().AddPlace("template:file-content").AddError(err).Extract()...)
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
