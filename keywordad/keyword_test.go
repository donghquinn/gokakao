package keywordad

import (
	"net/http"
	"slices"
	"strings"
	"testing"
)

const (
	keywordTestJSON1 = `{"adGroupId":"4444444441","id":"55555555551","text":"키워드1","config":"ON",` +
		`"landingInfo":{"rspvLandingUrl":"https://www.daum.net","pcLandingUrl":null,"mobileLandingUrl":null},` +
		`"bidStrategy":{"type":"MANUAL","bidAmount":1000},` +
		`"status":["WAITING","OFF_BY_BIZ_CHANNEL_WAITING","OFF"],"reviewStatus":"WAITING"}`
	keywordTestJSON2 = `{"adGroupId":"4444444441","id":"55555555552","text":"키워드2","config":"ON",` +
		`"landingInfo":{"rspvLandingUrl":null,"pcLandingUrl":"https://www.daum.net","mobileLandingUrl":"https://www.daum.net"},` +
		`"bidStrategy":{"type":"AD_GROUP","bidAmount":null},` +
		`"status":["OFF_BY_BIZ_CHANNEL_WAITING","WAITING"],"reviewStatus":"WAITING"}`
)

func checkKeywords(t *testing.T, ks []Keyword) {
	t.Helper()
	if len(ks) != 2 {
		t.Fatalf("keywords = %+v", ks)
	}
	k := ks[0]
	if k.AdGroupID != 4444444441 || k.ID != 55555555551 || k.Text != "키워드1" || k.Config != ConfigOn ||
		k.LandingInfo != (LandingInfo{RspvLandingURL: "https://www.daum.net"}) ||
		k.BidStrategy != (BidStrategy{Type: BidManual, BidAmount: 1000}) ||
		!slices.Equal(k.Status, []Status{StatusWaiting, StatusOffByBizChannelWaiting, StatusOff}) ||
		k.ReviewStatus != ReviewWaiting {
		t.Errorf("keyword 0 = %+v", k)
	}
	k = ks[1]
	if k.ID != 55555555552 || k.LandingInfo.PCLandingURL != "https://www.daum.net" || k.LandingInfo.MobileLandingURL != "https://www.daum.net" ||
		k.BidStrategy != (BidStrategy{Type: BidByAdGroup}) {
		t.Errorf("keyword 1 = %+v", k)
	}
}

func TestKeywordService_List(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v1/keywords", 200, `[`+keywordTestJSON1+`,`+keywordTestJSON2+`]`)

	ks, err := c.Keywords.List(t.Context(), 1111, 4444444441, &KeywordListOptions{Configs: []Config{ConfigOn}})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodGet || req.Query.Get("adGroupId") != "4444444441" || !slices.Equal(req.Query["config"], []string{"ON"}) {
		t.Errorf("request = %s %v", req.Method, req.Query)
	}
	checkKeywords(t, ks)

	if _, err := c.Keywords.List(t.Context(), 1111, 4444444441, nil); err != nil {
		t.Fatal(err)
	}
	if q := srv.LastRequest().Query; len(q) != 1 {
		t.Errorf("nil options query = %v", q)
	}
	if _, err := c.Keywords.List(t.Context(), 1111, 0, nil); err == nil {
		t.Error("missing ad group ID: want error")
	}
}

func TestKeywordService_Get(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v1/keywords/55555555551", 200, `{"adGroupId":"4444444441","id":"55555555551","text":"키워드1","config":"OFF",`+
		`"landingInfo":{"rspvLandingUrl":null,"pcLandingUrl":null,"mobileLandingUrl":null},"bidStrategy":{"type":"AD_GROUP","bidAmount":null},`+
		`"status":["WAITING","OFF_BY_BIZ_CHANNEL_WAITING","OFF"],"reviewStatus":"WAITING"}`)

	k, err := c.Keywords.Get(t.Context(), 1111, 55555555551)
	if err != nil {
		t.Fatal(err)
	}
	assertRequest(t, srv, "1111")
	if k.ID != 55555555551 || k.Config != ConfigOff || k.LandingInfo != (LandingInfo{}) || k.BidStrategy.Type != BidByAdGroup ||
		len(k.Status) != 3 || k.ReviewStatus != ReviewWaiting {
		t.Errorf("keyword = %+v", k)
	}
}

