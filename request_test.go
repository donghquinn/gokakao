package kakao

import (
	"bytes"
	"context"
	"errors"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/donghquinn/gokakao/internal/kakaotest"
)

func TestClient_do_auth(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		auth  authKind
		token string
		want  string
	}{
		{"none", authNone, "", ""},
		{"REST key", authRESTKey, "", "KakaoAK rest"},
		{"admin key", authAdminKey, "", "KakaoAK admin"},
		{"bearer", authBearer, "tok", "Bearer tok"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c, srv := newTestClient(t, WithRESTKey("rest"), WithAdminKey("admin"), WithUserAgent("test"))
			srv.HandleJSON("GET /x", 200, `{}`)
			if err := c.do(t.Context(), &request{method: http.MethodGet, host: HostAPI, path: "/x", auth: tt.auth, token: tt.token}, nil); err != nil {
				t.Fatal(err)
			}
			req := srv.LastRequest()
			if got := req.Header.Get("Authorization"); got != tt.want {
				t.Errorf("Authorization = %q, want %q", got, tt.want)
			}
			if got := req.Header.Get("User-Agent"); got != "test gokakao/"+Version {
				t.Errorf("User-Agent = %q", got)
			}
		})
	}
}

func TestClient_do_missingCredential(t *testing.T) {
	t.Parallel()
	c, _ := newTestClient(t)
	for _, auth := range []authKind{authRESTKey, authAdminKey, authBearer} {
		err := c.do(t.Context(), &request{method: http.MethodGet, host: HostAPI, path: "/x", auth: auth}, nil)
		if !errors.Is(err, ErrMissingCredential) {
			t.Errorf("auth %d: err = %v, want ErrMissingCredential", auth, err)
		}
	}
	if err := c.do(t.Context(), &request{method: http.MethodGet, host: HostAPI, path: "/x", auth: authKind(99)}, nil); err == nil {
		t.Error("unknown auth kind: want error")
	}
}

func TestClient_do_bodies(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /form", 200, `{}`)
	srv.HandleJSON("POST /json", 200, `{}`)
	srv.HandleJSON("GET /query", 200, `{}`)

	form := url.Values{"template_object": {`{"object_type":"text"}`}}
	if err := c.do(t.Context(), &request{method: http.MethodPost, host: HostAPI, path: "/form", form: form}, nil); err != nil {
		t.Fatal(err)
	}
	req := srv.LastRequest()
	if ct := req.Header.Get("Content-Type"); ct != contentTypeForm {
		t.Errorf("form Content-Type = %q", ct)
	}
	if got := req.Form().Get("template_object"); got != `{"object_type":"text"}` {
		t.Errorf("form value = %q", got)
	}

	body := map[string]any{"channel_public_id": "_ZeUTxl"}
	if err := c.do(t.Context(), &request{method: http.MethodPost, host: HostAPI, path: "/json", json: body}, nil); err != nil {
		t.Fatal(err)
	}
	req = srv.LastRequest()
	if ct := req.Header.Get("Content-Type"); ct != contentTypeJSON {
		t.Errorf("json Content-Type = %q", ct)
	}
	if got := string(req.Body); got != `{"channel_public_id":"_ZeUTxl"}` {
		t.Errorf("json body = %s", got)
	}

	q := url.Values{"query": {"판교 카페"}, "size": {"15"}}
	if err := c.do(t.Context(), &request{method: http.MethodGet, host: HostAPI, path: "/query", query: q}, nil); err != nil {
		t.Fatal(err)
	}
	req = srv.LastRequest()
	if req.Query.Get("query") != "판교 카페" || req.Query.Get("size") != "15" {
		t.Errorf("query = %v", req.Query)
	}
	if len(req.Body) != 0 || req.Header.Get("Content-Type") != "" {
		t.Error("GET without body sent a body or Content-Type")
	}

	err := c.do(t.Context(), &request{method: http.MethodPost, host: HostAPI, path: "/x", form: form, json: body}, nil)
	if err == nil || !strings.Contains(err.Error(), "more than one body") {
		t.Errorf("both bodies: err = %v", err)
	}
	err = c.do(t.Context(), &request{method: http.MethodPost, host: HostAPI, path: "/x", json: make(chan int)}, nil)
	if err == nil || !strings.Contains(err.Error(), "encode") {
		t.Errorf("unencodable JSON: err = %v", err)
	}
}

