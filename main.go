package main

import (
	"context"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	serverName    = "ticktick-mcp"
	serverVersion = "0.1.0"
)

func main() {
	client, err := newTickTickClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	server := mcp.NewServer(
		&mcp.Implementation{Name: serverName, Version: serverVersion},
		nil,
	)

	registerTools(server, &tickTickService{client: client})

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
