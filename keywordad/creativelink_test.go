package keywordad

import (
	"net/http"
	"slices"
	"testing"
)

func TestCreativeLinkService_lists(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v1/creativeLinks", 200, `[
  {"creativeLinkId": "7777777771", "adGroupId": "4444444441", "creativeId": "66666666661", "config": "ON", "status": ["OFF_BY_BIZ_CHANNEL_WAITING"]},
  {"creativeLinkId": "7777777772", "adGroupId": "4444444441", "creativeId": "66666666662", "config": "ON", "status": ["OFF_BY_BIZ_CHANNEL_WAITING"]}
]`)
	ctx := t.Context()

	ls, err := c.CreativeLinks.ListByAdGroup(ctx, 1111111111, 4444444441, &CreativeLinkListOptions{Configs: []Config{ConfigOn, ConfigOff}})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111111111")
	if req.Query.Get("adGroupId") != "4444444441" || !slices.Equal(req.Query["config"], []string{"ON", "OFF"}) || req.Query.Has("creativeId") {
		t.Errorf("query = %v", req.Query)
	}
	if len(ls) != 2 || ls[1].CreativeLinkID != 7777777772 || ls[1].CreativeID != 66666666662 || ls[1].AdGroupID != 4444444441 ||
		ls[1].Config != ConfigOn || !slices.Equal(ls[1].Status, []Status{StatusOffByBizChannelWaiting}) {
		t.Errorf("links = %+v", ls)
	}

	if _, err := c.CreativeLinks.ListByCreative(ctx, 1111111111, 66666666661, nil); err != nil {
		t.Fatal(err)
	}
	req = assertRequest(t, srv, "1111111111")
	if req.Query.Get("creativeId") != "66666666661" || len(req.Query) != 1 {
		t.Errorf("query = %v", req.Query)
	}

	if _, err := c.CreativeLinks.ListByAdGroup(ctx, 1, 0, nil); err == nil {
		t.Error("missing ad group: want error")
	}
	if _, err := c.CreativeLinks.ListByCreative(ctx, 1, 0, nil); err == nil {
		t.Error("missing creative: want error")
	}
}

func TestCreativeLinkService_Get(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v1/creativeLinks/7777777771", 200, `{"creativeLinkId": "7777777771", "adGroupId": "4444444441", "creativeId": "66666666661", "config": "ON", "status": ["OFF_BY_BIZ_CHANNEL_WAITING"]}`)

	l, err := c.CreativeLinks.Get(t.Context(), 1111111111, 7777777771)
	if err != nil {
		t.Fatal(err)
	}
	assertRequest(t, srv, "1111111111")
	if l.CreativeLinkID != 7777777771 || l.AdGroupID != 4444444441 || l.CreativeID != 66666666661 || l.Config != ConfigOn || len(l.Status) != 1 {
		t.Errorf("link = %+v", l)
	}
}

func TestCreativeLinkService_writes(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v1/creativeLinks", 200, `{"creativeLinkId": "7777777771", "adGroupId": "4444444441", "creativeId": "66666666661", "config": "ON"}`)
	srv.HandleJSON("PATCH /openapi/v1/creativeLinks/7777777771/onOff", 200, ``)
	srv.HandleJSON("DELETE /openapi/v1/creativeLinks/7777777771", 200, ``)
	ctx := t.Context()

	l, err := c.CreativeLinks.Link(ctx, 1111111111, 66666666661, 4444444441)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111111111")
	if string(req.Body) != `{"creativeId":66666666661,"adGroupId":4444444441}` {
		t.Errorf("body = %s", req.Body)
	}
	if l.CreativeLinkID != 7777777771 || l.Config != ConfigOn {
		t.Errorf("link = %+v", l)
	}

	if err := c.CreativeLinks.SetConfig(ctx, 1111111111, 7777777771, ConfigOff); err != nil {
		t.Fatal(err)
	}
	req = assertRequest(t, srv, "1111111111")
	if req.Method != http.MethodPatch || string(req.Body) != `{"config":"OFF"}` {
		t.Errorf("SetConfig %s body = %s", req.Method, req.Body)
	}

	if err := c.CreativeLinks.Unlink(ctx, 1111111111, 7777777771); err != nil {
		t.Fatal(err)
	}
	if req := assertRequest(t, srv, "1111111111"); req.Method != http.MethodDelete {
		t.Errorf("Unlink method = %s", req.Method)
	}

	if _, err := c.CreativeLinks.Link(ctx, 1, 0, 4444444441); err == nil {
		t.Error("missing creative: want error")
	}
	if _, err := c.CreativeLinks.Link(ctx, 1, 66666666661, 0); err == nil {
		t.Error("missing ad group: want error")
	}
}
