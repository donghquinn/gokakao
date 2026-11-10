package keywordad

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	kakao "github.com/donghquinn/gokakao"
)

// DatePreset is a predefined report period.
type DatePreset string

// Date presets.
const (
	DateToday      DatePreset = "TODAY"
	DateYesterday  DatePreset = "YESTERDAY"
	DateThisWeek   DatePreset = "THIS_WEEK"
	DateLastWeek   DatePreset = "LAST_WEEK"
	DateLast7Days  DatePreset = "LAST_7DAY"
	DateLast14Days DatePreset = "LAST_14DAY"
	DateLast30Days DatePreset = "LAST_30DAY"
	DateThisMonth  DatePreset = "THIS_MONTH"
	DateLastMonth  DatePreset = "LAST_MONTH"
)

// TimeUnit is the period of each report row.
type TimeUnit string

// Time units.
const (
	TimeUnitNone  TimeUnit = "NONE"
	TimeUnitDay   TimeUnit = "DAY"   // periods up to 31 days
	TimeUnitWeek  TimeUnit = "WEEK"  // periods up to 62 days
	TimeUnitMonth TimeUnit = "MONTH" // periods up to 12 months
)

// ReportDimension breaks a report down by an attribute.
type ReportDimension string

// Report dimensions. Each lists the reports that support it.
const (
	DimensionNone             ReportDimension = "NONE"
	DimensionHour             ReportDimension = "HOUR"              // all reports
	DimensionDevice           ReportDimension = "DEVICE"            // all but ad account
	DimensionPlacement        ReportDimension = "PLACEMENT"         // campaign, ad group, creative
	DimensionTargetingType    ReportDimension = "TARGETING_TYPE"    // campaign, ad group, creative
	DimensionKeywordEx        ReportDimension = "KEYWORD_EX"        // keyword
	DimensionAssetType        ReportDimension = "ASSET_TYPE"        // creative (premium link)
	DimensionTemplateCategory ReportDimension = "TEMPLATE_CATEGORY" // creative (talk channel)
)

// Product is a campaign type in reports.
type Product string

// Products.
const (
	ProductPremiumLink Product = "PREMIUM_LINK"
	ProductTalkChannel Product = "TALK_CHANNEL"
)

// ReportQuery selects the metrics and period of a report.
type ReportQuery struct {
	// MetricsGroups is required. Ad account reports support only
	// MetricsBasic.
	MetricsGroups []MetricsGroup
	// Start and End are "yyyyMMdd" dates within the last two years. If
	// either is empty, DatePreset applies, and DateToday if that is empty
	// too.
	Start, End string
	DatePreset DatePreset
	// Dimension defaults to DimensionNone.
	Dimension ReportDimension
	// TimeUnit defaults to TimeUnitNone.
	TimeUnit TimeUnit
	// AdGroupID narrows keyword and creative reports to an ad group. Other
	// reports ignore it.
	AdGroupID int64
}

func (q *ReportQuery) values() (url.Values, error) {
	if len(q.MetricsGroups) == 0 {
		return nil, errors.New("keywordad: metrics groups are required")
	}
	gs := make([]string, len(q.MetricsGroups))
	for i, g := range q.MetricsGroups {
		gs[i] = string(g)
	}
	v := url.Values{"metricsGroups": {strings.Join(gs, ",")}}
	set := func(k, s string) {
		if s != "" {
			v.Set(k, s)
		}
	}
	set("start", q.Start)
	set("end", q.End)
	set("datePreset", string(q.DatePreset))
	set("dimension", string(q.Dimension))
	set("timeUnit", string(q.TimeUnit))
	return v, nil
}

// BasicMetrics are the MetricsBasic report metrics.
type BasicMetrics struct {
	// Imp is impressions and Click clicks.
	Imp   int64 `json:"imp"`
	Click int64 `json:"click"`
	// CTR is the click-through rate in percent.
	CTR float64 `json:"ctr"`
	// Spending is the cost in KRW.
	Spending float64 `json:"spending"`
}

// AdditionMetrics are the MetricsAddition report metrics.
type AdditionMetrics struct {
	// RImp is total impressions.
	RImp int64 `json:"rimp"`
	// PPC is the cost per click in KRW.
	PPC float64 `json:"ppc"`
	// Rank is the average position; premium link only.
	Rank float64 `json:"rank"`
	// CostPerImp is the cost per impression; talk channel search only.
	CostPerImp float64 `json:"costperimp"`
}

// PixelSDKConversionMetrics are the MetricsPixelSDKConversion report
// metrics: Pixel & SDK conversions within 1 and 7 days.
type PixelSDKConversionMetrics struct {
	ConvCmptReg1d       int64 `json:"convCmptReg1d"`  // registrations completed
	ConvCmptReg7d       int64 `json:"convCmptReg7d"`  // registrations completed
	ConvViewCart1d      int64 `json:"convViewCart1d"` // cart views
	ConvViewCart7d      int64 `json:"convViewCart7d"` // cart views
	ConvPurchase1d      int64 `json:"convPurchase1d"`
	ConvPurchase7d      int64 `json:"convPurchase7d"`
	ConvPurchaseP1d     int64 `json:"convPurchaseP1d"` // purchase amount in KRW
	ConvPurchaseP7d     int64 `json:"convPurchaseP7d"` // purchase amount in KRW
	ConvParticipation1d int64 `json:"convParticipation1d"`
	ConvParticipation7d int64 `json:"convParticipation7d"`
	ConvSignup1d        int64 `json:"convSignup1d"`
	ConvSignup7d        int64 `json:"convSignup7d"`
	ConvAppInstall1d    int64 `json:"convAppInstall1d"`
	ConvAppInstall7d    int64 `json:"convAppInstall7d"`
}

