package quotahistory

import (
	"testing"
	"time"
)

func TestParseClaudeUsageKeepsWindowsAndScopes(t *testing.T) {
	body := []byte(`{
		"five_hour": {"utilization": 7.0, "resets_at": "2026-11-17T17:00:00.160235+00:00"},
		"seven_day": {"utilization": 100.0, "resets_at": "2026-11-18T02:00:00.160263+00:00"},
		"seven_day_opus": null,
		"limits": [
			{"kind": "weekly_scoped", "percent": 40, "resets_at": "2026-11-20T02:00:00Z", "is_active": true,
			 "scope": {"model": {"display_name": "Fable"}}}
		]
	}`)

	windows, errParse := parseClaudeUsage(body)
	if errParse != nil {
		t.Fatal(errParse)
	}
	want := []struct {
		id    string
		used  float64
		scope string
	}{{"five-hour", 7, ScopeAll}, {"seven-day", 100, ScopeAll}, {"seven-day-fable", 40, ScopeModel}}
	if len(windows) != len(want) {
		t.Fatalf("windows = %+v", windows)
	}
	for i, w := range want {
		got := windows[i]
		if got.ID != w.id || got.UsedPercent != w.used || got.Scope != w.scope || got.ResetsAt == nil {
			t.Fatalf("window %d = %+v, want %+v", i, got, w)
		}
	}
	if windows[0].ResetsAt.Format(time.RFC3339) != "2026-11-17T17:00:00Z" {
		t.Fatalf("five-hour reset = %s", windows[0].ResetsAt)
	}
}

func TestParseClaudeUsageRejectsUnexpectedBody(t *testing.T) {
	if _, errParse := parseClaudeUsage([]byte(`{"error":"nope"}`)); errParse == nil {
		t.Fatal("expected an error for a body without windows")
	}
}

func TestParseCodexUsageNamesWindowsByDuration(t *testing.T) {
	now := time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)
	body := []byte(`{"rate_limit": {
		"primary_window": {"used_percent": 30, "limit_window_seconds": 604800, "reset_at": 1795000000},
		"secondary_window": {"used_percent": 12.5, "limit_window_seconds": 18000, "reset_after_seconds": 600}
	}}`)

	windows, errParse := parseCodexUsage(body, now)
	if errParse != nil {
		t.Fatal(errParse)
	}
	if len(windows) != 2 || windows[0].ID != "weekly" || windows[1].ID != "five-hour" {
		t.Fatalf("windows = %+v", windows)
	}
	if !windows[0].ResetsAt.Equal(time.Unix(1795000000, 0)) {
		t.Fatalf("weekly reset = %s", windows[0].ResetsAt)
	}
	if !windows[1].ResetsAt.Equal(now.Add(10 * time.Minute)) {
		t.Fatalf("five-hour reset = %s", windows[1].ResetsAt)
	}
}
