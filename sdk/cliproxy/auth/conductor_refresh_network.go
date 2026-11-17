package auth

import (
	"net"
	"sync"
	"time"
)

const networkRefreshRetry = 15 * time.Second

type offlineRefreshLadder struct {
	mu       sync.Mutex
	failures map[string]int
}

var offlineRefreshes = &offlineRefreshLadder{failures: map[string]int{}}

func (l *offlineRefreshLadder) next(id string, now time.Time, err error) time.Time {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !isPreConnectError(err) {
		delete(l.failures, id)
		return now.Add(refreshFailureBackoff)
	}
	l.failures[id]++
	return now.Add(min(networkRefreshRetry<<min(l.failures[id]-1, 5), refreshFailureBackoff))
}

func (l *offlineRefreshLadder) settle(id string, err error) {
	if !isPreConnectError(err) {
		l.forget(id)
	}
}

func (l *offlineRefreshLadder) forget(id string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.failures, id)
}

func isPreConnectError(err error) bool {
	switch e := err.(type) {
	case nil:
		return false
	case *net.DNSError:
		return true
	case *net.OpError:
		if e.Op == "dial" || e.Op == "proxyconnect" {
			return true
		}
	}
	switch wrapped := err.(type) {
	case interface{ Unwrap() error }:
		return isPreConnectError(wrapped.Unwrap())
	case interface{ Unwrap() []error }:
		for _, inner := range wrapped.Unwrap() {
			if isPreConnectError(inner) {
				return true
			}
		}
	}
	return false
}
