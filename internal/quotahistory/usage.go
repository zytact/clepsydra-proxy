package quotahistory

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/MoeForge/moego/json"
)

type Scope string

const (
	ScopeAll   Scope = "all"
	ScopeModel Scope = "model"

	fiveHourSeconds = int64(5 * 3600)
	weekSeconds     = int64(7 * 24 * 3600)
	minMonthSeconds = int64(28 * 24 * 3600)
	maxMonthSeconds = int64(31 * 24 * 3600)
)

type claudeWindow struct {
	Utilization      *float64 `json:"utilization"`
	ResetsAt         *string  `json:"resets_at"`
	LimitDollars     *float64 `json:"limit_dollars"`
	UsedDollars      *float64 `json:"used_dollars"`
	RemainingDollars *float64 `json:"remaining_dollars"`
}

func (w *claudeWindow) dollarDenominated() bool {
	return w.LimitDollars != nil || w.UsedDollars != nil || w.RemainingDollars != nil
}

type claudeLimit struct {
	Kind     string   `json:"kind"`
	Percent  *float64 `json:"percent"`
	ResetsAt *string  `json:"resets_at"`
	IsActive bool     `json:"is_active"`
	Scope    *struct {
		Model *struct {
			DisplayName string `json:"display_name"`
		} `json:"model"`
	} `json:"scope"`
}

type claudeUsage struct {
	FiveHour          *claudeWindow `json:"five_hour"`
	SevenDay          *claudeWindow `json:"seven_day"`
	SevenDayOAuthApps *claudeWindow `json:"seven_day_oauth_apps"`
	SevenDayOpus      *claudeWindow `json:"seven_day_opus"`
	SevenDaySonnet    *claudeWindow `json:"seven_day_sonnet"`
	SevenDayCowork    *claudeWindow `json:"seven_day_cowork"`
	SevenDayFable     *claudeWindow `json:"iguana_necktie"`
	Limits            []claudeLimit `json:"limits"`
}

func parseClaudeUsage(body []byte) ([]Window, error) {
	var usage claudeUsage
	if errUnmarshal := json.Unmarshal(body, &usage); errUnmarshal != nil {
		return nil, fmt.Errorf("decode claude usage: %w", errUnmarshal)
	}
	if usage.FiveHour == nil && usage.SevenDay == nil {
		return nil, fmt.Errorf("claude usage has no five_hour or seven_day window")
	}
	windows := []Window{}
	add := func(id string, window *claudeWindow, seconds int64, scope Scope) {
		if window == nil || window.Utilization == nil {
			return
		}
		windows = append(windows, Window{
			ID:            id,
			UsedPercent:   clampPercent(*window.Utilization),
			ResetsAt:      parseTime(window.ResetsAt),
			WindowSeconds: seconds,
			Scope:         scope,
		})
	}
	add("five-hour", usage.FiveHour, fiveHourSeconds, ScopeAll)
	add("seven-day", usage.SevenDay, weekSeconds, ScopeAll)
	add("seven-day-oauth-apps", usage.SevenDayOAuthApps, weekSeconds, ScopeModel)
	add("seven-day-opus", usage.SevenDayOpus, weekSeconds, ScopeModel)
	add("seven-day-sonnet", usage.SevenDaySonnet, weekSeconds, ScopeModel)
	add("seven-day-cowork", usage.SevenDayCowork, weekSeconds, ScopeModel)
	fable := fableLimit(usage.Limits)
	if fable == nil && usage.SevenDayFable != nil && !usage.SevenDayFable.dollarDenominated() {
		fable = usage.SevenDayFable
	}
	add("seven-day-fable", fable, weekSeconds, ScopeModel)
	return windows, nil
}

func fableLimit(limits []claudeLimit) *claudeWindow {
	var found *claudeWindow
	for _, limit := range limits {
		if limit.Kind != "weekly_scoped" || limit.Percent == nil || limit.Scope == nil || limit.Scope.Model == nil {
			continue
		}
		name := strings.ToLower(strings.TrimSpace(limit.Scope.Model.DisplayName))
		if name != "fable" && name != "fable 5" {
			continue
		}
		window := &claudeWindow{Utilization: limit.Percent, ResetsAt: limit.ResetsAt}
		if limit.IsActive {
			return window
		}
		if found == nil {
			found = window
		}
	}
	return found
}

