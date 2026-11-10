package kakao

import (
	"errors"
	"strings"
	"testing"
)

var setHeader = map[string]string{"alg": "RS256", "kid": "k1", "typ": "secevent+jwt"}

func (f *oidcFixture) setClaims() map[string]any {
	return map[string]any{
		"iss": Issuer, "aud": "rest", "sub": "1234", "app_id": "5678", "jti": "jti-1",
		"txn": "txn-1", "toe": f.now.Unix() - 5, "iat": f.now.Unix(),
		"events": map[string]any{
			EventUserUnlinked: map[string]any{
				"subject": map[string]any{"subject_type": "iss-sub", "iss": Issuer, "sub": "1234"},
				"reason":  "UNLINK_FROM_APPS",
			},
		},
	}
}

func TestAuthService_VerifySecurityEvent(t *testing.T) {
	t.Parallel()
	f := newOIDCFixture(t)

	ev, err := f.c.Auth.VerifySecurityEvent(t.Context(), signJWT(t, testKeys()[0], setHeader, f.setClaims()))
	if err != nil {
		t.Fatal(err)
	}
	if ev.Subject != "1234" || ev.AppID != "5678" || ev.JTI != "jti-1" || ev.TransactionID != "txn-1" || ev.TimeOfEvent != f.now.Unix()-5 {
		t.Errorf("event = %+v", ev)
	}
	d, ok := ev.Events[EventUserUnlinked]
	if !ok || d.Reason != "UNLINK_FROM_APPS" || d.Subject == nil || d.Subject.Sub != "1234" || d.Subject.SubjectType != "iss-sub" {
		t.Errorf("detail = %+v", d)
	}

	claims := f.setClaims()
	delete(claims, "txn")
	claims["txm"] = "txm-1"
	claims["events"] = map[string]any{
		EventIdentifierChanged: map[string]any{
			"subject":   map[string]any{"subject_type": "account_email", "account_email": "old@kakao.com"},
			"new_value": "new@kakao.com",
		},
		EventBusinessTokenIssued: map[string]any{
			"subject":       map[string]any{"subject_type": "oauth_token", "token_type": "business_access_token", "token_identifier_alg": "hash_sha256", "token": "abc"},
			"token_subject": map[string]any{"subject_type": "iss-sub", "iss": "https://kauth.kakao.com/", "sub": "99"},
			"token_id":      "tid", "token_class": "business",
		},
	}
	ev, err = f.c.Auth.VerifySecurityEvent(t.Context(), signJWT(t, testKeys()[0], setHeader, claims))
	if err != nil {
		t.Fatal(err)
	}
	if ev.TransactionID != "txm-1" {
		t.Errorf("txm fallback: %q", ev.TransactionID)
	}
	if d := ev.Events[EventIdentifierChanged]; d.NewValue != "new@kakao.com" || d.Subject.AccountEmail != "old@kakao.com" {
		t.Errorf("identifier changed = %+v", d)
	}
	if d := ev.Events[EventBusinessTokenIssued]; d.TokenClass != "business" || d.TokenSubject.Sub != "99" || d.Subject.Token != "abc" {
		t.Errorf("business token issued = %+v", d)
	}
}

func TestAuthService_VerifySecurityEvent_invalid(t *testing.T) {
	t.Parallel()
	f := newOIDCFixture(t)
	key := testKeys()[0]
	with := func(k string, v any) string {
		c := f.setClaims()
		c[k] = v
		return signJWT(t, key, setHeader, c)
	}
	valid := signJWT(t, key, setHeader, f.setClaims())
	parts := strings.Split(valid, ".")

	tests := []struct {
		name     string
		token    string
		wantCode string
	}{
		{"malformed", "abc", "invalid_request"},
		{"id token typ", signJWT(t, key, map[string]string{"alg": "RS256", "kid": "k1", "typ": "JWT"}, f.setClaims()), "invalid_request"},
		{"bad payload encoding", parts[0] + ".!!!." + parts[2], "invalid_request"},
		{"payload not JSON", signJWT(t, key, setHeader, "x"), "invalid_request"},
		{"other key", signJWT(t, testKeys()[1], setHeader, f.setClaims()), "invalid_key"},
		{"unknown kid", signJWT(t, key, map[string]string{"alg": "RS256", "kid": "nope", "typ": "secevent+jwt"}, f.setClaims()), "invalid_key"},
		{"wrong issuer", with("iss", "https://evil.example"), "invalid_issuer"},
		{"wrong audience", with("aud", "other"), "invalid_audience"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := f.c.Auth.VerifySecurityEvent(t.Context(), tt.token)
			var se *SETError
			if !errors.As(err, &se) || se.Code != tt.wantCode {
				t.Fatalf("err = %v, want code %s", err, tt.wantCode)
			}
			if !strings.Contains(se.Error(), tt.wantCode) {
				t.Errorf("Error() = %q", se.Error())
			}
		})
	}
}

func TestAuthService_VerifySecurityEvent_keysUnavailable(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithRESTKey("rest"))
	srv.HandleJSON("GET /.well-known/jwks.json", 503, ``)
	_, err := c.Auth.VerifySecurityEvent(t.Context(), signJWT(t, testKeys()[0], setHeader, map[string]any{}))
	var se *SETError
	if err == nil || errors.As(err, &se) {
		t.Errorf("err = %v, want a non-SETError so the delivery is retried", err)
	}
}
