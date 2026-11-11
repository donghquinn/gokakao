package kakao

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestAuthService_AuthCodeURL(t *testing.T) {
	t.Parallel()
	c, err := NewClient(WithRESTKey("rest"))
	if err != nil {
		t.Fatal(err)
	}

	raw, err := c.Auth.AuthCodeURL("https://example.com/cb", &AuthCodeOptions{
		Scopes:        []string{"openid", "account_email"},
		Prompts:       []Prompt{PromptLogin, PromptSelectAccount},
		LoginHint:     "ryan@kakao.com",
		ServiceTerms:  []string{"tag1", "tag2"},
		State:         "xyz",
		Nonce:         "n-0S6",
		CodeChallenge: "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
	})
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if u.Scheme+"://"+u.Host+u.Path != "https://kauth.kakao.com/oauth/authorize" {
		t.Errorf("URL = %s", raw)
	}
	for k, want := range map[string]string{
		"client_id": "rest", "redirect_uri": "https://example.com/cb", "response_type": "code",
		"scope": "openid,account_email", "prompt": "login,select_account", "login_hint": "ryan@kakao.com",
		"service_terms": "tag1,tag2", "state": "xyz", "nonce": "n-0S6",
		"code_challenge": "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM", "code_challenge_method": "S256",
	} {
		if got := u.Query().Get(k); got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}

	raw, err = c.Auth.AuthCodeURL("https://example.com/cb", nil)
	if err != nil {
		t.Fatal(err)
	}
	if u, _ := url.Parse(raw); len(u.Query()) != 3 {
		t.Errorf("nil options: %s", raw)
	}

	noKey, _ := NewClient()
	if _, err := noKey.Auth.AuthCodeURL("x", nil); !errors.Is(err, ErrMissingCredential) {
		t.Errorf("no key: err = %v", err)
	}
}

func TestAuthService_LogoutURL(t *testing.T) {
	t.Parallel()
	c, _ := NewClient(WithRESTKey("rest"))
	raw, err := c.Auth.LogoutURL("https://example.com/bye", "s1")
	if err != nil {
		t.Fatal(err)
	}
	want := "https://kauth.kakao.com/oauth/logout?client_id=rest&logout_redirect_uri=https%3A%2F%2Fexample.com%2Fbye&state=s1"
	if raw != want {
		t.Errorf("got  %s\nwant %s", raw, want)
	}
	if raw, _ := c.Auth.LogoutURL("https://example.com/bye", ""); strings.Contains(raw, "state=") {
		t.Errorf("empty state sent: %s", raw)
	}
	noKey, _ := NewClient()
	if _, err := noKey.Auth.LogoutURL("x", ""); !errors.Is(err, ErrMissingCredential) {
		t.Errorf("no key: err = %v", err)
	}
}

const tokenResponse = `{"token_type":"bearer","access_token":"AT","id_token":"IDT","expires_in":43199,"refresh_token":"RT","refresh_token_expires_in":5184000,"scope":"account_email profile openid"}`

func fixedNow(c *Client, t time.Time) {
	c.now = func() time.Time { return t }
}

func TestAuthService_Exchange(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithRESTKey("rest"), WithClientSecret("secret"))
	now := time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)
	fixedNow(c, now)
	srv.HandleJSON("POST /oauth/token", 200, tokenResponse)

	tok, err := c.Auth.Exchange(t.Context(), "CODE", "https://example.com/cb", &ExchangeOptions{CodeVerifier: "verifier"})
	if err != nil {
		t.Fatal(err)
	}
	req := srv.LastRequest()
	if req.Header.Get("Authorization") != "" {
		t.Errorf("token request must not send Authorization, got %q", req.Header.Get("Authorization"))
	}
	f := req.Form()
	for k, want := range map[string]string{
		"grant_type": "authorization_code", "client_id": "rest", "client_secret": "secret",
		"redirect_uri": "https://example.com/cb", "code": "CODE", "code_verifier": "verifier",
	} {
		if got := f.Get(k); got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}
	if tok.AccessToken != "AT" || tok.IDToken != "IDT" || tok.RefreshToken != "RT" || tok.TokenType != "bearer" || tok.Scope != "account_email profile openid" {
		t.Errorf("token = %+v", tok)
	}
	if !tok.Expiry.Equal(now.Add(43199*time.Second)) || !tok.RefreshTokenExpiry.Equal(now.Add(60*24*time.Hour)) {
		t.Errorf("expiry = %v, refresh expiry = %v", tok.Expiry, tok.RefreshTokenExpiry)
	}

	if _, err := c.Auth.Exchange(t.Context(), "", "x", nil); err == nil {
		t.Error("empty code: want error")
	}
}

func TestAuthService_Exchange_noSecret(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithRESTKey("rest"))
	srv.HandleJSON("POST /oauth/token", 200, tokenResponse)
	if _, err := c.Auth.Exchange(t.Context(), "CODE", "x", nil); err != nil {
		t.Fatal(err)
	}
	if f := srv.LastRequest().Form(); f.Has("client_secret") || f.Has("code_verifier") {
		t.Errorf("unexpected fields: %v", f)
	}

	noKey, _ := newTestClient(t)
	if _, err := noKey.Auth.Exchange(t.Context(), "CODE", "x", nil); !errors.Is(err, ErrMissingCredential) {
		t.Errorf("no key: err = %v", err)
	}
}

