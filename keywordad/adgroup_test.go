package keywordad

import (
	"net/http"
	"slices"
	"testing"
)

const (
	adGroupTestDay      = `["1","1","1","1","1","1","1","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","1","1"]`
	adGroupTestSchedule = `{"beginDate":"2021-01-01","endDate":"2021-02-20","detailTime":true,` +
		`"mondayTime":` + adGroupTestDay + `,"tuesdayTime":` + adGroupTestDay + `,"wednesdayTime":` + adGroupTestDay +
		`,"thursdayTime":` + adGroupTestDay + `,"fridayTime":` + adGroupTestDay + `,"saturdayTime":` + adGroupTestDay +
		`,"sundayTime":` + adGroupTestDay + `}`
	adGroupTestJSON = `{"campaignId":"3333333331","id":"4444444441","name":"광고그룹1","config":"ON",` +
		`"status":["OFF_BY_BIZ_CHANNEL_WAITING","WAITING"],` +
		`"placements":["PREMIUM_LINK_SEARCH_KAKAO","PREMIUM_LINK_CONTENT_KAKAO","PREMIUM_LINK_CONTENT_PARTNER"],` +
		`"deviceTypes":["PC","MOBILE"],"useKeywordExtension":true,"exceptionKeywords":["키워드1","키워드2"],` +
		`"dailyBudgetAmount":10000,"dailyBudgetAmountOver":false,"bidAmount":70,"mobileBidWeight":150,` +
		`"pcBidWeight":140,"keywordExtensionBidWeight":100,"contentBidAmount":70,"schedule":` + adGroupTestSchedule + `}`
)

func adGroupTestScheduleValue() Schedule {
	day := []string{"1", "1", "1", "1", "1", "1", "1", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "1", "1"}
	return Schedule{
		BeginDate: "2021-01-01", EndDate: "2021-02-20", DetailTime: true,
		MondayTime: day, TuesdayTime: day, WednesdayTime: day, ThursdayTime: day,
		FridayTime: day, SaturdayTime: day, SundayTime: day,
	}
}

func checkAdGroup(t *testing.T, g *AdGroup) {
	t.Helper()
	if g.ID != 4444444441 || g.CampaignID != 3333333331 || g.Name != "광고그룹1" || g.Config != ConfigOn ||
		!slices.Equal(g.Status, []Status{StatusOffByBizChannelWaiting, StatusWaiting}) ||
		!slices.Equal(g.Placements, []Placement{PlacementSearchKakao, PlacementContentKakao, PlacementContentPartner}) ||
		!slices.Equal(g.DeviceTypes, []DeviceType{DevicePC, DeviceMobile}) || !g.UseKeywordExtension ||
		!slices.Equal(g.ExceptionKeywords, []string{"키워드1", "키워드2"}) ||
		g.DailyBudgetAmount != 10000 || g.DailyBudgetAmountOver || g.BidAmount != 70 || g.MobileBidWeight != 150 ||
		g.PCBidWeight != 140 || g.KeywordExtensionBidWeight != 100 || g.ContentBidAmount != 70 {
		t.Errorf("ad group = %+v", g)
	}
	if s := g.Schedule; s.BeginDate != "2021-01-01" || s.EndDate != "2021-02-20" || !s.DetailTime ||
		len(s.MondayTime) != 24 || s.SundayTime[23] != "1" || s.WednesdayTime[7] != "0" {
		t.Errorf("schedule = %+v", s)
	}
}

func TestAdGroupService_List(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v1/adGroups", 200, `[`+adGroupTestJSON+`,{"campaignId":"3333333331","id":"4444444442","name":"광고그룹2","config":"OFF"}]`)

	gs, err := c.AdGroups.List(t.Context(), 1111, 3333333331, &AdGroupListOptions{Configs: []Config{ConfigOn, ConfigOff}})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodGet || req.Query.Get("campaignId") != "3333333331" || !slices.Equal(req.Query["config"], []string{"ON", "OFF"}) {
		t.Errorf("request = %s %v", req.Method, req.Query)
	}
	if len(gs) != 2 || gs[1].ID != 4444444442 || gs[1].Config != ConfigOff {
		t.Fatalf("ad groups = %+v", gs)
	}
	checkAdGroup(t, &gs[0])

	if _, err := c.AdGroups.List(t.Context(), 1111, 3333333331, nil); err != nil {
		t.Fatal(err)
	}
	if q := srv.LastRequest().Query; len(q) != 1 || q.Get("campaignId") != "3333333331" {
		t.Errorf("nil options query = %v", q)
	}

	if _, err := c.AdGroups.List(t.Context(), 1111, 0, nil); err == nil {
		t.Error("missing campaign ID: want error")
	}
}

