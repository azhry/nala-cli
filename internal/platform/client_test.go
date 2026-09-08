package platform

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/azhry/nala-cli/internal/config"
)

func TestClientOperationsUseSharedSessionAndServiceBoundaries(t *testing.T) {
	store := &config.SessionStore{Path: filepath.Join(t.TempDir(), "session.json")}
	if err := store.Save("shared-session-token"); err != nil {
		t.Fatalf("save session: %v", err)
	}

	var labsRequests atomic.Int32
	labs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer shared-session-token" {
			t.Errorf("labs Authorization = %q", got)
		}
		labsRequests.Add(1)
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/apps":
			if r.URL.Query().Get("page") != "3" || r.URL.Query().Get("pageSize") != "7" {
				t.Errorf("list query = %s, want page=3&pageSize=7", r.URL.RawQuery)
			}
			writeTestJSON(t, w, `{"apps":[{"id":42,"ownerId":"user-1","name":"demo","repositoryUrl":"https://example.test/demo","createdAt":"2026-01-02T03:04:05Z","updatedAt":"2026-01-02T03:04:05Z"}],"pagination":{"page":3,"pageSize":7,"total":1,"hasNextPage":false}}`)
		case r.Method == http.MethodGet && r.URL.Path == "/api/apps/42":
			writeTestJSON(t, w, `{"app":{"id":42,"ownerId":"user-1","name":"demo","displayName":"Demo","repositoryUrl":"https://example.test/demo","componentType":"web","descriptorVersion":1,"domain":{"subdomain":"demo","configuredDomain":null},"health":"healthy","latestDeployment":null}}`)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/apps/42":
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	defer labs.Close()

	var svcRequests atomic.Int32
	svc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer shared-session-token" {
			t.Errorf("svc Authorization = %q", got)
		}
		svcRequests.Add(1)
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/deployments":
			var request deploymentRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Errorf("decode deployment request: %v", err)
			}
			if request != (deploymentRequest{AppID: 42, SourceRef: "main", IdempotencyKey: "retry-1"}) {
				t.Errorf("deployment request = %+v", request)
			}
			writeTestJSON(t, w, `{"id":9,"appId":42,"status":"queued","stage":"queued","sourceRef":"main","createdAt":"2026-01-02T03:04:05Z","updatedAt":"2026-01-02T03:04:05Z"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/api/deployments/9":
			writeTestJSON(t, w, `{"id":9,"appId":42,"status":"running","stage":"building","sourceRef":"main","createdAt":"2026-01-02T03:04:05Z","updatedAt":"2026-01-02T03:04:06Z"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer svc.Close()

	client := NewClient(labs.URL, svc.URL, store)
	apps, err := client.ListApps(context.Background(), 3, 7)
	if err != nil {
		t.Fatalf("ListApps() error = %v", err)
	}
	if len(apps.Apps) != 1 || apps.Apps[0].ID != 42 || apps.Pagination.Page != 3 {
		t.Fatalf("ListApps() = %+v", apps)
	}

	detail, err := client.GetApp(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetApp() error = %v", err)
	}
	if detail.App.ID != 42 || detail.App.Health != "healthy" {
		t.Fatalf("GetApp() = %+v", detail)
	}

	created, err := client.CreateDeployment(context.Background(), 42, " main ", " retry-1 ")
	if err != nil {
		t.Fatalf("CreateDeployment() error = %v", err)
	}
	if created.ID != 9 || created.Status != "queued" {
		t.Fatalf("CreateDeployment() = %+v", created)
	}

	deployment, err := client.GetDeployment(context.Background(), 9)
	if err != nil {
		t.Fatalf("GetDeployment() error = %v", err)
	}
	if deployment.ID != 9 || deployment.Stage != "building" {
		t.Fatalf("GetDeployment() = %+v", deployment)
	}

	if err := client.DeleteApp(context.Background(), 42); err != nil {
		t.Fatalf("DeleteApp() error = %v", err)
	}
	if labsRequests.Load() != 3 {
		t.Fatalf("labs request count = %d, want 3", labsRequests.Load())
	}
	if svcRequests.Load() != 2 {
		t.Fatalf("svc request count = %d, want 2", svcRequests.Load())
	}
}

func TestFollowDeploymentEventsParsesSSEAndIgnoresKeepAlive(t *testing.T) {
	store := &config.SessionStore{Path: filepath.Join(t.TempDir(), "session.json")}
	if err := store.Save("shared-session-token"); err != nil {
		t.Fatalf("save session: %v", err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/deployments/9/events" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("cursor") != "4" {
			t.Errorf("cursor = %q, want 4", r.URL.Query().Get("cursor"))
		}
		if got := r.Header.Get("Accept"); got != "text/event-stream" {
			t.Errorf("Accept = %q", got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer shared-session-token" {
			t.Errorf("Authorization = %q", got)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("id: 5\nevent: deployment.stage\ndata: {\"deploymentId\":9,\"status\":\"running\",\"stage\":\"building\",\"message\":\"building\"}\n\n: keep-alive\n\nid: 6\ndata: {\"id\":6,\"deploymentId\":9,\"event\":\"deployment.completed\",\"status\":\"succeeded\"}\n\n"))
	}))
	defer server.Close()

	var events []DeploymentEvent
	err := NewClient(server.URL, server.URL, store).FollowDeploymentEvents(context.Background(), 9, 4, func(event DeploymentEvent) error {
		events = append(events, event)
		return nil
	})
	if err != nil {
		t.Fatalf("FollowDeploymentEvents() error = %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("events = %+v, want two events", events)
	}
	if events[0].ID != 5 || events[0].Type != "deployment.stage" || events[0].Stage != "building" {
		t.Fatalf("first event = %+v", events[0])
	}
	if events[1].ID != 6 || events[1].Type != "deployment.completed" || events[1].Status != "succeeded" {
		t.Fatalf("second event = %+v", events[1])
	}
}

func TestClientValidatesInputsAndDoesNotRequest(t *testing.T) {
	store := &config.SessionStore{Path: filepath.Join(t.TempDir(), "session.json")}
	if err := store.Save("shared-session-token"); err != nil {
		t.Fatalf("save session: %v", err)
	}
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { requests.Add(1) }))
	defer server.Close()
	client := NewClient(server.URL, server.URL, store)

	checks := []struct {
		name string
		call func() error
	}{
		{"list page", func() error { _, err := client.ListApps(context.Background(), 0, 20); return err }},
		{"list page size", func() error { _, err := client.ListApps(context.Background(), 1, 101); return err }},
		{"get app", func() error { _, err := client.GetApp(context.Background(), 0); return err }},
		{"deploy app", func() error { _, err := client.CreateDeployment(context.Background(), 0, "main", "retry"); return err }},
		{"deploy source ref", func() error { _, err := client.CreateDeployment(context.Background(), 1, " ", "retry"); return err }},
		{"monitor cursor", func() error {
			return client.FollowDeploymentEvents(context.Background(), 1, -1, func(DeploymentEvent) error { return nil })
		}},
		{"delete app", func() error { return client.DeleteApp(context.Background(), 0) }},
	}
	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			if err := check.call(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
	if requests.Load() != 0 {
		t.Fatalf("request count = %d, want 0", requests.Load())
	}
}

func TestResponseErrorDoesNotExposeResponseBody(t *testing.T) {
	store := &config.SessionStore{Path: filepath.Join(t.TempDir(), "session.json")}
	if err := store.Save("shared-session-token"); err != nil {
		t.Fatalf("save session: %v", err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthenticated","detail":"secret-response-body"}`))
	}))
	defer server.Close()

	_, err := NewClient(server.URL, server.URL, store).GetApp(context.Background(), 42)
	if err == nil || !strings.Contains(err.Error(), "unauthenticated") {
		t.Fatalf("GetApp() error = %v, want unauthenticated", err)
	}
	if strings.Contains(err.Error(), "secret-response-body") || strings.Contains(err.Error(), "shared-session-token") {
		t.Fatalf("GetApp() error exposed sensitive data: %v", err)
	}
}

func writeTestJSON(t *testing.T, w http.ResponseWriter, payload string) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write([]byte(payload)); err != nil {
		t.Errorf("write response: %v", err)
	}
}
