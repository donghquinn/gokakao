package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	kakao "github.com/donghquinn/gokakao"
)

// SecurityEventVerifier verifies Security Event Tokens. [*kakao.AuthService]
// implements it.
type SecurityEventVerifier interface {
	VerifySecurityEvent(ctx context.Context, token string) (*kakao.SecurityEvent, error)
}

var _ SecurityEventVerifier = (*kakao.AuthService)(nil)

// SecurityEventHandler returns a handler for the account status webhook
// (계정 상태 변경 웹훅). It verifies the Security Event Token in the request
// body with v, typically client.Auth, and calls fn with the verified
// event. It replies:
//
//   - 202 Accepted after fn returns;
//   - 400 Bad Request with an RFC 8935 error body when the token is
//     invalid, which tells Kakao not to retry;
//   - 503 Service Unavailable when Kakao's signing keys cannot be
//     fetched, so Kakao retries the delivery.
//
// Kakao retries undelivered events up to 5 times over about 2 hours; use
// [kakao.SecurityEvent.JTI] to drop duplicates.
//
// Docs: https://developers.kakao.com/docs/ko/kakaologin/callback
func SecurityEventHandler(v SecurityEventVerifier, fn func(r *http.Request, e *kakao.SecurityEvent)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		b, err := readBody(r)
		if err != nil {
			writeSETError(w, &kakao.SETError{Code: "invalid_request", Description: err.Error()})
			return
		}
		e, err := v.VerifySecurityEvent(r.Context(), string(bytes.TrimSpace(b)))
		if err != nil {
			var se *kakao.SETError
			if errors.As(err, &se) {
				writeSETError(w, se)
				return
			}
			http.Error(w, "cannot verify security event", http.StatusServiceUnavailable)
			return
		}
		fn(r, e)
		w.WriteHeader(http.StatusAccepted)
	})
}

func writeSETError(w http.ResponseWriter, e *kakao.SETError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]string{"err": e.Code, "description": e.Description})
}