func TestClient_do_decode(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /ok", 200, `{"id":123,"expires_in":43199}`)
	srv.HandleJSON("POST /empty", 200, ``)
	srv.HandleJSON("GET /bad", 200, `{"id":`)
	srv.Handle("GET /img", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte{0x89, 'P', 'N', 'G'})
	})

	var info struct {
		ID        int64 `json:"id"`
		ExpiresIn int   `json:"expires_in"`
	}
	if err := c.do(t.Context(), &request{method: http.MethodGet, host: HostAPI, path: "/ok"}, &info); err != nil {
		t.Fatal(err)
	}
	if info.ID != 123 || info.ExpiresIn != 43199 {
		t.Errorf("decoded %+v", info)
	}

	if err := c.do(t.Context(), &request{method: http.MethodPost, host: HostAPI, path: "/empty"}, &info); err != nil {
		t.Errorf("empty body: %v", err)
	}

	err := c.do(t.Context(), &request{method: http.MethodGet, host: HostAPI, path: "/bad"}, &info)
	if err == nil || !strings.Contains(err.Error(), "decode GET /bad") {
		t.Errorf("bad JSON: err = %v", err)
	}

	var raw rawBody
	if err := c.do(t.Context(), &request{method: http.MethodGet, host: HostDAPI, path: "/img"}, &raw); err != nil {
		t.Fatal(err)
	}
	if raw.ContentType != "image/png" || string(raw.Data) != "\x89PNG" {
		t.Errorf("raw = %+v", raw)
	}
}

func TestClient_do_errorResponse(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /v2/user/me", 401, `{"msg":"this access token does not exist","code":-401}`)
	err := c.do(t.Context(), &request{method: http.MethodGet, host: HostAPI, path: "/v2/user/me", auth: authBearer, token: "t"}, nil)
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("err = %v, want ErrInvalidToken", err)
	}
	var e *Error
	if !errors.As(err, &e) || e.HTTPStatus != 401 {
		t.Errorf("err = %#v", err)
	}
}

func TestClient_do_bodyLimit(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	c.maxBodySize = 8
	srv.HandleJSON("GET /big", 200, `{"a":"0123456789"}`)
	err := c.do(t.Context(), &request{method: http.MethodGet, host: HostAPI, path: "/big"}, nil)
	if err == nil || !strings.Contains(err.Error(), "exceeds 8 bytes") {
		t.Errorf("err = %v", err)
	}
}

// flaky registers a handler that fails with status and body for the first
// n calls and then succeeds. It returns the call counter.
func flaky(srv *kakaotest.Server, pattern string, n int32, status int, body string) *atomic.Int32 {
	var calls atomic.Int32
	srv.Handle(pattern, func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) <= n {
			w.WriteHeader(status)
			_, _ = w.Write([]byte(body))
			return
		}
		_, _ = w.Write([]byte(`{}`))
	})
	return &calls
}

func TestClient_do_retry(t *testing.T) {
	t.Parallel()
	fast := WithRetry(RetryPolicy{MaxAttempts: 3, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond})
	tests := []struct {
		name      string
		opts      []Option
		method    string
		idem      bool
		failures  int32
		status    int
		body      string
		wantCalls int32
		wantErr   bool
	}{
		{"GET recovers from 500", []Option{fast}, http.MethodGet, false, 2, 500, ``, 3, false},
		{"GET recovers from code -1", []Option{fast}, http.MethodGet, false, 1, 400, `{"code":-1,"msg":"internal"}`, 2, false},
		{"GET recovers from 429", []Option{fast}, http.MethodGet, false, 1, 429, `{"code":-10,"msg":"limit"}`, 2, false},
		{"GET gives up after MaxAttempts", []Option{fast}, http.MethodGet, false, 5, 503, ``, 3, true},
		{"no retry on -10 quota with 400", []Option{fast}, http.MethodGet, false, 1, 400, `{"code":-10,"msg":"limit"}`, 1, true},
		{"no retry on -401", []Option{fast}, http.MethodGet, false, 1, 401, `{"code":-401,"msg":"bad"}`, 1, true},
		{"no retry on POST", []Option{fast}, http.MethodPost, false, 1, 500, ``, 1, true},
		{"retry idempotent POST", []Option{fast}, http.MethodPost, true, 1, 500, ``, 2, false},
		{"no retry when disabled", nil, http.MethodGet, false, 1, 500, ``, 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c, srv := newTestClient(t, tt.opts...)
			calls := flaky(srv, tt.method+" /x", tt.failures, tt.status, tt.body)
			err := c.do(t.Context(), &request{method: tt.method, host: HostAPI, path: "/x", idempotent: tt.idem}, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got := calls.Load(); got != tt.wantCalls {
				t.Errorf("calls = %d, want %d", got, tt.wantCalls)
			}
		})
	}
}

