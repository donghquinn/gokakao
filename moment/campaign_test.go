package moment

import (
	"errors"
	"net/http"
	"slices"
	"testing"
	"time"

	kakao "github.com/donghquinn/gokakao"
)

const campaignDetailJSON = `{
  "id": 1234,
  "name": "캠페인1",
  "campaignTypeGoal": {"campaignType": "DISPLAY", "goal": "VISITING"},
  "objective": null,
  "dailyBudgetAmount": null,
  "config": "ON",
  "statusDescription": "운영중",
  "trackId": "5360727718927577526",
  "adAccountId": 10000,
  "status": ["LIVE"],
  "systemConfig": "ON",
  "isDailyBudgetAmountOver": false
}`

func campaignAssertDetail(t *testing.T, cp *CampaignDetail) {
	t.Helper()
	if cp.ID != 1234 || cp.Name != "캠페인1" ||
		cp.CampaignTypeGoal != (CampaignTypeGoal{CampaignType: CampaignTypeDisplay, Goal: GoalVisiting}) ||
		cp.Objective != nil || cp.DailyBudgetAmount != 0 || cp.Config != ConfigOn || cp.StatusDescription != "운영중" ||
		cp.TrackID != "5360727718927577526" || cp.AdAccountID != 10000 ||
		!slices.Equal(cp.Status, []CampaignStatus{CampaignStatusLive}) || cp.SystemConfig != SystemConfigOn ||
		cp.DailyBudgetAmountOver || cp.Kclid {
		t.Errorf("campaign = %+v", cp)
	}
}

