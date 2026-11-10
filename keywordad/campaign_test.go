package keywordad

import (
	"net/http"
	"slices"
	"testing"
)

const campaignJSON = `{"bizChannelId":"2222222221","id":"3333333331","name":"캠페인1","config":"ON","status":["OFF_BY_BIZ_CHANNEL_WAITING"],"trackId":null,"traceUrl":true,"dailyBudgetAmount":500000,"dailyBudgetAmountOver":false}`

func TestCampaignService_List(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v1/campaigns", 200, `[`+campaignJSON+`,{"bizChannelId":"2222222221","id":"3333333332","name":"캠페인2","config":"ON","status":["OFF_BY_BIZ_CHANNEL_WAITING"],"trackId":null,"traceUrl":true,"dailyBudgetAmount":500000,"dailyBudgetAmountOver":false}]`)

	cs, err := c.Campaigns.List(t.Context(), 1111111111, &CampaignListOptions{Configs: []Config{ConfigOn}})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111111111")
	if req.Method != http.MethodGet || req.Path != "/openapi/v1/campaigns" || !slices.Equal(req.Query["config"], []string{"ON"}) {
		t.Errorf("request = %s %s?%v", req.Method, req.Path, req.Query)
	}
	if len(cs) != 2 || cs[1].ID != 3333333332 || cs[1].Name != "캠페인2" || cs[1].BizChannelID != 2222222221 {
		t.Errorf("campaigns = %+v", cs)
	}

	if _, err := c.Campaigns.List(t.Context(), 1111111111, nil); err != nil {
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
	srv.HandleJSON("GET /openapi/v1/campaigns/3333333331", 200, campaignJSON)

	cp, err := c.Campaigns.Get(t.Context(), 1111111111, 3333333331)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111111111")
	if req.Method != http.MethodGet || req.Path != "/openapi/v1/campaigns/3333333331" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	if cp.ID != 3333333331 || cp.BizChannelID != 2222222221 || cp.Name != "캠페인1" || cp.Config != ConfigOn ||
		!slices.Equal(cp.Status, []Status{StatusOffByBizChannelWaiting}) || cp.TrackID != "" || !cp.TraceURL ||
		cp.DailyBudgetAmount != 500000 || cp.DailyBudgetAmountOver {
		t.Errorf("campaign = %+v", cp)
	}

	if _, err := c.Campaigns.Get(t.Context(), 1111111111, 0); err == nil {
		t.Error("missing campaign: want error")
	}
}

func TestCampaignService_Create(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v1/campaigns", 200, `{"bizChannelId":"2222222221","id":"3333333331","name":"캠페인1","config":"ON","status":["OFF_BY_BIZ_CHANNEL_WAITING"],"trackId":null,"traceUrl":true,"dailyBudgetAmount":30000,"dailyBudgetAmountOver":false}`)

	traceURL := true
	cp, err := c.Campaigns.Create(t.Context(), 1111111111, NewCampaign{BizChannelID: 2222222221, Name: "캠페인1", TraceURL: &traceURL, DailyBudgetAmount: 30000})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111111111")
	if req.Method != http.MethodPost || req.Path != "/openapi/v1/campaigns" ||
		string(req.Body) != `{"bizChannelId":2222222221,"name":"캠페인1","traceUrl":true,"dailyBudgetAmount":30000}` {
		t.Errorf("request = %s %s %s", req.Method, req.Path, req.Body)
	}
	if cp.ID != 3333333331 || cp.DailyBudgetAmount != 30000 {
		t.Errorf("campaign = %+v", cp)
	}

	if _, err := c.Campaigns.Create(t.Context(), 1, NewCampaign{BizChannelID: 2, Name: "c", TrackID: "1234567890"}); err != nil {
		t.Fatal(err)
	}
	if got := string(srv.LastRequest().Body); got != `{"bizChannelId":2,"name":"c","trackId":"1234567890"}` {
		t.Errorf("body = %s", got)
	}

	for _, nc := range []NewCampaign{{Name: "c"}, {BizChannelID: 2}} {
		if _, err := c.Campaigns.Create(t.Context(), 1, nc); err == nil {
			t.Errorf("Create(%+v): want error", nc)
		}
	}
	if _, err := c.Campaigns.Create(t.Context(), 0, NewCampaign{BizChannelID: 2, Name: "c"}); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestCampaignService_Update(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PUT /openapi/v1/campaigns/3333333331", 200, `{"bizChannelId":"2222222221","id":"3333333331","name":"캠페인2","config":"ON","status":["OFF_BY_BIZ_CHANNEL_WAITING"],"trackId":"1234567890","traceUrl":false,"dailyBudgetAmount":30000,"dailyBudgetAmountOver":false}`)

	trackID, traceURL, budget := "1234567890", false, int64(30000)
	cp, err := c.Campaigns.Update(t.Context(), 1111111111, 3333333331, CampaignChanges{Name: "캠페인2", TrackID: &trackID, TraceURL: &traceURL, DailyBudgetAmount: &budget})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111111111")
	if req.Method != http.MethodPut || req.Path != "/openapi/v1/campaigns/3333333331" ||
		string(req.Body) != `{"name":"캠페인2","trackId":"1234567890","traceUrl":false,"dailyBudgetAmount":30000}` {
		t.Errorf("request = %s %s %s", req.Method, req.Path, req.Body)
	}
	if cp.Name != "캠페인2" || cp.TrackID != "1234567890" || cp.TraceURL {
		t.Errorf("campaign = %+v", cp)
	}

	// Changes keeps the current values.
	if _, err := c.Campaigns.Update(t.Context(), 1111111111, 3333333331, cp.Changes()); err != nil {
		t.Fatal(err)
	}
	if got := string(srv.LastRequest().Body); got != `{"name":"캠페인2","trackId":"1234567890","traceUrl":false,"dailyBudgetAmount":30000}` {
		t.Errorf("Changes body = %s", got)
	}
	unset := (&Campaign{Name: "n", TraceURL: true}).Changes()
	if unset.TrackID != nil || unset.DailyBudgetAmount != nil || unset.TraceURL == nil || !*unset.TraceURL {
		t.Errorf("Changes of unset campaign = %+v", unset)
	}

	if _, err := c.Campaigns.Update(t.Context(), 1, 2, CampaignChanges{}); err == nil {
		t.Error("missing name: want error")
	}
	if _, err := c.Campaigns.Update(t.Context(), 1, 0, CampaignChanges{Name: "n"}); err == nil {
		t.Error("missing campaign: want error")
	}
}

func TestCampaignService_SetConfig(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PATCH /openapi/v1/campaigns/3333333331/onOff", 200, ``)

	if err := c.Campaigns.SetConfig(t.Context(), 1111111111, 3333333331, ConfigOn); err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111111111")
	if req.Method != http.MethodPatch || string(req.Body) != `{"id":3333333331,"config":"ON"}` {
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
	srv.HandleJSON("PATCH /openapi/v1/campaigns/3333333331/dailyBudget", 200, ``)

	if err := c.Campaigns.UpdateDailyBudget(t.Context(), 1111111111, 3333333331, 20000); err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111111111")
	if req.Method != http.MethodPatch || string(req.Body) != `{"dailyBudgetAmount":20000}` {
		t.Errorf("request = %s %s", req.Method, req.Body)
	}

	if err := c.Campaigns.UpdateDailyBudget(t.Context(), 1111111111, 3333333331, 0); err != nil {
		t.Fatal(err)
	}
	if got := string(srv.LastRequest().Body); got != `{"dailyBudgetAmount":null}` {
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
	srv.HandleJSON("DELETE /openapi/v1/campaigns/3333333331", 200, ``)

	if err := c.Campaigns.Delete(t.Context(), 1111111111, 3333333331); err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111111111")
	if req.Method != http.MethodDelete || req.Path != "/openapi/v1/campaigns/3333333331" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}

	if err := c.Campaigns.Delete(t.Context(), 0, 1); err == nil {
		t.Error("missing ad account: want error")
	}
}
