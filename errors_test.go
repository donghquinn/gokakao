package kakao

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestError_Error(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{"kapi", &Error{HTTPStatus: 401, Code: -401, Msg: "this access token does not exist"}, "kakao: HTTP 401, code -401: this access token does not exist"},
		{"dapi", &Error{HTTPStatus: 400, Type: "MissingParameter", Msg: "query parameter required"}, "kakao: HTTP 400, MissingParameter: query parameter required"},
		{"bare", &Error{HTTPStatus: 502}, "kakao: HTTP 502"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestError_Is(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		err    error
		target error
		want   bool
	}{
		{"matching code", &Error{HTTPStatus: 401, Code: -401}, ErrInvalidToken, true},
		{"other code", &Error{HTTPStatus: 401, Code: -401}, ErrConsentRequired, false},
		{"rate limit by code", &Error{HTTPStatus: 400, Code: -10}, ErrRateLimited, true},
		{"rate limit by status", &Error{HTTPStatus: 429}, ErrRateLimited, true},
		{"zero code never matches", &Error{HTTPStatus: 400}, codeError(0), false},
		{"non-sentinel target", &Error{Code: -401}, errors.New("x"), false},
		{"oauth KOE237", &OAuthError{HTTPStatus: 400, KOECode: "KOE237"}, ErrRateLimited, true},
		{"oauth other", &OAuthError{HTTPStatus: 400, KOECode: "KOE320"}, ErrRateLimited, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := errors.Is(tt.err, tt.target); got != tt.want {
				t.Errorf("errors.Is(%v, %v) = %v, want %v", tt.err, tt.target, got, tt.want)
			}
		})
	}
}

func TestOAuthError_Error(t *testing.T) {
	t.Parallel()
	e := &OAuthError{HTTPStatus: 400, Err: "invalid_grant", Description: "authorization code not found", KOECode: "KOE320"}
	want := "kakao: oauth HTTP 400: invalid_grant (KOE320): authorization code not found"
	if got := e.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestCodeError_Error(t *testing.T) {
	t.Parallel()
	if got, want := ErrInvalidToken.Error(), "kakao: code -401: invalid token"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := codeError(-12345).Error(), "kakao: code -12345"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseError(t *testing.T) {
	t.Parallel()
	t.Run("kapi with scopes", func(t *testing.T) {
		t.Parallel()
		body := `{"msg":"[talk_message] scope required","code":-402,"api_type":"TALK_MEMO_SEND","required_scopes":["talk_message"],"allowed_scopes":["profile"]}`
		var e *Error
		if !errors.As(parseError(HostAPI, 403, []byte(body)), &e) {
			t.Fatal("not *Error")
		}
		if e.HTTPStatus != 403 || e.Code != -402 || e.APIType != "TALK_MEMO_SEND" ||
			!slices.Equal(e.RequiredScopes, []string{"talk_message"}) || !slices.Equal(e.AllowedScopes, []string{"profile"}) {
			t.Errorf("unexpected %+v", e)
		}
		if !errors.Is(e, ErrConsentRequired) {
			t.Error("want ErrConsentRequired")
		}
	})
	t.Run("dapi", func(t *testing.T) {
		t.Parallel()
		body := `{"errorType":"MissingParameter","message":"query parameter required"}`
		var e *Error
		if !errors.As(parseError(HostDAPI, 400, []byte(body)), &e) {
			t.Fatal("not *Error")
		}
		if e.Type != "MissingParameter" || e.Msg != "query parameter required" {
			t.Errorf("unexpected %+v", e)
		}
	})
	t.Run("kauth", func(t *testing.T) {
		t.Parallel()
		body := `{"error":"invalid_grant","error_description":"authorization code not found for code=x","error_code":"KOE320"}`
		var e *OAuthError
		if !errors.As(parseError(HostAuth, 400, []byte(body)), &e) {
			t.Fatal("not *OAuthError")
		}
		if e.Err != "invalid_grant" || e.KOECode != "KOE320" || e.HTTPStatus != 400 {
			t.Errorf("unexpected %+v", e)
		}
	})
	t.Run("kauth with kapi-style body", func(t *testing.T) {
		t.Parallel()
		var e *Error
		if !errors.As(parseError(HostAuth, 401, []byte(`{"msg":"bad","code":-401}`)), &e) || e.Code != -401 {
			t.Errorf("want *Error with code -401, got %v", e)
		}
	})
	t.Run("non-JSON", func(t *testing.T) {
		t.Parallel()
		var e *Error
		if !errors.As(parseError(HostAPI, 502, []byte("<html>Bad Gateway</html>")), &e) {
			t.Fatal("not *Error")
		}
		if e.HTTPStatus != 502 || e.Msg != "<html>Bad Gateway</html>" {
			t.Errorf("unexpected %+v", e)
		}
	})
	t.Run("long body truncated", func(t *testing.T) {
		t.Parallel()
		var e *Error
		errors.As(parseError(HostAPI, 500, []byte(strings.Repeat("x", 1000))), &e)
		if len(e.Msg) != 203 {
			t.Errorf("len(Msg) = %d, want 203", len(e.Msg))
		}
	})
	t.Run("unknown JSON", func(t *testing.T) {
		t.Parallel()
		var e *Error
		errors.As(parseError(HostAPI, 500, []byte(`{"foo":1}`)), &e)
		if e.Msg != `{"foo":1}` {
			t.Errorf("Msg = %q", e.Msg)
		}
	})
}

func FuzzParseError(f *testing.F) {
	f.Add(`{"msg":"x","code":-401}`, 0)
	f.Add(`{"errorType":"X","message":"y"}`, 2)
	f.Add(`{"error":"invalid_grant","error_code":"KOE320"}`, 1)
	f.Add(`not json`, 1)
	f.Fuzz(func(t *testing.T, body string, h int) {
		host := Host(h%int(numHosts-1)) + 1
		if host <= hostUnknown {
			host = HostAPI
		}
		if err := parseError(host, 400, []byte(body)); err == nil || err.Error() == "" {
			t.Fatal("parseError returned empty error")
		}
	})
}
