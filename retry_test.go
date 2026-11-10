package kakao

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestRetryPolicy_backoff(t *testing.T) {
	t.Parallel()
	p := RetryPolicy{BaseDelay: 100 * time.Millisecond, MaxDelay: time.Second}
	p.setDefaults()

	for attempt := 1; attempt <= 10; attempt++ {
		ceiling := min(p.BaseDelay<<(attempt-1), p.MaxDelay)
		for range 50 {
			if d := p.backoff(attempt, nil); d <= 0 || d > ceiling {
				t.Fatalf("attempt %d: backoff %v outside (0, %v]", attempt, d, ceiling)
			}
		}
	}
	if d := p.backoff(100, nil); d <= 0 || d > p.MaxDelay {
		t.Errorf("overflowing shift: backoff %v", d)
	}

	header := http.Header{"Retry-After": {"0"}}
	if d := p.backoff(1, header); d != 0 {
		t.Errorf("Retry-After 0: backoff %v", d)
	}
	header.Set("Retry-After", "120")
	if d := p.backoff(1, header); d != p.MaxDelay {
		t.Errorf("Retry-After 120: backoff %v, want capped %v", d, p.MaxDelay)
	}
}

func TestRetryable(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"transport", errors.New("connection reset"), true},
		{"429", &Error{HTTPStatus: 429}, true},
		{"503", &Error{HTTPStatus: 503}, true},
		{"code -1", &Error{HTTPStatus: 400, Code: -1}, true},
		{"code -7", &Error{HTTPStatus: 400, Code: -7}, true},
		{"code -9798", &Error{HTTPStatus: 503, Code: -9798}, true},
		{"code -2", &Error{HTTPStatus: 400, Code: -2}, false},
		{"code -10 with 400", &Error{HTTPStatus: 400, Code: -10}, false},
		{"oauth 400", &OAuthError{HTTPStatus: 400, KOECode: "KOE237"}, false},
		{"oauth 500", &OAuthError{HTTPStatus: 500}, true},
		{"deadline", context.DeadlineExceeded, true}, // caller checks ctx.Err first
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := retryable(tt.err); got != tt.want {
				t.Errorf("retryable(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