// KakaoFriendMetrics are the MetricsKakaoFriend report metrics: KakaoTalk
// Channel friends added within 1 and 7 days. Talk channel search only.
type KakaoFriendMetrics struct {
	ConvKfPfAdd1d        int64   `json:"convKfPfAdd1d"`
	ConvKfPfAdd7d        int64   `json:"convKfPfAdd7d"`
	CostPerConvKfPfAdd1d float64 `json:"costPerConvKfPfAdd1d"`
	CostPerConvKfPfAdd7d float64 `json:"costPerConvKfPfAdd7d"`
	ConvKfPfAdd1dRate    float64 `json:"convKfPfAdd1dRate"`
	ConvKfPfAdd7dRate    float64 `json:"convKfPfAdd7dRate"`
}

// ReportMetrics holds the metrics of the requested groups; the others are
// zero.
type ReportMetrics struct {
	BasicMetrics
	AdditionMetrics
	PixelSDKConversionMetrics
	KakaoFriendMetrics
}

// ReportDimensions identifies what a report row covers. Fields that do not
// apply to the report are zero.
type ReportDimensions struct {
	AdAccountID    kakao.Int64 `json:"adAccountId"`
	CampaignID     kakao.Int64 `json:"campaignId"`
	AdGroupID      kakao.Int64 `json:"adGroupId"`
	KeywordID      kakao.Int64 `json:"keywordId"`
	CreativeLinkID kakao.Int64 `json:"creativeLinkId"`
	Product        Product     `json:"product"`
	// Hour is "00" to "23" with DimensionHour.
	Hour string `json:"hour"`
	// Device is "MOBILE" or "PC" with DimensionDevice.
	Device string `json:"device"`
	// Placement is e.g. "SEARCH_KAKAO" with DimensionPlacement.
	Placement string `json:"placement"`
	// TargetingType is "EXACT", "PREFIX" or "CATEGORY" with
	// DimensionTargetingType.
	TargetingType string `json:"targetingType"`
	// KeywordEx is set with DimensionKeywordEx.
	KeywordEx string `json:"keywordEx"`
	// AssetType is set with DimensionAssetType.
	AssetType AssetType `json:"assetType"`
}

// ReportRow is one row of a report.
type ReportRow struct {
	// Start and End are the row's period. Kakao's examples use both
	// "yyyyMMdd" and "yyyy-MM-dd".
	Start      string           `json:"start"`
	End        string           `json:"end"`
	Dimensions ReportDimensions `json:"dimensions"`
	Metrics    ReportMetrics    `json:"metrics"`
}

func (s *ReportService) report(ctx context.Context, adAccountID int64, path string, campaignID int64, q ReportQuery, withAdGroup bool) ([]ReportRow, error) {
	v, err := q.values()
	if err != nil {
		return nil, err
	}
	if campaignID != 0 {
		v.Set("campaignId", strconv.FormatInt(campaignID, 10))
	}
	if withAdGroup && q.AdGroupID != 0 {
		v.Set("adGroupId", strconv.FormatInt(q.AdGroupID, 10))
	}
	var res struct {
		Data []ReportRow `json:"data"`
	}
	if err := s.c.do(ctx, call{method: http.MethodGet, path: path, adAccountID: adAccountID, query: v}, &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}

// AdAccount returns an ad account's report. Kakao allows five report calls
// per second.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/report
func (s *ReportService) AdAccount(ctx context.Context, adAccountID int64, q ReportQuery) ([]ReportRow, error) {
	return s.report(ctx, adAccountID, "/v1/adAccounts/report", 0, q, false)
}

// Campaigns returns a report on an ad account's campaigns. Kakao allows
// five report calls per second.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/report
func (s *ReportService) Campaigns(ctx context.Context, adAccountID int64, q ReportQuery) ([]ReportRow, error) {
	return s.report(ctx, adAccountID, "/v1/campaigns/report", 0, q, false)
}

// AdGroups returns a report on a campaign's ad groups. Kakao allows five
// report calls per second.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/report
func (s *ReportService) AdGroups(ctx context.Context, adAccountID, campaignID int64, q ReportQuery) ([]ReportRow, error) {
	if campaignID == 0 {
		return nil, errors.New("keywordad: campaign ID is required")
	}
	return s.report(ctx, adAccountID, "/v1/adGroups/report", campaignID, q, false)
}

// Keywords returns a report on a premium link campaign's keywords, or an
// ad group's when q.AdGroupID is set. Talk channel search campaigns have
// no keyword report. Kakao allows five report calls per second.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/report
func (s *ReportService) Keywords(ctx context.Context, adAccountID, campaignID int64, q ReportQuery) ([]ReportRow, error) {
	if campaignID == 0 {
		return nil, errors.New("keywordad: campaign ID is required")
	}
	return s.report(ctx, adAccountID, "/v1/keywords/report", campaignID, q, true)
}

// Creatives returns a report on a campaign's creative links, or an ad
// group's when q.AdGroupID is set. Kakao allows five report calls per
// second.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/report
func (s *ReportService) Creatives(ctx context.Context, adAccountID, campaignID int64, q ReportQuery) ([]ReportRow, error) {
	if campaignID == 0 {
		return nil, errors.New("keywordad: campaign ID is required")
	}
	return s.report(ctx, adAccountID, "/v1/creatives/report", campaignID, q, true)
}
