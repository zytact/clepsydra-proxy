package auth

import (
	"errors"
	"net"
	"time"
)

const networkRefreshRetry = 15 * time.Second

func expiredTokenRefreshRetry(now time.Time, err error, failures int) (int, time.Time) {
	if !isPreConnectError(err) {
		return 0, now.Add(refreshFailureBackoff)
	}
	failures++
	delay := min(networkRefreshRetry<<min(failures-1, 5), refreshFailureBackoff)
	return failures, now.Add(delay)
}

func isPreConnectError(err error) bool {
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return true
	}
	var opErr *net.OpError
	return errors.As(err, &opErr) && (opErr.Op == "dial" || opErr.Op == "proxyconnect")
}
