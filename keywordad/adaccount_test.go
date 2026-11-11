package keywordad

import (
	"errors"
	"slices"
	"testing"

	kakao "github.com/donghquinn/gokakao"
)

func TestAdAccountService_List(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v1/adAccounts/pages", 200, `{"content":[{"id":111,"name":"테스트계정1","memberType":"MASTER","config":"ON"},{"id":"222","name":"테스트계정2","memberType":"MEMBER","config":"OFF"}],"size":2,"totalElements":92,"totalPages":46,"page":0}`)

	page, err := c.AdAccounts.List(t.Context(), &AdAccountListOptions{Configs: []Config{ConfigOn, ConfigOff}, PageOptions: PageOptions{Page: 1, Size: 2}})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "")
	if !slices.Equal(req.Query["config"], []string{"ON", "OFF"}) || req.Query.Get("page") != "1" || req.Query.Get("size") != "2" {
		t.Errorf("query = %v", req.Query)
	}
	if page.TotalElements != 92 || page.TotalPages != 46 || len(page.Content) != 2 ||
		page.Content[1] != (AdAccount{ID: 222, Name: "테스트계정2", MemberType: MemberMember, Config: ConfigOff}) {
		t.Errorf("page = %+v", page)
	}

	if _, err := c.AdAccounts.List(t.Context(), nil); err != nil {
		t.Fatal(err)
	}
	if q := srv.LastRequest().Query; len(q) != 0 {
		t.Errorf("nil options query = %v", q)
	}
}

func TestAdAccountService_reads(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	detail := `{"id":"1111111111","name":"카카오키워드광고_광고계정1","type":"BUSINESS","config":"ON","adminStop":false,"outOfBalance":false,"statusDescription":"운영 가능","bizWalletId":"11111","advertiser":{"businessRegistrationNumber":"120-81-47521","name":"주식회사 카카오"}}`
	srv.HandleJSON("GET /openapi/v1/adAccounts/1111111111", 200, detail)
	srv.HandleJSON("GET /openapi/v1/adAccounts/balance", 200, `{"id":"1111111111","bizWalletId":"11111","cash":5000000.0,"freeCash":100000.0}`)
	srv.HandleJSON("GET /openapi/v1/adAccounts/bizRight", 200, `{"id":1111111111,"bizRightCompanyName":"주식회사 카카오"}`)
	srv.HandleJSON("GET /openapi/v1/adAccounts/channel/profiles", 200, `[{"uuid":"@kakao1","name":"카카오채널1"}]`)
	ctx := t.Context()

	a, err := c.AdAccounts.Get(ctx, 1111111111)
	if err != nil {
		t.Fatal(err)
	}
	assertRequest(t, srv, "1111111111")
	if a.ID != 1111111111 || a.Type != AdAccountBusiness || a.BizWalletID != 11111 || a.StatusDescription != "운영 가능" ||
		a.Advertiser == nil || a.Advertiser.Name != "주식회사 카카오" {
		t.Errorf("account = %+v", a)
	}

	b, err := c.AdAccounts.Balance(ctx, 1111111111)
	if err != nil || b.Cash != 5000000 || b.FreeCash != 100000 || b.BizWalletID != 11111 {
		t.Errorf("balance = %+v, %v", b, err)
	}
	assertRequest(t, srv, "1111111111")

	r, err := c.AdAccounts.BizRight(ctx, 1111111111)
	if err != nil || r.BizRightCompanyName != "주식회사 카카오" {
		t.Errorf("biz right = %+v, %v", r, err)
	}

	ps, err := c.AdAccounts.ChannelProfiles(ctx)
	if err != nil || len(ps) != 1 || ps[0] != (ChannelProfile{UUID: "@kakao1", Name: "카카오채널1"}) {
		t.Errorf("profiles = %+v, %v", ps, err)
	}
	assertRequest(t, srv, "")
}

func TestAdAccountService_writes(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v1/adAccounts", 200, `{"id":"1111111111","name":"테스트_광고계정","type":"BUSINESS","config":"ON"}`)
	srv.HandleJSON("PUT /openapi/v1/adAccounts", 200, ``)

	a, err := c.AdAccounts.Create(t.Context(), NewAdAccount{Name: "테스트_광고계정", AdvertiserCompanyID: 34})
	if err != nil || a.ID != 1111111111 {
		t.Fatalf("Create = %+v, %v", a, err)
	}
	req := assertRequest(t, srv, "")
	if string(req.Body) != `{"name":"테스트_광고계정","advertiserCompanyId":34}` || req.Header.Get("Content-Type") != "application/json" {
		t.Errorf("body = %s", req.Body)
	}

	if err := c.AdAccounts.SetConfig(t.Context(), 1111111111, ConfigOff); err != nil {
		t.Fatal(err)
	}
	req = assertRequest(t, srv, "1111111111")
	if string(req.Body) != `{"config":"OFF"}` {
		t.Errorf("body = %s", req.Body)
	}

	if _, err := c.AdAccounts.Create(t.Context(), NewAdAccount{Name: "x"}); err == nil {
		t.Error("missing advertiser: want error")
	}
}

func TestClient_errors(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v1/adAccounts/balance", 400, `{"code":-820,"msg":"KakaoKeywordAdException","extras":{"detailCode":103005,"detailMsg":"없는 비즈채널입니다."}}`)
	_, err := c.AdAccounts.Balance(t.Context(), 1)
	var e *kakao.Error
	if !errors.As(err, &e) || e.Extras == nil || e.Extras.DetailCode != 103005 {
		t.Errorf("err = %v", err)
	}

	noToken := New(c.kc, "")
	if _, err := noToken.AdAccounts.ChannelProfiles(t.Context()); !errors.Is(err, kakao.ErrMissingCredential) {
		t.Errorf("no token: err = %v", err)
	}
}
