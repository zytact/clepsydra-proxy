package management

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
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
		ID:        "antigravity-cli",
		Provider:  "antigravity",
		Label:     "cli@example.com",
		Status:    coreauth.StatusActive,
		QuotaOnly: true,
		Attributes: map[string]string{
			coreauth.AttributeRuntimeOnly: "true",
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
		"relogin":  serveQuotaOnly(h.StartOAuthV8, http.MethodGet, "/v8/management/oauth/auth-url?provider=antigravity&name=antigravity-cli", ""),
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

func TestQuotaOnlyAPICallAllowlist(t *testing.T) {
	auth := &coreauth.Auth{QuotaOnly: true, Attributes: map[string]string{"base_url": "https://proxy.example.com/agy/"}}
	for target, want := range map[string]bool{
		"https://daily-cloudcode-pa.googleapis.com/v1internal:retrieveUserQuotaSummary":         true,
		"https://daily-cloudcode-pa.sandbox.googleapis.com/v1internal:retrieveUserQuotaSummary": true,
		"https://cloudcode-pa.googleapis.com/v1internal:retrieveUserQuotaSummary":               true,
		"https://daily-cloudcode-pa.googleapis.com/v1internal:loadCodeAssist":                   true,
		"https://cloudcode-pa.googleapis.com/v1internal:fetchAvailableModels":                   true,
		"https://proxy.example.com/agy/v1internal:retrieveUserQuotaSummary":                     true,
		"https://daily-cloudcode-pa.googleapis.com/v1internal:streamGenerateContent":            false,
		"https://daily-cloudcode-pa.googleapis.com/v1internal:onboardUser":                      false,
		"https://daily-cloudcode-pa.googleapis.com/v1internal:loadCodeAssist?alt=sse":           false,
		"http://daily-cloudcode-pa.googleapis.com/v1internal:loadCodeAssist":                    false,
		"https://evil.example.com/v1internal:retrieveUserQuotaSummary":                          false,
		"https://user@cloudcode-pa.googleapis.com/v1internal:loadCodeAssist":                    false,
	} {
		parsed, errParse := url.Parse(target)
		if errParse != nil {
			t.Fatal(errParse)
		}
		if got := quotaOnlyAPICallAllowed(auth, http.MethodPost, parsed); got != want {
			t.Errorf("POST %s allowed = %v, want %v", target, got, want)
		}
	}
	parsed, _ := url.Parse("https://cloudcode-pa.googleapis.com/v1internal:loadCodeAssist")
	if quotaOnlyAPICallAllowed(auth, http.MethodGet, parsed) {
		t.Error("GET allowed for a quota-only credential")
	}
}

func TestAPICallRejectsQuotaOnlyCredentialOutsideTheAllowlist(t *testing.T) {
	h, manager := newQuotaOnlyHandler(t)
	auth, _ := manager.GetByID("antigravity-cli")
	body := `{"method":"POST","url":"https://daily-cloudcode-pa.googleapis.com/v1internal:streamGenerateContent","auth_index":"` + auth.EnsureIndex() + `","header":{"Authorization":"Bearer $TOKEN$"}}`
	rec := serveQuotaOnly(h.APICall, http.MethodPost, "/v8/management/api-call", body)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("api-call status %d, want %d: %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}
