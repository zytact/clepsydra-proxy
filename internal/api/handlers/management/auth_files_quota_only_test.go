package management

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MoeForge/moego/json"
	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

func newQuotaOnlyHandler(t *testing.T) (*Handler, *coreauth.Manager) {
	t.Helper()
	t.Setenv("MANAGEMENT_PASSWORD", "")
	manager := coreauth.NewManager(nil, nil, nil)
	record := &coreauth.Auth{
		ID:       "antigravity-cli",
		Provider: "antigravity",
		Label:    "cli@example.com",
		Status:   coreauth.StatusActive,
		Attributes: map[string]string{
			coreauth.AttributeRuntimeOnly: "true",
			coreauth.AttributeQuotaOnly:   "true",
			coreauth.AttributeLoginSource: "antigravity-cli",
		},
		Metadata: map[string]any{
			"type":          "antigravity",
			"email":         "cli@example.com",
			"project_id":    "project-1",
			"refresh_token": "refresh",
		},
	}
	if _, errRegister := manager.Register(coreauth.WithSkipPersist(context.Background()), record); errRegister != nil {
		t.Fatal(errRegister)
	}
	return NewHandlerWithoutConfigFilePath(&config.Config{AuthDir: t.TempDir()}, manager), manager
}

func serveQuotaOnly(handler gin.HandlerFunc, method, target, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(method, target, strings.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	handler(ctx)
	return rec
}

func TestListAuthFilesReportsQuotaOnlyCredential(t *testing.T) {
	h, _ := newQuotaOnlyHandler(t)
	rec := serveQuotaOnly(h.ListAuthFiles, http.MethodGet, "/v8/management/credentials", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("list status %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Files []map[string]any `json:"files"`
	}
	if errDecode := json.Unmarshal(rec.Body.Bytes(), &body); errDecode != nil {
		t.Fatal(errDecode)
	}
	if len(body.Files) != 1 {
		t.Fatalf("files = %v", body.Files)
	}
	entry := body.Files[0]
	for key, want := range map[string]any{
		"quota_only":   true,
		"login_source": "antigravity-cli",
		"runtime_only": true,
		"provider":     "antigravity",
		"email":        "cli@example.com",
		"project_id":   "project-1",
	} {
		if entry[key] != want {
			t.Errorf("entry[%q] = %v, want %v", key, entry[key], want)
		}
	}
}

func TestQuotaOnlyCredentialRejectsMutations(t *testing.T) {
	h, manager := newQuotaOnlyHandler(t)
	for name, rec := range map[string]*httptest.ResponseRecorder{
		"download": serveQuotaOnly(h.DownloadAuthFile, http.MethodGet, "/v8/management/credentials/download?name=antigravity-cli", ""),
		"delete":   serveQuotaOnly(h.DeleteAuthFile, http.MethodDelete, "/v8/management/credentials?name=antigravity-cli", ""),
		"status":   serveQuotaOnly(h.PatchAuthFileStatus, http.MethodPatch, "/v8/management/credentials/status", `{"name":"antigravity-cli","disabled":true}`),
		"fields":   serveQuotaOnly(h.PatchAuthFileFields, http.MethodPatch, "/v8/management/credentials/fields", `{"name":"antigravity-cli","note":"x"}`),
	} {
		if rec.Code != http.StatusConflict {
			t.Errorf("%s status %d, want %d: %s", name, rec.Code, http.StatusConflict, rec.Body.String())
		}
	}
	current, ok := manager.GetByID("antigravity-cli")
	if !ok || current.Disabled || current.Attributes["note"] != "" {
		t.Fatalf("quota-only credential changed: ok=%v auth=%+v", ok, current)
	}
}
