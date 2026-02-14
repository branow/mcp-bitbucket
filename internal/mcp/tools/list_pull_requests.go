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

// ListPullRequests provides the list_pull_requests tool for listing pull requests
// in a repository.
type ListPullRequests struct {
	bitbucket *bb.Service
	input     *jsonschema.Schema
	output    *jsonschema.Schema
}

// NewListPullRequests creates a new tool for listing pull requests.
//
// Parameters:
//   - bitbucket: The Bitbucket service for making API requests
//
// Returns a configured ListPullRequestsTool.
func NewListPullRequests(bitbucket *bb.Service) (*ListPullRequests, error) {
	input, err := jsonschema.For[bb.ListPullRequestsOptions](nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate input schema: %w", err)
	}

	output, err := jsonschema.For[bb.Page[bb.PullRequestSummary]](nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate output schema: %w", err)
	}

	return &ListPullRequests{
		bitbucket: bitbucket,
		input:     input,
		output:    output,
	}, nil
}

// GetDefinition returns the MCP tool definition for listing pull requests.
func (t *ListPullRequests) GetDefinition() *mcp.Tool {
	return &mcp.Tool{
		Name:         "list_pull_requests",
		Title:        "List Pull Requests",
		Description:  "List pull requests in a repository filtered by state",
		InputSchema:  t.input,
		OutputSchema: t.output,
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}
}

// Handler processes tool call requests for listing pull requests.
// It validates parameters using schema utils, calls the Bitbucket service,
// and returns the pull requests as structured output.
//
// Parameters:
//   - ctx: The request context
//   - req: The MCP call tool request
//
// Returns:
//   - CallToolResult: nil when returning structured output
//   - Output: The paginated list of pull requests
//   - error: Any error that occurred during processing
func (t *ListPullRequests) Handler(
	ctx context.Context, req *mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {

	var options *bb.ListPullRequestsOptions
	if err := json.Unmarshal(req.Params.Arguments, &options); err != nil {
		return nil, util.NewInvalidParamsError("arguments do not conform to input schema: " + err.Error())
	}

	res, err := t.bitbucket.ListPullRequests(ctx, *options)
	if err != nil {
		return nil, err
	}

	bytes, err := json.Marshal(res)
	if err != nil {
		slog.Error("Failed to marshal list pull requests response", util.NewLogArgsExtractor().AddPlace("tool:list_pull_requests").AddError(err).Extract()...)
		return nil, util.NewInternalError()
	}

	// for backwards competibily we should return both json object and text response
	return &mcp.CallToolResult{
		Content:           []mcp.Content{&mcp.TextContent{Text: string(bytes)}},
		StructuredContent: res,
	}, nil
}