func TestCampaignService_List(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/campaigns", 200, `{
  "content": [
    {"id": 1111, "name": "캠페인 1", "config": "ON", "userConfig": "ON", "systemConfig": "ADMIN_STOP"},
    {"id": 1112, "name": "캠페인 2", "config": "OFF", "userConfig": "OFF", "systemConfig": " ON"}
  ]
}`)

	cs, err := c.Campaigns.List(t.Context(), 10000, &CampaignListOptions{Config: ConfigOn})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "10000")
	if req.Method != http.MethodGet || req.Query.Get("config") != "ON" {
		t.Errorf("request = %s %s?%v", req.Method, req.Path, req.Query)
	}
	if len(cs) != 2 || cs[0] != (CampaignSummary{ID: 1111, Name: "캠페인 1", Config: ConfigOn, UserConfig: ConfigOn, SystemConfig: SystemConfigAdminStop}) ||
		cs[1].ID != 1112 || cs[1].Config != ConfigOff {
		t.Errorf("campaigns = %+v", cs)
	}

	if _, err := c.Campaigns.List(t.Context(), 10000, nil); err != nil {
		t.Fatal(err)
	}
	if q := srv.LastRequest().Query; len(q) != 0 {
		t.Errorf("nil options query = %v", q)
	}
	if _, err := c.Campaigns.List(t.Context(), 0, nil); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestCampaignService_Get(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/campaigns/1234", 200, campaignDetailJSON)

	cp, err := c.Campaigns.Get(t.Context(), 10000, 1234)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "10000")
	if req.Method != http.MethodGet {
		t.Errorf("method = %s", req.Method)
	}
	campaignAssertDetail(t, cp)

	if _, err := c.Campaigns.Get(t.Context(), 10000, 0); err == nil {
		t.Error("missing campaign: want error")
	}
}

func TestCampaignService_Create(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v4/campaigns", 200, campaignDetailJSON)

	cp, err := c.Campaigns.Create(t.Context(), 10000, NewCampaign{
		Name:              "첫번째_캠페인",
		CampaignTypeGoal:  CampaignTypeGoal{CampaignType: CampaignTypeDisplay, Goal: GoalConversion},
		Objective:         &Objective{Type: ObjectiveTalkChannel, Value: "abcd1234"},
		DailyBudgetAmount: 200000,
	})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "10000")
	want := `{"name":"첫번째_캠페인","campaignTypeGoal":{"campaignType":"DISPLAY","goal":"CONVERSION"},"objective":{"type":"TALK_CHANNEL","value":"abcd1234"},"dailyBudgetAmount":200000}`
	if req.Method != http.MethodPost || string(req.Body) != want {
		t.Errorf("request = %s %s", req.Method, req.Body)
	}
	campaignAssertDetail(t, cp)

	kclid := true
	if _, err := c.Campaigns.Create(t.Context(), 1, NewCampaign{CampaignTypeGoal: CampaignTypeGoal{CampaignType: CampaignTypeVideo, Goal: GoalView}, TrackID: "1", Kclid: &kclid}); err != nil {
		t.Fatal(err)
	}
	if got := string(srv.LastRequest().Body); got != `{"campaignTypeGoal":{"campaignType":"VIDEO","goal":"VIEW"},"trackId":"1","kclid":true}` {
		t.Errorf("body = %s", got)
	}

	if _, err := c.Campaigns.Create(t.Context(), 1, NewCampaign{CampaignTypeGoal: CampaignTypeGoal{CampaignType: CampaignTypeVideo}}); err == nil {
		t.Error("missing goal: want error")
	}
	if _, err := c.Campaigns.Create(t.Context(), 0, NewCampaign{}); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestCampaignService_Update(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PUT /openapi/v4/campaigns", 200, campaignDetailJSON)

	budget := int64(100000)
	cp, err := c.Campaigns.Update(t.Context(), 10000, 5678, CampaignChanges{Name: "첫번째_캠페인_수정", DailyBudgetAmount: &budget})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "10000")
	if req.Method != http.MethodPut || string(req.Body) != `{"id":5678,"name":"첫번째_캠페인_수정","dailyBudgetAmount":100000}` {
		t.Errorf("request = %s %s", req.Method, req.Body)
	}
	campaignAssertDetail(t, cp)

	// Changes keeps the current values.
	if _, err := c.Campaigns.Update(t.Context(), 10000, 1234, cp.Changes()); err != nil {
		t.Fatal(err)
	}
	if got := string(srv.LastRequest().Body); got != `{"id":1234,"name":"캠페인1","trackId":"5360727718927577526","kclid":false}` {
		t.Errorf("Changes body = %s", got)
	}
	ch := (&CampaignDetail{Name: "n", DailyBudgetAmount: 50000, Kclid: true}).Changes()
	if ch.TrackID != nil || ch.DailyBudgetAmount == nil || *ch.DailyBudgetAmount != 50000 || ch.Kclid == nil || !*ch.Kclid {
		t.Errorf("Changes = %+v", ch)
	}

	if _, err := c.Campaigns.Update(t.Context(), 1, 0, CampaignChanges{}); err == nil {
		t.Error("missing campaign: want error")
	}
}

func TestCampaignService_SetConfig(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PUT /openapi/v4/campaigns/onOff", 200, ``)

	if err := c.Campaigns.SetConfig(t.Context(), 10000, 5678, ConfigOn); err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "10000")
	if req.Method != http.MethodPut || string(req.Body) != `{"id":5678,"config":"ON"}` {
		t.Errorf("request = %s %s", req.Method, req.Body)
	}

	if err := c.Campaigns.SetConfig(t.Context(), 1, 2, ConfigDel); err == nil {
		t.Error("DEL: want error")
	}
	if err := c.Campaigns.SetConfig(t.Context(), 0, 2, ConfigOff); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestCampaignService_UpdateDailyBudget(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PUT /openapi/v4/campaigns/dailyBudgetAmount", 200, ``)

	if err := c.Campaigns.UpdateDailyBudget(t.Context(), 10000, 5678, 5000000); err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "10000")
	if req.Method != http.MethodPut || string(req.Body) != `{"id":5678,"dailyBudgetAmount":5000000}` {
		t.Errorf("request = %s %s", req.Method, req.Body)
	}

	if err := c.Campaigns.UpdateDailyBudget(t.Context(), 10000, 5678, 0); err != nil {
		t.Fatal(err)
	}
	if got := string(srv.LastRequest().Body); got != `{"id":5678,"dailyBudgetAmount":null}` {
		t.Errorf("unset body = %s", got)
	}

	if err := c.Campaigns.UpdateDailyBudget(t.Context(), 1, 2, -1); err == nil {
		t.Error("negative: want error")
	}
	if err := c.Campaigns.UpdateDailyBudget(t.Context(), 1, 0, 1); err == nil {
		t.Error("missing campaign: want error")
	}
}

func TestCampaignService_Delete(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("DELETE /openapi/v4/campaigns/5678", 200, ``)

	if err := c.Campaigns.Delete(t.Context(), 10000, 5678); err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "10000")
	if req.Method != http.MethodDelete || req.Path != "/openapi/v4/campaigns/5678" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}

	if err := c.Campaigns.Delete(t.Context(), 0, 1); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestCampaignService_LatestSystemStopReason(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/campaigns/1234/latestSystemConfigHistory", 200, `{
  "id": 1234,
  "systemConfig": "EXTERNAL_SERVICE_STOP",
  "reason": "채널 파트너센터에서 규제됨",
  "detailReason": "메시지 집행 가이드, 운영정책 위반으로 yyyy년 mm월 dd일까지 메시지 발송이 불가",
  "createdDate": "2021-01-01T00:00:00",
  "lastModifiedDate": "2021-01-01T00:00:00"
}`)

	r, err := c.Campaigns.LatestSystemStopReason(t.Context(), 10000, 1234)
	if err != nil {
		t.Fatal(err)
	}
	assertRequest(t, srv, "10000")
	want := time.Date(2021, 1, 1, 0, 0, 0, 0, time.FixedZone("KST", 9*60*60))
	if r.ID != 1234 || r.SystemConfig != SystemConfigExternalServiceStop || r.Reason != "채널 파트너센터에서 규제됨" ||
		r.DetailReason == "" || !r.CreatedDate.Equal(want) || !r.LastModifiedDate.Equal(want) {
		t.Errorf("reason = %+v", r)
	}

	if _, err := c.Campaigns.LatestSystemStopReason(t.Context(), 10000, 0); err == nil {
		t.Error("missing campaign: want error")
	}
}

func TestCampaignService_SystemStopReasons(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/campaigns/1234/systemConfigHistories", 200, `[
  {
    "id": 1235,
    "systemConfig": "ADMIN_STOP",
    "reason": "해당 캠페인유형은 지원 종료되었습니다.",
    "createdDate": "2021-01-01T00:00:00",
    "lastModifiedDate": "2021-01-01T00:00:00"
  },
  {
    "id": 1234,
    "systemConfig": "EXTERNAL_SERVICE_STOP",
    "reason": "채널 파트너센터에서 규제됨",
    "detailReason": "메시지 집행 가이드, 운영정책 위반으로 yyyy년 mm월 dd일까지 메시지 발송이 불가",
    "createdDate": "2021-01-01T00:00:00",
    "lastModifiedDate": "2021-01-01T00:00:00"
  }
]`)

	rs, err := c.Campaigns.SystemStopReasons(t.Context(), 10000, 1234)
	if err != nil {
		t.Fatal(err)
	}
	assertRequest(t, srv, "10000")
	if len(rs) != 2 || rs[0].ID != 1235 || rs[0].SystemConfig != SystemConfigAdminStop || rs[0].DetailReason != "" ||
		rs[1].SystemConfig != SystemConfigExternalServiceStop || rs[1].CreatedDate.Year() != 2021 {
		t.Errorf("reasons = %+v", rs)
	}

	if _, err := c.Campaigns.SystemStopReasons(t.Context(), 0, 1234); err == nil {
		t.Error("missing ad account: want error")
	}
}

// TestCampaignService_apiErrors checks that every campaign method returns
// API errors.
func TestCampaignService_apiErrors(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("/", 400, `{"code":-813,"msg":"KakaoMomentException","extras":{"detailCode":31001,"detailMsg":"캠페인이 존재하지 않습니다."}}`)
	ctx := t.Context()
	nc := NewCampaign{CampaignTypeGoal: CampaignTypeGoal{CampaignType: CampaignTypeDisplay, Goal: GoalVisiting}}
	calls := map[string]func() error{
		"List":                   func() error { _, err := c.Campaigns.List(ctx, 1, nil); return err },
		"Get":                    func() error { _, err := c.Campaigns.Get(ctx, 1, 2); return err },
		"Create":                 func() error { _, err := c.Campaigns.Create(ctx, 1, nc); return err },
		"Update":                 func() error { _, err := c.Campaigns.Update(ctx, 1, 2, CampaignChanges{}); return err },
		"LatestSystemStopReason": func() error { _, err := c.Campaigns.LatestSystemStopReason(ctx, 1, 2); return err },
		"SystemStopReasons":      func() error { _, err := c.Campaigns.SystemStopReasons(ctx, 1, 2); return err },
	}
	for name, f := range calls {
		var e *kakao.Error
		if err := f(); !errors.As(err, &e) || e.Code != -813 {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}
