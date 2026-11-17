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
	if next, ok := nextRefreshCheckAt(time.Now(), updated, time.Second); !ok || next.After(before.Add(networkRefreshRetry+5*time.Second)) {
		t.Fatalf("next refresh check = %s, %v", next, ok)
	}
}

func TestExpiredTokenRefreshRetryAt_OtherErrorsKeepBackoff(t *testing.T) {
	now := time.Now()
	if got := expiredTokenRefreshRetryAt(now, fmt.Errorf("upstream 503 service unavailable")); !got.Equal(now.Add(refreshFailureBackoff)) {
		t.Fatalf("retry at = %s, want %s", got, now.Add(refreshFailureBackoff))
	}
}
