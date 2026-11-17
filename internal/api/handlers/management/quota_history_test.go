package management

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/MoeForge/moego/json"
	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/quotahistory"
)

func TestGetQuotaHistoryReturnsStoredReadings(t *testing.T) {
	gin.SetMode(gin.TestMode)
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	recent := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
	old := time.Now().UTC().Add(-48 * time.Hour)
	store := quotahistory.NewStore(quotahistory.Dir(configPath))
	if errAppend := store.Append(
		quotahistory.Reading{At: old, Provider: "claude", AuthID: "old"},
		quotahistory.Reading{At: recent, Provider: "claude", AuthID: "recent", Windows: []quotahistory.Window{{ID: "five-hour", UsedPercent: 40, Scope: quotahistory.ScopeAll}}},
	); errAppend != nil {
		t.Fatal(errAppend)
	}
	h := &Handler{configFilePath: configPath}

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	since := time.Now().UTC().Add(-2 * time.Hour).Format(time.RFC3339)
	c.Request = httptest.NewRequest(http.MethodGet, "/v8/management/quota/history?since="+since, nil)
	h.GetQuotaHistory(c)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	var body struct {
		IntervalSeconds int                    `json:"interval_seconds"`
		Readings        []quotahistory.Reading `json:"readings"`
	}
	if errDecode := json.Unmarshal(recorder.Body.Bytes(), &body); errDecode != nil {
		t.Fatal(errDecode)
	}
	if body.IntervalSeconds != 300 || len(body.Readings) != 1 || body.Readings[0].AuthID != "recent" {
		t.Fatalf("body = %+v", body)
	}
}

func TestGetQuotaHistoryRejectsBadSince(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/v8/management/quota/history?since=yesterday", nil)
	(&Handler{configFilePath: filepath.Join(t.TempDir(), "config.yaml")}).GetQuotaHistory(c)
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), "RFC 3339") {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
}