func TestAuthService_Exchange_error(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithRESTKey("rest"))
	srv.HandleJSON("POST /oauth/token", 400, `{"error":"invalid_grant","error_description":"authorization code not found for code=CODE","error_code":"KOE320"}`)
	_, err := c.Auth.Exchange(t.Context(), "CODE", "x", nil)
	var oe *OAuthError
	if !errors.As(err, &oe) || oe.KOECode != "KOE320" || oe.Err != "invalid_grant" {
		t.Errorf("err = %v", err)
	}
}

func TestAuthService_Refresh(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithRESTKey("rest"))
	srv.HandleJSON("POST /oauth/token", 200, `{"token_type":"bearer","access_token":"AT2","expires_in":43199}`)
	tok, err := c.Auth.Refresh(t.Context(), "RT")
	if err != nil {
		t.Fatal(err)
	}
	f := srv.LastRequest().Form()
	if f.Get("grant_type") != "refresh_token" || f.Get("refresh_token") != "RT" || f.Get("client_id") != "rest" {
		t.Errorf("form = %v", f)
	}
	if tok.AccessToken != "AT2" || tok.RefreshToken != "" || !tok.RefreshTokenExpiry.IsZero() {
		t.Errorf("token = %+v", tok)
	}
	if _, err := c.Auth.Refresh(t.Context(), ""); err == nil {
		t.Error("empty refresh token: want error")
	}
}

func TestPKCE(t *testing.T) {
	t.Parallel()
	// RFC 7636 appendix B.
	if got := S256Challenge("dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"); got != "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM" {
		t.Errorf("S256Challenge = %q", got)
	}
	v1, v2 := GenerateVerifier(), GenerateVerifier()
	if v1 == v2 || len(v1) < 43 || len(v1) > 128 {
		t.Errorf("verifiers %q, %q", v1, v2)
	}
	for _, r := range v1 {
		if !strings.ContainsRune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~", r) {
			t.Errorf("verifier has invalid character %q", r)
		}
	}
}

func TestTokenSource(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithRESTKey("rest"))
	now := time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)
	var clock atomic.Pointer[time.Time]
	clock.Store(&now)
	c.now = func() time.Time { return *clock.Load() }

	var refreshes atomic.Int32
	srv.HandleJSON("POST /oauth/token", 200, `{"token_type":"bearer","access_token":"AT2","expires_in":3600}`)

	var saved []*Token
	var mu sync.Mutex
	ts := c.Auth.TokenSource(&Token{
		AccessToken: "AT1", RefreshToken: "RT1", Expiry: now.Add(10 * time.Minute),
		RefreshTokenExpiry: now.Add(30 * 24 * time.Hour),
	}, &TokenSourceOptions{OnRefresh: func(tok *Token) {
		refreshes.Add(1)
		mu.Lock()
		saved = append(saved, tok)
		mu.Unlock()
	}})

	tok, err := ts.Token(t.Context())
	if err != nil || tok.AccessToken != "AT1" {
		t.Fatalf("valid token: %v, %+v", err, tok)
	}

	later := now.Add(9*time.Minute + 30*time.Second) // within the refresh leeway
	clock.Store(&later)
	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			tok, err := ts.Token(context.Background())
			if err != nil || tok.AccessToken != "AT2" {
				t.Errorf("refreshed token: %v, %+v", err, tok)
			}
		})
	}
	wg.Wait()
	if n := refreshes.Load(); n != 1 {
		t.Errorf("refreshes = %d, want 1", n)
	}
	if got := saved[0]; got.RefreshToken != "RT1" || !got.RefreshTokenExpiry.Equal(now.Add(30*24*time.Hour)) {
		t.Errorf("old refresh token not carried over: %+v", got)
	}
	if f := srv.LastRequest().Form(); f.Get("refresh_token") != "RT1" {
		t.Errorf("form = %v", f)
	}

	tok.AccessToken = "mutated"
	if again, _ := ts.Token(t.Context()); again.AccessToken != "AT2" {
		t.Error("Token returned a shared pointer")
	}
}

func TestTokenSource_rotatesRefreshTokenAndReportsErrors(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithRESTKey("rest"))
	srv.HandleJSON("POST /oauth/token", 200, `{"access_token":"AT2","expires_in":3600,"refresh_token":"RT2","refresh_token_expires_in":5184000}`)
	ts := c.Auth.TokenSource(&Token{RefreshToken: "RT1"}, nil) // no access token yet
	tok, err := ts.Token(t.Context())
	if err != nil || tok.RefreshToken != "RT2" {
		t.Fatalf("%v, %+v", err, tok)
	}

	c2, srv2 := newTestClient(t, WithRESTKey("rest"))
	srv2.HandleJSON("POST /oauth/token", 400, `{"error":"invalid_grant","error_code":"KOE322"}`)
	ts2 := c2.Auth.TokenSource(&Token{AccessToken: "x", RefreshToken: "bad", Expiry: time.Now().Add(-time.Hour)}, nil)
	var oe *OAuthError
	if _, err := ts2.Token(t.Context()); !errors.As(err, &oe) {
		t.Errorf("err = %v", err)
	}

	ts3 := c2.Auth.TokenSource(&Token{AccessToken: "x"}, nil) // zero expiry: assumed valid
	if tok, err := ts3.Token(t.Context()); err != nil || tok.AccessToken != "x" {
		t.Errorf("%v, %+v", err, tok)
	}
}
