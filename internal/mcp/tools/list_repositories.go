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

// ListRepositories provides the list_repositories tool for listing bitbucket
// repositories.
type ListRepositories struct {
	bitbucket *bb.Service
	input     *jsonschema.Schema
	output    *jsonschema.Schema
}

// NewListRepositories creates a new tool for listing repositories.
//
// Parameters:
//   - bitbucket: The Bitbucket service for making API requests
//
// Returns a configured ListRepositoriesTool.
func NewListRepositories(bitbucket *bb.Service) (*ListRepositories, error) {
	input, err := jsonschema.For[bb.ListRepositoriesOptions](nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate input schema: %w", err)
	}

	output, err := jsonschema.For[bb.Page[bb.Repository]](nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate output schema: %w", err)
	}

	return &ListRepositories{
		bitbucket: bitbucket,
		input:     input,
		output:    output,
	}, nil
}

// GetDefinition returns the MCP tool definition for listing repositories.
func (t *ListRepositories) GetDefinition() *mcp.Tool {
	return &mcp.Tool{
		Name:         "list_repositories",
		Title:        "List Repositories",
		Description:  "Retrieves a list of repositories from the configured Bitbucket workspace, including metadata such as repository name, slug, and visibility.",
		InputSchema:  t.input,
		OutputSchema: t.output,
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}
}

// Handler processes tool call requests for listing repositories.
// It validates parameters using schema utils, calls the Bitbucket service,
// and returns the repositories as structured output.
//
// Parameters:
//   - ctx: The request context
//   - req: The MCP call tool request
//   - options: The input parameters
//
// Returns:
//   - CallToolResult: nil when returning structured output
//   - Output: The paginated list of repositories
//   - error: Any error that occurred during processing
func (t *ListRepositories) Handler(
	ctx context.Context, req *mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {

	var options *bb.ListRepositoriesOptions
	if err := json.Unmarshal(req.Params.Arguments, &options); err != nil {
		return nil, util.NewInvalidParamsError("arguments do not conform to input schema: " + err.Error())
	}

	res, err := t.bitbucket.ListRepositories(ctx, *options)
	if err != nil {
		return nil, err
	}

	bytes, err := json.Marshal(res)
	if err != nil {
		slog.Error("Failed to marshal get repository response", util.NewLogArgsExtractor().AddPlace("tool:list_repositories").AddError(err).Extract()...)
		return nil, util.NewInternalError()
	}

	// for backwards competibily we should return both json object and text response
	return &mcp.CallToolResult{
		Content:           []mcp.Content{&mcp.TextContent{Text: string(bytes)}},
		StructuredContent: res,
	}, nil
}
