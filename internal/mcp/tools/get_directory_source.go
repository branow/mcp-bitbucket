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

// GetDirectorySource implements the ToolProvider interface
// for retrieving directory contents from a Bitbucket repository.
type GetDirectorySource struct {
	bitbucket *bb.Service
	input     *jsonschema.Schema
	output    *jsonschema.Schema
}

// NewGetDirectorySource creates a new tool for retrieving directory contents.
//
// Parameters:
//   - bitbucket: The Bitbucket service for making API requests
//
// Returns a configured GetDirectorySourceTool.
func NewGetDirectorySource(bitbucket *bb.Service) (*GetDirectorySource, error) {
	input, err := jsonschema.For[bb.GetDirectorySourceOptions](nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate input schema: %w", err)
	}

	output, err := jsonschema.For[bb.Page[bb.SourceItem]](nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate output schema: %w", err)
	}

	return &GetDirectorySource{
		bitbucket: bitbucket,
		input:     input,
		output:    output,
	}, nil
}

// GetDefinition returns the MCP tool definition for retrieving directory contents.
func (t *GetDirectorySource) GetDefinition() *mcp.Tool {
	return &mcp.Tool{
		Name:         "get_directory_source",
		Title:        "Get Directory Source",
		Description:  "Retrieves the contents of a directory in a Bitbucket repository, listing files and subdirectories at the specified path and optional branch/commit ref.",
		InputSchema:  t.input,
		OutputSchema: t.output,
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}
}

// Handler processes tool call requests for retrieving directory contents.
// It parses and validates the parameters, calls the Bitbucket service,
// and returns the directory listing as JSON.
//
// Parameters:
//   - ctx: The request context
//   - req: The MCP call tool request
//
// Returns:
//   - CallToolResult containing the directory listing as JSON
//   - InvalidParamsError if parameter parsing or validation fails
//   - ResourceNotFoundError if the directory doesn't exist
//   - InternalError if internal logic fails
func (t *GetDirectorySource) Handler(
	ctx context.Context,
	req *mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {

	var options *bb.GetDirectorySourceOptions
	if err := json.Unmarshal(req.Params.Arguments, &options); err != nil {
		return nil, util.NewInvalidParamsError("arguments do not conform to input schema: " + err.Error())
	}

	res, err := t.bitbucket.GetDirectorySource(ctx, *options)
	if err != nil {
		return nil, err
	}

	bytes, err := json.Marshal(res)
	if err != nil {
		slog.Error("Failed to marshal get directory source response", util.NewLogArgsExtractor().AddPlace("tool:get_directory_source").AddError(err).Extract()...)
		return nil, util.NewInternalError()
	}

	// for backwards compatibility we should return both json object and text response
	return &mcp.CallToolResult{
		Content:           []mcp.Content{&mcp.TextContent{Text: string(bytes)}},
		StructuredContent: res,
	}, nil
}
