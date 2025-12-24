package mcp

import "github.com/modelcontextprotocol/go-sdk/mcp"

type Dispatcher[T any] interface {
	Dispatch(*mcp.Server)
}
