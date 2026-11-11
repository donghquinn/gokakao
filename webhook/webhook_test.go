package webhook

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	kakao "github.com/donghquinn/gokakao"
)

const adminKey = "primary-admin-key"

func newRequest(method, target, contentType, body string, auth bool) *http.Request {
	r := httptest.NewRequest(method, target, strings.NewReader(body))
	if contentType != "" {
		r.Header.Set("Content-Type", contentType)
	}
	if auth {
		r.Header.Set("Authorization", "KakaoAK "+adminKey)
	}
	r.Header.Set("X-Kakao-Resource-ID", "res-1")
	r.Header.Set("User-Agent", "KakaoOpenAPI/1.0")
	return r
}

func TestUnlinkHandler(t *testing.T) {
	t.Parallel()
	want := UnlinkEvent{AppID: 123, UserID: 4567, ReferrerType: "UNLINK_FROM_APPS"}
	tests := []struct {
		name       string
		req        *http.Request
		wantStatus int
		wantEvent  *UnlinkEvent
	}{
		{"GET", newRequest(http.MethodGet, "/unlink?app_id=123&user_id=4567&referrer_type=UNLINK_FROM_APPS", "", "", true), 200, &want},
		{"POST form", newRequest(http.MethodPost, "/unlink", "application/x-www-form-urlencoded", "app_id=123&user_id=4567&referrer_type=UNLINK_FROM_APPS", true), 200, &want},
		{
			"group app", newRequest(http.MethodGet, "/unlink?app_id=1&user_id=2&referrer_type=ACCOUNT_DELETE&group_user_token=g", "", "", true), 200,
			&UnlinkEvent{AppID: 1, UserID: 2, ReferrerType: "ACCOUNT_DELETE", GroupUserToken: "g"},
		},
		{"no auth", newRequest(http.MethodGet, "/unlink?app_id=1&user_id=2", "", "", false), 401, nil},
		{"bad user_id", newRequest(http.MethodGet, "/unlink?app_id=1&user_id=x", "", "", true), 400, nil},
		{"bad app_id", newRequest(http.MethodGet, "/unlink?app_id=x&user_id=1", "", "", true), 400, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got *UnlinkEvent
			h := UnlinkHandler(adminKey, func(_ *http.Request, e *UnlinkEvent) error { got = e; return nil })
			w := httptest.NewRecorder()
			h.ServeHTTP(w, tt.req)
			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if (got == nil) != (tt.wantEvent == nil) || (got != nil && *got != *tt.wantEvent) {
				t.Errorf("event = %+v, want %+v", got, tt.wantEvent)
			}
		})
	}
}

func TestAuthorize(t *testing.T) {
	t.Parallel()
	r := newRequest(http.MethodGet, "/", "", "", true)
	if err := authorize(r, ""); err == nil || errors.Is(err, ErrUnauthorized) {
		t.Errorf("empty key: err = %v, want a configuration error", err)
	}
	r.Header.Set("Authorization", "KakaoAK wrong")
	if err := authorize(r, adminKey); !errors.Is(err, ErrUnauthorized) {
		t.Errorf("wrong key: err = %v", err)
	}
	r.Header.Set("Authorization", "Bearer "+adminKey)
	if err := authorize(r, adminKey); !errors.Is(err, ErrUnauthorized) {
		t.Errorf("wrong scheme: err = %v", err)
	}
}

