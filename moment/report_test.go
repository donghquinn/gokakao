package moment

import (
	"errors"
	"net/http"
	"net/url"
	"testing"

	kakao "github.com/donghquinn/gokakao"
)

func reportJSON(levelKey string) string {
	return `{
  "code": 200,
  "message": "Success",
  "data": [
    {
      "start": "2020-01-01",
      "end": "2020-01-01",
      "dimensions": {"creative_format": "IMAGE BANNER", "` + levelKey + `": "1234"},
      "metrics": {"imp": 4, "click": 0, "ctr": 0.0, "cost": 0.0}
    },
    {}
  ]
}`
}

func reportAssertRows(t *testing.T, rows []ReportRow, levelKey string) {
	t.Helper()
	if len(rows) != 2 {
		t.Fatalf("rows = %+v", rows)
	}
	r := rows[0]
	if r.Start != "2020-01-01" || r.End != "2020-01-01" ||
		r.Dimensions["creative_format"] != "IMAGE BANNER" || r.Dimensions[levelKey] != "1234" ||
		r.Metrics["imp"] != 4 || r.Metrics["click"] != 0 || len(r.Metrics) != 4 {
		t.Errorf("row = %+v", r)
	}
	if rows[1].Dimensions != nil || rows[1].Metrics != nil {
		t.Errorf("empty row = %+v", rows[1])
	}
}

func reportAssertQuery(t *testing.T, got url.Values, want map[string]string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("query = %v, want %v", got, want)
	}
	for k, v := range want {
		if got.Get(k) != v {
			t.Errorf("query %s = %q, want %q", k, got.Get(k), v)
		}
	}
}

func TestReportService_AdAccount(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/adAccounts/report", 200, reportJSON("ad_account_id"))

	rows, err := c.Reports.AdAccount(t.Context(), 1234, ReportQuery{
		MetricsGroups: []MetricsGroup{MetricsBasic},
		DatePreset:    DatePresetToday,
		Level:         ReportLevelAdAccount,
		Dimension:     DimensionCreativeFormat,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1234")
	if req.Method != http.MethodGet {
		t.Errorf("method = %s", req.Method)
	}
	reportAssertQuery(t, req.Query, map[string]string{
		"adAccountId": "1234", "datePreset": "TODAY", "level": "AD_ACCOUNT", "dimension": "CREATIVE_FORMAT", "metricsGroup": "BASIC",
	})
	reportAssertRows(t, rows, "ad_account_id")

	// Several accounts, without the header.
	if _, err := c.Reports.AdAccount(t.Context(), 0, ReportQuery{
		MetricsGroups: []MetricsGroup{MetricsBasic, MetricsAddition},
		Start:         "20200101", End: "20200101", TimeUnit: TimeUnitAll,
	}, &AdAccountReportOptions{AdAccountIDs: []int64{1, 2}}); err != nil {
		t.Fatal(err)
	}
	req = assertRequest(t, srv, "")
	reportAssertQuery(t, req.Query, map[string]string{
		"adAccountId": "1,2", "start": "20200101", "end": "20200101", "timeUnit": "ALL", "metricsGroup": "BASIC,ADDITION",
	})

	q := ReportQuery{MetricsGroups: []MetricsGroup{MetricsBasic}}
	if _, err := c.Reports.AdAccount(t.Context(), 0, q, nil); err == nil {
		t.Error("missing ad account: want error")
	}
	if _, err := c.Reports.AdAccount(t.Context(), 1, q, &AdAccountReportOptions{AdAccountIDs: []int64{1, 2, 3, 4, 5, 6}}); err == nil {
		t.Error("6 ad accounts: want error")
	}
	if _, err := c.Reports.AdAccount(t.Context(), 1, ReportQuery{}, nil); err == nil {
		t.Error("missing metrics groups: want error")
	}
}

func TestReportService_Campaigns(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/campaigns/report", 200, reportJSON("campaign_id"))

	rows, err := c.Reports.Campaigns(t.Context(), 10000, []int64{1, 2, 3, 4, 5}, ReportQuery{
		MetricsGroups: []MetricsGroup{MetricsBasic},
		DatePreset:    DatePresetToday,
		Level:         ReportLevelCampaign,
		Dimension:     DimensionCreativeFormat,
	})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "10000")
	reportAssertQuery(t, req.Query, map[string]string{
		"campaignId": "1,2,3,4,5", "datePreset": "TODAY", "level": "CAMPAIGN", "dimension": "CREATIVE_FORMAT", "metricsGroup": "BASIC",
	})
	reportAssertRows(t, rows, "campaign_id")

	q := ReportQuery{MetricsGroups: []MetricsGroup{MetricsBasic}}
	if _, err := c.Reports.Campaigns(t.Context(), 1, nil, q); err == nil {
		t.Error("no campaigns: want error")
	}
	if _, err := c.Reports.Campaigns(t.Context(), 1, make([]int64, 6), q); err == nil {
		t.Error("6 campaigns: want error")
	}
	if _, err := c.Reports.Campaigns(t.Context(), 0, []int64{1}, q); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestReportService_AdGroups(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/adGroups/report", 200, reportJSON("ad_group_id"))

	rows, err := c.Reports.AdGroups(t.Context(), 10000, []int64{15970}, ReportQuery{
		MetricsGroups: []MetricsGroup{MetricsBasic},
		Start:         "20200501",
		End:           "20200501",
		Level:         ReportLevelAdGroup,
		Dimension:     DimensionCreativeFormat,
	})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "10000")
	reportAssertQuery(t, req.Query, map[string]string{
		"adGroupId": "15970", "start": "20200501", "end": "20200501", "level": "AD_GROUP", "dimension": "CREATIVE_FORMAT", "metricsGroup": "BASIC",
	})
	reportAssertRows(t, rows, "ad_group_id")

	if _, err := c.Reports.AdGroups(t.Context(), 1, make([]int64, 41), ReportQuery{MetricsGroups: []MetricsGroup{MetricsBasic}}); err == nil {
		t.Error("41 ad groups: want error")
	}
}

func TestReportService_Creatives(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/creatives/report", 200, reportJSON("creative_id"))

	rows, err := c.Reports.Creatives(t.Context(), 10000, []int64{40068, 40065}, ReportQuery{
		MetricsGroups: []MetricsGroup{MetricsBasic},
		DatePreset:    DatePresetToday,
		Level:         ReportLevelCreative, // ignored
		Dimension:     DimensionCreativeFormat,
	})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "10000")
	reportAssertQuery(t, req.Query, map[string]string{
		"creativeId": "40068,40065", "datePreset": "TODAY", "dimension": "CREATIVE_FORMAT", "metricsGroup": "BASIC",
	})
	reportAssertRows(t, rows, "creative_id")

	if _, err := c.Reports.Creatives(t.Context(), 1, make([]int64, 101), ReportQuery{MetricsGroups: []MetricsGroup{MetricsBasic}}); err == nil {
		t.Error("101 creatives: want error")
	}
}

func TestReportService_apiErrors(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("/", 400, `{"code":-813,"msg":"KakaoMomentException"}`)
	q := ReportQuery{MetricsGroups: []MetricsGroup{MetricsBasic}}
	_, err := c.Reports.Campaigns(t.Context(), 1, []int64{2}, q)
	var e *kakao.Error
	if !errors.As(err, &e) || e.Code != -813 {
		t.Errorf("err = %v", err)
	}
}
