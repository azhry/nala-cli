package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/azhry/nala-cli/internal/config"
)

const defaultLoginTimeout = 10 * time.Minute

type Entitlements struct {
	MaxDeployments *int   `json:"maxDeployments"`
	MaxDatabases   *int   `json:"maxDatabases"`
	Expires        bool   `json:"expires"`
	Policy         string `json:"policy"`
}

type User struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Email        string       `json:"email"`
	Tier         string       `json:"tier"`
	Entitlements Entitlements `json:"entitlements"`
	Permissions  []string     `json:"permissions,omitempty"`
}

type sessionResponse struct {
	Authenticated bool `json:"authenticated"`
	User          User `json:"user"`
}

type exchangeResponse struct {
	Authenticated bool   `json:"authenticated"`
	Token         string `json:"token"`
	User          User   `json:"user"`
}

type exchangeRequest struct {
	Code string `json:"code"`
}

type Client struct {
	BaseURL      string
	HTTPClient   *http.Client
	Store        *config.SessionStore
	OpenBrowser  func(string) error
	Listen       func(network, address string) (net.Listener, error)
	LoginTimeout time.Duration
}

func NewClient(baseURL string, store *config.SessionStore) *Client {
	return &Client{
		BaseURL:      strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		Store:        store,
		OpenBrowser:  OpenURL,
		Listen:       net.Listen,
		LoginTimeout: defaultLoginTimeout,
	}
}

func NewClientFromEnvironment() (*Client, error) {
	baseURL := strings.TrimSpace(getenv("NALA_API_BASE_URL", "http://127.0.0.1:8080"))
	store, err := config.NewSessionStore()
	if err != nil {
		return nil, err
	}
	return NewClient(baseURL, store), nil
}

func (c *Client) Login(ctx context.Context) (User, error) {
	if c == nil || c.Store == nil {
		return User{}, errors.New("client is not configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if c.LoginTimeout <= 0 {
		c.LoginTimeout = defaultLoginTimeout
	}
	loginContext, cancel := context.WithTimeout(ctx, c.LoginTimeout)
	defer cancel()
	listener, err := c.listen("tcp", "127.0.0.1:0")
	if err != nil {
		return User{}, fmt.Errorf("start callback server: %w", err)
	}
	defer listener.Close()

	state, err := randomValue(32)
	if err != nil {
		return User{}, errors.New("create login state")
	}
	callbackURL := "http://" + listener.Addr().String() + "/callback"
	callbackResult := make(chan callback, 1)
	server := &http.Server{Handler: c.callbackHandler(state, callbackResult)}
	serverErrors := make(chan error, 1)
	go func() {
		if serveErr := server.Serve(listener); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			serverErrors <- serveErr
		}
	}()
	defer func() {
		shutdownContext, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
		_ = server.Shutdown(shutdownContext)
		shutdownCancel()
	}()

	authorizationURL, err := c.authorizationURL(loginContext, callbackURL, state)
	if err != nil {
		return User{}, err
	}
	if err := c.openBrowser(authorizationURL); err != nil {
		return User{}, fmt.Errorf("open browser: %w", err)
	}

	var result callback
	select {
	case result = <-callbackResult:
	case err := <-serverErrors:
		return User{}, fmt.Errorf("callback server stopped: %w", err)
	case <-loginContext.Done():
		return User{}, errors.New("login timed out or was canceled")
	}
	if result.err != "" {
		return User{}, errors.New("login was denied")
	}

	exchanged, err := c.exchange(loginContext, result.code)
	if err != nil {
		return User{}, err
	}
	if err := c.Store.Save(exchanged.Token); err != nil {
		return User{}, fmt.Errorf("save session: %w", err)
	}
	return exchanged.User, nil
}

