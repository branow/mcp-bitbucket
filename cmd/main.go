package main

import (
	"log"

	"github.com/branow/mcp-bitbucket/internal/config"
	"github.com/branow/mcp-bitbucket/internal/server"
)

func main() {
	cfg := config.NewGlobal()
	server := server.NewMcpServer(cfg)
	log.Fatal(server.Run())
}
