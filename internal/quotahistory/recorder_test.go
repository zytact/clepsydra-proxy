package quotahistory

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

type staticAuths []*auth.Auth

func (s staticAuths) List() []*auth.Auth { return s }

func TestPollRecordsEachEnabledAccount(t *testing.T) {
	var gotAuthorization string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuthorization = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"five_hour":{"utilization":25,"resets_at":"2026-11-17T17:00:00Z"},"seven_day":{"utilization":50,"resets_at":null}}`))
	}))
	defer server.Close()
	limited := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer limited.Close()

	now := time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)
	store := NewStore(t.TempDir())
	recorder := NewRecorder(staticAuths{
		{ID: "claude-a.json", Provider: "claude", Metadata: map[string]any{"access_token": "token-a", "email": "a@example.com"}},
		{ID: "codex-b.json", Provider: "codex", Metadata: map[string]any{"access_token": "token-b"}},
		{ID: "claude-off.json", Provider: "claude", Disabled: true, Metadata: map[string]any{"access_token": "token-c"}},
		{ID: "gemini.json", Provider: "gemini", Metadata: map[string]any{"access_token": "token-d"}},
	}, store, func() string { return "" })
	recorder.now = func() time.Time { return now }
	recorder.urls = map[string]string{"claude": server.URL, "codex": limited.URL}

	recorder.Poll(context.Background())

	readings, errSince := store.Since(now.Add(-time.Minute))
	if errSince != nil {
		t.Fatal(errSince)
	}
	if len(readings) != 2 {
		t.Fatalf("readings = %+v, want the claude and codex accounts", readings)
	}
	claude, codex := readings[0], readings[1]
	if claude.AuthID != "claude-a.json" || claude.Email != "a@example.com" || len(claude.Windows) != 2 || claude.Error != "" {
		t.Fatalf("claude reading = %+v", claude)
	}
	if claude.AuthIndex == "" {
		t.Fatal("claude reading has no auth index")
	}
	if gotAuthorization != "Bearer token-a" {
		t.Fatalf("Authorization = %q", gotAuthorization)
	}
	if codex.Error != "usage request returned status 429" || len(codex.Windows) != 0 {
		t.Fatalf("codex reading = %+v", codex)
	}
}

func TestPollRecordsExpiredTokenWithoutRequest(t *testing.T) {
	requested := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { requested = true }))
	defer server.Close()
	now := time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)
	store := NewStore(t.TempDir())
	recorder := NewRecorder(staticAuths{{
		ID:       "claude-a.json",
		Provider: "claude",
		Metadata: map[string]any{"access_token": "old", "expired": now.Add(-time.Hour).Format(time.RFC3339)},
	}}, store, func() string { return "" })
	recorder.now = func() time.Time { return now }
	recorder.urls = map[string]string{"claude": server.URL}

	recorder.Poll(context.Background())

	readings, _ := store.Since(now.Add(-time.Minute))
	if requested || len(readings) != 1 || readings[0].Error != "access token expired" {
		t.Fatalf("requested = %v, readings = %+v", requested, readings)
	}
}