func TestAdGroupService_Get(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v1/adGroups/4444444441", 200, adGroupTestJSON)

	g, err := c.AdGroups.Get(t.Context(), 1111, 4444444441)
	if err != nil {
		t.Fatal(err)
	}
	assertRequest(t, srv, "1111")
	checkAdGroup(t, g)
}

func TestAdGroupService_Create(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v1/adGroups", 200, adGroupTestJSON)

	useExt := true
	sch := adGroupTestScheduleValue()
	g, err := c.AdGroups.Create(t.Context(), 1111, NewAdGroup{
		CampaignID:          3333333331,
		Name:                "광고그룹1",
		Placements:          []Placement{PlacementSearchKakao, PlacementContentKakao, PlacementContentPartner},
		DeviceTypes:         []DeviceType{DevicePC, DeviceMobile},
		UseKeywordExtension: &useExt,
		ExceptionKeywords:   []string{"키워드1", "키워드2"},
		BidAmount:           70,
		DailyBudgetAmount:   100000,
		MobileBidWeight:     150,
		PCBidWeight:         140,
		ContentBidAmount:    70,
		Schedule:            &sch,
	})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	want := `{"campaignId":3333333331,"name":"광고그룹1",` +
		`"placements":["PREMIUM_LINK_SEARCH_KAKAO","PREMIUM_LINK_CONTENT_KAKAO","PREMIUM_LINK_CONTENT_PARTNER"],` +
		`"deviceTypes":["PC","MOBILE"],"useKeywordExtension":true,"exceptionKeywords":["키워드1","키워드2"],` +
		`"bidAmount":70,"dailyBudgetAmount":100000,"mobileBidWeight":150,"pcBidWeight":140,"contentBidAmount":70,` +
		`"schedule":` + adGroupTestSchedule + `}`
	if req.Method != http.MethodPost || string(req.Body) != want || req.Header.Get("Content-Type") != "application/json" {
		t.Errorf("body = %s", req.Body)
	}
	checkAdGroup(t, g)

	if _, err := c.AdGroups.Create(t.Context(), 1111, NewAdGroup{CampaignID: 1, Name: "x"}); err != nil {
		t.Fatal(err)
	}
	if body := string(srv.LastRequest().Body); body != `{"campaignId":1,"name":"x"}` {
		t.Errorf("minimal body = %s", body)
	}

	if _, err := c.AdGroups.Create(t.Context(), 1111, NewAdGroup{Name: "x"}); err == nil {
		t.Error("missing campaign ID: want error")
	}
	if _, err := c.AdGroups.Create(t.Context(), 1111, NewAdGroup{CampaignID: 1}); err == nil {
		t.Error("missing name: want error")
	}
}

func TestAdGroupService_Update(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PUT /openapi/v1/adGroups/4444444441", 200, adGroupTestJSON)
	srv.HandleJSON("GET /openapi/v1/adGroups/4444444441", 200, adGroupTestJSON)

	cur, err := c.AdGroups.Get(t.Context(), 1111, 4444444441)
	if err != nil {
		t.Fatal(err)
	}
	ch := cur.Changes()
	ch.BidAmount = 1000
	ch.DailyBudgetAmount = 2000
	g, err := c.AdGroups.Update(t.Context(), 1111, 4444444441, ch)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	want := `{"name":"광고그룹1",` +
		`"placements":["PREMIUM_LINK_SEARCH_KAKAO","PREMIUM_LINK_CONTENT_KAKAO","PREMIUM_LINK_CONTENT_PARTNER"],` +
		`"deviceTypes":["PC","MOBILE"],"useKeywordExtension":true,"exceptionKeywords":["키워드1","키워드2"],` +
		`"bidAmount":1000,"dailyBudgetAmount":2000,"mobileBidWeight":150,"pcBidWeight":140,` +
		`"keywordExtensionBidWeight":100,"contentBidAmount":70,"schedule":` + adGroupTestSchedule + `}`
	if req.Method != http.MethodPut || string(req.Body) != want {
		t.Errorf("body = %s", req.Body)
	}
	checkAdGroup(t, g)

	if ch := (&AdGroup{Name: "x"}).Changes(); ch.ContentBidAmount != nil {
		t.Errorf("unset content bid: changes = %+v", ch)
	}
	if _, err := c.AdGroups.Update(t.Context(), 1111, 4444444441, AdGroupChanges{}); err == nil {
		t.Error("missing name: want error")
	}
}

