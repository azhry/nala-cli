package auth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

func TestLoginDeviceCodePollsValidatesAndStoresSessionWithoutOpeningBrowser(t *testing.T) {
	store := &config.SessionStore{Path: filepath.Join(t.TempDir(), "session.json")}
	deviceCode := strings.Repeat("D", 43)
	accessToken := "header.payload.signature"
	var pollCount int
	var waits []time.Duration
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/cli/device/code":
			if r.Method != http.MethodPost || r.URL.RawQuery != "" || r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
				t.Errorf("device code request = %s %s with content type %q", r.Method, r.URL, r.Header.Get("Content-Type"))
			}
			if err := r.ParseForm(); err != nil || r.PostForm.Get("client_id") != deviceClientID || len(r.PostForm) != 1 {
				t.Errorf("device code form = %v, err = %v", r.PostForm, err)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(deviceAuthorizationResponse{
				DeviceCode: deviceCode, UserCode: "BCDFGHJK", VerificationURI: "https://nala.example/device", ExpiresIn: 600, Interval: 5,
			})
		case "/api/auth/cli/device/token":
			pollCount++
			if err := r.ParseForm(); err != nil {
				t.Errorf("parse token form: %v", err)
			}
			if r.PostForm.Get("grant_type") != deviceCodeGrantType || r.PostForm.Get("client_id") != deviceClientID || r.PostForm.Get("device_code") != deviceCode {
				t.Errorf("token form = %v", r.PostForm)
			}
			w.Header().Set("Content-Type", "application/json")
			switch pollCount {
			case 1:
				w.WriteHeader(http.StatusBadRequest)
				_, _ = io.WriteString(w, `{"error":"authorization_pending"}`)
			case 2:
				w.WriteHeader(http.StatusBadRequest)
				_, _ = io.WriteString(w, `{"error":"slow_down"}`)
			case 3:
				w.WriteHeader(http.StatusBadRequest)
				_, _ = io.WriteString(w, `{"error":"authorization_pending"}`)
			default:
				_ = json.NewEncoder(w).Encode(deviceTokenResponse{AccessToken: accessToken, TokenType: "Bearer"})
			}
		case "/api/auth/session":
			if r.Header.Get("Authorization") != "Bearer "+accessToken {
				t.Errorf("session Authorization = %q", r.Header.Get("Authorization"))
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"authenticated":true,"user":{"id":"user-1","name":"Test User","email":"user@example.test","tier":"developer"}}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, store)
	browserOpened := false
	client.OpenBrowser = func(string) error {
		browserOpened = true
		return nil
	}
	client.Wait = func(_ context.Context, duration time.Duration) error {
		waits = append(waits, duration)
		return nil
	}

	var output strings.Builder
	user, err := client.LoginDeviceCode(context.Background(), &output)
	if err != nil {
		t.Fatalf("LoginDeviceCode() error = %v", err)
	}
	if user.ID != "user-1" || user.Name != "Test User" || pollCount != 4 || browserOpened {
		t.Fatalf("user/polls/browser = %+v/%d/%v", user, pollCount, browserOpened)
	}
	if len(waits) != 4 || waits[0] != 5*time.Second || waits[1] != 5*time.Second || waits[2] != 10*time.Second || waits[3] != 10*time.Second {
		t.Fatalf("poll waits = %v", waits)
	}
	if got := output.String(); !strings.Contains(got, "https://nala.example/device") || !strings.Contains(got, "BCDFGHJK") || strings.Contains(got, deviceCode) || strings.Contains(got, accessToken) {
		t.Fatalf("device login output contains unexpected values: %q", got)
	}
	storedToken, err := store.Token()
	if err != nil {
		t.Fatalf("read stored token: %v", err)
	}
	if storedToken != accessToken {
		t.Fatal("stored session does not match the validated Nala Labs token")
	}
}

