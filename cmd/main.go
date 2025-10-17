package main

import (
	"log"

	"github.com/branow/mcp-bitbucket/cmd/mcp"
)

func main() {
	server := mcp.NewMcpServer(":8080")
	log.Fatal(server.Run())
}
