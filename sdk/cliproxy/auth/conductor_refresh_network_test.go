package auth

import (
	"context"
	"errors"
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

func TestOfflineRefreshLadder_BacksOffWhileOfflineAndResetsOnOtherErrors(t *testing.T) {
	ladder := &offlineRefreshLadder{failures: map[string]int{}}
	offline := &net.OpError{Op: "proxyconnect", Net: "tcp", Err: fmt.Errorf("connection refused")}
	now := time.Now()
	var delays []time.Duration
	for range 7 {
		at := ladder.next("a", now, offline)
		delays = append(delays, at.Sub(now))
		now = at
	}
	want := []time.Duration{15 * time.Second, 30 * time.Second, time.Minute, 2 * time.Minute, 4 * time.Minute, refreshFailureBackoff, refreshFailureBackoff}
	for i := range want {
		if delays[i] != want[i] {
			t.Fatalf("delays = %v, want %v", delays, want)
		}
	}
	if at := ladder.next("a", now, fmt.Errorf("upstream 503 service unavailable")); !at.Equal(now.Add(refreshFailureBackoff)) {
		t.Fatalf("retry at = %s after a non-network error", at)
	}
	if at := ladder.next("a", now, offline); !at.Equal(now.Add(networkRefreshRetry)) {
		t.Fatalf("retry at = %s, want a fresh ladder after the streak ended", at)
	}
}

func TestOfflineRefreshLadder_KeepsClimbingThroughSlowFailuresAndStartsOverAfterSuccess(t *testing.T) {
	ladder := &offlineRefreshLadder{failures: map[string]int{}}
	offline := &net.DNSError{Err: "i/o timeout", Name: "platform.claude.com", IsTimeout: true}
	now := time.Now()
	ladder.next("a", now, offline)
	now = now.Add(2 * time.Minute)
	if at := ladder.next("a", now, offline); !at.Equal(now.Add(2 * networkRefreshRetry)) {
		t.Fatalf("retry at = %s, want the second rung after a slow failure", at)
	}
	ladder.forget("a")
	if at := ladder.next("a", now, offline); !at.Equal(now.Add(networkRefreshRetry)) {
		t.Fatalf("retry at = %s, want the first rung after a successful refresh", at)
	}
}

func TestIsPreConnectError_SeesADialInsideASocksConnect(t *testing.T) {
	refused := &net.OpError{Op: "dial", Net: "tcp", Err: fmt.Errorf("connection refused")}
	socks := &net.OpError{Op: "socks connect", Net: "tcp", Err: refused}
	if !isPreConnectError(fmt.Errorf("token refresh request failed: %w", socks)) {
		t.Fatal("a refused dial to a SOCKS proxy should count as offline")
	}
	handshake := &net.OpError{Op: "socks connect", Net: "tcp", Err: fmt.Errorf("unknown error general SOCKS server failure")}
	if isPreConnectError(handshake) {
		t.Fatal("a SOCKS handshake failure is not a pre-connect error")
	}
	joined := errors.Join(fmt.Errorf("outer: %w", &net.DNSError{Err: "no such host", IsNotFound: true}), fmt.Errorf("siwc: %w", refused))
	if !isPreConnectError(fmt.Errorf("refresh failed: %w", joined)) {
		t.Fatal("a joined pair of offline partition errors should count as offline")
	}
}
