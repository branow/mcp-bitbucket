// Package util provides utility functions and error handling for the MCP Bitbucket server.
//
// This package defines standard JSON-RPC error codes and provides helper functions
// for creating consistent error responses throughout the application.
package util

import "github.com/modelcontextprotocol/go-sdk/jsonrpc"

// JSON-RPC error codes used throughout the application.
const (
  CodeInvalidParamsErr int64 = jsonrpc.CodeInvalidParams
  CodeResourceNotFoundErr int64 = -32002
  CodeResourceUnavailableErr int64 = -32802
  CodeInternalErr int64 = jsonrpc.CodeInternalError
)

// NewInvalidParamsError creates a JSON-RPC error for invalid parameters.
// This should be used when request parameters fail validation.
func NewInvalidParamsError(message string) error {
  return &jsonrpc.Error{
    Code:    CodeInvalidParamsErr,
    Message: message,
  }
}

// NewResourceNotFoundError creates a JSON-RPC error for missing resources.
// This should be used when a requested resource (repository, PR, file, etc.) is not found.
func NewResourceNotFoundError(message string) error {
  return &jsonrpc.Error{
    Code:    CodeResourceNotFoundErr,
    Message: message,
  }
}

// NewResourceUnavailableError creates a JSON-RPC error for unavailable resources.
// This should be used when a service or resource is temporarily unavailable (e.g., 5xx HTTP errors).
func NewResourceUnavailableError(message string) error {
  return &jsonrpc.Error{
    Code:    CodeResourceUnavailableErr,
    Message: message,
  }
}

// NewInternalError creates a JSON-RPC error for internal server errors.
// This should be used when an unexpected error occurs during processing.
func NewInternalError() error {
  return &jsonrpc.Error{
    Code:    CodeInternalErr,
    Message: "Internal server error",
  }
}