func (c *Client) UserInfo(ctx context.Context) (User, error) {
	if c == nil || c.Store == nil {
		return User{}, errors.New("client is not configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	token, err := c.Store.Token()
	if err != nil {
		return User{}, errors.New("not authenticated")
	}
	endpoint, err := c.endpoint("/api/auth/session")
	if err != nil {
		return User{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return User{}, err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	response, err := c.noRedirectClient().Do(request)
	if err != nil {
		return User{}, fmt.Errorf("request current user: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusUnauthorized {
		return User{}, errors.New("not authenticated")
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return User{}, fmt.Errorf("current-user request failed with status %s", response.Status)
	}
	var session sessionResponse
	if err := decodeJSON(response.Body, &session); err != nil {
		return User{}, errors.New("current-user response is invalid")
	}
	if !session.Authenticated || session.User.ID == "" {
		return User{}, errors.New("not authenticated")
	}
	return session.User, nil
}

type callback struct {
	code string
	err  string
}

func (c *Client) callbackHandler(expectedState string, result chan<- callback) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Query().Get("state") != expectedState {
			http.Error(w, "invalid callback", http.StatusBadRequest)
			return
		}
		if providerError := strings.TrimSpace(r.URL.Query().Get("error")); providerError != "" {
			select {
			case result <- callback{err: providerError}:
			default:
			}
			_, _ = io.WriteString(w, "Login was canceled. You can return to the CLI.\n")
			return
		}
		if strings.TrimSpace(r.URL.Query().Get("code")) == "" {
			http.Error(w, "invalid callback", http.StatusBadRequest)
			return
		}
		select {
		case result <- callback{code: r.URL.Query().Get("code")}:
		default:
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = io.WriteString(w, "Login complete. You can return to the CLI.\n")
	})
	return mux
}

func (c *Client) authorizationURL(ctx context.Context, callbackURL, state string) (string, error) {
	endpoint, err := c.endpoint("/api/auth/cli/authorize")
	if err != nil {
		return "", err
	}
	query := url.Values{}
	query.Set("redirect_uri", callbackURL)
	query.Set("state", state)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+query.Encode(), nil)
	if err != nil {
		return "", err
	}
	response, err := c.noRedirectClient().Do(request)
	if err != nil {
		return "", fmt.Errorf("request login: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusFound {
		return "", fmt.Errorf("login request failed with status %s", response.Status)
	}
	location := strings.TrimSpace(response.Header.Get("Location"))
	if location == "" {
		return "", errors.New("login response did not provide an authorization URL")
	}
	parsedLocation, err := url.Parse(location)
	if err != nil || (parsedLocation.Scheme != "http" && parsedLocation.Scheme != "https") || parsedLocation.Host == "" {
		return "", errors.New("login response provided an invalid authorization URL")
	}
	return location, nil
}

func (c *Client) exchange(ctx context.Context, code string) (exchangeResponse, error) {
	endpoint, err := c.endpoint("/api/auth/cli/exchange")
	if err != nil {
		return exchangeResponse{}, err
	}
	payload, err := json.Marshal(exchangeRequest{Code: code})
	if err != nil {
		return exchangeResponse{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(payload)))
	if err != nil {
		return exchangeResponse{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := c.httpClient().Do(request)
	if err != nil {
		return exchangeResponse{}, fmt.Errorf("exchange login: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusUnauthorized {
		return exchangeResponse{}, errors.New("login exchange was rejected")
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return exchangeResponse{}, fmt.Errorf("login exchange failed with status %s", response.Status)
	}
	var exchanged exchangeResponse
	if err := decodeJSON(response.Body, &exchanged); err != nil || !exchanged.Authenticated || strings.TrimSpace(exchanged.Token) == "" || exchanged.User.ID == "" {
		return exchangeResponse{}, errors.New("login exchange response is invalid")
	}
	return exchanged, nil
}

func (c *Client) endpoint(path string) (string, error) {
	base, err := url.Parse(c.BaseURL)
	if err != nil || (base.Scheme != "http" && base.Scheme != "https") || base.Host == "" || base.RawQuery != "" || base.Fragment != "" {
		return "", errors.New("NALA_API_BASE_URL must be an http(s) URL")
	}
	return strings.TrimRight(base.String(), "/") + path, nil
}

func (c *Client) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return &http.Client{Timeout: 30 * time.Second}
}

func (c *Client) noRedirectClient() *http.Client {
	client := *c.httpClient()
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }
	return &client
}

func (c *Client) listen(network, address string) (net.Listener, error) {
	if c.Listen == nil {
		return net.Listen(network, address)
	}
	return c.Listen(network, address)
}

func (c *Client) openBrowser(rawURL string) error {
	if c.OpenBrowser == nil {
		return OpenURL(rawURL)
	}
	return c.OpenBrowser(rawURL)
}

func OpenURL(rawURL string) error {
	var command string
	var args []string
	switch runtime.GOOS {
	case "windows":
		command = "rundll32.exe"
		args = []string{"url.dll,FileProtocolHandler", rawURL}
	case "darwin":
		command = "open"
		args = []string{rawURL}
	default:
		command = "xdg-open"
		args = []string{rawURL}
	}
	return exec.Command(command, args...).Start()
}

func randomValue(size int) (string, error) {
	value := make([]byte, size)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func decodeJSON(reader io.Reader, target any) error {
	decoder := json.NewDecoder(io.LimitReader(reader, 1<<20))
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("response contains multiple JSON values")
	}
	return nil
}

func getenv(key, fallback string) string {
	if value := strings.TrimSpace(strings.TrimSpace(os.Getenv(key))); value != "" {
		return value
	}
	return fallback
}