func TestAdGroupService_patches(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	for _, p := range []string{"onOff", "bidAmount", "dailyBudget", "bidWeight", "contentBidAmount"} {
		srv.HandleJSON("PATCH /openapi/v1/adGroups/4444444441/"+p, 200, ``)
	}
	srv.HandleJSON("DELETE /openapi/v1/adGroups/4444444441", 200, ``)
	ctx := t.Context()

	tests := []struct {
		name   string
		call   func() error
		method string
		path   string
		body   string
	}{
		{
			"SetConfig", func() error { return c.AdGroups.SetConfig(ctx, 1111, 4444444441, ConfigOff) },
			http.MethodPatch, "/onOff", `{"config":"OFF"}`,
		},
		{
			"UpdateBidAmount", func() error { return c.AdGroups.UpdateBidAmount(ctx, 1111, 4444444441, 1000) },
			http.MethodPatch, "/bidAmount", `{"bidAmount":1000}`,
		},
		{
			"UpdateDailyBudget", func() error { return c.AdGroups.UpdateDailyBudget(ctx, 1111, 4444444441, 2000) },
			http.MethodPatch, "/dailyBudget", `{"dailyBudgetAmount":2000}`,
		},
		{"UpdateBidWeights", func() error {
			return c.AdGroups.UpdateBidWeights(ctx, 1111, 4444444441, BidWeights{Mobile: 100, PC: 110, KeywordExtension: 120})
		}, http.MethodPatch, "/bidWeight", `{"mobileBidWeight":100,"pcBidWeight":110,"keywordExtensionBidWeight":120}`},
		{
			"UpdateContentBidAmount", func() error { return c.AdGroups.UpdateContentBidAmount(ctx, 1111, 4444444441, 1000) },
			http.MethodPatch, "/contentBidAmount", `{"contentBidAmount":1000}`,
		},
		{
			"UpdateContentBidAmount unset", func() error { return c.AdGroups.UpdateContentBidAmount(ctx, 1111, 4444444441, 0) },
			http.MethodPatch, "/contentBidAmount", `{"contentBidAmount":null}`,
		},
		{
			"Delete", func() error { return c.AdGroups.Delete(ctx, 1111, 4444444441) },
			http.MethodDelete, "", ``,
		},
	}
	for _, tt := range tests {
		if err := tt.call(); err != nil {
			t.Fatalf("%s: %v", tt.name, err)
		}
		req := assertRequest(t, srv, "1111")
		if req.Method != tt.method || req.Path != "/openapi/v1/adGroups/4444444441"+tt.path || string(req.Body) != tt.body {
			t.Errorf("%s: request = %s %s %s", tt.name, req.Method, req.Path, req.Body)
		}
	}
}

func TestAdGroupService_apiErrors(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	const errBody = `{"code":-813,"msg":"KakaoKeywordAdException","extras":{"detailCode":31001,"detailMsg":"광고그룹이 존재하지 않습니다."}}`
	srv.HandleJSON("/openapi/v1/adGroups", 400, errBody)
	srv.HandleJSON("/openapi/v1/adGroups/1", 400, errBody)
	ctx := t.Context()

	if _, err := c.AdGroups.List(ctx, 9, 1, nil); err == nil {
		t.Error("List: want error")
	}
	if _, err := c.AdGroups.Get(ctx, 9, 1); err == nil {
		t.Error("Get: want error")
	}
	if _, err := c.AdGroups.Create(ctx, 9, NewAdGroup{CampaignID: 1, Name: "x"}); err == nil {
		t.Error("Create: want error")
	}
	if _, err := c.AdGroups.Update(ctx, 9, 1, AdGroupChanges{Name: "x"}); err == nil {
		t.Error("Update: want error")
	}
}
