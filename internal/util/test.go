package util

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertJsonRpcError checks if the error is a jsonrpc.Error with the expected code
func AssertJsonRpcError(t *testing.T, err error, expectedCode int64, msgAndArgs ...interface{}) {
	t.Helper()
	var jsonrpcErr *jsonrpc.Error
	require.ErrorAs(t, err, &jsonrpcErr, msgAndArgs...)
	assert.Equal(t, expectedCode, jsonrpcErr.Code, msgAndArgs...)
}
