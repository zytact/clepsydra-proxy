package auth

import (
	"errors"
	"net"
	"time"
)

const networkRefreshRetry = 15 * time.Second

func expiredTokenRefreshRetryAt(now time.Time, err error) time.Time {
	if isNetworkUnreachableError(err) {
		return now.Add(networkRefreshRetry)
	}
	return now.Add(refreshFailureBackoff)
}

func isNetworkUnreachableError(err error) bool {
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return true
	}
	var opErr *net.OpError
	return errors.As(err, &opErr) && opErr.Op == "dial"
}
