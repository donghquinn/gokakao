package keywordad

import (
	"errors"
	"testing"

	kakao "github.com/donghquinn/gokakao"
)

// TestServices_apiErrors checks that the creative, creative link, image,
// pixel and report methods return API errors.
func TestServices_apiErrors(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("/", 400, `{"code":-813,"msg":"KakaoKeywordAdException"}`)
	ctx := t.Context()
	q := ReportQuery{MetricsGroups: []MetricsGroup{MetricsBasic}}
	nc := NewCreative{BizChannelID: 2, Name: "n", Title: "t", Description: "d"}
	calls := map[string]func() error{
		"Creatives.List":              func() error { _, err := c.Creatives.List(ctx, 1, 2, nil); return err },
		"Creatives.Get":               func() error { _, err := c.Creatives.Get(ctx, 1, 2); return err },
		"Creatives.Create":            func() error { _, err := c.Creatives.Create(ctx, 1, nc); return err },
		"Creatives.Delete":            func() error { return c.Creatives.Delete(ctx, 1, 2) },
		"CreativeLinks.ListByAdGroup": func() error { _, err := c.CreativeLinks.ListByAdGroup(ctx, 1, 2, nil); return err },
		"CreativeLinks.Get":           func() error { _, err := c.CreativeLinks.Get(ctx, 1, 2); return err },
		"CreativeLinks.Link":          func() error { _, err := c.CreativeLinks.Link(ctx, 1, 2, 3); return err },
		"Images.List":                 func() error { _, err := c.Images.List(ctx, 1, nil); return err },
		"Images.Get":                  func() error { _, err := c.Images.Get(ctx, 1, 2); return err },
		"Images.Upload":               func() error { _, err := c.Images.Upload(ctx, 1, "a.png", []byte("x")); return err },
		"Images.UploadURL":            func() error { _, err := c.Images.UploadURL(ctx, 1, "https://x"); return err },
		"Pixels.List":                 func() error { _, err := c.Pixels.List(ctx, 1); return err },
		"Pixels.Available":            func() error { _, err := c.Pixels.Available(ctx, 1); return err },
		"Reports.AdAccount":           func() error { _, err := c.Reports.AdAccount(ctx, 1, q); return err },
	}
	for name, f := range calls {
		var e *kakao.Error
		if err := f(); !errors.As(err, &e) || e.Code != -813 {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}

const fullMetricsJSON = `{
  "imp": 105, "click": 10, "spending": 700.0, "ctr": 10.5, "rimp": 5000, "ppc": 70.0, "rank": 3,
  "convCmptReg1d": 1, "convCmptReg7d": 2, "convViewCart1d": 3, "convViewCart7d": 4,
  "convPurchase1d": 5, "convPurchase7d": 6, "convPurchaseP1d": 7, "convPurchaseP7d": 8,
  "convParticipation1d": 9, "convParticipation7d": 10, "convSignup1d": 11, "convSignup7d": 12,
  "convAppInstall1d": 13, "convAppInstall7d": 14
}`

func assertFullMetrics(t *testing.T, m ReportMetrics) {
	t.Helper()
	if m.BasicMetrics != (BasicMetrics{Imp: 105, Click: 10, CTR: 10.5, Spending: 700}) ||
		m.AdditionMetrics != (AdditionMetrics{RImp: 5000, PPC: 70, Rank: 3}) ||
		m.ConvCmptReg1d != 1 || m.ConvPurchaseP7d != 8 || m.ConvAppInstall7d != 14 {
		t.Errorf("metrics = %+v", m)
	}
}

func TestReportService_AdAccount(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v1/adAccounts/report", 200, `{"data": [{"start": "20210101", "end": "20210131", "dimensions": {"adAccountId": "1111111111"}, "metrics": {"imp": 105, "click": 10, "spending": 700.0, "ctr": 10.5}}]}`)

	rows, err := c.Reports.AdAccount(t.Context(), 1111111111, ReportQuery{
		MetricsGroups: []MetricsGroup{MetricsBasic},
		Start:         "20210101",
		End:           "20210131",
		AdGroupID:     4444444441, // ignored
	})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111111111")
	if q := req.Query; len(q) != 3 || q.Get("metricsGroups") != "BASIC" || q.Get("start") != "20210101" || q.Get("end") != "20210131" {
		t.Errorf("query = %v", q)
	}
	if len(rows) != 1 || rows[0].Start != "20210101" || rows[0].End != "20210131" || rows[0].Dimensions.AdAccountID != 1111111111 ||
		rows[0].Metrics.BasicMetrics != (BasicMetrics{Imp: 105, Click: 10, CTR: 10.5, Spending: 700}) {
		t.Errorf("rows = %+v", rows)
	}
}

func TestReportService_Campaigns(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v1/campaigns/report", 200, `{"data": [{"start": "20210101", "end": "20210131",
  "dimensions": {"adAccountId": "1111111111", "campaignId": "3333333331", "product": "PREMIUM_LINK"}, "metrics": `+fullMetricsJSON+`}]}`)

	rows, err := c.Reports.Campaigns(t.Context(), 1111111111, ReportQuery{
		MetricsGroups: []MetricsGroup{MetricsBasic, MetricsAddition, MetricsPixelSDKConversion},
		DatePreset:    DateLast7Days,
		Dimension:     DimensionDevice,
		TimeUnit:      TimeUnitDay,
	})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111111111")
	if q := req.Query; q.Get("metricsGroups") != "BASIC,ADDITION,PIXEL_SDK_CONVERSION" || q.Get("datePreset") != "LAST_7DAY" ||
		q.Get("dimension") != "DEVICE" || q.Get("timeUnit") != "DAY" || q.Has("start") || q.Has("campaignId") {
		t.Errorf("query = %v", q)
	}
	if len(rows) != 1 || rows[0].Dimensions.CampaignID != 3333333331 || rows[0].Dimensions.Product != ProductPremiumLink {
		t.Fatalf("rows = %+v", rows)
	}
	assertFullMetrics(t, rows[0].Metrics)

	if _, err := c.Reports.Campaigns(t.Context(), 1, ReportQuery{}); err == nil {
		t.Error("missing metrics groups: want error")
	}
}

func TestReportService_AdGroups(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v1/adGroups/report", 200, `{"data": [{"start": "20210101", "end": "20210131",
  "dimensions": {"adAccountId": "1111111111", "campaignId": "3333333331", "adGroupId": "4444444441", "product": "PREMIUM_LINK"}, "metrics": `+fullMetricsJSON+`}]}`)

	rows, err := c.Reports.AdGroups(t.Context(), 1111111111, 3333333331, ReportQuery{
		MetricsGroups: []MetricsGroup{MetricsBasic, MetricsAddition, MetricsPixelSDKConversion},
		Start:         "20210101", End: "20210131",
		AdGroupID: 4444444441, // ignored
	})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111111111")
	if q := req.Query; q.Get("campaignId") != "3333333331" || q.Has("adGroupId") || q.Get("metricsGroups") != "BASIC,ADDITION,PIXEL_SDK_CONVERSION" {
		t.Errorf("query = %v", q)
	}
	if len(rows) != 1 || rows[0].Dimensions.AdGroupID != 4444444441 {
		t.Fatalf("rows = %+v", rows)
	}
	assertFullMetrics(t, rows[0].Metrics)

	if _, err := c.Reports.AdGroups(t.Context(), 1, 0, ReportQuery{MetricsGroups: []MetricsGroup{MetricsBasic}}); err == nil {
		t.Error("missing campaign: want error")
	}
}

func TestReportService_Keywords(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v1/keywords/report", 200, `{"data": [{"start": "2021-01-01", "end": "2021-01-31",
  "dimensions": {"adAccountId": "1111111111", "campaignId": "3333333331", "adGroupId": "4444444441", "keywordId": "55555555551", "product": "PREMIUM_LINK"}, "metrics": `+fullMetricsJSON+`}]}`)

	rows, err := c.Reports.Keywords(t.Context(), 1111111111, 3333333331, ReportQuery{
		MetricsGroups: []MetricsGroup{MetricsBasic, MetricsAddition, MetricsPixelSDKConversion},
		Start:         "20210101", End: "20210101",
		AdGroupID: 4444444441,
	})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111111111")
	if q := req.Query; q.Get("campaignId") != "3333333331" || q.Get("adGroupId") != "4444444441" {
		t.Errorf("query = %v", q)
	}
	if len(rows) != 1 || rows[0].Start != "2021-01-01" || rows[0].Dimensions.KeywordID != 55555555551 {
		t.Fatalf("rows = %+v", rows)
	}
	assertFullMetrics(t, rows[0].Metrics)

	if _, err := c.Reports.Keywords(t.Context(), 1, 0, ReportQuery{MetricsGroups: []MetricsGroup{MetricsBasic}}); err == nil {
		t.Error("missing campaign: want error")
	}
}

func TestReportService_Creatives(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v1/creatives/report", 200, `{"data": [{"start": "2021-01-01", "end": "2021-01-31",
  "dimensions": {"adAccountId": "1111111111", "campaignId": "3333333331", "adGroupId": "4444444441", "creativeLinkId": "7777777771", "product": "TALK_CHANNEL", "assetType": "THUMBNAIL"},
  "metrics": {"imp": 1, "costperimp": 2.5, "convKfPfAdd1d": 3, "convKfPfAdd7d": 4, "costPerConvKfPfAdd1d": 5.5, "costPerConvKfPfAdd7d": 6.5, "convKfPfAdd1dRate": 0.1, "convKfPfAdd7dRate": 0.2}}]}`)

	rows, err := c.Reports.Creatives(t.Context(), 1111111111, 3333333331, ReportQuery{
		MetricsGroups: []MetricsGroup{MetricsBasic, MetricsAddition, MetricsKakaoFriend},
		Dimension:     DimensionAssetType,
	})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111111111")
	if q := req.Query; q.Get("campaignId") != "3333333331" || q.Has("adGroupId") || q.Get("metricsGroups") != "BASIC,ADDITION,KAKAO_FRIEND" || q.Get("dimension") != "ASSET_TYPE" {
		t.Errorf("query = %v", q)
	}
	want := KakaoFriendMetrics{ConvKfPfAdd1d: 3, ConvKfPfAdd7d: 4, CostPerConvKfPfAdd1d: 5.5, CostPerConvKfPfAdd7d: 6.5, ConvKfPfAdd1dRate: 0.1, ConvKfPfAdd7dRate: 0.2}
	if len(rows) != 1 || rows[0].Dimensions.CreativeLinkID != 7777777771 || rows[0].Dimensions.Product != ProductTalkChannel ||
		rows[0].Dimensions.AssetType != AssetThumbnail || rows[0].Metrics.CostPerImp != 2.5 || rows[0].Metrics.KakaoFriendMetrics != want {
		t.Errorf("rows = %+v", rows)
	}

	if _, err := c.Reports.Creatives(t.Context(), 1, 0, ReportQuery{MetricsGroups: []MetricsGroup{MetricsBasic}}); err == nil {
		t.Error("missing campaign: want error")
	}
}
