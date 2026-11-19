package auth

import (
	"context"
	"net/http"
	"slices"
	"sync"
	"testing"

	internalconfig "github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/registry"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/executor"
)

type creditsRecordingExecutor struct {
	mu      sync.Mutex
	authIDs []string
}

func (e *creditsRecordingExecutor) record(auth *Auth) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.authIDs = append(e.authIDs, auth.ID)
}

func (e *creditsRecordingExecutor) ran(id string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return slices.Contains(e.authIDs, id)
}

func (e *creditsRecordingExecutor) Identifier() string { return "antigravity" }

func creditsServe(ctx context.Context, auth *Auth) bool {
	return AntigravityCreditsRequested(ctx) && auth.ID != "ag-normal"
}

func (e *creditsRecordingExecutor) Execute(ctx context.Context, auth *Auth, _ cliproxyexecutor.Request, _ cliproxyexecutor.Options) (cliproxyexecutor.Response, error) {
	e.record(auth)
	if !creditsServe(ctx, auth) {
		return cliproxyexecutor.Response{}, &Error{HTTPStatus: http.StatusTooManyRequests, Message: "quota exhausted"}
	}
	return cliproxyexecutor.Response{Payload: []byte(auth.ID)}, nil
}

func (e *creditsRecordingExecutor) ExecuteStream(ctx context.Context, auth *Auth, _ cliproxyexecutor.Request, _ cliproxyexecutor.Options) (*cliproxyexecutor.StreamResult, error) {
	e.record(auth)
	ch := make(chan cliproxyexecutor.StreamChunk, 1)
	if creditsServe(ctx, auth) {
		ch <- cliproxyexecutor.StreamChunk{Payload: []byte(auth.ID)}
	} else {
		ch <- cliproxyexecutor.StreamChunk{Err: &Error{HTTPStatus: http.StatusTooManyRequests, Message: "quota exhausted"}}
	}
	close(ch)
	return &cliproxyexecutor.StreamResult{Headers: http.Header{}, Chunks: ch}, nil
}

func (e *creditsRecordingExecutor) Refresh(_ context.Context, auth *Auth) (*Auth, error) {
	return auth, nil
}

func (e *creditsRecordingExecutor) CountTokens(_ context.Context, auth *Auth, _ cliproxyexecutor.Request, _ cliproxyexecutor.Options) (cliproxyexecutor.Response, error) {
	e.record(auth)
	return cliproxyexecutor.Response{}, nil
}

func (e *creditsRecordingExecutor) PrepareRequest(req *http.Request, auth *Auth) error {
	e.record(auth)
	req.Header.Set("Authorization", "Bearer token")
	return nil
}

func (e *creditsRecordingExecutor) HttpRequest(_ context.Context, auth *Auth, _ *http.Request) (*http.Response, error) {
	e.record(auth)
	return nil, nil
}

func newCreditsManager(t *testing.T, model string) (*Manager, *creditsRecordingExecutor) {
	t.Helper()
	executor := &creditsRecordingExecutor{}
	manager := NewManager(nil, nil, nil)
	manager.SetConfig(&internalconfig.Config{QuotaExceeded: internalconfig.QuotaExceeded{AntigravityCredits: true}})
	manager.RegisterExecutor(executor)
	for _, auth := range []*Auth{{ID: "ag-normal", Provider: "antigravity", Status: StatusActive}, newQuotaOnlyAuth("antigravity-cli")} {
		registry.GetGlobalRegistry().RegisterClient(auth.ID, "antigravity", []*registry.ModelInfo{{ID: model}})
		t.Cleanup(func() { registry.GetGlobalRegistry().UnregisterClient(auth.ID) })
		if _, errRegister := manager.Register(WithSkipPersist(context.Background()), auth); errRegister != nil {
			t.Fatal(errRegister)
		}
	}
	return manager, executor
}

func TestCreditsFallbackNeverRunsQuotaOnlyAuth(t *testing.T) {
	const model = "claude-opus-4-6-thinking"
	request := cliproxyexecutor.Request{Model: model}

	manager, executor := newCreditsManager(t, model)
	if resp, errExecute := manager.Execute(context.Background(), []string{"antigravity"}, request, cliproxyexecutor.Options{}); errExecute == nil {
		t.Fatalf("Execute succeeded with %q; only the quota-only auth could have served it", resp.Payload)
	}
	if !executor.ran("ag-normal") || executor.ran("antigravity-cli") {
		t.Fatalf("non-streaming credits fallback ran the quota-only auth: %v", executor.authIDs)
	}

	manager, executor = newCreditsManager(t, model)
	if stream, errStream := manager.ExecuteStream(context.Background(), []string{"antigravity"}, request, cliproxyexecutor.Options{}); errStream == nil {
		for chunk := range stream.Chunks {
			if chunk.Err == nil {
				t.Fatalf("stream produced %q; only the quota-only auth could have served it", chunk.Payload)
			}
		}
	}
	if !executor.ran("ag-normal") || executor.ran("antigravity-cli") {
		t.Fatalf("streaming credits fallback ran the quota-only auth: %v", executor.authIDs)
	}
}

func TestPinnedQuotaOnlyAuthNeverRuns(t *testing.T) {
	const model = "claude-opus-4-6-thinking"
	manager, executor := newCreditsManager(t, model)
	pinned := cliproxyexecutor.Options{Metadata: map[string]any{cliproxyexecutor.PinnedAuthMetadataKey: "antigravity-cli"}}
	request := cliproxyexecutor.Request{Model: model}

	if _, errExecute := manager.Execute(context.Background(), []string{"antigravity"}, request, pinned); errExecute == nil {
		t.Fatal("pinned Execute succeeded on the quota-only auth")
	}
	if stream, errStream := manager.ExecuteStream(context.Background(), []string{"antigravity"}, request, pinned); errStream == nil {
		for chunk := range stream.Chunks {
			if chunk.Err == nil {
				t.Fatal("pinned ExecuteStream produced output from the quota-only auth")
			}
		}
	}
	if _, errCount := manager.ExecuteCount(context.Background(), []string{"antigravity"}, request, pinned); errCount == nil {
		t.Fatal("pinned ExecuteCount succeeded on the quota-only auth")
	}
	if executor.ran("antigravity-cli") {
		t.Fatalf("pinned request ran the quota-only auth: %v", executor.authIDs)
	}
}

func TestExecutionGuardRefusesQuotaOnlyAuth(t *testing.T) {
	executor := &creditsRecordingExecutor{}
	auth := newQuotaOnlyAuth("antigravity-cli")
	ctx := context.Background()
	request := cliproxyexecutor.Request{Model: "m"}
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://example.com", nil)

	_, errExecute := guardedExecute(ctx, executor, auth, request, cliproxyexecutor.Options{})
	_, errStream := guardedExecuteStream(ctx, executor, auth, request, cliproxyexecutor.Options{})
	_, errCount := guardedCountTokens(ctx, executor, auth, request, cliproxyexecutor.Options{})
	_, errHTTP := guardedHttpRequest(ctx, executor, auth, httpReq)
	errPrepare := guardedPrepareRequest(executor, httpReq, auth)
	for name, err := range map[string]error{"execute": errExecute, "stream": errStream, "count": errCount, "http": errHTTP, "prepare": errPrepare} {
		guardErr, ok := err.(*Error)
		if !ok || guardErr.Code != "auth_quota_only" || guardErr.Retryable {
			t.Errorf("%s: err = %v, want non-retryable auth_quota_only", name, err)
		}
	}
	if len(executor.authIDs) != 0 {
		t.Fatalf("guard let the executor run: %v", executor.authIDs)
	}
}
