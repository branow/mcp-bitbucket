package main

import (
	"fmt"
	"log"

	"github.com/branow/mcp-bitbucket/internal/config"
	"github.com/branow/mcp-bitbucket/internal/server"
)

func main() {
	addr := fmt.Sprintf(":%d", config.McpServerPort())
	server := server.NewMcpServer(addr)
	log.Fatal(server.Run())
}