func TestChannelHandler(t *testing.T) {
	t.Parallel()
	body := `{"event":"added","id":"1111","id_type":"app_user_id","channel_public_id":"_FLX","channel_uuid":"@ad","updated_at":"2019-11-21T10:14:07Z"}`
	var got *ChannelEvent
	h := ChannelHandler(adminKey, func(_ *http.Request, e *ChannelEvent) error { got = e; return nil })

	w := httptest.NewRecorder()
	h.ServeHTTP(w, newRequest(http.MethodPost, "/channel", "application/json", body, true))
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
	want := ChannelEvent{
		Event: "added", ID: "1111", IDType: "app_user_id", ChannelPublicID: "_FLX", ChannelUUID: "@ad",
		UpdatedAt: kakao.Time{Time: time.Date(2019, 11, 21, 10, 14, 7, 0, time.UTC)}, ResourceID: "res-1",
	}
	if got == nil || *got != want {
		t.Errorf("event = %+v", got)
	}

	for name, tc := range map[string]struct {
		req  *http.Request
		code int
	}{
		"no auth":   {newRequest(http.MethodPost, "/channel", "application/json", body, false), 401},
		"bad JSON":  {newRequest(http.MethodPost, "/channel", "application/json", "{", true), 400},
		"too large": {newRequest(http.MethodPost, "/channel", "application/json", strings.Repeat("x", maxBodySize+1), true), 400},
	} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, tc.req)
		if w.Code != tc.code {
			t.Errorf("%s: status = %d, want %d", name, w.Code, tc.code)
		}
	}
}

func TestShareHandler(t *testing.T) {
	t.Parallel()
	var got *ShareEvent
	h := ShareHandler(adminKey, func(_ *http.Request, e *ShareEvent) error { got = e; return nil })

	w := httptest.NewRecorder()
	h.ServeHTTP(w, newRequest(http.MethodGet, "/share?CHAT_TYPE=MemoChat&HASH_CHAT_ID=h1&TEMPLATE_ID=10000&IS_SINGLE_CHATROOM=true&custom_parameter_key=v", "", "", true))
	if w.Code != 200 || got == nil {
		t.Fatalf("status = %d, event = %v", w.Code, got)
	}
	if got.ChatType != "MemoChat" || got.HashChatID != "h1" || got.TemplateID != 10000 || got.IsSingleChatroom == nil || !*got.IsSingleChatroom ||
		got.ResourceID != "res-1" || len(got.Args) != 1 || got.Args["custom_parameter_key"] != "v" {
		t.Errorf("GET event = %+v", got)
	}

	w = httptest.NewRecorder()
	body := `{"CHAT_TYPE":"MultiChat","HASH_CHAT_ID":"h2","TEMPLATE_ID":10000,"IS_SINGLE_CHATROOM":false,"CHAT_MEMBER_COUNT_RANGE":"5+","user_id":"abcd","product_id":1234}`
	h.ServeHTTP(w, newRequest(http.MethodPost, "/share", "application/json", body, true))
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
	if got.ChatType != "MultiChat" || got.TemplateID != 10000 || *got.IsSingleChatroom || got.ChatMemberCountRange != "5+" ||
		got.Args["user_id"] != "abcd" || got.Args["product_id"] != float64(1234) || len(got.Args) != 2 {
		t.Errorf("POST event = %+v", got)
	}

	w = httptest.NewRecorder()
	h.ServeHTTP(w, newRequest(http.MethodPost, "/share", "application/json", `{"CHAT_TYPE":"MemoChat","HASH_CHAT_ID":"h"}`, true))
	if w.Code != 200 || got.TemplateID != 0 || got.IsSingleChatroom != nil {
		t.Errorf("optional fields: %+v", got)
	}

	for name, tc := range map[string]struct {
		req  *http.Request
		code int
	}{
		"no auth":         {newRequest(http.MethodGet, "/share?CHAT_TYPE=MemoChat", "", "", false), 401},
		"bad template ID": {newRequest(http.MethodGet, "/share?TEMPLATE_ID=x", "", "", true), 400},
		"bad single flag": {newRequest(http.MethodGet, "/share?IS_SINGLE_CHATROOM=maybe", "", "", true), 400},
		"bad JSON":        {newRequest(http.MethodPost, "/share", "application/json", "[", true), 400},
		"too large":       {newRequest(http.MethodPost, "/share", "application/json", strings.Repeat("x", maxBodySize+1), true), 400},
	} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, tc.req)
		if w.Code != tc.code {
			t.Errorf("%s: status = %d, want %d", name, w.Code, tc.code)
		}
	}
}

