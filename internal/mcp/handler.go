package mcp

import (
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var handler *mcp.StreamableHTTPHandler

func init() {
	server := mcp.NewServer(&mcp.Implementation{
		Title:   "Bitbucket MCP",
		Version: "1.0.0",
	}, nil)

	handler = mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return server
	}, nil)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	handler.ServeHTTP(w, r)
}
