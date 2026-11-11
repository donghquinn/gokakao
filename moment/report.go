package moment

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// ReportQuery selects the metrics, period and grouping of a report.
//
// A report covers today only with DatePresetToday; other presets and
// Start/End exclude today. A day's figures may change until 08:00 the
// next day.
type ReportQuery struct {
	// MetricsGroups is required. Display entities report no MESSAGE,
	// MESSAGE_ADDITION or MESSAGE_CLICK metrics, and message entities no
	// BASIC, ADDITION, PLUS_FRIEND or VIDEO metrics.
	MetricsGroups []MetricsGroup
	// Start and End are "yyyyMMdd" dates up to yesterday, at most 31 days
	// apart. If either is empty, DatePreset applies, and
	// DatePresetToday if that is empty too.
	Start, End string
	DatePreset DatePreset
	// TimeUnit defaults to TimeUnitDay. Use DimensionHour for hourly rows.
	TimeUnit TimeUnit
	// Level groups the rows: AD_ACCOUNT (default) or CAMPAIGN for ad
	// account reports, CAMPAIGN (default) or AD_GROUP for campaign
	// reports, AD_GROUP (default) or CREATIVE for ad group reports.
	// Creative reports ignore it.
	Level ReportLevel
	// Dimension breaks the rows down further. Age, gender, location,
	// device and placement only have values for display ads.
	Dimension ReportDimension
}

func (q *ReportQuery) values(withLevel bool) (url.Values, error) {
	if len(q.MetricsGroups) == 0 {
		return nil, errors.New("moment: metrics groups are required")
	}
	gs := make([]string, len(q.MetricsGroups))
	for i, g := range q.MetricsGroups {
		gs[i] = string(g)
	}
	v := url.Values{"metricsGroup": {strings.Join(gs, ",")}}
	set := func(k, s string) {
		if s != "" {
			v.Set(k, s)
		}
	}
	set("start", q.Start)
	set("end", q.End)
	set("datePreset", string(q.DatePreset))
	set("timeUnit", string(q.TimeUnit))
	if withLevel {
		set("level", string(q.Level))
	}
	set("dimension", string(q.Dimension))
	return v, nil
}

// ReportRow is one row of a report. A row without data has empty
// Dimensions and Metrics.
type ReportRow struct {
	// Start and End are the row's period. Kakao documents "yyyyMMdd" but
	// its examples use "yyyy-MM-dd".
	Start string `json:"start"`
	End   string `json:"end"`
	// Dimensions holds the level ID (e.g. "campaign_id") and the
	// dimension value (e.g. "creative_format").
	Dimensions Dimensions `json:"dimensions"`
	// Metrics holds the metrics of the requested groups.
	Metrics Metrics `json:"metrics"`
}

// AdAccountReportOptions are the optional parameters of
// [ReportService.AdAccount].
type AdAccountReportOptions struct {
	// AdAccountIDs reports on up to 5 ad accounts at once instead of the
	// one in the adAccountId header. The rate limit then applies per app.
	AdAccountIDs []int64
}

func reportJoinIDs(ids []int64) string {
	s := make([]string, len(ids))
	for i, id := range ids {
		s[i] = strconv.FormatInt(id, 10)
	}
	return strings.Join(s, ",")
}

func (s *ReportService) report(ctx context.Context, adAccountID int64, path, idKey string, ids []int64, q ReportQuery, withLevel bool) ([]ReportRow, error) {
	v, err := q.values(withLevel)
	if err != nil {
		return nil, err
	}
	v.Set(idKey, reportJoinIDs(ids))
	var res struct {
		Data []ReportRow `json:"data"`
	}
	if err := s.c.do(ctx, call{method: http.MethodGet, path: path, adAccountID: adAccountID, query: v}, &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}

// AdAccount returns a report on an ad account, covering display and
// message campaigns. adAccountID may be 0 when opts lists the accounts.
// Kakao allows one call per 5 seconds per ad account (per app with
// several accounts).
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/report
func (s *ReportService) AdAccount(ctx context.Context, adAccountID int64, q ReportQuery, opts *AdAccountReportOptions) ([]ReportRow, error) {
	ids := []int64{adAccountID}
	if opts != nil && len(opts.AdAccountIDs) > 0 {
		ids = opts.AdAccountIDs
	}
	if len(ids) > 5 {
		return nil, errors.New("moment: at most 5 ad account IDs are allowed")
	}
	if ids[0] == 0 {
		return nil, errors.New("moment: ad account ID is required")
	}
	return s.report(ctx, adAccountID, "/v4/adAccounts/report", "adAccountId", ids, q, true)
}

func reportCheckIDs(adAccountID int64, ids []int64, limit int, what string) error {
	if adAccountID == 0 {
		return errors.New("moment: ad account ID is required")
	}
	if len(ids) == 0 || len(ids) > limit {
		return fmt.Errorf("moment: 1 to %d %s IDs are required", limit, what)
	}
	return nil
}

// Campaigns returns a report on up to 5 campaigns. Kakao allows one call
// per 5 seconds per ad account and app.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/report
func (s *ReportService) Campaigns(ctx context.Context, adAccountID int64, campaignIDs []int64, q ReportQuery) ([]ReportRow, error) {
	if err := reportCheckIDs(adAccountID, campaignIDs, 5, "campaign"); err != nil {
		return nil, err
	}
	return s.report(ctx, adAccountID, "/v4/campaigns/report", "campaignId", campaignIDs, q, true)
}

// AdGroups returns a report on up to 40 ad groups. Kakao allows one call
// per second per ad account and app.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/report
func (s *ReportService) AdGroups(ctx context.Context, adAccountID int64, adGroupIDs []int64, q ReportQuery) ([]ReportRow, error) {
	if err := reportCheckIDs(adAccountID, adGroupIDs, 40, "ad group"); err != nil {
		return nil, err
	}
	return s.report(ctx, adAccountID, "/v4/adGroups/report", "adGroupId", adGroupIDs, q, true)
}

// Creatives returns a report on up to 100 creatives, by original creative
// ID. Rows are always at the creative level; q.Level is ignored. Kakao
// allows one call per 5 seconds per ad account and app.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/report
func (s *ReportService) Creatives(ctx context.Context, adAccountID int64, creativeIDs []int64, q ReportQuery) ([]ReportRow, error) {
	if err := reportCheckIDs(adAccountID, creativeIDs, 100, "creative"); err != nil {
		return nil, err
	}
	return s.report(ctx, adAccountID, "/v4/creatives/report", "creativeId", creativeIDs, q, false)
}
