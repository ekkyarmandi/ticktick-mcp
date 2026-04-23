package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	serverName    = "ticktick-mcp"
	serverVersion = "0.2.0"
)

func main() {
	cfg := parseRuntimeConfig()

	client, err := newTickTickClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	server := mcp.NewServer(
		&mcp.Implementation{Name: serverName, Version: serverVersion},
		nil,
	)

	svc := &tickTickService{
		client:             client,
		excludedGroupIDs:   parseCSVSet(os.Getenv("EXCLUDED_GROUP_IDS")),
		excludedProjectIDs: parseCSVSet(os.Getenv("EXCLUDED_PROJECT_IDS")),
	}
	registerTools(server, svc)

	switch cfg.Transport {
	case "stdio":
		if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
			log.Fatal(err)
		}
	case "http":
		if err := runHTTPServer(server, cfg); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("unsupported transport %q, expected stdio or http", cfg.Transport)
	}
}

type runtimeConfig struct {
	Transport     string
	Addr          string
	Path          string
	StatelessHTTP bool
	ServerToken   string
}

func parseRuntimeConfig() runtimeConfig {
	defaultPort := envOrDefault("PORT", "8080")

	transport := flag.String("transport", envOrDefault("MCP_TRANSPORT", "stdio"), "MCP transport to use: stdio or http")
	addr := flag.String("addr", envOrDefault("MCP_HTTP_ADDR", ":"+defaultPort), "HTTP listen address for MCP HTTP transport")
	path := flag.String("path", envOrDefault("MCP_HTTP_PATH", "/mcp"), "HTTP path for the MCP endpoint")
	stateless := flag.Bool("stateless-http", envBool("MCP_HTTP_STATELESS"), "run Streamable HTTP in stateless mode")
	serverToken := flag.String("server-token", strings.TrimSpace(os.Getenv("MCP_SERVER_TOKEN")), "optional bearer token required for HTTP access")
	flag.Parse()

	return runtimeConfig{
		Transport:     strings.ToLower(strings.TrimSpace(*transport)),
		Addr:          strings.TrimSpace(*addr),
		Path:          normalizeHTTPPath(*path),
		StatelessHTTP: *stateless,
		ServerToken:   strings.TrimSpace(*serverToken),
	}
}

func runHTTPServer(server *mcp.Server, cfg runtimeConfig) error {
	mcpHandler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return server
	}, &mcp.StreamableHTTPOptions{
		Stateless: cfg.StatelessHTTP,
	})

	mux := http.NewServeMux()
	mux.Handle(cfg.Path, authMiddleware(cfg.ServerToken, mcpHandler))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":        true,
			"name":      serverName,
			"version":   serverVersion,
			"transport": "http",
			"path":      cfg.Path,
			"time":      time.Now().UTC().Format(time.RFC3339),
		})
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"name":         serverName,
			"version":      serverVersion,
			"mcp_endpoint": cfg.Path,
			"health":       "/healthz",
			"transport":    "streamable-http",
		})
	})

	httpServer := &http.Server{
		Addr:              cfg.Addr,
		Handler:           loggingMiddleware(mux),
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("TickTick MCP listening on %s", cfg.Addr)
	log.Printf("MCP endpoint available at %s", cfg.Path)
	if cfg.ServerToken != "" {
		log.Printf("HTTP bearer token protection is enabled")
	}

	return httpServer.ListenAndServe()
}

func authMiddleware(token string, next http.Handler) http.Handler {
	if token == "" {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !requestHasValidAuth(r, token) {
			w.Header().Set("WWW-Authenticate", `Bearer realm="ticktick-mcp"`)
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "missing or invalid bearer token",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func requestHasValidAuth(r *http.Request, token string) bool {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	expectedBearer := "Bearer " + token
	if authHeader == expectedBearer {
		return true
	}

	xAPIKey := strings.TrimSpace(r.Header.Get("X-API-Key"))
	if xAPIKey == token {
		return true
	}

	apiKey := strings.TrimSpace(r.Header.Get("Api-Key"))
	if apiKey == token {
		return true
	}

	return false
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write json response: %v", err)
	}
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envBool(key string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func normalizeHTTPPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "/mcp"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}

func init() {
	flag.CommandLine.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [flags]\n\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "TickTick MCP server. Defaults to stdio for local MCP clients.\n\n")
		flag.PrintDefaults()
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
