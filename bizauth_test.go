package kakao

import (
	"errors"
	"net/url"
	"slices"
	"testing"
	"time"
)

func TestBusinessService_AuthCodeURL(t *testing.T) {
	t.Parallel()
	c, err := NewClient(WithRESTKey("rest"))
	if err != nil {
		t.Fatal(err)
	}

	raw, err := c.Business.AuthCodeURL("https://example.com/biz", []string{"moment_create", "moment_management", "moment_delete"},
		&BusinessAuthOptions{State: "s1", ResourceIDs: []string{"moment:*", "keyword:*"}})
	if err != nil {
		t.Fatal(err)
	}
	u, _ := url.Parse(raw)
	if u.Host != "kauth.kakao.com" || u.Path != "/oauth/business/authorize" {
		t.Errorf("URL = %s", raw)
	}
	q := u.Query()
	if q.Get("client_id") != "rest" || q.Get("response_type") != "code" || q.Get("redirect_uri") != "https://example.com/biz" ||
		q.Get("scope") != "moment_create,moment_management,moment_delete" || q.Get("state") != "s1" ||
		!slices.Equal(q["resource_ids"], []string{"moment:*", "keyword:*"}) {
		t.Errorf("query = %v", q)
	}

	raw, err = c.Business.AuthCodeURL("https://example.com/biz", []string{"moment_management"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if u, _ := url.Parse(raw); len(u.Query()) != 4 {
		t.Errorf("nil options: %s", raw)
	}

	for name, scopes := range map[string][]string{
		"no scopes":            nil,
		"email alone":          {"biz_account_email"},
		"create without asset": {"keyword_create", "keyword_management"},
	} {
		if _, err := c.Business.AuthCodeURL("x", scopes, nil); err == nil {
			t.Errorf("%s: want error", name)
		}
	}
	noKey, _ := NewClient()
	if _, err := noKey.Business.AuthCodeURL("x", []string{"moment_management"}, nil); !errors.Is(err, ErrMissingCredential) {
		t.Errorf("no key: err = %v", err)
	}
}

func TestBusinessService_Exchange(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithRESTKey("rest"), WithClientSecret("secret"))
	srv.HandleJSON("POST /oauth/business/token", 200, `{"access_token":"BT","token_type":"bearer","scope":"moment_create moment_management moment_delete"}`)

	tok, err := c.Business.Exchange(t.Context(), "CODE", "https://example.com/biz")
	if err != nil {
		t.Fatal(err)
	}
	want := url.Values{"grant_type": {"authorization_code"}, "client_id": {"rest"}, "code": {"CODE"}, "redirect_uri": {"https://example.com/biz"}, "client_secret": {"secret"}}
	if f := srv.LastRequest().Form(); !equalValues(f, want) {
		t.Errorf("form = %v", f)
	}
	if *tok != (BusinessToken{TokenType: "bearer", AccessToken: "BT", Scope: "moment_create moment_management moment_delete"}) {
		t.Errorf("token = %+v", tok)
	}

	if _, err := c.Business.Exchange(t.Context(), "", "x"); err == nil {
		t.Error("empty code: want error")
	}
	noKey, _ := newTestClient(t)
	if _, err := noKey.Business.Exchange(t.Context(), "CODE", "x"); !errors.Is(err, ErrMissingCredential) {
		t.Errorf("no key: err = %v", err)
	}

	bad, srv2 := newTestClient(t, WithRESTKey("rest"))
	srv2.HandleJSON("POST /oauth/business/token", 400, `{"error":"invalid_grant","error_code":"KOE320"}`)
	var oe *OAuthError
	if _, err := bad.Business.Exchange(t.Context(), "CODE", "x"); !errors.As(err, &oe) {
		t.Errorf("err = %v", err)
	}
}

func TestBusinessService_TokenInfoAndRevoke(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /v1/business/tokeninfo", 200, `{"id":1234567890,"token_user_id":"tu1","biz_agreements":[
		{"group":"personal","scope":["biz_account_email"]},
		{"group":"keyword","scope":["keyword_management","keyword_create","keyword_delete"],"ad_account_ids":["*"]}],
		"status":"ACTIVE","updated_at":"2024-08-21T01:56:59Z"}`)
	srv.HandleJSON("POST /v1/business/revoke", 200, `{"id":1234567890,"token_user_id":"tu1"}`)

	info, err := c.Business.TokenInfo(t.Context(), "BT")
	if err != nil {
		t.Fatal(err)
	}
	if srv.LastRequest().Header.Get("Authorization") != "Bearer BT" {
		t.Error("missing business token")
	}
	if info.ID != 1234567890 || info.TokenUserID != "tu1" || info.Status != "ACTIVE" || len(info.BizAgreements) != 2 ||
		info.BizAgreements[1].Group != "keyword" || !slices.Equal(info.BizAgreements[1].AdAccountIDs, []string{"*"}) ||
		!info.UpdatedAt.Equal(time.Date(2024, 8, 21, 1, 56, 59, 0, time.UTC)) {
		t.Errorf("info = %+v", info)
	}

	rev, err := c.Business.Revoke(t.Context(), "BT")
	if err != nil || *rev != (RevokedBusinessToken{ID: 1234567890, TokenUserID: "tu1"}) {
		t.Errorf("revoke = %+v, %v", rev, err)
	}

	if _, err := c.Business.TokenInfo(t.Context(), ""); !errors.Is(err, ErrMissingCredential) {
		t.Errorf("empty token: err = %v", err)
	}
	if _, err := c.Business.Revoke(t.Context(), ""); !errors.Is(err, ErrMissingCredential) {
		t.Errorf("empty token: err = %v", err)
	}
}