type codexWindow struct {
	UsedPercent        *float64        `json:"used_percent"`
	ResetAt            json.RawMessage `json:"reset_at"`
	ResetAfterSeconds  *float64        `json:"reset_after_seconds"`
	LimitWindowSeconds *float64        `json:"limit_window_seconds"`
}

type codexRateLimit struct {
	PrimaryWindow   *codexWindow `json:"primary_window"`
	SecondaryWindow *codexWindow `json:"secondary_window"`
	LimitReached    bool         `json:"limit_reached"`
	Allowed         *bool        `json:"allowed"`
}

func (limit codexRateLimit) reached() bool {
	return limit.LimitReached || (limit.Allowed != nil && !*limit.Allowed)
}

func anyWindowReported(windows []*codexWindow) bool {
	for _, window := range windows {
		if window != nil && window.UsedPercent != nil {
			return true
		}
	}
	return false
}

type codexUsage struct {
	RateLimit *codexRateLimit `json:"rate_limit"`
}

func parseCodexUsage(body []byte, now time.Time) ([]Window, error) {
	var usage codexUsage
	if errUnmarshal := json.Unmarshal(body, &usage); errUnmarshal != nil {
		return nil, fmt.Errorf("decode codex usage: %w", errUnmarshal)
	}
	if usage.RateLimit == nil {
		return []Window{}, nil
	}
	candidates := []*codexWindow{usage.RateLimit.PrimaryWindow, usage.RateLimit.SecondaryWindow}
	unexplainedLimit := usage.RateLimit.reached() && !anyWindowReported(candidates)
	windows := []Window{}
	seen := map[string]bool{}
	for position, window := range candidates {
		if window == nil {
			continue
		}
		used := 0.0
		switch {
		case window.UsedPercent != nil:
			used = *window.UsedPercent
		case unexplainedLimit:
			used = 100
		default:
			continue
		}
		seconds := int64(0)
		if window.LimitWindowSeconds != nil {
			seconds = int64(*window.LimitWindowSeconds)
		}
		id := codexWindowID(seconds, position)
		if seen[id] {
			continue
		}
		seen[id] = true
		windows = append(windows, Window{
			ID:            id,
			UsedPercent:   clampPercent(used),
			ResetsAt:      codexReset(window, now),
			WindowSeconds: seconds,
			Scope:         ScopeAll,
		})
	}
	return windows, nil
}

func codexWindowID(seconds int64, position int) string {
	switch {
	case seconds == fiveHourSeconds:
		return "five-hour"
	case seconds == weekSeconds:
		return "weekly"
	case seconds >= minMonthSeconds && seconds <= maxMonthSeconds:
		return "monthly"
	case seconds == 0 && position == 0:
		return "five-hour"
	case seconds == 0:
		return "weekly"
	default:
		return "window-" + strconv.FormatInt(seconds, 10)
	}
}

func codexReset(window *codexWindow, now time.Time) *time.Time {
	raw := strings.Trim(strings.TrimSpace(string(window.ResetAt)), `"`)
	if raw != "" && raw != "null" {
		if epoch, errParse := strconv.ParseFloat(raw, 64); errParse == nil && epoch > 0 {
			at := time.Unix(int64(epoch), 0).UTC()
			return &at
		}
		if at := parseTime(&raw); at != nil {
			return at
		}
	}
	if window.ResetAfterSeconds != nil && *window.ResetAfterSeconds >= 0 {
		at := now.Add(time.Duration(*window.ResetAfterSeconds * float64(time.Second))).UTC()
		return &at
	}
	return nil
}

func parseTime(raw *string) *time.Time {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil
	}
	at, errParse := time.Parse(time.RFC3339Nano, strings.TrimSpace(*raw))
	if errParse != nil {
		return nil
	}
	at = at.UTC()
	return &at
}

func clampPercent(value float64) float64 {
	if math.IsNaN(value) {
		return 0
	}
	return math.Max(0, math.Min(100, value))
}
