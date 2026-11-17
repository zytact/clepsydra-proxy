package auth

import (
	"errors"
	"net"
	"sync"
	"time"
)

const networkRefreshRetry = 15 * time.Second

type offlineStreak struct {
	failures int
	lastAt   time.Time
	delay    time.Duration
}

type offlineRefreshLadder struct {
	mu      sync.Mutex
	streaks map[string]offlineStreak
}

var offlineRefreshes = &offlineRefreshLadder{streaks: map[string]offlineStreak{}}

func (l *offlineRefreshLadder) next(id string, now time.Time, err error) time.Time {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !isPreConnectError(err) {
		delete(l.streaks, id)
		return now.Add(refreshFailureBackoff)
	}
	streak := l.streaks[id]
	if streak.failures > 0 && now.Sub(streak.lastAt) > streak.delay+networkRefreshRetry {
		streak = offlineStreak{}
	}
	streak.failures++
	streak.delay = min(networkRefreshRetry<<min(streak.failures-1, 5), refreshFailureBackoff)
	streak.lastAt = now
	l.streaks[id] = streak
	return now.Add(streak.delay)
}

func isPreConnectError(err error) bool {
	for ; err != nil; err = errors.Unwrap(err) {
		if _, ok := err.(*net.DNSError); ok {
			return true
		}
		if opErr, ok := err.(*net.OpError); ok && (opErr.Op == "dial" || opErr.Op == "proxyconnect") {
			return true
		}
	}
	return false
}
