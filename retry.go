package kakao

import (
	"errors"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"
)

// RetryPolicy controls retries of idempotent requests. Pass it to
// [WithRetry]. Zero fields take the defaults noted below.
//
// A request is retried on network errors, HTTP 429 and 5xx responses, and
// Kakao codes -1 (internal error), -7 (service unavailable) and -9798
// (maintenance). Retries wait with exponential backoff and full jitter, or
// honor a Retry-After header when present, and stop when the request's
// context is done.
type RetryPolicy struct {
	// MaxAttempts is the total number of attempts, including the first.
	// Default 3.
	MaxAttempts int
	// BaseDelay is the backoff before the first retry. Default 200ms.
	BaseDelay time.Duration
	// MaxDelay caps a single wait. Default 5s.
	MaxDelay time.Duration
}

func (p *RetryPolicy) setDefaults() {
	if p.MaxAttempts <= 0 {
		p.MaxAttempts = 3
	}
	if p.BaseDelay <= 0 {
		p.BaseDelay = 200 * time.Millisecond
	}
	if p.MaxDelay <= 0 {
		p.MaxDelay = 5 * time.Second
	}
}

// backoff returns the wait before retry number attempt (1-based).
func (p *RetryPolicy) backoff(attempt int, header http.Header) time.Duration {
	if header != nil {
		if s, err := strconv.Atoi(header.Get("Retry-After")); err == nil && s >= 0 {
			return min(time.Duration(s)*time.Second, p.MaxDelay)
		}
	}
	d := p.BaseDelay << (attempt - 1)
	if d <= 0 || d > p.MaxDelay {
		d = p.MaxDelay
	}
	return rand.N(d) + 1 //nolint:gosec // jitter needs no cryptographic randomness
}

// retryable reports whether a failed attempt may be retried.
func retryable(err error) bool {
	var e *Error
	if errors.As(err, &e) {
		switch {
		case e.HTTPStatus == http.StatusTooManyRequests, e.HTTPStatus >= 500:
			return true
		case e.Code == int(ErrInternal), e.Code == int(ErrServiceUnavailable), e.Code == int(ErrMaintenance):
			return true
		}
		return false
	}
	var oe *OAuthError
	if errors.As(err, &oe) {
		return oe.HTTPStatus >= 500
	}
	// Transport error. Context errors are checked by the caller.
	return true
}
