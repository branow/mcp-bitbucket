package main

import (
	"fmt"
	"log"

	"github.com/branow/mcp-bitbucket/cmd/mcp"
	"github.com/branow/mcp-bitbucket/internal/config"
)

func main() {
	addr := fmt.Sprintf(":%d", config.McpServerPort())
	server := mcp.NewMcpServer(addr)
	log.Fatal(server.Run())
}