func TestClient_do_retryStopsOnContextCancel(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithRetry(RetryPolicy{MaxAttempts: 5, BaseDelay: time.Hour, MaxDelay: time.Hour}))
	calls := flaky(srv, "GET /x", 10, 500, ``)
	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()
	start := time.Now()
	err := c.do(ctx, &request{method: http.MethodGet, host: HostAPI, path: "/x"}, nil)
	var e *Error
	if !errors.As(err, &e) || e.HTTPStatus != 500 {
		t.Errorf("err = %v, want the last *Error", err)
	}
	if calls.Load() != 1 {
		t.Errorf("calls = %d, want 1", calls.Load())
	}
	if time.Since(start) > 5*time.Second {
		t.Error("retry wait ignored context cancellation")
	}
}

func TestClient_do_transportError(t *testing.T) {
	t.Parallel()
	c, err := NewClient(WithBaseURL(HostAPI, "http://127.0.0.1:1"))
	if err != nil {
		t.Fatal(err)
	}
	if err := c.do(t.Context(), &request{method: http.MethodGet, host: HostAPI, path: "/x"}, nil); err == nil {
		t.Error("want transport error")
	}
	if err := c.do(t.Context(), &request{method: "BAD METHOD", host: HostAPI, path: "/x"}, nil); err == nil ||
		!strings.Contains(err.Error(), "build request") {
		t.Errorf("err = %v", err)
	}
}

func TestClient_Do(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v1/images", 200, `{"id":7}`)
	srv.HandleJSON("GET /openapi/v1/adAccounts/pages", 400,
		`{"code":-820,"msg":"KakaoKeywordAdException","extras":{"detailCode":103005,"detailMsg":"없는 비즈채널입니다."}}`)

	var out struct {
		ID int64 `json:"id"`
	}
	err := c.Do(t.Context(), &Request{
		Method: http.MethodPost, Host: HostKeywordAd, Path: "/openapi/v1/images",
		Header:      http.Header{"Adaccountid": {"111"}},
		BearerToken: "BT",
		Multipart: &Multipart{
			Fields: url.Values{"name": {"배너"}},
			Files:  []MultipartFile{{Field: "imageFile", Filename: "a.png", ContentType: "image/png", Data: []byte("\x89PNG")}, {Field: "doc", Filename: "b.bin", Data: []byte("x")}},
		},
	}, &out)
	if err != nil || out.ID != 7 {
		t.Fatalf("Do = %+v, %v", out, err)
	}
	req := srv.LastRequest()
	if req.Header.Get("Authorization") != "Bearer BT" || req.Header.Get("adAccountId") != "111" {
		t.Errorf("headers = %v", req.Header)
	}
	mt, params, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil || mt != "multipart/form-data" {
		t.Fatalf("Content-Type = %q", req.Header.Get("Content-Type"))
	}
	mr := multipart.NewReader(bytes.NewReader(req.Body), params["boundary"])
	form, err := mr.ReadForm(1 << 20)
	if err != nil {
		t.Fatal(err)
	}
	if form.Value["name"][0] != "배너" || form.File["imageFile"][0].Filename != "a.png" ||
		form.File["imageFile"][0].Header.Get("Content-Type") != "image/png" || form.File["doc"][0].Header.Get("Content-Type") != "application/octet-stream" {
		t.Errorf("form = %+v / %+v", form.Value, form.File)
	}

	err = c.Do(t.Context(), &Request{Method: http.MethodGet, Host: HostKeywordAd, Path: "/openapi/v1/adAccounts/pages"}, nil)
	var e *Error
	if !errors.As(err, &e) || e.Code != -820 || e.Extras == nil || e.Extras.DetailCode != 103005 {
		t.Fatalf("err = %#v", err)
	}
	if want := "kakao: HTTP 400, code -820: KakaoKeywordAdException (detail 103005: 없는 비즈채널입니다.)"; e.Error() != want {
		t.Errorf("Error() = %q", e.Error())
	}
	if srv.LastRequest().Header.Get("Authorization") != "" {
		t.Error("no bearer token: Authorization must be empty")
	}

	if err := c.Do(t.Context(), nil, nil); err == nil {
		t.Error("nil request: want error")
	}
	if err := c.Do(t.Context(), &Request{Host: Host(0)}, nil); err == nil {
		t.Error("unknown host: want error")
	}
	err = c.Do(t.Context(), &Request{Method: http.MethodPost, Host: HostKeywordAd, Path: "/x", JSON: 1, Multipart: &Multipart{}}, nil)
	if err == nil || !strings.Contains(err.Error(), "more than one body") {
		t.Errorf("two bodies: err = %v", err)
	}
}