func TestPushFailureHandler(t *testing.T) {
	t.Parallel()
	var got *PushFailure
	h := PushFailureHandler(func(_ *http.Request, f *PushFailure) error { got = f; return nil })

	w := httptest.NewRecorder()
	h.ServeHTTP(w, newRequest(http.MethodPost, "/push_fail", "application/x-www-form-urlencoded", "userId=123&pushToken=abcdefg&date=1396582208000&newPushToken=AAbbccddee", false))
	if w.Code != 200 || got == nil {
		t.Fatalf("status = %d, failure = %v", w.Code, got)
	}
	want := PushFailure{UserID: 123, PushToken: "abcdefg", Date: time.UnixMilli(1396582208000), NewPushToken: "AAbbccddee"}
	if *got != want {
		t.Errorf("failure = %+v", got)
	}

	w = httptest.NewRecorder()
	h.ServeHTTP(w, newRequest(http.MethodPost, "/push_fail", "application/x-www-form-urlencoded", "userId=1&date=yesterday", false))
	if w.Code != 400 {
		t.Errorf("bad date: status = %d", w.Code)
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, newRequest(http.MethodPost, "/push_fail", "application/x-www-form-urlencoded", "userId=abc", false))
	if w.Code != 400 {
		t.Errorf("bad userId: status = %d", w.Code)
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, newRequest(http.MethodPost, "/push_fail", "application/x-www-form-urlencoded", "%zz", false))
	if w.Code != 400 {
		t.Errorf("bad form: status = %d", w.Code)
	}
}

// fakeVerifier stands in for kakao.AuthService.
type fakeVerifier struct {
	event *kakao.SecurityEvent
	err   error
	token string
}

func (f *fakeVerifier) VerifySecurityEvent(_ context.Context, token string) (*kakao.SecurityEvent, error) {
	f.token = token
	return f.event, f.err
}

func TestSecurityEventHandler(t *testing.T) {
	t.Parallel()
	ev := &kakao.SecurityEvent{JTI: "j1"}
	tests := []struct {
		name       string
		verifier   *fakeVerifier
		req        *http.Request
		wantStatus int
		wantErr    string
		wantCalled bool
	}{
		{"valid", &fakeVerifier{event: ev}, newRequest(http.MethodPost, "/ssf", "application/secevent+jwt", "a.b.c\n", false), 202, "", true},
		{
			"invalid token", &fakeVerifier{err: &kakao.SETError{Code: "invalid_key", Description: "bad signature"}},
			newRequest(http.MethodPost, "/ssf", "application/secevent+jwt", "a.b.c", false), 400, "invalid_key", false,
		},
		{
			"keys unavailable", &fakeVerifier{err: errors.New("fetch failed")},
			newRequest(http.MethodPost, "/ssf", "application/secevent+jwt", "a.b.c", false), 503, "", false,
		},
		{
			"too large", &fakeVerifier{event: ev},
			newRequest(http.MethodPost, "/ssf", "application/secevent+jwt", strings.Repeat("x", maxBodySize+1), false), 400, "invalid_request", false,
		},
		{"GET", &fakeVerifier{event: ev}, newRequest(http.MethodGet, "/ssf", "", "", false), 405, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got *kakao.SecurityEvent
			h := SecurityEventHandler(tt.verifier, func(_ *http.Request, e *kakao.SecurityEvent) error { got = e; return nil })
			w := httptest.NewRecorder()
			h.ServeHTTP(w, tt.req)
			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if (got != nil) != tt.wantCalled {
				t.Errorf("callback called = %v, want %v", got != nil, tt.wantCalled)
			}
			if tt.wantCalled && tt.verifier.token != "a.b.c" {
				t.Errorf("token = %q, want trimmed body", tt.verifier.token)
			}
			if tt.wantErr != "" {
				var body struct{ Err, Description string }
				if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil || body.Err != tt.wantErr || w.Header().Get("Content-Type") != "application/json" {
					t.Errorf("error body = %s (%v), content type %q", w.Body, err, w.Header().Get("Content-Type"))
				}
			}
		})
	}
}

