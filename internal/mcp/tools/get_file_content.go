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

// GetFileContent implements the ToolProvider interface
// for retrieving file content from a Bitbucket repository.
type GetFileContent struct {
	bitbucket *bb.Service
	input     *jsonschema.Schema
	output    *jsonschema.Schema
}

// NewGetFileContent creates a new tool for retrieving file content.
//
// Parameters:
//   - bitbucket: The Bitbucket service for making API requests
//
// Returns a configured GetFileContentTool.
func NewGetFileContent(bitbucket *bb.Service) (*GetFileContent, error) {
	input, err := jsonschema.For[bb.GetFileContentOptions](nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate input schema: %w", err)
	}

	output, err := jsonschema.For[bb.FileContent](nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate output schema: %w", err)
	}

	return &GetFileContent{
		bitbucket: bitbucket,
		input:     input,
		output:    output,
	}, nil
}

// GetDefinition returns the MCP tool definition for retrieving file content.
func (t *GetFileContent) GetDefinition() *mcp.Tool {
	return &mcp.Tool{
		Name:         "get_file_content",
		Title:        "Get File Content",
		Description:  "Get file content from a repository by path and optional branch/commit ref",
		InputSchema:  t.input,
		OutputSchema: t.output,
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}
}

// Handler processes tool call requests for retrieving file content.
// It parses and validates the parameters, calls the Bitbucket service,
// and returns the file content as JSON.
//
// Parameters:
//   - ctx: The request context
//   - req: The MCP call tool request
//
// Returns:
//   - CallToolResult containing the file content as JSON
//   - InvalidParamsError if parameter parsing or validation fails
//   - ResourceNotFoundError if the file doesn't exist
//   - InternalError if internal logic fails
func (t *GetFileContent) Handler(
	ctx context.Context,
	req *mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {

	var options *bb.GetFileContentOptions
	if err := json.Unmarshal(req.Params.Arguments, &options); err != nil {
		return nil, util.NewInvalidParamsError("arguments do not conform to input schema: " + err.Error())
	}

	res, err := t.bitbucket.GetFileContent(ctx, *options)
	if err != nil {
		return nil, err
	}

	bytes, err := json.Marshal(res)
	if err != nil {
		slog.Error("Failed to marshal get file content response", util.NewLogArgsExtractor().AddPlace("tool:get_file_content").AddError(err).Extract()...)
		return nil, util.NewInternalError()
	}

	// for backwards compatibility we should return both json object and text response
	return &mcp.CallToolResult{
		Content:           []mcp.Content{&mcp.TextContent{Text: string(bytes)}},
		StructuredContent: res,
	}, nil
}
