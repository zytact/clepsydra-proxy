package quotahistory

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

type staticAuths []*auth.Auth

func (s staticAuths) List() []*auth.Auth { return s }

func withURLs(urls map[string]string) map[string]usageSource {
	sources := map[string]usageSource{}
	for provider, source := range defaultSources() {
		if url, ok := urls[provider]; ok {
			source.urls = fixedURLs(url)
			sources[provider] = source
		}
	}
	return sources
}

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
	recorder.sources = withURLs(map[string]string{"claude": server.URL, "codex": limited.URL})

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
	recorder.sources = withURLs(map[string]string{"claude": server.URL})

	recorder.Poll(context.Background())

	readings, _ := store.Since(now.Add(-time.Minute))
	if requested || len(readings) != 1 || readings[0].Error != "access token expired" {
		t.Fatalf("requested = %v, readings = %+v", requested, readings)
	}
}

func TestPollPostsAntigravityProjectAndFallsBackToTheSecondHost(t *testing.T) {
	daily := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer daily.Close()
	var gotMethod, gotBody, gotContentType, gotUserAgent string
	prod := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotContentType = r.Header.Get("Content-Type")
		gotUserAgent = r.Header.Get("User-Agent")
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		_, _ = w.Write([]byte(`{"groups":[{"buckets":[{"bucketId":"gemini-5h","window":"5h","remainingFraction":0.5}]}]}`))
	}))
	defer prod.Close()

	now := time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)
	store := NewStore(t.TempDir())
	recorder := NewRecorder(staticAuths{
		{ID: "antigravity-a.json", Provider: "antigravity", Metadata: map[string]any{"access_token": "token-a", "project_id": "project-a"}},
	}, store, func() string { return "" })
	recorder.now = func() time.Time { return now }
	source := defaultSources()["antigravity"]
	source.urls = fixedURLs(daily.URL, prod.URL)
	recorder.sources = map[string]usageSource{"antigravity": source}

	recorder.Poll(context.Background())

	readings, errSince := store.Since(now.Add(-time.Minute))
	if errSince != nil {
		t.Fatal(errSince)
	}
	if len(readings) != 1 || readings[0].Error != "" || len(readings[0].Windows) != 1 || readings[0].Windows[0].UsedPercent != 50 {
		t.Fatalf("readings = %+v", readings)
	}
	if gotMethod != http.MethodPost || gotBody != `{"project":"project-a"}` || gotContentType != "application/json" {
		t.Fatalf("request = %s %q content-type %q", gotMethod, gotBody, gotContentType)
	}
	if !strings.HasPrefix(gotUserAgent, "antigravity/") {
		t.Fatalf("User-Agent = %q", gotUserAgent)
	}
}

func pollAntigravity(t *testing.T, account *auth.Auth, urls ...string) Reading {
	t.Helper()
	now := time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)
	store := NewStore(t.TempDir())
	recorder := NewRecorder(staticAuths{account}, store, func() string { return "" })
	recorder.now = func() time.Time { return now }
	source := defaultSources()["antigravity"]
	if len(urls) > 0 {
		source.urls = fixedURLs(urls...)
	}
	recorder.sources = map[string]usageSource{"antigravity": source}
	recorder.Poll(context.Background())
	readings, errSince := store.Since(now.Add(-time.Minute))
	if errSince != nil {
		t.Fatal(errSince)
	}
	if len(readings) != 1 {
		t.Fatalf("readings = %+v", readings)
	}
	return readings[0]
}

func antigravityAccount(metadata map[string]any) *auth.Auth {
	metadata["access_token"] = "token-a"
	return &auth.Auth{ID: "antigravity-a.json", Provider: "antigravity", Metadata: metadata}
}

func respond(status int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
}

func TestPollFallsBackWhenTheFirstHostAnswersAnUnreadableBody(t *testing.T) {
	daily := respond(http.StatusOK, `{"groups":[]}`)
	defer daily.Close()
	prod := respond(http.StatusOK, `{"buckets":[{"bucketId":"gemini-5h","window":"5h","remainingFraction":0.5}]}`)
	defer prod.Close()

	reading := pollAntigravity(t, antigravityAccount(map[string]any{}), daily.URL, prod.URL)

	if reading.Error != "" || len(reading.Windows) != 1 {
		t.Fatalf("reading = %+v", reading)
	}
}

func TestPollStopsOnAnAuthErrorFromTheFirstHost(t *testing.T) {
	for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden} {
		daily := respond(status, "")
		prodRequested := false
		prod := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { prodRequested = true }))

		reading := pollAntigravity(t, antigravityAccount(map[string]any{}), daily.URL, prod.URL)
		daily.Close()
		prod.Close()

		if prodRequested {
			t.Fatalf("status %d: fell back to the second host", status)
		}
		if reading.Error != (statusError{status: status}).Error() {
			t.Fatalf("status %d: reading error = %q", status, reading.Error)
		}
	}
}

func TestPollReportsTheFirstHostErrorWhenEveryHostFails(t *testing.T) {
	daily := respond(http.StatusServiceUnavailable, "")
	defer daily.Close()
	prod := respond(http.StatusNotFound, "")
	defer prod.Close()

	reading := pollAntigravity(t, antigravityAccount(map[string]any{}), daily.URL, prod.URL)

	if reading.Error != "usage request returned status 503" {
		t.Fatalf("reading error = %q", reading.Error)
	}
}

func TestAntigravityQuotaURLsPinACustomBaseURL(t *testing.T) {
	custom := antigravityAccount(map[string]any{"base_url": "https://example.test/"})
	if got := antigravityQuotaURLs(custom); len(got) != 1 || got[0] != "https://example.test/v1internal:retrieveUserQuotaSummary" {
		t.Fatalf("custom urls = %v", got)
	}
	if got := antigravityQuotaURLs(antigravityAccount(map[string]any{})); len(got) != 2 ||
		!strings.HasPrefix(got[0], "https://daily-cloudcode-pa.") || !strings.HasPrefix(got[1], "https://cloudcode-pa.") {
		t.Fatalf("default urls = %v", got)
	}
}

func TestPollUsesOnlyTheCustomBaseURL(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	reading := pollAntigravity(t, antigravityAccount(map[string]any{"base_url": server.URL}))

	if gotPath != "/v1internal:retrieveUserQuotaSummary" || reading.Error != "usage request returned status 503" {
		t.Fatalf("path %q, reading = %+v", gotPath, reading)
	}
}
