package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/branow/mcp-bitbucket/internal/util/web"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetResource(t *testing.T, client *mcp.ClientSession, uri string, file string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := client.ReadResource(ctx, &mcp.ReadResourceParams{URI: uri})
	require.NoError(t, err, "failed to read resource")
	require.NotNil(t, result, "resource response must not be nil")
	require.Len(t, result.Contents, 1, "expected one content entry in resource response")

	content := result.Contents[0]
	assert.Equal(t, uri, content.URI)
	assert.Equal(t, "application/json", content.MIMEType)

	expectedData := ReadMcpServerFile(t, file)
	assert.JSONEq(t, string(expectedData), content.Text)
}

func TestGetResourceError(t *testing.T, client *mcp.ClientSession, uri string, code int64, error string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := client.ReadResource(ctx, &mcp.ReadResourceParams{URI: uri})
	require.Error(t, err, "expected read resource error")
	assert.Nil(t, result, "expected nil resource respond")

	var jsonrpcErr *jsonrpc.Error
	require.ErrorAs(t, err, &jsonrpcErr, "error must be a JSON-RPC error")
	assert.Equal(t, code, jsonrpcErr.Code, "unexpected error code")
	assert.Contains(t, jsonrpcErr.Message, error, "unexpected error message")
}

func TestCallTool(t *testing.T, client *mcp.ClientSession, name string, args any, file string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := client.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
	require.NoError(t, err, "failed to call tool")
	require.NotNil(t, result, "tool response must not be nil")

	require.Falsef(t, result.IsError, "tool respond with error")
	require.Len(t, result.Content, 1, "expected one content entry in tool response")

	expectedJson := ReadMcpServerFile(t, file)

	content := result.Content[0]
	textContent, ok := content.(*mcp.TextContent)
	require.True(t, ok, "expected TextContent, got %T: %v", content, content)
	assert.JSONEq(t, string(expectedJson), textContent.Text)

	actualStructured := result.StructuredContent
	require.NotNil(t, actualStructured, "expected structured content to be present")

	actualStrucutredJson, err := json.Marshal(result.StructuredContent)
	require.NoError(t, err, "failed to marshal structured content")
	assert.JSONEq(t, string(expectedJson), string(actualStrucutredJson))
}

func TestCallToolError(t *testing.T, client *mcp.ClientSession, name string, args any, code int64, error string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := client.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
	require.Errorf(t, err, "expected call tool error")
	assert.Nil(t, result, "expected nil resource respond")

	var jsonrpcErr *jsonrpc.Error
	require.ErrorAs(t, err, &jsonrpcErr, "error must be a JSON-RPC error")
	assert.Equal(t, code, jsonrpcErr.Code, "unexpected error code")
	assert.Contains(t, jsonrpcErr.Message, error, "unexpected error message")
}

func ReadMcpServerFile(t *testing.T, filename string) []byte {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("data", "mcpserver", filename))
	require.NoError(t, err, fmt.Sprintf("failed to read test data file %s", filename))
	return data
}

func ReadBitbucketFile(t *testing.T, filename string) []byte {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("data", "bitbucket", filename))
	require.NoError(t, err, fmt.Sprintf("failed to read test data file %s", filename))
	return data
}

func ReadResponseBody(t *testing.T, resp *http.Response) string {
	t.Helper()

	var body string
	err := web.ReadResponseText(resp, &body)
	require.NoError(t, err, "failed to read response body")
	return body
}
