package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/azhry/nala-cli/internal/config"
)

func TestLoginCompletesLoopbackCallbackAndStoresSession(t *testing.T) {
	store := &config.SessionStore{Path: filepath.Join(t.TempDir(), "session.json")}
	var openedURL string
	var authorizationQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/cli/authorize":
			authorizationQuery = r.URL.Query()
			redirect, err := url.Parse(r.URL.Query().Get("redirect_uri"))
			if err != nil {
				t.Fatalf("parse redirect URI: %v", err)
			}
			if redirect.Hostname() != "127.0.0.1" || redirect.Port() == "" {
				t.Fatalf("redirect URI = %q, want loopback", redirect.String())
			}
			query := redirect.Query()
			query.Set("code", "one-time-provider-code")
			query.Set("state", r.URL.Query().Get("state"))
			redirect.RawQuery = query.Encode()
			http.Redirect(w, r, redirect.String(), http.StatusFound)
		case "/api/auth/cli/exchange":
			var request struct {
				Code string `json:"code"`
			}
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatalf("decode exchange request: %v", err)
			}
			if request.Code != "one-time-provider-code" {
				t.Fatalf("exchange code = %q", request.Code)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(exchangeResponse{
				Authenticated: true,
				Token:         "session-token-for-test",
				User:          User{ID: "user-1", Name: "Test User", Email: "user@example.test", Tier: "free"},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, store)
	client.OpenBrowser = func(rawURL string) error {
		openedURL = rawURL
		go func() {
			response, err := http.Get(rawURL)
			if err == nil {
				_ = response.Body.Close()
			}
		}()
		return nil
	}

	user, err := client.Login(context.Background())
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if user.ID != "user-1" || user.Name != "Test User" {
		t.Fatalf("user = %+v", user)
	}
	if authorizationQuery.Get("redirect_uri") == "" || len(authorizationQuery.Get("state")) < 32 {
		t.Fatalf("authorization query = %v", authorizationQuery)
	}
	if !strings.Contains(openedURL, "code=one-time-provider-code") {
		t.Fatalf("opened authorization URL = %q", openedURL)
	}
	token, err := store.Token()
	if err != nil {
		t.Fatalf("read stored token: %v", err)
	}
	if token != "session-token-for-test" {
		t.Fatalf("stored token = %q", token)
	}
}

func TestUserInfoUsesStoredBearerToken(t *testing.T) {
	store := &config.SessionStore{Path: filepath.Join(t.TempDir(), "session.json")}
	if err := store.Save("session-token-for-test"); err != nil {
		t.Fatalf("save token: %v", err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/auth/session" {
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer session-token-for-test" {
			t.Fatalf("Authorization = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"authenticated":true,"user":{"id":"user-1","name":"Test User","email":"user@example.test","tier":"developer","entitlements":{"maxDeployments":3,"maxDatabases":2,"expires":false,"policy":"developer"}}}`))
	}))
	defer server.Close()

	user, err := NewClient(server.URL, store).UserInfo(context.Background())
	if err != nil {
		t.Fatalf("UserInfo() error = %v", err)
	}
	if user.ID != "user-1" || user.Tier != "developer" {
		t.Fatalf("user = %+v", user)
	}
}

func TestUserInfoWithoutSessionDoesNotMakeRequest(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true }))
	defer server.Close()

	store := &config.SessionStore{Path: filepath.Join(t.TempDir(), "missing-session.json")}
	_, err := NewClient(server.URL, store).UserInfo(context.Background())
	if err == nil || err.Error() != "not authenticated" {
		t.Fatalf("UserInfo() error = %v, want not authenticated", err)
	}
	if called {
		t.Fatal("UserInfo() made a request without a stored session")
	}
}
