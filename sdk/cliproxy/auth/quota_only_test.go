package auth

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/registry"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/executor"
)

type quotaOnlyRecordingExecutor struct {
	schedulerTestExecutor
	executed *int
}

func (e quotaOnlyRecordingExecutor) Execute(context.Context, *Auth, cliproxyexecutor.Request, cliproxyexecutor.Options) (cliproxyexecutor.Response, error) {
	*e.executed++
	return cliproxyexecutor.Response{}, nil
}

func (e quotaOnlyRecordingExecutor) ExecuteStream(context.Context, *Auth, cliproxyexecutor.Request, cliproxyexecutor.Options) (*cliproxyexecutor.StreamResult, error) {
	*e.executed++
	return nil, nil
}

func newQuotaOnlyAuth(id string) *Auth {
	return &Auth{
		ID:         id,
		Provider:   "antigravity",
		Status:     StatusActive,
		Attributes: map[string]string{AttributeRuntimeOnly: "true", AttributeQuotaOnly: "true"},
		Metadata: map[string]any{
			"type":          "antigravity",
			"access_token":  "token",
			"refresh_token": "refresh",
			"expired":       time.Now().Add(time.Hour).Format(time.RFC3339),
		},
	}
}

func TestQuotaOnlyAuthIsNeverAvailable(t *testing.T) {
	if IsAuthAvailableForModel(newQuotaOnlyAuth("quota-only"), "gemini-3-pro", time.Now()) {
		t.Fatal("quota-only auth is available for routing")
	}
}

func TestQuotaOnlyAuthNeverServesRequests(t *testing.T) {
	const model = "quota-only-model"
	executed := 0
	manager := NewManager(nil, nil, nil)
	manager.RegisterExecutor(quotaOnlyRecordingExecutor{schedulerTestExecutor: schedulerTestExecutor{provider: "antigravity"}, executed: &executed})
	auth := newQuotaOnlyAuth(t.Name())
	registry.GetGlobalRegistry().RegisterClient(auth.ID, "antigravity", []*registry.ModelInfo{{ID: model}})
	t.Cleanup(func() { registry.GetGlobalRegistry().UnregisterClient(auth.ID) })
	if _, errRegister := manager.Register(WithSkipPersist(context.Background()), auth); errRegister != nil {
		t.Fatal(errRegister)
	}

	if manager.HasProviderAuth("antigravity") {
		t.Fatal("HasProviderAuth(antigravity) = true with only a quota-only auth")
	}
	request := cliproxyexecutor.Request{Model: model, Payload: []byte(`{}`)}
	pinned := cliproxyexecutor.Options{Metadata: map[string]any{cliproxyexecutor.PinnedAuthMetadataKey: auth.ID}, Headers: http.Header{}}
	for _, opts := range []cliproxyexecutor.Options{{}, pinned} {
		if _, errExecute := manager.Execute(context.Background(), []string{"antigravity"}, request, opts); errExecute == nil {
			t.Fatal("Execute succeeded through a quota-only auth")
		}
		if _, errStream := manager.ExecuteStream(context.Background(), []string{"antigravity"}, request, opts); errStream == nil {
			t.Fatal("ExecuteStream succeeded through a quota-only auth")
		}
	}
	if executed != 0 {
		t.Fatalf("executor ran %d times through a quota-only auth", executed)
	}
	if _, errVerify := manager.VerifyInference(context.Background(), auth.ID, model); errVerify == nil {
		t.Fatal("VerifyInference succeeded on a quota-only auth")
	}
}

func TestQuotaOnlyAuthIsNotPersisted(t *testing.T) {
	store := &countingStore{}
	manager := NewManager(store, nil, nil)
	auth := newQuotaOnlyAuth(t.Name())
	if _, errRegister := manager.Register(context.Background(), auth); errRegister != nil {
		t.Fatal(errRegister)
	}
	auth.Metadata["access_token"] = "refreshed"
	if _, errUpdate := manager.Update(context.Background(), auth); errUpdate != nil {
		t.Fatal(errUpdate)
	}
	if saves := store.saveCount.Load(); saves != 0 {
		t.Fatalf("store saved a quota-only auth %d times", saves)
	}
}
