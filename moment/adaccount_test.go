package moment

import (
	"errors"
	"net/http"
	"testing"

	kakao "github.com/donghquinn/gokakao"
)

func TestAdAccountService_List(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/adAccounts/pages", 200, `{
  "content": [
    {"id": 111, "name": "테스트계정1", "memberType": "MASTER", "config": "ON"},
    {"id": 222, "name": "테스트계정2", "memberType": "MASTER", "config": "ON"}
  ],
  "number": 0, "size": 2, "totalElements": 92, "totalPages": 46, "page": 0
}`)

	page, err := c.AdAccounts.List(t.Context(), &AdAccountListOptions{PageOptions: PageOptions{Page: 1, Size: 2}, Configs: []Config{ConfigOn, ConfigOff}})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "")
	if req.Method != http.MethodGet || req.Query.Get("config") != "ON,OFF" || req.Query.Get("page") != "1" || req.Query.Get("size") != "2" {
		t.Errorf("request = %s %s?%v", req.Method, req.Path, req.Query)
	}
	if len(page.Content) != 2 || page.Content[1].ID != 222 || page.Content[1].Name != "테스트계정2" ||
		page.Content[0].MemberType != MemberMaster || page.Content[0].Config != ConfigOn ||
		page.TotalElements != 92 || page.TotalPages != 46 || page.Size != 2 {
		t.Errorf("page = %+v", page)
	}

	if _, err := c.AdAccounts.List(t.Context(), nil); err != nil {
		t.Fatal(err)
	}
	if q := srv.LastRequest().Query; len(q) != 0 {
		t.Errorf("nil options query = %v", q)
	}
}

func TestAdAccountService_Get(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/adAccounts/1111", 200, `{
  "id": 1111,
  "name": "카카오모먼트_광고계정",
  "advertiser": {"businessRegistrationNumber": "120-81-47521", "name": "주식회사 카카오"},
  "type": "BUSINESS",
  "config": "ON",
  "isAdminStop": false,
  "isOutOfBalance": false,
  "statusDescription": "운영중",
  "bizWalletId": 1111
}`)

	a, err := c.AdAccounts.Get(t.Context(), 1111)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodGet {
		t.Errorf("method = %s", req.Method)
	}
	if a.ID != 1111 || a.Name != "카카오모먼트_광고계정" || a.Type != AdAccountBusiness || a.Config != ConfigOn ||
		a.AdminStop || a.OutOfBalance || a.StatusDescription != "운영중" || a.BizWalletID != 1111 ||
		a.Advertiser == nil || a.Advertiser.BusinessRegistrationNumber != "120-81-47521" || a.Advertiser.Name != "주식회사 카카오" {
		t.Errorf("ad account = %+v", a)
	}

	if _, err := c.AdAccounts.Get(t.Context(), 0); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestAdAccountService_Get_error(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/adAccounts/1", 400, `{"code":-813,"msg":"KakaoMomentException","extras":{"detailCode":21006,"detailMsg":"존재하지 않는 광고계정입니다."}}`)

	_, err := c.AdAccounts.Get(t.Context(), 1)
	var e *kakao.Error
	if !errors.As(err, &e) || e.Code != -813 {
		t.Errorf("err = %v", err)
	}
}

func TestAdAccountService_Balance(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/adAccounts/balance", 200, `{"id": 1111, "bizWalletId": 1111, "cash": 5000000, "freeCash": 100000}`)

	b, err := c.AdAccounts.Balance(t.Context(), 1111)
	if err != nil {
		t.Fatal(err)
	}
	assertRequest(t, srv, "1111")
	if *b != (AdAccountBalance{ID: 1111, BizWalletID: 1111, Cash: 5000000, FreeCash: 100000}) {
		t.Errorf("balance = %+v", b)
	}

	if _, err := c.AdAccounts.Balance(t.Context(), 0); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestAdAccountService_BizRight(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/adAccounts/bizRight", 200, `{"id": 1111, "bizRightCompanyName": "주식회사 카카오"}`)

	b, err := c.AdAccounts.BizRight(t.Context(), 1111)
	if err != nil {
		t.Fatal(err)
	}
	assertRequest(t, srv, "1111")
	if *b != (BizRight{ID: 1111, BizRightCompanyName: "주식회사 카카오"}) {
		t.Errorf("biz right = %+v", b)
	}

	if _, err := c.AdAccounts.BizRight(t.Context(), 0); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestAdAccountService_ApproveBizRight(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v4/adAccounts/1111/bizRight", 200, `{"id": 1111, "bizRightCompanyName": "주식회사 카카오"}`)

	b, err := c.AdAccounts.ApproveBizRight(t.Context(), 1111)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodPost || len(req.Body) != 0 {
		t.Errorf("request = %s %s", req.Method, req.Body)
	}
	if b.ID != 1111 || b.BizRightCompanyName != "주식회사 카카오" {
		t.Errorf("biz right = %+v", b)
	}

	if _, err := c.AdAccounts.ApproveBizRight(t.Context(), 0); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestAdAccountService_ChannelProfiles(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/adAccounts/channel/profiles", 200, `[
  {"id": "_ZQxd", "name": "카카오채널", "searchId": "카카오채널"},
  {"id": "_Xxju", "name": "카카오채널2", "searchId": "카카오채널2"}
]`)

	ps, err := c.AdAccounts.ChannelProfiles(t.Context(), 1111)
	if err != nil {
		t.Fatal(err)
	}
	assertRequest(t, srv, "1111")
	if len(ps) != 2 || ps[1] != (ChannelProfile{ID: "_Xxju", Name: "카카오채널2", SearchID: "카카오채널2"}) {
		t.Errorf("profiles = %+v", ps)
	}

	if _, err := c.AdAccounts.ChannelProfiles(t.Context(), 0); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestAdAccountService_Trackers(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/adAccounts/trackers", 200, `[{"id": "1", "name": "트래커1"}, {"id": "2", "name": "트래커2"}]`)

	ts, err := c.AdAccounts.Trackers(t.Context(), 1111, &TrackerListOptions{CampaignType: CampaignTypeTalkBizBoard, Goal: GoalVisiting})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Query.Get("campaignType") != "TALK_BIZ_BOARD" || req.Query.Get("goal") != "VISITING" {
		t.Errorf("query = %v", req.Query)
	}
	if len(ts) != 2 || ts[1] != (Tracker{ID: "2", Name: "트래커2"}) {
		t.Errorf("trackers = %+v", ts)
	}

	if _, err := c.AdAccounts.Trackers(t.Context(), 1111, nil); err != nil {
		t.Fatal(err)
	}
	if q := srv.LastRequest().Query; len(q) != 0 {
		t.Errorf("nil options query = %v", q)
	}
	if _, err := c.AdAccounts.Trackers(t.Context(), 0, nil); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestAdAccountService_Create(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v4/adAccounts", 200, `{
  "id": 123,
  "name": "테스트_광고계정",
  "advertiser": {"businessRegistrationNumber": "678-67-67890", "name": "광고주 사업자"},
  "type": "BUSINESS",
  "config": "ON",
  "isAdminStop": false,
  "isOutOfBalance": true,
  "statusDescription": "잔액부족"
}`)

	a, err := c.AdAccounts.Create(t.Context(), NewAdAccount{Name: "테스트_광고계정", AdvertiserCompanyID: 34})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "")
	if req.Method != http.MethodPost || string(req.Body) != `{"name":"테스트_광고계정","advertiserCompanyId":34}` {
		t.Errorf("request = %s %s", req.Method, req.Body)
	}
	if a.ID != 123 || !a.OutOfBalance || a.StatusDescription != "잔액부족" || a.Advertiser.Name != "광고주 사업자" {
		t.Errorf("ad account = %+v", a)
	}

	for _, na := range []NewAdAccount{{Name: "n"}, {AdvertiserCompanyID: 34}} {
		if _, err := c.AdAccounts.Create(t.Context(), na); err == nil {
			t.Errorf("Create(%+v): want error", na)
		}
	}
}

func TestAdAccountService_SetConfig(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PUT /openapi/v4/adAccounts/onOff", 200, ``)

	if err := c.AdAccounts.SetConfig(t.Context(), 1111, ConfigOn); err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodPut || string(req.Body) != `{"config":"ON"}` {
		t.Errorf("request = %s %s", req.Method, req.Body)
	}

	if err := c.AdAccounts.SetConfig(t.Context(), 1111, ConfigDel); err == nil {
		t.Error("DEL: want error")
	}
	if err := c.AdAccounts.SetConfig(t.Context(), 0, ConfigOff); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestAdAccountService_Company(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/companies", 200, `{
  "id": 680,
  "businessRegistrationNumber": "123-12-12345",
  "name": "주식회사 사업자",
  "representationName": "사업자대표",
  "businessCategory": "서비스",
  "businessItem": "서비스",
  "address": {"zipCode": "13494", "baseAddress": "경기 성남시 분당구 판교역로 235", "detailAddress": "H스퀘어"}
}`)

	co, err := c.AdAccounts.Company(t.Context(), "120-81-47521")
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "")
	if req.Method != http.MethodGet || req.Query.Get("bizNum") != "120-81-47521" {
		t.Errorf("request = %s ?%v", req.Method, req.Query)
	}
	if co.ID != 680 || co.BusinessRegistrationNumber != "123-12-12345" || co.Name != "주식회사 사업자" ||
		co.RepresentationName != "사업자대표" || co.BusinessCategory != "서비스" || co.BusinessItem != "서비스" ||
		co.Address == nil || *co.Address != (CompanyAddress{ZipCode: "13494", BaseAddress: "경기 성남시 분당구 판교역로 235", DetailAddress: "H스퀘어"}) {
		t.Errorf("company = %+v", co)
	}

	if _, err := c.AdAccounts.Company(t.Context(), ""); err == nil {
		t.Error("missing biz num: want error")
	}
}

// TestAdAccountService_apiErrors checks that every ad account method
// returns API errors.
func TestAdAccountService_apiErrors(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("/", 400, `{"code":-813,"msg":"KakaoMomentException"}`)
	ctx := t.Context()
	calls := map[string]func() error{
		"List":            func() error { _, err := c.AdAccounts.List(ctx, nil); return err },
		"Balance":         func() error { _, err := c.AdAccounts.Balance(ctx, 1); return err },
		"BizRight":        func() error { _, err := c.AdAccounts.BizRight(ctx, 1); return err },
		"ApproveBizRight": func() error { _, err := c.AdAccounts.ApproveBizRight(ctx, 1); return err },
		"ChannelProfiles": func() error { _, err := c.AdAccounts.ChannelProfiles(ctx, 1); return err },
		"Trackers":        func() error { _, err := c.AdAccounts.Trackers(ctx, 1, nil); return err },
		"Create": func() error {
			_, err := c.AdAccounts.Create(ctx, NewAdAccount{Name: "n", AdvertiserCompanyID: 1})
			return err
		},
		"Company": func() error { _, err := c.AdAccounts.Company(ctx, "1"); return err },
	}
	for name, f := range calls {
		var e *kakao.Error
		if err := f(); !errors.As(err, &e) || e.Code != -813 {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}
