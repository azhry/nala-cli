package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/azhry/nala-cli/internal/config"
)

func TestRunAppListUsesFlagsAndPrintsJSON(t *testing.T) {
	server := newCLIService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/apps" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("page") != "2" || r.URL.Query().Get("pageSize") != "5" {
			t.Errorf("query = %s, want page=2&pageSize=5", r.URL.RawQuery)
		}
		if r.Header.Get("Authorization") != "Bearer cli-test-token" {
			t.Errorf("Authorization = %q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"apps":[{"id":7,"name":"demo"}],"pagination":{"page":2,"pageSize":5,"total":1,"hasNextPage":false}}`))
	})
	defer server.Close()

	var stdout, stderr bytes.Buffer
	err := run([]string{"app", "list", "--page", "2", "--page-size", "5"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run() error = %v; stderr = %s", err, stderr.String())
	}
	var response struct {
		Apps []struct {
			ID int64 `json:"id"`
		} `json:"apps"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("decode stdout: %v", err)
	}
	if len(response.Apps) != 1 || response.Apps[0].ID != 7 {
		t.Fatalf("stdout = %s", stdout.String())
	}
}

func TestRunAppDeployAndMonitorFollowUseNalaSVC(t *testing.T) {
	server := newCLIService(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/deployments":
			var payload struct {
				AppID          int64  `json:"appId"`
				SourceRef      string `json:"sourceRef"`
				IdempotencyKey string `json:"idempotencyKey"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("decode deploy payload: %v", err)
			}
			if payload.AppID != 7 || payload.SourceRef != "main" || payload.IdempotencyKey != "cli-retry-1" {
				t.Errorf("deploy payload = %+v", payload)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":12,"appId":7,"status":"queued","stage":"queued"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/deployments/12/events":
			if r.URL.Query().Get("cursor") != "3" {
				t.Errorf("cursor = %q, want 3", r.URL.Query().Get("cursor"))
			}
			if r.Header.Get("Accept") != "text/event-stream" {
				t.Errorf("Accept = %q", r.Header.Get("Accept"))
			}
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = w.Write([]byte("id: 4\nevent: deployment.completed\ndata: {\"deploymentId\":12,\"status\":\"succeeded\",\"stage\":\"succeeded\"}\n\n"))
		default:
			http.NotFound(w, r)
		}
	})
	defer server.Close()

	var deployOut, deployErr bytes.Buffer
	if err := run([]string{"app", "deploy", "--id", "7", "--source-ref", "main", "--idempotency-key", "cli-retry-1"}, &deployOut, &deployErr); err != nil {
		t.Fatalf("deploy run() error = %v; stderr = %s", err, deployErr.String())
	}
	if !bytes.Contains(deployOut.Bytes(), []byte(`"id":12`)) {
		t.Fatalf("deploy stdout = %s", deployOut.String())
	}

	var monitorOut, monitorErr bytes.Buffer
	if err := run([]string{"app", "monitor", "--deployment-id", "12", "--cursor", "3", "--follow"}, &monitorOut, &monitorErr); err != nil {
		t.Fatalf("monitor run() error = %v; stderr = %s", err, monitorErr.String())
	}
	var event struct {
		ID     int64  `json:"id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(monitorOut.Bytes(), &event); err != nil {
		t.Fatalf("decode monitor stdout: %v", err)
	}
	if event.ID != 4 || event.Status != "succeeded" {
		t.Fatalf("monitor stdout = %s", monitorOut.String())
	}
}

func TestRunAppGetAndDeleteUseNalaLabs(t *testing.T) {
	server := newCLIService(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/apps/7":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"app":{"id":7,"name":"demo","health":"healthy"}}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/api/apps/7":
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	})
	defer server.Close()

	var getOut, getErr bytes.Buffer
	if err := run([]string{"app", "get", "--id", "7"}, &getOut, &getErr); err != nil {
		t.Fatalf("get run() error = %v; stderr = %s", err, getErr.String())
	}
	if !bytes.Contains(getOut.Bytes(), []byte(`"id":7`)) {
		t.Fatalf("get stdout = %s", getOut.String())
	}

	var deleteOut, deleteErr bytes.Buffer
	if err := run([]string{"app", "delete", "--id", "7"}, &deleteOut, &deleteErr); err != nil {
		t.Fatalf("delete run() error = %v; stderr = %s", err, deleteErr.String())
	}
	if got := deleteOut.String(); got != "Deleted app 7\n" {
		t.Fatalf("delete stdout = %q", got)
	}
}

func TestRunAppRejectsInvalidIDBeforeRequest(t *testing.T) {
	var requests int
	server := newCLIService(t, func(http.ResponseWriter, *http.Request) { requests++ })
	defer server.Close()

	var stdout, stderr bytes.Buffer
	err := run([]string{"app", "get", "--id", "0"}, &stdout, &stderr)
	if err == nil || err.Error() != "app id must be positive" {
		t.Fatalf("run() error = %v, want local ID validation", err)
	}
	if requests != 0 {
		t.Fatalf("request count = %d, want 0", requests)
	}
}

func newCLIService(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Setenv("NALA_API_BASE_URL", server.URL)
	t.Setenv("NALA_SVC_BASE_URL", server.URL)
	t.Setenv("NALA_CONFIG_DIR", filepath.Join(t.TempDir(), "config"))
	store, err := config.NewSessionStore()
	if err != nil {
		t.Fatalf("new session store: %v", err)
	}
	if err := store.Save("cli-test-token"); err != nil {
		t.Fatalf("save session: %v", err)
	}
	return server
}
