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

// PullRequest implements the TemplateProvider interface
// for retrieving a single Bitbucket pull request with optional commits, diff,
// and comments.
type PullRequest struct {
	bitbucket *bb.Service
	template  string
	uriParser *util.UriTemplateParser
}

// NewPullRequest creates a new provider for retrieving a single pull request.
// The provider supports the URI template:
// mcp://bitbucket/{workspace}/repositories/{repository}/pullrequests/{id}?commits={commits}&diff={diff}&comments={comments}
//
// Parameters:
//   - bitbucket: The Bitbucket service for making API requests
//
// Returns a configured PullRequestProvider.
func NewPullRequest(bitbucket *bb.Service) (*PullRequest, error) {
	template := "bitbucket://api/{workspace}/repositories/{repository}/pullrequests/{id}{?commits,diff,comments}"
	parser, err := util.NewUriTemplateParser(template)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URI template %w", err)
	}

	return &PullRequest{
		bitbucket: bitbucket,
		template:  template,
		uriParser: parser,
	}, nil
}

// GetDefinition returns the MCP resource template definition for retrieving
// a pull request.
// The template includes URI pattern, title, description, and MIME type.
func (p *PullRequest) GetDefinition() *mcp.ResourceTemplate {
	return &mcp.ResourceTemplate{
		Name:        "pullRequest",
		URITemplate: p.template,
		Title:       "Pull Request",
		Description: "Retrieves a pull request from the configured Bitbucket workspace, including metadata such as title, state, and reviewers. Optionally includes commits (commits=true), diff (diff=true), and comments (comments=true).",
		MIMEType:    string(web.MimeApplicationJson),
	}
}

// Handler processes read resource requests for retrieving a single pull request.
// It parses and validates the URI parameters, calls the Bitbucket service,
// and returns the pull request details as JSON.
//
// URI Parameters:
//   - workspace: The workspace slug or username (required, must not be blank)
//   - repository: The repository name/slug (required, must not be blank)
//   - id: The pull request ID (required, must be positive)
//   - commits: Include commits (optional, defaults to false)
//   - diff: Include diff (optional, defaults to false)
//   - comments: Include comments (optional, defaults to false)
//
// Returns:
//   - ReadResourceResult containing the pull request details as JSON
//   - InvalidParamsError if URI parsing or validation fails
//   - ResourceNotFoundError if the pull request doesn't exist
//   - InternalError if internal logic fails
func (p *PullRequest) Handler(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	params, err := p.uriParser.Parse(req.Params.URI)
	if err != nil {
		return nil, util.NewInvalidParamsError("uri: " + err.Error())
	}

	workspace := params.Path["workspace"]
	repository := params.Path["repository"]

	id, err := sch.Int().Parse(params.Path["id"])
	if err != nil {
		return nil, util.NewInvalidParamsError("id: " + err.Error())
	}

	commits := sch.Bool().Optional(false).Parse(params.Query["commits"])
	diff := sch.Bool().Optional(false).Parse(params.Query["diff"])
	comments := sch.Bool().Optional(false).Parse(params.Query["comments"])

	res, err := p.bitbucket.GetPullRequest(ctx, bb.GetPullRequestOptions{
		Workspace:       workspace,
		Repository:      repository,
		Id:              id,
		IncludeCommits:  commits,
		IncludeDiff:     diff,
		IncludeComments: comments,
	})
	if err != nil {
		return nil, err
	}

	bytes, err := json.Marshal(res)
	if err != nil {
		slog.Error("Failed to marshal get pull request response",
			util.NewLogArgsExtractor().AddPlace("template:pull_request").AddError(err).Extract()...,
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
