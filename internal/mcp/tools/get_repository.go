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

// GetRepository implements the ToolProvider interface
// for retrieving a single Bitbucket repository with optional source listing
// and README.
type GetRepository struct {
	bitbucket *bb.Service
	input     *jsonschema.Schema
	output    *jsonschema.Schema
}

// NewGetRepository creates a new tool for retrieving a single repository.
//
// Parameters:
//   - bitbucket: The Bitbucket service for making API requests
//
// Returns a configured GetRepositoryTool.
func NewGetRepository(bitbucket *bb.Service) (*GetRepository, error) {
	input, err := jsonschema.For[bb.GetRepositoryOptions](nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate input schema: %w", err)
	}

	output, err := jsonschema.For[bb.RepositoryDetails](nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate output schema: %w", err)
	}

	return &GetRepository{
		bitbucket: bitbucket,
		input:     input,
		output:    output,
	}, nil
}

// GetDefinition returns the MCP tool definition for retrieving a repository.
func (t *GetRepository) GetDefinition() *mcp.Tool {
	return &mcp.Tool{
		Name:         "get_repository",
		Title:        "Get Repository",
		Description:  "Retrieves a repository from the configured Bitbucket workspace, including metadata such as repository name, slug, and visibility. Optionally includes root-level source listing and README file content.",
		InputSchema:  t.input,
		OutputSchema: t.output,
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}
}

// Handler processes tool call requests for retrieving a single repository.
// It parses and validates the parameters, calls the Bitbucket service,
// and returns the repository details as JSON.
//
// Parameters:
//   - ctx: The request context
//   - req: The MCP call tool request
//
// Returns:
//   - CallToolResult containing the repository details as JSON
//   - InvalidParamsError if parameter parsing or validation fails
//   - ResourceNotFoundError if the repository doesn't exist
//   - InternalError if internal logic fails
func (t *GetRepository) Handler(
	ctx context.Context,
	req *mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {

	var options *bb.GetRepositoryOptions
	if err := json.Unmarshal(req.Params.Arguments, &options); err != nil {
		return nil, util.NewInvalidParamsError("arguments do not conform to input schema: " + err.Error())
	}

	res, err := t.bitbucket.GetRepository(ctx, *options)
	if err != nil {
		return nil, err
	}

	bytes, err := json.Marshal(res)
	if err != nil {
		slog.Error("Failed to marshal get repository response", util.NewLogArgsExtractor().AddPlace("tool:get_repository").AddError(err).Extract()...)
		return nil, util.NewInternalError()
	}

	// for backwards compatibility we should return both json object and text response
	return &mcp.CallToolResult{
		Content:           []mcp.Content{&mcp.TextContent{Text: string(bytes)}},
		StructuredContent: res,
	}, nil
}
