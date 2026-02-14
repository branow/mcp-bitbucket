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

// PullRequests implements the TemplateProvider interface
// for listing Bitbucket pull requests.
type PullRequests struct {
	bitbucket *bb.Service
	template  string
	uriParser *util.UriTemplateParser
}

// NewPullRequests creates a new provider for listing pull requests.
// The provider supports the URI template:
// bitbucket://api/{workspace}/repositories/{repository}/pullrequests{?state,author,reviewer,page,pageSize}
//
// Parameters:
//   - bitbucket: The Bitbucket service for making API requests
//
// Returns a configured PullRequestsProvider.
func NewPullRequests(bitbucket *bb.Service) (*PullRequests, error) {
	template := "bitbucket://api/{workspace}/repositories/{repository}/pullrequests{?state,author,reviewer,page,pageSize}"
	parser, err := util.NewUriTemplateParser(template)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URI template %w", err)
	}

	return &PullRequests{
		bitbucket: bitbucket,
		template:  template,
		uriParser: parser,
	}, nil
}

// GetDefinition returns the MCP resource template definition for listing pull requests.
// The template includes URI pattern, title, description, and MIME type.
func (p *PullRequests) GetDefinition() *mcp.ResourceTemplate {
	return &mcp.ResourceTemplate{
		Name:        "pull-pequests",
		URITemplate: p.template,
		Title:       "List Pull Requests",
		Description: "List pull requests in a repository filtered by state",
		MIMEType:    string(web.MimeApplicationJson),
	}
}

// Handler processes read resource requests for listing pull requests.
// It parses and validates the URI parameters, calls the Bitbucket service,
// and returns the pull requests as JSON.
//
// URI Parameters:
//   - workspace: The workspace slug or username (required, must not be blank)
//   - repository: The repository name/slug (required, must not be blank)
//   - state: Comma-separated list of states (optional, defaults to OPEN)
//   - page: The page number (optional, defaults to 1, must be positive)
//   - pageSize: The number of items per page (optional, defaults to 25, must be positive)
//
// Returns
//   - ReadResourceResult containing the list of pull requests as JSON
//   - InvalidParamsError if URI parsing or validation fails
//   - ResourceNotFoundError if the repository doesn't exist
//   - InternalError if internal logic fails
func (p *PullRequests) Handler(
	ctx context.Context,
	req *mcp.ReadResourceRequest,
) (*mcp.ReadResourceResult, error) {

	params, err := p.uriParser.Parse(req.Params.URI)
	if err != nil {
		return nil, util.NewInvalidParamsError("uri: " + err.Error())
	}

	workspace := params.Path["workspace"]
	repository := params.Path["repository"]
	page := sch.Int().Optional(0).Parse(params.Query["page"])
	size := sch.Int().Optional(0).Parse(params.Query["pageSize"])
	state := sch.List(",").Optional([]string{}).Parse(params.Query["state"])

	res, err := p.bitbucket.ListPullRequests(ctx, bb.ListPullRequestsOptions{
		Workspace:  workspace,
		Repository: repository,
		State:      state,
		Page:       page,
		PageSize:   size,
	})
	if err != nil {
		return nil, err
	}

	bytes, err := json.Marshal(res)
	if err != nil {
		slog.Error("Failed to marshal list pull requests response", util.NewLogArgsExtractor().AddPlace("template:pull-requests").AddError(err).Extract()...,
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
