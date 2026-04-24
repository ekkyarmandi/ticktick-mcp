package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	tickTickAuthorizeURL = "https://ticktick.com/oauth/authorize"
	tickTickTokenURL     = "https://ticktick.com/oauth/token"
	tokenFileName        = "tokens.json"
)

// oauthConfig holds the OAuth client credentials.
type oauthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Scopes       string
}

// oauthTokens represents the stored OAuth tokens.
type oauthTokens struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// tokenStore manages persistent OAuth tokens with thread-safe access.
type tokenStore struct {
	mu       sync.RWMutex
	tokens   *oauthTokens
	filePath string
	oauth    *oauthConfig
}

func newTokenStore(cfg *oauthConfig, filePath string) *tokenStore {
	ts := &tokenStore{
		filePath: filePath,
		oauth:    cfg,
	}
	ts.load()
	return ts
}

// load reads tokens from disk. Silent on missing file.
func (ts *tokenStore) load() {
	data, err := os.ReadFile(ts.filePath)
	if err != nil {
		return
	}
	var tok oauthTokens
	if err := json.Unmarshal(data, &tok); err != nil {
		log.Printf("oauth: failed to parse %s: %v", ts.filePath, err)
		return
	}
	ts.tokens = &tok
}

// save writes tokens to disk.
func (ts *tokenStore) save() error {
	data, err := json.MarshalIndent(ts.tokens, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ts.filePath, data, 0600)
}

// HasTokens returns true if we have stored tokens.
func (ts *tokenStore) HasTokens() bool {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.tokens != nil && ts.tokens.AccessToken != ""
}

// AccessToken returns a valid access token, refreshing if needed.
func (ts *tokenStore) AccessToken() (string, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.tokens == nil {
		return "", fmt.Errorf("not authenticated -- visit /auth/login to connect TickTick")
	}

	// Refresh if token expires within 60 seconds
	if time.Until(ts.tokens.ExpiresAt) < 60*time.Second {
		if err := ts.refresh(); err != nil {
			return "", fmt.Errorf("token refresh failed: %w", err)
		}
	}

	return ts.tokens.AccessToken, nil
}

// Exchange trades an authorization code for tokens.
func (ts *tokenStore) Exchange(code string) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	form := url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {ts.oauth.RedirectURI},
	}

	tok, err := ts.tokenRequest(form)
	if err != nil {
		return err
	}

	ts.tokens = tok
	return ts.save()
}

// refresh uses the refresh token to get new tokens. Must be called with lock held.
func (ts *tokenStore) refresh() error {
	if ts.tokens.RefreshToken == "" {
		return fmt.Errorf("no refresh token available -- re-authenticate at /auth/login")
	}

	form := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {ts.tokens.RefreshToken},
	}

	tok, err := ts.tokenRequest(form)
	if err != nil {
		return err
	}

	ts.tokens = tok
	log.Printf("oauth: token refreshed, expires at %s", tok.ExpiresAt.Format(time.RFC3339))
	return ts.save()
}

// tokenRequest sends a POST to the token endpoint with Basic auth.
func (ts *tokenStore) tokenRequest(form url.Values) (*oauthTokens, error) {
	req, err := http.NewRequest(http.MethodPost, tickTickTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(ts.oauth.ClientID, ts.oauth.ClientSecret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse token response: %w", err)
	}

	return &oauthTokens{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(result.ExpiresIn) * time.Second),
	}, nil
}

// randomState generates a random hex string for CSRF protection.
func randomState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// oauthLoginHandler redirects the user to TickTick's authorization page.
func oauthLoginHandler(cfg *oauthConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := randomState()

		params := url.Values{
			"client_id":     {cfg.ClientID},
			"redirect_uri":  {cfg.RedirectURI},
			"response_type": {"code"},
			"scope":         {cfg.Scopes},
			"state":         {state},
		}

		authURL := tickTickAuthorizeURL + "?" + params.Encode()

		// Store state in a short-lived cookie for validation
		http.SetCookie(w, &http.Cookie{
			Name:     "oauth_state",
			Value:    state,
			Path:     "/auth",
			MaxAge:   600,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		http.Redirect(w, r, authURL, http.StatusFound)
	}
}

// oauthCallbackHandler handles the redirect from TickTick after authorization.
func oauthCallbackHandler(store *tokenStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check for error from TickTick
		if errParam := r.URL.Query().Get("error"); errParam != "" {
			desc := r.URL.Query().Get("error_description")
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error":       errParam,
				"description": desc,
			})
			return
		}

		// Validate state
		stateCookie, err := r.Cookie("oauth_state")
		if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid or missing state parameter",
			})
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "missing authorization code",
			})
			return
		}

		if err := store.Exchange(code); err != nil {
			log.Printf("oauth: token exchange failed: %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "token exchange failed: " + err.Error(),
			})
			return
		}

		// Clear state cookie
		http.SetCookie(w, &http.Cookie{
			Name:   "oauth_state",
			Path:   "/auth",
			MaxAge: -1,
		})

		log.Printf("oauth: authentication successful")
		writeJSON(w, http.StatusOK, map[string]string{
			"message": "TickTick connected successfully. You can close this page.",
		})
	}
}

// oauthStatusHandler returns current auth status.
func oauthStatusHandler(store *tokenStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store.mu.RLock()
		defer store.mu.RUnlock()

		status := map[string]any{
			"authenticated": store.tokens != nil && store.tokens.AccessToken != "",
		}

		if store.tokens != nil && !store.tokens.ExpiresAt.IsZero() {
			status["expires_at"] = store.tokens.ExpiresAt.Format(time.RFC3339)
			status["expires_in_seconds"] = int(time.Until(store.tokens.ExpiresAt).Seconds())
		}

		writeJSON(w, http.StatusOK, status)
	}
}
