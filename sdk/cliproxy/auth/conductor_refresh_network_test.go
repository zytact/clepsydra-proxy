package auth

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"
)

type offlineRefreshTestExecutor struct {
	schedulerProviderTestExecutor
}

func (e offlineRefreshTestExecutor) Refresh(ctx context.Context, auth *Auth) (*Auth, error) {
	dnsErr := &net.DNSError{Err: "no such host", Name: "platform.claude.com", IsNotFound: true}
	return nil, fmt.Errorf("token refresh failed after 3 attempts: token refresh request failed: %w",
		&net.OpError{Op: "dial", Net: "tcp", Err: dnsErr})
}

func TestManager_RefreshAuthOffline_ExpiredTokenRetriesSoon(t *testing.T) {
	ctx := context.Background()
	manager := NewManager(nil, &RoundRobinSelector{}, nil)
	manager.RegisterExecutor(offlineRefreshTestExecutor{
		schedulerProviderTestExecutor: schedulerProviderTestExecutor{provider: "claude"},
	})
	auth := &Auth{
		ID:       "offline-refresh-expired-token",
		Provider: "claude",
		Status:   StatusActive,
		Metadata: map[string]any{
			"email":         "offline@example.com",
			"access_token":  "expired-access-token",
			"refresh_token": "refresh-token",
			"expired":       time.Now().Add(-8 * time.Hour).Format(time.RFC3339),
		},
	}
	if _, errRegister := manager.Register(ctx, auth); errRegister != nil {
		t.Fatalf("register auth: %v", errRegister)
	}

	before := time.Now()
	manager.refreshAuth(ctx, auth.ID)

	updated, ok := manager.GetByID(auth.ID)
	if !ok {
		t.Fatalf("expected auth %q after refresh", auth.ID)
	}
	if updated.StatusMessage != "token expired" {
		t.Fatalf("StatusMessage = %q, want token expired", updated.StatusMessage)
	}
	if limit := before.Add(networkRefreshRetry + 5*time.Second); updated.NextRefreshAfter.After(limit) {
		t.Fatalf("NextRefreshAfter = %s, want by %s so the token recovers soon after the network returns", updated.NextRefreshAfter, limit)
	}
}

func TestExpiredTokenRefreshRetry_BacksOffWhileOfflineAndResetsOnOtherErrors(t *testing.T) {
	now := time.Now()
	offline := &net.OpError{Op: "proxyconnect", Net: "tcp", Err: fmt.Errorf("connection refused")}
	failures := 0
	var delays []time.Duration
	for range 7 {
		var at time.Time
		failures, at = expiredTokenRefreshRetry(now, offline, failures)
		delays = append(delays, at.Sub(now))
	}
	want := []time.Duration{15 * time.Second, 30 * time.Second, time.Minute, 2 * time.Minute, 4 * time.Minute, refreshFailureBackoff, refreshFailureBackoff}
	for i := range want {
		if delays[i] != want[i] {
			t.Fatalf("delays = %v, want %v", delays, want)
		}
	}
	failures, at := expiredTokenRefreshRetry(now, fmt.Errorf("upstream 503 service unavailable"), failures)
	if failures != 0 || !at.Equal(now.Add(refreshFailureBackoff)) {
		t.Fatalf("failures = %d, retry at = %s", failures, at)
	}
}
