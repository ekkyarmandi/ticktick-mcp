package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// bearerTokenKey is the context key for per-request Bearer tokens.
type bearerTokenKey struct{}

// bearerTokenFromContext retrieves the Bearer token stored in context by the
// bearer-injection middleware. Returns empty string if none.
func bearerTokenFromContext(ctx context.Context) string {
	tok, _ := ctx.Value(bearerTokenKey{}).(string)
	return tok
}

// bearerContextMiddleware extracts the Bearer token from the Authorization
// header and stores it in the request context so tool handlers can use it.
func bearerContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if tok := extractBearerToken(r); tok != "" {
			ctx := context.WithValue(r.Context(), bearerTokenKey{}, tok)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}

// mcpOAuthProxyConfig holds the settings for the OAuth proxy endpoints.
type mcpOAuthProxyConfig struct {
	// Issuer is the external base URL of this server (e.g. https://ticktick-mcp.ekky.dev).
	Issuer string
	// OAuth client credentials for TickTick.
	ClientID     string
	ClientSecret string
	// Scopes supported.
	Scopes []string
}

func newMCPOAuthProxyConfig() *mcpOAuthProxyConfig {
	issuer := strings.TrimRight(os.Getenv("MCP_OAUTH_ISSUER"), "/")
	if issuer == "" {
		// Fall back to redirect URI's origin
		redirectURI := os.Getenv("TICKTICK_REDIRECT_URI")
		if redirectURI != "" {
			if u, err := url.Parse(redirectURI); err == nil {
				issuer = u.Scheme + "://" + u.Host
			}
		}
	}
	if issuer == "" {
		issuer = "http://localhost:8080"
	}

	scopes := strings.Fields(envOrDefault("TICKTICK_SCOPES", "tasks:read tasks:write"))

	return &mcpOAuthProxyConfig{
		Issuer:       issuer,
		ClientID:     os.Getenv("TICKTICK_CLIENT_ID"),
		ClientSecret: os.Getenv("TICKTICK_CLIENT_SECRET"),
		Scopes:       scopes,
	}
}

// protectedResourceHandler serves /.well-known/oauth-protected-resource (RFC 9728).
func protectedResourceHandler(cfg *mcpOAuthProxyConfig) http.HandlerFunc {
	resp, _ := json.Marshal(map[string]any{
		"resource":                cfg.Issuer,
		"authorization_servers":   []string{cfg.Issuer},
		"bearer_methods_supported": []string{"header"},
	})
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(resp)
	}
}

// authorizationServerHandler serves /.well-known/oauth-authorization-server (RFC 8414).
func authorizationServerHandler(cfg *mcpOAuthProxyConfig) http.HandlerFunc {
	resp, _ := json.Marshal(map[string]any{
		"issuer":                             cfg.Issuer,
		"authorization_endpoint":             cfg.Issuer + "/oauth/authorize",
		"token_endpoint":                     cfg.Issuer + "/oauth/token",
		"response_types_supported":           []string{"code"},
		"grant_types_supported":              []string{"authorization_code", "refresh_token"},
		"code_challenge_methods_supported":   []string{"S256"},
		"scopes_supported":                   cfg.Scopes,
		"token_endpoint_auth_methods_supported": []string{"client_secret_basic", "client_secret_post"},
	})
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(resp)
	}
}

// oauthAuthorizeProxyHandler redirects to TickTick's authorize URL, passing
// through all query parameters from the client (including PKCE params).
func oauthAuthorizeProxyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		target := tickTickAuthorizeURL + "?" + r.URL.RawQuery
		log.Printf("oauth proxy: redirecting to TickTick authorize")
		http.Redirect(w, r, target, http.StatusFound)
	}
}

// oauthTokenProxyHandler proxies token requests to TickTick's token endpoint.
// If the client doesn't send client credentials, the server injects its own.
func oauthTokenProxyHandler(cfg *mcpOAuthProxyConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "failed to read request body"})
			return
		}

		// Parse form to check if client_id is present
		formValues, _ := url.ParseQuery(string(body))

		// Build upstream request
		upstreamReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, tickTickTokenURL, strings.NewReader(formValues.Encode()))
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to build upstream request"})
			return
		}
		upstreamReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		// Check if client sent Basic auth or client_id in body
		_, _, hasBasicAuth := r.BasicAuth()
		hasClientInBody := formValues.Get("client_id") != ""

		if hasBasicAuth {
			// Forward client's own Basic auth
			upstreamReq.Header.Set("Authorization", r.Header.Get("Authorization"))
		} else if hasClientInBody {
			// Client sent credentials in body -- TickTick expects Basic auth
			clientID := formValues.Get("client_id")
			clientSecret := formValues.Get("client_secret")
			// Remove from body, set as Basic auth
			formValues.Del("client_id")
			formValues.Del("client_secret")
			upstreamReq.Body = io.NopCloser(strings.NewReader(formValues.Encode()))
			upstreamReq.ContentLength = int64(len(formValues.Encode()))
			upstreamReq.SetBasicAuth(clientID, clientSecret)
		} else {
			// No client credentials from client -- use server's own
			upstreamReq.SetBasicAuth(cfg.ClientID, cfg.ClientSecret)
		}

		resp, err := http.DefaultClient.Do(upstreamReq)
		if err != nil {
			log.Printf("oauth proxy: token request failed: %v", err)
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "upstream token request failed"})
			return
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)

		// Forward response as-is
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(resp.StatusCode)
		w.Write(respBody)
	}
}
