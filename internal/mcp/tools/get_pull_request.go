package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	bb "github.com/branow/mcp-bitbucket/internal/bitbucket"
	"github.com/branow/mcp-bitbucket/internal/util"
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetPullRequest implements the ToolProvider interface
// for retrieving a single Bitbucket pull request with optional commits, diff,
// and comments.
type GetPullRequest struct {
	bitbucket *bb.Service
	input     *jsonschema.Schema
	output    *jsonschema.Schema
}

// NewGetPullRequest creates a new tool for retrieving a single pull request.
//
// Parameters:
//   - bitbucket: The Bitbucket service for making API requests
//
// Returns a configured GetPullRequestTool.
func NewGetPullRequest(bitbucket *bb.Service) (*GetPullRequest, error) {
	input, err := jsonschema.For[bb.GetPullRequestOptions](nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate input schema: %w", err)
	}

	output, err := jsonschema.For[bb.PullRequestDetails](nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate output schema: %w", err)
	}

	return &GetPullRequest{
		bitbucket: bitbucket,
		input:     input,
		output:    output,
	}, nil
}

// GetDefinition returns the MCP tool definition for retrieving a pull request.
func (t *GetPullRequest) GetDefinition() *mcp.Tool {
	return &mcp.Tool{
		Name:         "get_pull_request",
		Title:        "Get Pull Request",
		Description:  "Retrieves a pull request from the configured Bitbucket workspace, including metadata such as title, state, and reviewers. Optionally includes commits, diff, and comments.",
		InputSchema:  t.input,
		OutputSchema: t.output,
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}
}

// Handler processes tool call requests for retrieving a single pull request.
// It parses and validates the parameters, calls the Bitbucket service,
// and returns the pull request details as JSON.
//
// Parameters:
//   - ctx: The request context
//   - req: The MCP call tool request
//
// Returns:
//   - CallToolResult containing the pull request details as JSON
//   - InvalidParamsError if parameter parsing or validation fails
//   - ResourceNotFoundError if the pull request doesn't exist
//   - InternalError if internal logic fails
func (t *GetPullRequest) Handler(
	ctx context.Context,
	req *mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {

	var options *bb.GetPullRequestOptions
	if err := json.Unmarshal(req.Params.Arguments, &options); err != nil {
		return nil, util.NewInvalidParamsError("arguments do not conform to input schema: " + err.Error())
	}

	res, err := t.bitbucket.GetPullRequest(ctx, *options)
	if err != nil {
		return nil, err
	}

	bytes, err := json.Marshal(res)
	if err != nil {
		slog.Error("Failed to marshal get pull request response", util.NewLogArgsExtractor().AddPlace("tool:get_pull_request").AddError(err).Extract()...)
		return nil, util.NewInternalError()
	}

	// for backwards compatibility we should return both json object and text response
	return &mcp.CallToolResult{
		Content:           []mcp.Content{&mcp.TextContent{Text: string(bytes)}},
		StructuredContent: res,
	}, nil
}
