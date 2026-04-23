package main

import (
	"context"
	"log"
	"os"
	"strings"

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

	svc := &tickTickService{
		client:            client,
		excludedGroupIDs:  parseCSVSet(os.Getenv("EXCLUDED_GROUP_IDS")),
		excludedProjectIDs: parseCSVSet(os.Getenv("EXCLUDED_PROJECT_IDS")),
	}
	registerTools(server, svc)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}

func parseCSVSet(val string) map[string]bool {
	set := make(map[string]bool)
	for _, s := range strings.Split(val, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			set[s] = true
		}
	}
	return set
}