func TestKeywordService_Create(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v1/keywords", 200, `[`+keywordTestJSON1+`,`+keywordTestJSON2+`]`)

	ks, err := c.Keywords.Create(t.Context(), 1111, NewKeywords{
		AdGroupID: 4444444441, Texts: []string{"키워드1", "키워드2"}, BidStrategy: BidStrategy{Type: BidByAdGroup},
	})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if want := `{"adGroupId":4444444441,"texts":["키워드1","키워드2"],"bidStrategy":{"type":"AD_GROUP"}}`; req.Method != http.MethodPost || string(req.Body) != want {
		t.Errorf("body = %s", req.Body)
	}
	checkKeywords(t, ks)

	for name, k := range map[string]NewKeywords{
		"no ad group":   {Texts: []string{"a"}, BidStrategy: BidStrategy{Type: BidByAdGroup}},
		"no texts":      {AdGroupID: 1, BidStrategy: BidStrategy{Type: BidByAdGroup}},
		"no strategy":   {AdGroupID: 1, Texts: []string{"a"}},
		"manual no bid": {AdGroupID: 1, Texts: []string{"a"}, BidStrategy: BidStrategy{Type: BidManual}},
	} {
		if _, err := c.Keywords.Create(t.Context(), 1111, k); err == nil || !strings.HasPrefix(err.Error(), "keywordad: ") {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}

func TestKeywordService_CreateIndividual(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v1/keywords/individual", 200, `[`+keywordTestJSON1+`,`+keywordTestJSON2+`]`)

	ks, err := c.Keywords.CreateIndividual(t.Context(), 1111, NewIndividualKeywords{
		AdGroupID: 4444444441,
		Keywords: []NewIndividualKeyword{
			{Text: "키워드1", BidStrategy: BidStrategy{Type: BidByAdGroup}, LandingInfo: &LandingInfo{RspvLandingURL: "https://www.daum.net"}},
			{Text: "키워드2", BidStrategy: BidStrategy{Type: BidManual, BidAmount: 1000}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	want := `{"adGroupId":4444444441,"individualKeywords":[` +
		`{"text":"키워드1","bidStrategy":{"type":"AD_GROUP"},"landingInfo":{"rspvLandingUrl":"https://www.daum.net"}},` +
		`{"text":"키워드2","bidStrategy":{"type":"MANUAL","bidAmount":1000}}]}`
	if req.Method != http.MethodPost || string(req.Body) != want {
		t.Errorf("body = %s", req.Body)
	}
	checkKeywords(t, ks)

	for name, k := range map[string]NewIndividualKeywords{
		"no ad group": {Keywords: []NewIndividualKeyword{{Text: "a", BidStrategy: BidStrategy{Type: BidByAdGroup}}}},
		"no keywords": {AdGroupID: 1},
		"no text":     {AdGroupID: 1, Keywords: []NewIndividualKeyword{{BidStrategy: BidStrategy{Type: BidByAdGroup}}}},
		"no strategy": {AdGroupID: 1, Keywords: []NewIndividualKeyword{{Text: "a"}}},
	} {
		if _, err := c.Keywords.CreateIndividual(t.Context(), 1111, k); err == nil {
			t.Errorf("%s: want error", name)
		}
	}
}

func TestKeywordService_Update(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PUT /openapi/v1/keywords/5555555551", 200, `{"adGroupId":"4444444441","id":"5555555551","text":"키워드1","config":"ON",`+
		`"landingInfo":{"rspvLandingUrl":null,"pcLandingUrl":null,"mobileLandingUrl":null},"bidStrategy":{"type":"AD_GROUP","bidAmount":null},`+
		`"status":["OFF_BY_BIZ_CHANNEL_WAITING","WAITING"],"reviewStatus":"WAITING"}`)

	k, err := c.Keywords.Update(t.Context(), 1111, 5555555551, KeywordChanges{BidStrategy: BidStrategy{Type: BidByAdGroup}})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodPut || string(req.Body) != `{"bidStrategy":{"type":"AD_GROUP"}}` {
		t.Errorf("body = %s", req.Body)
	}
	if k.ID != 5555555551 || k.Text != "키워드1" || k.BidStrategy.Type != BidByAdGroup {
		t.Errorf("keyword = %+v", k)
	}

	if _, err := c.Keywords.Update(t.Context(), 1111, 5555555551, KeywordChanges{
		LandingInfo: &LandingInfo{PCLandingURL: "https://a", MobileLandingURL: "https://b"},
		BidStrategy: BidStrategy{Type: BidManual, BidAmount: 500},
	}); err != nil {
		t.Fatal(err)
	}
	if body := string(srv.LastRequest().Body); body != `{"landingInfo":{"pcLandingUrl":"https://a","mobileLandingUrl":"https://b"},"bidStrategy":{"type":"MANUAL","bidAmount":500}}` {
		t.Errorf("body = %s", body)
	}

	if _, err := c.Keywords.Update(t.Context(), 1111, 5555555551, KeywordChanges{}); err == nil {
		t.Error("missing bid strategy: want error")
	}
}

func TestKeywordService_patches(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	for _, p := range []string{"onOff", "bidStrategy", "landingInfo"} {
		srv.HandleJSON("PATCH /openapi/v1/keywords/55555555551/"+p, 200, ``)
	}
	srv.HandleJSON("DELETE /openapi/v1/keywords/55555555551", 200, ``)
	ctx := t.Context()

	tests := []struct {
		name   string
		call   func() error
		method string
		path   string
		body   string
	}{
		{
			"SetConfig", func() error { return c.Keywords.SetConfig(ctx, 1111, 55555555551, ConfigOff) },
			http.MethodPatch, "/onOff", `{"config":"OFF"}`,
		},
		{"UpdateBidStrategy", func() error {
			return c.Keywords.UpdateBidStrategy(ctx, 1111, 55555555551, BidStrategy{Type: BidManual, BidAmount: 3000})
		}, http.MethodPatch, "/bidStrategy", `{"bidStrategy":{"type":"MANUAL","bidAmount":3000}}`},
		{"UpdateLandingInfo", func() error {
			return c.Keywords.UpdateLandingInfo(ctx, 1111, 55555555551, LandingInfo{PCLandingURL: "https://www.daum.net", MobileLandingURL: "https://www.daum.net"})
		}, http.MethodPatch, "/landingInfo", `{"landingInfo":{"pcLandingUrl":"https://www.daum.net","mobileLandingUrl":"https://www.daum.net"}}`},
		{
			"UpdateLandingInfo clear", func() error { return c.Keywords.UpdateLandingInfo(ctx, 1111, 55555555551, LandingInfo{}) },
			http.MethodPatch, "/landingInfo", `{"landingInfo":{}}`,
		},
		{"Delete", func() error { return c.Keywords.Delete(ctx, 1111, 55555555551) }, http.MethodDelete, "", ``},
	}
	for _, tt := range tests {
		if err := tt.call(); err != nil {
			t.Fatalf("%s: %v", tt.name, err)
		}
		req := assertRequest(t, srv, "1111")
		if req.Method != tt.method || req.Path != "/openapi/v1/keywords/55555555551"+tt.path || string(req.Body) != tt.body {
			t.Errorf("%s: request = %s %s %s", tt.name, req.Method, req.Path, req.Body)
		}
	}

	if err := c.Keywords.UpdateBidStrategy(ctx, 1111, 55555555551, BidStrategy{Type: BidManual}); err == nil {
		t.Error("manual without bid: want error")
	}
}

func TestKeywordService_reviewAndQuality(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v1/keywords/55555555551/reviewReason", 200,
		`{"id":"55555555551","reasons":["신청하신 키워드는 사이트와 관련성이 부족한 키워드로 확인되어 광고 등록이 불가합니다.","신청하신 키워드는 광고 소재 부적합 키워드로 분류되어 광고 등록이 불가합니다.\n - 비속어, 은어, 성인금칙어, 무리한 축약어"]}`)
	srv.HandleJSON("GET /openapi/v1/keywords/55555555551/quality", 200, `{"id":"55555555551","text":"키워드1","quality":2}`)

	r, err := c.Keywords.ReviewHoldReasons(t.Context(), 1111, 55555555551)
	if err != nil {
		t.Fatal(err)
	}
	assertRequest(t, srv, "1111")
	if r.ID != 55555555551 || len(r.Reasons) != 2 || !strings.HasSuffix(r.Reasons[1], "무리한 축약어") {
		t.Errorf("reasons = %+v", r)
	}

	q, err := c.Keywords.QualityIndex(t.Context(), 1111, 55555555551)
	if err != nil {
		t.Fatal(err)
	}
	if req := assertRequest(t, srv, "1111"); req.Method != http.MethodGet || req.Path != "/openapi/v1/keywords/55555555551/quality" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	if *q != (KeywordQuality{ID: 55555555551, Text: "키워드1", Quality: 2}) {
		t.Errorf("quality = %+v", q)
	}
}

func TestKeywordService_apiErrors(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	const errBody = `{"code":-813,"msg":"KakaoKeywordAdException","extras":{"detailCode":41001,"detailMsg":"키워드가 존재하지 않습니다."}}`
	for _, p := range []string{"", "/individual", "/1", "/1/reviewReason", "/1/quality"} {
		srv.HandleJSON("/openapi/v1/keywords"+p, 400, errBody)
	}
	ctx := t.Context()
	bs := BidStrategy{Type: BidByAdGroup}

	if _, err := c.Keywords.List(ctx, 9, 1, nil); err == nil {
		t.Error("List: want error")
	}
	if _, err := c.Keywords.Get(ctx, 9, 1); err == nil {
		t.Error("Get: want error")
	}
	if _, err := c.Keywords.Create(ctx, 9, NewKeywords{AdGroupID: 1, Texts: []string{"a"}, BidStrategy: bs}); err == nil {
		t.Error("Create: want error")
	}
	if _, err := c.Keywords.CreateIndividual(ctx, 9, NewIndividualKeywords{AdGroupID: 1, Keywords: []NewIndividualKeyword{{Text: "a", BidStrategy: bs}}}); err == nil {
		t.Error("CreateIndividual: want error")
	}
	if _, err := c.Keywords.Update(ctx, 9, 1, KeywordChanges{BidStrategy: bs}); err == nil {
		t.Error("Update: want error")
	}
	if _, err := c.Keywords.ReviewHoldReasons(ctx, 9, 1); err == nil {
		t.Error("ReviewHoldReasons: want error")
	}
	if _, err := c.Keywords.QualityIndex(ctx, 9, 1); err == nil {
		t.Error("QualityIndex: want error")
	}
}
