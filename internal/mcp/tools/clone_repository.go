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

// CloneRepository implements the ToolProvider interface for cloning a Bitbucket repository.
type CloneRepository struct {
	bitbucket *bb.Service
	input     *jsonschema.Schema
	output    *jsonschema.Schema
}

// NewCloneRepository creates a new tool for cloning a Bitbucket repository.
func NewCloneRepository(bitbucket *bb.Service) (*CloneRepository, error) {
	input, err := jsonschema.For[bb.CloneRepositoryOptions](nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate input schema: %w", err)
	}

	output, err := jsonschema.For[bb.CloneRepositoryResult](nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate output schema: %w", err)
	}

	return &CloneRepository{
		bitbucket: bitbucket,
		input:     input,
		output:    output,
	}, nil
}

// GetDefinition returns the MCP tool definition for cloning a repository.
func (t *CloneRepository) GetDefinition() *mcp.Tool {
	destructive := true
	return &mcp.Tool{
		Name:         "clone_repository",
		Title:        "Clone Repository",
		Description:  "Clone a Bitbucket repository to a local path using server-managed credentials.",
		InputSchema:  t.input,
		OutputSchema: t.output,
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: &destructive,
		},
	}
}

// Handler processes tool call requests for cloning a repository.
//
// Parameters:
//   - ctx: The request context, used for authentication and cancellation
//   - req: The MCP tool call request containing JSON-encoded CloneRepositoryOptions
//
// Returns a CallToolResult with the resolved absolute local path of the cloned
// repository, or an error if the arguments are invalid or the clone fails.
func (t *CloneRepository) Handler(
	ctx context.Context,
	req *mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {

	var options bb.CloneRepositoryOptions
	if err := json.Unmarshal(req.Params.Arguments, &options); err != nil {
		return nil, util.NewInvalidParamsError("arguments do not conform to input schema: " + err.Error())
	}

	res, err := t.bitbucket.CloneRepository(ctx, options)
	if err != nil {
		return nil, err
	}

	bytes, err := json.Marshal(res)
	if err != nil {
		slog.Error("Failed to marshal clone repository response", util.NewLogArgsExtractor().AddPlace("tool:clone_repository").AddError(err).Extract()...)
		return nil, util.NewInternalError()
	}

	return &mcp.CallToolResult{
		Content:           []mcp.Content{&mcp.TextContent{Text: string(bytes)}},
		StructuredContent: res,
	}, nil
}
