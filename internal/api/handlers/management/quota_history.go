package management

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/quotahistory"
)

const defaultQuotaHistorySpan = 15 * 24 * time.Hour

func (h *Handler) GetQuotaHistory(c *gin.Context) {
	now := time.Now()
	since := now.Add(-defaultQuotaHistorySpan)
	if raw := strings.TrimSpace(c.Query("since")); raw != "" {
		parsed, errParse := time.Parse(time.RFC3339, raw)
		if errParse != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "since must be an RFC 3339 time"})
			return
		}
		since = parsed
	}
	if oldest := now.Add(-quotahistory.Retention); since.Before(oldest) {
		since = oldest
	}
	dir := quotahistory.Dir(h.configFilePath)
	readings := []quotahistory.Reading{}
	if dir != "" {
		stored, errRead := quotahistory.NewStore(dir).Since(since)
		if errRead != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read quota history"})
			return
		}
		readings = stored
	}
	c.JSON(http.StatusOK, gin.H{
		"interval_seconds": int(quotahistory.Interval / time.Second),
		"since":            since.UTC().Format(time.RFC3339),
		"readings":         readings,
	})
}
