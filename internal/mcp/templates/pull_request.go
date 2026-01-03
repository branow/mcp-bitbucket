package templates

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/branow/mcp-bitbucket/internal/bitbucket"
	"github.com/branow/mcp-bitbucket/internal/util"
	sch "github.com/branow/mcp-bitbucket/internal/util/schema"
	"github.com/branow/mcp-bitbucket/internal/util/web"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// PullRequestProvider implements the ResourceTemplateProvider interface
// for retrieving a single Bitbucket pull request with optional commits, diff, and comments.
type PullRequestProvider struct {
	bitbucket *bitbucket.Service
	template  string
	uriParser *util.UriTemplateParser
}

// NewPullRequestProvider creates a new provider for retrieving a single pull request.
// The provider supports the URI template:
// mcp://bitbucket/{namespace}/repositories/{repository}/pullrequests/{pullRequestId}?commits={commits}&diff={diff}&comments={comments}
//
// Parameters:
//   - bitbucket: The Bitbucket service for making API requests
//
// Returns a configured PullRequestProvider.
func NewPullRequestProvider(bitbucket *bitbucket.Service) *PullRequestProvider {
	template := "mcp://bitbucket/{namespace}/repositories/{repository}/pullrequests/{pullRequestId}{?commits,diff,comments}"
	parser, err := util.NewUriTemplateParser(template)
	if err != nil {
		panic(err)
	}

	return &PullRequestProvider{
		bitbucket: bitbucket,
		template:  template,
		uriParser: parser,
	}
}

// GetDefinition returns the MCP resource template definition for retrieving a pull request.
// The template includes URI pattern, title, description, and MIME type.
func (p *PullRequestProvider) GetDefinition() *mcp.ResourceTemplate {
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
//   - namespace: The workspace slug or username (required, must not be blank)
//   - repository: The repository name/slug (required, must not be blank)
//   - pullRequestId: The pull request ID (required, must be positive)
//   - commits: Include commits (optional, defaults to false)
//   - diff: Include diff (optional, defaults to false)
//   - comments: Include comments (optional, defaults to false)
//
// Returns:
//   - ReadResourceResult containing the pull request details as JSON
//   - InvalidParamsError if URI parsing or validation fails
//   - ResourceNotFoundError if the pull request doesn't exist
//   - InternalError if internal logic fails
func (p *PullRequestProvider) Handler(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
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

	pullRequestId, err := sch.Int().Must(sch.Positive()).Parse(params.Path["pullRequestId"])
	if err != nil {
		return nil, util.NewInvalidParamsError(fmt.Sprintf("pullRequestId: %s", err.Error()))
	}

	commits := sch.Bool().Optional(false).Parse(params.Query["commits"])
	diff := sch.Bool().Optional(false).Parse(params.Query["diff"])
	comments := sch.Bool().Optional(false).Parse(params.Query["comments"])

	res, err := p.bitbucket.GetPullRequest(ctx, namespace, repository, pullRequestId, bitbucket.GetPullRequestOptions{
		IncludeCommits:  commits,
		IncludeDiff:     diff,
		IncludeComments: comments,
	})
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
