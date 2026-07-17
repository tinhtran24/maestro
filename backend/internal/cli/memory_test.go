package cli

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestMemoryListRendersTasks(t *testing.T) {
	cfg := setConfigEnv(t)
	var requests []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appendPrimaryRequest(&requests, r)
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/acme/memory/tasks" {
			_, _ = io.WriteString(w, `{"tasks":[
				{"id":"acme-7","sessionId":"acme-7","kind":"worker","intent":"add the widget","taskType":"feature","branch":"feat/widget","occurredAt":"2026-07-17T08:00:00Z","changedFiles":["a.go","b.go"],"changedTests":["a_test.go"]}
			]}`)
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)
	writeRunFileFor(t, cfg, srv)

	out, errOut, err := executeCLI(t, Deps{ProcessAlive: func(int) bool { return true }}, "memory", "ls", "acme")
	if err != nil {
		t.Fatalf("memory ls failed: %v stderr=%s", err, errOut)
	}
	for _, want := range []string{"acme-7", "feature", "worker", "add the widget"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
	if want := []string{"GET /api/v1/projects/acme/memory/tasks"}; !reflect.DeepEqual(requests, want) {
		t.Fatalf("requests=%#v want %#v", requests, want)
	}
}

func TestMemoryListEmpty(t *testing.T) {
	cfg := setConfigEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"tasks":[]}`)
	}))
	t.Cleanup(srv.Close)
	writeRunFileFor(t, cfg, srv)

	out, errOut, err := executeCLI(t, Deps{ProcessAlive: func(int) bool { return true }}, "memory", "ls", "acme")
	if err != nil {
		t.Fatalf("memory ls failed: %v stderr=%s", err, errOut)
	}
	if !strings.Contains(out, "No task memory") {
		t.Fatalf("expected empty-state message, got:\n%s", out)
	}
}

func TestMemoryListSurfacesDaemonErrorEnvelope(t *testing.T) {
	cfg := setConfigEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":"not_found","code":"PROJECT_NOT_FOUND","message":"Unknown project","requestId":"req-123"}`)
	}))
	t.Cleanup(srv.Close)
	writeRunFileFor(t, cfg, srv)

	_, errOut, err := executeCLI(t, Deps{ProcessAlive: func(int) bool { return true }}, "memory", "ls", "ghost")
	if err == nil {
		t.Fatal("expected an error for a not-found project")
	}
	for _, want := range []string{"Unknown project", "PROJECT_NOT_FOUND", "req-123"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error should surface the daemon envelope %q, got: %v", want, err)
		}
	}
	_ = errOut
}

func TestMemoryContextRequiresRole(t *testing.T) {
	cfg := setConfigEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "should not be called", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	writeRunFileFor(t, cfg, srv)

	_, _, err := executeCLI(t, Deps{ProcessAlive: func(int) bool { return true }}, "memory", "context", "acme")
	if err == nil {
		t.Fatal("expected a usage error when --role is omitted")
	}
	if !strings.Contains(err.Error(), "role") {
		t.Fatalf("error should mention the required role flag, got: %v", err)
	}
}