// TestSecurityEventHandler_endToEnd wires the handler to a real client
// whose keys come from a fake Kakao server.
func TestSecurityEventHandler_endToEnd(t *testing.T) {
	t.Parallel()
	kakaoSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"keys":[]}`))
	}))
	t.Cleanup(kakaoSrv.Close)
	c, err := kakao.NewClient(kakao.WithRESTKey("rest"), kakao.WithBaseURL(kakao.HostAuth, kakaoSrv.URL))
	if err != nil {
		t.Fatal(err)
	}
	h := SecurityEventHandler(c.Auth, func(*http.Request, *kakao.SecurityEvent) error { t.Error("callback must not run"); return nil })
	w := httptest.NewRecorder()
	h.ServeHTTP(w, newRequest(http.MethodPost, "/ssf", "application/secevent+jwt",
		"eyJhbGciOiJSUzI1NiIsImtpZCI6ImsxIiwidHlwIjoic2VjZXZlbnQrand0In0.e30.c2ln", false))
	if w.Code != 400 || !strings.Contains(w.Body.String(), "invalid_key") {
		t.Errorf("status = %d, body = %s", w.Code, w.Body)
	}
}

func TestParseUnlink_bodyTooLarge(t *testing.T) {
	t.Parallel()
	body := "user_id=1&x=" + url.QueryEscape(strings.Repeat("a", maxBodySize))
	_, err := ParseUnlink(newRequest(http.MethodPost, "/", "application/x-www-form-urlencoded", body, true), adminKey)
	if err == nil {
		t.Error("want error for oversized body")
	}
}

func TestHandlers_callbackErrors(t *testing.T) {
	t.Parallel()
	boom := errors.New("boom")
	var reported []error
	onErr := WithErrorHandler(func(_ *http.Request, err error) { reported = append(reported, err) })
	tests := []struct {
		name string
		h    http.Handler
		req  *http.Request
		want int
	}{
		{
			"unlink still replies 200",
			UnlinkHandler(adminKey, func(*http.Request, *UnlinkEvent) error { return boom }, onErr),
			newRequest(http.MethodGet, "/?app_id=1&user_id=2", "", "", true), 200,
		},
		{
			"channel replies 500",
			ChannelHandler(adminKey, func(*http.Request, *ChannelEvent) error { return boom }, onErr),
			newRequest(http.MethodPost, "/", "application/json", `{"event":"added"}`, true), 500,
		},
		{
			"share replies 500",
			ShareHandler(adminKey, func(*http.Request, *ShareEvent) error { return boom }, onErr),
			newRequest(http.MethodGet, "/?CHAT_TYPE=MemoChat", "", "", true), 500,
		},
		{
			"push failure replies 500",
			PushFailureHandler(func(*http.Request, *PushFailure) error { return boom }, onErr),
			newRequest(http.MethodPost, "/", "application/x-www-form-urlencoded", "userId=1", false), 500,
		},
		{
			"security event replies 503 so Kakao retries",
			SecurityEventHandler(&fakeVerifier{event: &kakao.SecurityEvent{}}, func(*http.Request, *kakao.SecurityEvent) error { return boom }, onErr),
			newRequest(http.MethodPost, "/", "application/secevent+jwt", "a.b.c", false), 503,
		},
	}
	for _, tt := range tests {
		w := httptest.NewRecorder()
		tt.h.ServeHTTP(w, tt.req)
		if w.Code != tt.want {
			t.Errorf("%s: status = %d, want %d", tt.name, w.Code, tt.want)
		}
	}
	if len(reported) != len(tests) {
		t.Fatalf("reported %d errors, want %d", len(reported), len(tests))
	}
	for _, err := range reported {
		if !errors.Is(err, boom) {
			t.Errorf("reported %v", err)
		}
	}

	reported = nil
	h := ChannelHandler(adminKey, func(*http.Request, *ChannelEvent) error { return nil }, onErr)
	h.ServeHTTP(httptest.NewRecorder(), newRequest(http.MethodPost, "/", "application/json", "{}", false))
	h2 := SecurityEventHandler(&fakeVerifier{err: &kakao.SETError{Code: "invalid_key"}}, func(*http.Request, *kakao.SecurityEvent) error { return nil }, onErr)
	h2.ServeHTTP(httptest.NewRecorder(), newRequest(http.MethodPost, "/", "application/secevent+jwt", "x", false))
	if len(reported) != 2 || !errors.Is(reported[0], ErrUnauthorized) {
		t.Errorf("parse errors reported = %v", reported)
	}
}
