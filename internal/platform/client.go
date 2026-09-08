package platform

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/azhry/nala-cli/internal/config"
)

const (
	defaultNalaLabsURL = "http://127.0.0.1:8080"
	defaultNalaSVCURL  = "http://127.0.0.1:8081"
	defaultAppPage     = int64(1)
	defaultAppPageSize = int64(20)
	maxAppPageSize     = int64(100)
	maxResponseBytes   = 2 << 20
	maxSSELineBytes    = 1 << 20
)

type Client struct {
	NalaLabsBaseURL string
	NalaSVCBaseURL  string
	HTTPClient      *http.Client
	Store           *config.SessionStore
}

type App struct {
	ID            int64  `json:"id"`
	OwnerID       string `json:"ownerId"`
	Name          string `json:"name"`
	RepositoryURL string `json:"repositoryUrl"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

type AppPagination struct {
	Page        int64 `json:"page"`
	PageSize    int64 `json:"pageSize"`
	Total       int64 `json:"total"`
	HasNextPage bool  `json:"hasNextPage"`
}

type AppListResponse struct {
	Apps       []App         `json:"apps"`
	Pagination AppPagination `json:"pagination"`
}

type AppDomain struct {
	Subdomain        *string `json:"subdomain"`
	ConfiguredDomain *string `json:"configuredDomain"`
}

type DeploymentSummary struct {
	ID                   int64   `json:"id"`
	ComponentType        string  `json:"componentType"`
	ComponentName        string  `json:"componentName"`
	Status               string  `json:"status"`
	Stage                string  `json:"stage"`
	SourceRef            string  `json:"sourceRef"`
	RequestedSourceRef   string  `json:"requestedSourceRef"`
	ResolvedSourceRef    *string `json:"resolvedSourceRef"`
	SourceDir            string  `json:"sourceDir"`
	DescriptorPath       string  `json:"descriptorPath"`
	DescriptorVersion    int     `json:"descriptorVersion"`
	DescriptorChecksum   *string `json:"descriptorChecksum"`
	DescriptorContent    *string `json:"descriptorContent"`
	CommitSHA            *string `json:"commitSha"`
	ImageTag             *string `json:"imageTag"`
	ImageDigest          *string `json:"imageDigest"`
	TargetCluster        string  `json:"targetCluster"`
	TargetTier           *string `json:"targetTier"`
	Namespace            *string `json:"namespace"`
	NamespaceRef         *string `json:"namespaceRef"`
	DatabaseRef          *string `json:"databaseRef"`
	VaultRef             *string `json:"vaultRef"`
	IsolationRef         *string `json:"isolationRef"`
	ConfigurationVersion int     `json:"configurationVersion"`
	FailureCode          *string `json:"failureCode"`
	FailureMessage       *string `json:"failureMessage"`
	TransitionVersion    int64   `json:"transitionVersion"`
	CreatedAt            string  `json:"createdAt"`
	StartedAt            *string `json:"startedAt"`
	FinishedAt           *string `json:"finishedAt"`
	UpdatedAt            string  `json:"updatedAt"`
}

type AppDetail struct {
	ID                 int64              `json:"id"`
	OwnerID            string             `json:"ownerId"`
	Name               string             `json:"name"`
	DisplayName        string             `json:"displayName"`
	RepositoryURL      string             `json:"repositoryUrl"`
	ComponentType      string             `json:"componentType"`
	DescriptorVersion  int                `json:"descriptorVersion"`
	DescriptorChecksum *string            `json:"descriptorChecksum"`
	DescriptorContent  *string            `json:"descriptorContent"`
	CreatedAt          string             `json:"createdAt"`
	UpdatedAt          string             `json:"updatedAt"`
	Domain             AppDomain          `json:"domain"`
	Health             string             `json:"health"`
	LatestDeployment   *DeploymentSummary `json:"latestDeployment"`
}

type AppDetailResponse struct {
	App AppDetail `json:"app"`
}

type Deployment struct {
	ID                   int64   `json:"id"`
	AppID                int64   `json:"appId"`
	ComponentType        string  `json:"componentType"`
	ComponentName        string  `json:"componentName"`
	Status               string  `json:"status"`
	Stage                string  `json:"stage"`
	SourceRef            string  `json:"sourceRef"`
	RequestedSourceRef   string  `json:"requestedSourceRef"`
	ResolvedSourceRef    *string `json:"resolvedSourceRef"`
	SourceDir            string  `json:"sourceDir"`
	DescriptorPath       string  `json:"descriptorPath"`
	DescriptorVersion    int     `json:"descriptorVersion"`
	DescriptorChecksum   *string `json:"descriptorChecksum"`
	DescriptorContent    *string `json:"descriptorContent"`
	CommitSHA            *string `json:"commitSha"`
	ImageTag             *string `json:"imageTag"`
	ImageDigest          *string `json:"imageDigest"`
	TargetCluster        string  `json:"targetCluster"`
	TargetTier           *string `json:"targetTier"`
	Namespace            *string `json:"namespace"`
	NamespaceRef         *string `json:"namespaceRef"`
	DatabaseRef          *string `json:"databaseRef"`
	VaultRef             *string `json:"vaultRef"`
	IsolationRef         *string `json:"isolationRef"`
	ConfigurationVersion int     `json:"configurationVersion"`
	FailureCode          *string `json:"failureCode"`
	FailureMessage       *string `json:"failureMessage"`
	TransitionVersion    int64   `json:"transitionVersion"`
	CreatedAt            string  `json:"createdAt"`
	StartedAt            *string `json:"startedAt"`
	FinishedAt           *string `json:"finishedAt"`
	UpdatedAt            string  `json:"updatedAt"`
}

type DeploymentEvent struct {
	ID           int64  `json:"id"`
	DeploymentID int64  `json:"deploymentId"`
	Type         string `json:"event"`
	Status       string `json:"status,omitempty"`
	Stage        string `json:"stage,omitempty"`
	Message      string `json:"message,omitempty"`
	Stream       string `json:"stream,omitempty"`
	CreatedAt    string `json:"createdAt"`
}

type deploymentRequest struct {
	AppID          int64  `json:"appId"`
	SourceRef      string `json:"sourceRef"`
	IdempotencyKey string `json:"idempotencyKey"`
}

func NewClient(nalaLabsBaseURL, nalaSVCBaseURL string, store *config.SessionStore) *Client {
	return &Client{
		NalaLabsBaseURL: strings.TrimRight(strings.TrimSpace(nalaLabsBaseURL), "/"),
		NalaSVCBaseURL:  strings.TrimRight(strings.TrimSpace(nalaSVCBaseURL), "/"),
		Store:           store,
	}
}

func NewClientFromEnvironment() (*Client, error) {
	store, err := config.NewSessionStore()
	if err != nil {
		return nil, err
	}
	return NewClient(
		getenv("NALA_API_BASE_URL", defaultNalaLabsURL),
		getenv("NALA_SVC_BASE_URL", defaultNalaSVCURL),
		store,
	), nil
}

func (c *Client) ListApps(ctx context.Context, page, pageSize int64) (AppListResponse, error) {
	if page <= 0 {
		return AppListResponse{}, errors.New("app page must be positive")
	}
	if pageSize <= 0 || pageSize > maxAppPageSize {
		return AppListResponse{}, fmt.Errorf("app page size must be between 1 and %d", maxAppPageSize)
	}
	query := url.Values{}
	query.Set("page", strconv.FormatInt(page, 10))
	query.Set("pageSize", strconv.FormatInt(pageSize, 10))
	var result AppListResponse
	if err := c.doJSON(ctx, http.MethodGet, c.NalaLabsBaseURL, "/api/apps?"+query.Encode(), nil, &result, false); err != nil {
		return AppListResponse{}, fmt.Errorf("list apps: %w", err)
	}
	return result, nil
}

func (c *Client) GetApp(ctx context.Context, appID int64) (AppDetailResponse, error) {
	if appID <= 0 {
		return AppDetailResponse{}, errors.New("app id must be positive")
	}
	var result AppDetailResponse
	if err := c.doJSON(ctx, http.MethodGet, c.NalaLabsBaseURL, "/api/apps/"+strconv.FormatInt(appID, 10), nil, &result, false); err != nil {
		return AppDetailResponse{}, fmt.Errorf("get app: %w", err)
	}
	return result, nil
}

func (c *Client) CreateDeployment(ctx context.Context, appID int64, sourceRef, idempotencyKey string) (Deployment, error) {
	if appID <= 0 {
		return Deployment{}, errors.New("app id must be positive")
	}
	sourceRef = strings.TrimSpace(sourceRef)
	if sourceRef == "" || len([]rune(sourceRef)) > 255 {
		return Deployment{}, errors.New("source ref must be between 1 and 255 characters")
	}
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if idempotencyKey == "" || len([]rune(idempotencyKey)) > 128 {
		return Deployment{}, errors.New("idempotency key must be between 1 and 128 characters")
	}
	payload, err := json.Marshal(deploymentRequest{AppID: appID, SourceRef: sourceRef, IdempotencyKey: idempotencyKey})
	if err != nil {
		return Deployment{}, err
	}
	var result Deployment
	if err := c.doJSON(ctx, http.MethodPost, c.NalaSVCBaseURL, "/api/deployments", payload, &result, false); err != nil {
		return Deployment{}, fmt.Errorf("create deployment: %w", err)
	}
	return result, nil
}

func (c *Client) GetDeployment(ctx context.Context, deploymentID int64) (Deployment, error) {
	if deploymentID <= 0 {
		return Deployment{}, errors.New("deployment id must be positive")
	}
	var result Deployment
	path := "/api/deployments/" + strconv.FormatInt(deploymentID, 10)
	if err := c.doJSON(ctx, http.MethodGet, c.NalaSVCBaseURL, path, nil, &result, false); err != nil {
		return Deployment{}, fmt.Errorf("get deployment: %w", err)
	}
	return result, nil
}

func (c *Client) FollowDeploymentEvents(ctx context.Context, deploymentID, cursor int64, handle func(DeploymentEvent) error) error {
	if deploymentID <= 0 {
		return errors.New("deployment id must be positive")
	}
	if cursor < 0 {
		return errors.New("event cursor cannot be negative")
	}
	if handle == nil {
		return errors.New("event handler is required")
	}
	query := url.Values{}
	query.Set("cursor", strconv.FormatInt(cursor, 10))
	path := "/api/deployments/" + strconv.FormatInt(deploymentID, 10) + "/events?" + query.Encode()
	response, err := c.do(ctx, http.MethodGet, c.NalaSVCBaseURL, path, nil, true)
	if err != nil {
		return fmt.Errorf("follow deployment events: %w", err)
	}
	defer response.Body.Close()
	if !isSuccess(response.StatusCode) {
		return fmt.Errorf("follow deployment events: %w", responseError(response))
	}

	scanner := bufio.NewScanner(response.Body)
	scanner.Buffer(make([]byte, 64*1024), maxSSELineBytes)
	var record sseRecord
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			if record.hasData() {
				event, err := record.event()
				if err != nil {
					return fmt.Errorf("decode deployment event: %w", err)
				}
				if err := handle(event); err != nil {
					return err
				}
			}
			record = sseRecord{}
			continue
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		field, value, ok := strings.Cut(line, ":")
		if ok && strings.HasPrefix(value, " ") {
			value = value[1:]
		}
		switch field {
		case "id":
			record.id = strings.TrimSpace(value)
		case "event":
			record.eventType = strings.TrimSpace(value)
		case "data":
			record.data = append(record.data, value)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read deployment events: %w", err)
	}
	if record.hasData() {
		event, err := record.event()
		if err != nil {
			return fmt.Errorf("decode deployment event: %w", err)
		}
		if err := handle(event); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) DeleteApp(ctx context.Context, appID int64) error {
	if appID <= 0 {
		return errors.New("app id must be positive")
	}
	response, err := c.do(ctx, http.MethodDelete, c.NalaLabsBaseURL, "/api/apps/"+strconv.FormatInt(appID, 10), nil, false)
	if err != nil {
		return fmt.Errorf("delete app: %w", err)
	}
	defer response.Body.Close()
	if !isSuccess(response.StatusCode) {
		return fmt.Errorf("delete app: %w", responseError(response))
	}
	if response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("delete app: unexpected response status %s", response.Status)
	}
	return nil
}

type sseRecord struct {
	id        string
	eventType string
	data      []string
}

func (r sseRecord) hasData() bool {
	return len(r.data) > 0
}

func (r sseRecord) event() (DeploymentEvent, error) {
	var event DeploymentEvent
	if err := json.Unmarshal([]byte(strings.Join(r.data, "\n")), &event); err != nil {
		return DeploymentEvent{}, err
	}
	if event.Type == "" {
		event.Type = r.eventType
	}
	if event.ID == 0 && r.id != "" {
		id, err := strconv.ParseInt(r.id, 10, 64)
		if err != nil || id < 0 {
			return DeploymentEvent{}, errors.New("event id is invalid")
		}
		event.ID = id
	}
	return event, nil
}

func (c *Client) doJSON(ctx context.Context, method, baseURL, path string, body []byte, target any, stream bool) error {
	response, err := c.do(ctx, method, baseURL, path, body, stream)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if !isSuccess(response.StatusCode) {
		return responseError(response)
	}
	if target == nil {
		return nil
	}
	if err := decodeJSON(response.Body, target); err != nil {
		return errors.New("response is invalid")
	}
	return nil
}

func (c *Client) do(ctx context.Context, method, baseURL, path string, body []byte, stream bool) (*http.Response, error) {
	if c == nil || c.Store == nil {
		return nil, errors.New("client is not configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	token, err := c.Store.Token()
	if err != nil {
		return nil, errors.New("not authenticated")
	}
	endpoint, err := endpoint(baseURL, path)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	if stream {
		request.Header.Set("Accept", "text/event-stream")
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	return c.httpClient(stream).Do(request)
}

func (c *Client) httpClient(stream bool) *http.Client {
	var client http.Client
	if c != nil && c.HTTPClient != nil {
		client = *c.HTTPClient
	} else if stream {
		client = http.Client{}
	} else {
		client = http.Client{Timeout: 30 * time.Second}
	}
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return &client
}

func responseError(response *http.Response) error {
	defer response.Body.Close()
	var payload struct {
		Error string `json:"error"`
	}
	_ = decodeJSON(response.Body, &payload)
	code := strings.TrimSpace(payload.Error)
	if code == "" {
		return fmt.Errorf("request failed with status %s", response.Status)
	}
	return fmt.Errorf("request failed with status %s: %s", response.Status, code)
}

func endpoint(baseURL, path string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", errors.New("API base URL must be an http(s) URL without credentials, query, or fragment")
	}
	return strings.TrimRight(parsed.String(), "/") + path, nil
}

func decodeJSON(reader io.Reader, target any) error {
	decoder := json.NewDecoder(io.LimitReader(reader, maxResponseBytes))
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("response contains multiple JSON values")
	}
	return nil
}

func isSuccess(status int) bool {
	return status >= http.StatusOK && status < http.StatusMultipleChoices
}

func getenv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