func TestLoginDeviceCodeDoesNotSaveUnvalidatedSession(t *testing.T) {
	store := &config.SessionStore{Path: filepath.Join(t.TempDir(), "session.json")}
	deviceCode := strings.Repeat("E", 43)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/cli/device/code":
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"device_code":"`+deviceCode+`","user_code":"BCDFGHJK","verification_uri":"https://nala.example/device","expires_in":600,"interval":1}`)
		case "/api/auth/cli/device/token":
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"access_token":"header.payload.signature","token_type":"Bearer"}`)
		case "/api/auth/session":
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = io.WriteString(w, `{"authenticated":false}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, store)
	client.Wait = func(context.Context, time.Duration) error { return nil }
	_, err := client.LoginDeviceCode(context.Background(), io.Discard)
	if err == nil || !strings.Contains(err.Error(), "validate device session") {
		t.Fatalf("LoginDeviceCode() error = %v, want session validation error", err)
	}
	if _, err := store.Token(); err == nil {
		t.Fatal("unvalidated session was saved")
	}
}

func TestLoginDeviceCodeStopsOnDenialAndCancellation(t *testing.T) {
	for _, test := range []struct {
		name      string
		wait      func(context.Context, time.Duration) error
		wantErr   string
		wantPolls int
	}{
		{name: "denied", wait: func(context.Context, time.Duration) error { return nil }, wantErr: "device login was denied", wantPolls: 1},
		{name: "canceled", wait: func(context.Context, time.Duration) error { return context.Canceled }, wantErr: "timed out or was canceled", wantPolls: 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			polls := 0
			deviceCode := strings.Repeat("F", 43)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/auth/cli/device/code":
					w.Header().Set("Content-Type", "application/json")
					_, _ = io.WriteString(w, `{"device_code":"`+deviceCode+`","user_code":"BCDFGHJK","verification_uri":"https://nala.example/device","expires_in":600,"interval":1}`)
				case "/api/auth/cli/device/token":
					polls++
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusBadRequest)
					_, _ = io.WriteString(w, `{"error":"access_denied"}`)
				default:
					http.NotFound(w, r)
				}
			}))
			defer server.Close()

			store := &config.SessionStore{Path: filepath.Join(t.TempDir(), "session.json")}
			client := NewClient(server.URL, store)
			client.Wait = test.wait
			_, err := client.LoginDeviceCode(context.Background(), io.Discard)
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("LoginDeviceCode() error = %v, want %q", err, test.wantErr)
			}
			if polls != test.wantPolls {
				t.Fatalf("poll count = %d, want %d", polls, test.wantPolls)
			}
		})
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

func TestNewClientFromEnvironmentUsesSavedAPIURL(t *testing.T) {
	t.Setenv("NALA_CONFIG_DIR", t.TempDir())
	t.Setenv("NALA_API_BASE_URL", "")
	store, err := config.NewSettingsStore()
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Save(config.Settings{APIBaseURL: "http://127.0.0.1:8082"}); err != nil {
		t.Fatal(err)
	}

	client, err := NewClientFromEnvironment()
	if err != nil {
		t.Fatalf("NewClientFromEnvironment() error = %v", err)
	}
	if client.BaseURL != "http://127.0.0.1:8082" {
		t.Fatalf("BaseURL = %q, want saved API URL", client.BaseURL)
	}
}

func TestNewClientFromEnvironmentPrefersAPIEnvironmentOverride(t *testing.T) {
	t.Setenv("NALA_CONFIG_DIR", t.TempDir())
	t.Setenv("NALA_API_BASE_URL", "http://127.0.0.1:9090")
	store, err := config.NewSettingsStore()
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Save(config.Settings{APIBaseURL: "http://127.0.0.1:8082"}); err != nil {
		t.Fatal(err)
	}

	client, err := NewClientFromEnvironment()
	if err != nil {
		t.Fatalf("NewClientFromEnvironment() error = %v", err)
	}
	if client.BaseURL != "http://127.0.0.1:9090" {
		t.Fatalf("BaseURL = %q, want environment override", client.BaseURL)
	}
}
