package quotahistory

import (
	"math"
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
		scope Scope
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

func TestParseClaudeUsageKeepsTheCreditPoolOutOfFable(t *testing.T) {
	body := []byte(`{
		"five_hour": {"utilization": 1, "resets_at": null},
		"iguana_necktie": {"utilization": 60, "resets_at": null, "limit_dollars": 50, "used_dollars": 30}
	}`)
	windows, errParse := parseClaudeUsage(body)
	if errParse != nil {
		t.Fatal(errParse)
	}
	for _, window := range windows {
		if window.ID == "seven-day-fable" {
			t.Fatalf("credit pool recorded as Fable: %+v", window)
		}
	}
}

func TestParseCodexUsageCountsAReachedLimitAsFull(t *testing.T) {
	body := []byte(`{"rate_limit": {"limit_reached": true, "primary_window": {"limit_window_seconds": 18000, "reset_after_seconds": 60}}}`)
	windows, errParse := parseCodexUsage(body, time.Now())
	if errParse != nil {
		t.Fatal(errParse)
	}
	if len(windows) != 1 || windows[0].UsedPercent != 100 {
		t.Fatalf("windows = %+v", windows)
	}
}

func TestParseCodexUsageLeavesTheOtherWindowAloneWhenANearlyFullOneExplainsTheLimit(t *testing.T) {
	body := []byte(`{"rate_limit": {"limit_reached": true, "primary_window": {"used_percent": 99.6, "limit_window_seconds": 18000}, "secondary_window": {"limit_window_seconds": 604800}}}`)
	windows, errParse := parseCodexUsage(body, time.Now())
	if errParse != nil {
		t.Fatal(errParse)
	}
	if len(windows) != 1 || windows[0].ID != "five-hour" || windows[0].UsedPercent != 99.6 {
		t.Fatalf("windows = %+v", windows)
	}
}

func TestParseCodexUsageMarksTheSilentWindowFullWhenTheReportedOneHasRoom(t *testing.T) {
	body := []byte(`{"rate_limit": {"limit_reached": true, "primary_window": {"used_percent": 12, "limit_window_seconds": 18000}, "secondary_window": {"limit_window_seconds": 604800}}}`)
	windows, errParse := parseCodexUsage(body, time.Now())
	if errParse != nil {
		t.Fatal(errParse)
	}
	if len(windows) != 2 || windows[1].ID != "weekly" || windows[1].UsedPercent != 100 {
		t.Fatalf("windows = %+v", windows)
	}
}

func TestParseAntigravityQuotaKeepsOneWindowPerBucket(t *testing.T) {
	body := []byte(`{"groups": [
		{"displayName": "Gemini", "buckets": [
			{"bucketId": "gemini-5h", "window": "5h", "remainingFraction": 0.9825, "resetTime": "2026-11-19T19:35:19Z"},
			{"bucketId": "gemini-weekly", "window": "weekly", "remainingFraction": 0.995, "resetTime": "2026-11-26T14:35:19Z"},
			{"bucketId": "gemini-argon-wtu-weekly", "window": "weekly", "remainingFraction": 1}
		]},
		{"displayName": "Claude and GPT", "buckets": [
			{"bucketId": "3p-5h", "window": "5h", "remainingFraction": 1, "resetTime": "2026-11-19T20:38:51Z"},
			{"bucketId": "3p-weekly", "window": "weekly", "remainingFraction": 1.4},
			{"bucketId": "3p-monthly", "window": "monthly", "remainingFraction": 0.5, "disabled": true},
			{"bucketId": "3p-unknown", "window": "5h"}
		]}
	]}`)

	windows, errParse := parseAntigravityQuota(body)
	if errParse != nil {
		t.Fatal(errParse)
	}
	want := []struct {
		id      string
		used    float64
		seconds int64
		reset   bool
	}{
		{"gemini-5h", 1.75, fiveHourSeconds, true},
		{"gemini-weekly", 0.5, weekSeconds, true},
		{"gemini-argon-wtu-weekly", 0, weekSeconds, false},
		{"3p-5h", 0, fiveHourSeconds, true},
		{"3p-weekly", 0, weekSeconds, false},
	}
	if len(windows) != len(want) {
		t.Fatalf("windows = %+v", windows)
	}
	for i, w := range want {
		got := windows[i]
		if got.ID != w.id || math.Abs(got.UsedPercent-w.used) > 1e-9 || got.WindowSeconds != w.seconds ||
			got.Scope != ScopeModel || (got.ResetsAt != nil) != w.reset {
			t.Fatalf("window %d = %+v, want %+v", i, got, w)
		}
	}
}

func TestParseAntigravityQuotaFallsBackToTopLevelBuckets(t *testing.T) {
	windows, errParse := parseAntigravityQuota([]byte(`{"buckets": [{"bucketId": "gemini-5h", "window": "5h", "remainingFraction": 0.25}]}`))
	if errParse != nil {
		t.Fatal(errParse)
	}
	if len(windows) != 1 || windows[0].ID != "gemini-5h" || windows[0].UsedPercent != 75 {
		t.Fatalf("windows = %+v", windows)
	}
}

func TestParseAntigravityQuotaRejectsEmptySummary(t *testing.T) {
	if _, errParse := parseAntigravityQuota([]byte(`{"groups": []}`)); errParse == nil {
		t.Fatal("want an error for a summary without buckets")
	}
}
