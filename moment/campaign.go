package moment

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	kakao "github.com/donghquinn/gokakao"
)

// CampaignSummary is a campaign in a list.
type CampaignSummary struct {
	ID     kakao.Int64 `json:"id"`
	Name   string      `json:"name"`
	Config Config      `json:"config"`
	// UserConfig is the legacy name of Config, kept for history.
	UserConfig   Config       `json:"userConfig"`
	SystemConfig SystemConfig `json:"systemConfig"`
}

// CampaignListOptions are the optional parameters of
// [CampaignService.List].
type CampaignListOptions struct {
	// Config filters by status.
	Config Config
}

// List returns the ad account's campaigns.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/campaign
func (s *CampaignService) List(ctx context.Context, adAccountID int64, opts *CampaignListOptions) ([]CampaignSummary, error) {
	if adAccountID == 0 {
		return nil, errors.New("moment: ad account ID is required")
	}
	q := url.Values{}
	if opts != nil && opts.Config != "" {
		q.Set("config", string(opts.Config))
	}
	var res struct {
		Content []CampaignSummary `json:"content"`
	}
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v4/campaigns", adAccountID: adAccountID, query: q}, &res); err != nil {
		return nil, err
	}
	return res.Content, nil
}

// CampaignDetail is a campaign (캠페인).
type CampaignDetail struct {
	ID               kakao.Int64      `json:"id"`
	AdAccountID      kakao.Int64      `json:"adAccountId"`
	Name             string           `json:"name"`
	CampaignTypeGoal CampaignTypeGoal `json:"campaignTypeGoal"`
	Objective        *Objective       `json:"objective"`
	// DailyBudgetAmount is in KRW; 0 means no limit (한도 없음).
	DailyBudgetAmount kakao.Int64 `json:"dailyBudgetAmount"`
	Config            Config      `json:"config"`
	// StatusDescription combines Config and DailyBudgetAmountOver, e.g.
	// "운영중".
	StatusDescription string `json:"statusDescription"`
	// TrackID is the conversion tracking (전환추적) Pixel & SDK ID.
	TrackID               string           `json:"trackId"`
	Status                []CampaignStatus `json:"status"`
	SystemConfig          SystemConfig     `json:"systemConfig"`
	DailyBudgetAmountOver bool             `json:"isDailyBudgetAmountOver"`
	// Kclid reports whether the Kakao click ID (카카오 클릭 ID) is
	// appended to landing URLs.
	Kclid bool `json:"kclid"`
}

// Changes returns the [CampaignChanges] that keep c as it is, for editing
// before [CampaignService.Update].
func (c *CampaignDetail) Changes() CampaignChanges {
	ch := CampaignChanges{Name: c.Name, Kclid: &c.Kclid}
	if c.TrackID != "" {
		ch.TrackID = &c.TrackID
	}
	if c.DailyBudgetAmount > 0 {
		amount := int64(c.DailyBudgetAmount)
		ch.DailyBudgetAmount = &amount
	}
	return ch
}

// Get returns a campaign.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/campaign
func (s *CampaignService) Get(ctx context.Context, adAccountID, campaignID int64) (*CampaignDetail, error) {
	if adAccountID == 0 || campaignID == 0 {
		return nil, errors.New("moment: ad account ID and campaign ID are required")
	}
	var c CampaignDetail
	if err := s.c.do(ctx, call{method: http.MethodGet, path: idPath("/v4/campaigns/%d", campaignID), adAccountID: adAccountID}, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// NewCampaign describes a campaign to create.
type NewCampaign struct {
	// Name is at most 150 characters; Kakao names the campaign
	// "{유형}_{목표}_yyyyMMddHHmm" when empty.
	Name             string           `json:"name,omitempty"`
	CampaignTypeGoal CampaignTypeGoal `json:"campaignTypeGoal"`
	// Objective is required for the type and goal pairs that take one:
	// a KakaoTalk Channel profile for TALK_CHANNEL reach and
	// PERSONAL_MESSAGE, a Pixel & SDK or channel for CONVERSION.
	Objective *Objective `json:"objective,omitempty"`
	// DailyBudgetAmount is the daily budget in KRW; no limit when 0. Not
	// allowed for TALK_CHANNEL and PERSONAL_MESSAGE campaigns.
	DailyBudgetAmount int64 `json:"dailyBudgetAmount,omitempty"`
	// TrackID is a Pixel & SDK ID from [AdAccountService.Trackers] for
	// conversion tracking. Kakao sets it itself for Pixel & SDK
	// objectives.
	TrackID string `json:"trackId,omitempty"`
	// Kclid appends the Kakao click ID to landing URLs; test landings
	// before enabling it.
	Kclid *bool `json:"kclid,omitempty"`
}

// Create creates a campaign.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/campaign
func (s *CampaignService) Create(ctx context.Context, adAccountID int64, c NewCampaign) (*CampaignDetail, error) {
	if adAccountID == 0 {
		return nil, errors.New("moment: ad account ID is required")
	}
	if c.CampaignTypeGoal.CampaignType == "" || c.CampaignTypeGoal.Goal == "" {
		return nil, errors.New("moment: campaign type and goal are required")
	}
	var res CampaignDetail
	if err := s.c.do(ctx, call{method: http.MethodPost, path: "/v4/campaigns", adAccountID: adAccountID, body: c}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// CampaignChanges is the new state of a campaign for
// [CampaignService.Update].
//
// Kakao replaces the campaign's budget, tracking and click ID settings: a
// nil field resets it (no budget limit, no tracking ID, Kakao's default
// click ID setting) rather than keeping the current value. Only an empty
// Name keeps the current name. Start from [CampaignDetail.Changes] to keep
// the fields you do not change.
type CampaignChanges struct {
	// Name keeps the current name when empty.
	Name string `json:"name,omitempty"`
	// DailyBudgetAmount is the daily budget in KRW; nil means no limit.
	DailyBudgetAmount *int64 `json:"dailyBudgetAmount,omitempty"`
	// TrackID is the conversion tracking ID. For VISITING campaigns it is
	// a Pixel & SDK ID from [AdAccountService.Trackers]; for other goals
	// it must be the campaign's current ID, if it has one.
	TrackID *string `json:"trackId,omitempty"`
	// Kclid appends the Kakao click ID to landing URLs.
	Kclid *bool `json:"kclid,omitempty"`
}

// Update replaces a campaign's settings; see [CampaignChanges]. Only the
// name and track ID of TALK_CHANNEL and PERSONAL_MESSAGE reach campaigns
// can change.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/campaign
func (s *CampaignService) Update(ctx context.Context, adAccountID, campaignID int64, c CampaignChanges) (*CampaignDetail, error) {
	if adAccountID == 0 || campaignID == 0 {
		return nil, errors.New("moment: ad account ID and campaign ID are required")
	}
	body := struct {
		ID int64 `json:"id"`
		CampaignChanges
	}{campaignID, c}
	var res CampaignDetail
	if err := s.c.do(ctx, call{method: http.MethodPut, path: "/v4/campaigns", adAccountID: adAccountID, body: body}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// SetConfig turns a campaign ON or OFF. It only applies to campaigns that
// are ON or OFF; a TALK_CHANNEL reach campaign also needs all its ad
// groups OFF or canceled, and changing it pauses its pending messages.
// The API reference limits it to one call per second, the campaign page
// to one call per 5 seconds per user and ad account.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/campaign
func (s *CampaignService) SetConfig(ctx context.Context, adAccountID, campaignID int64, cfg Config) error {
	if adAccountID == 0 || campaignID == 0 {
		return errors.New("moment: ad account ID and campaign ID are required")
	}
	if cfg != ConfigOn && cfg != ConfigOff {
		return errors.New("moment: config must be ON or OFF")
	}
	body := struct {
		ID     int64  `json:"id"`
		Config Config `json:"config"`
	}{campaignID, cfg}
	return s.c.do(ctx, call{method: http.MethodPut, path: "/v4/campaigns/onOff", adAccountID: adAccountID, body: body}, nil)
}

// UpdateDailyBudget sets a display campaign's daily budget in KRW: 50,000
// to 1,000,000,000 in steps of 10, or 0 for no limit. It does not apply to
// reach campaigns. Lowering it also lowers ad group budgets and bids that
// would exceed it. The API reference limits it to one call per second,
// the campaign page to one call per 5 seconds per user and ad account.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/campaign
func (s *CampaignService) UpdateDailyBudget(ctx context.Context, adAccountID, campaignID, amount int64) error {
	if adAccountID == 0 || campaignID == 0 {
		return errors.New("moment: ad account ID and campaign ID are required")
	}
	if amount < 0 {
		return errors.New("moment: daily budget must not be negative")
	}
	body := struct {
		ID                int64  `json:"id"`
		DailyBudgetAmount *int64 `json:"dailyBudgetAmount"`
	}{ID: campaignID}
	if amount > 0 {
		body.DailyBudgetAmount = &amount
	}
	return s.c.do(ctx, call{method: http.MethodPut, path: "/v4/campaigns/dailyBudgetAmount", adAccountID: adAccountID, body: body}, nil)
}

// Delete deletes a campaign: it and its ad groups and creatives stop and
// can no longer be changed, but past reports remain. TALK_BIZ_BOARD reach
// and DAUM_SHOPPING campaigns cannot be deleted.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/campaign
func (s *CampaignService) Delete(ctx context.Context, adAccountID, campaignID int64) error {
	if adAccountID == 0 || campaignID == 0 {
		return errors.New("moment: ad account ID and campaign ID are required")
	}
	return s.c.do(ctx, call{method: http.MethodDelete, path: idPath("/v4/campaigns/%d", campaignID), adAccountID: adAccountID}, nil)
}

// SystemStopReason is why the system stopped a campaign, ad group or
// creative (시스템 정지 사유).
type SystemStopReason struct {
	ID           kakao.Int64  `json:"id"`
	SystemConfig SystemConfig `json:"systemConfig"`
	Reason       string       `json:"reason"`
	// DetailReason is empty when there are no details.
	DetailReason     string     `json:"detailReason"`
	CreatedDate      kakao.Time `json:"createdDate"`
	LastModifiedDate kakao.Time `json:"lastModifiedDate"`
}

// LatestSystemStopReason returns the latest reason the system stopped a
// campaign. Kakao only answers for campaigns whose system config is
// ADMIN_STOP or EXTERNAL_SERVICE_STOP.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/campaign
func (s *CampaignService) LatestSystemStopReason(ctx context.Context, adAccountID, campaignID int64) (*SystemStopReason, error) {
	if adAccountID == 0 || campaignID == 0 {
		return nil, errors.New("moment: ad account ID and campaign ID are required")
	}
	var r SystemStopReason
	if err := s.c.do(ctx, call{method: http.MethodGet, path: idPath("/v4/campaigns/%d/latestSystemConfigHistory", campaignID), adAccountID: adAccountID}, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// SystemStopReasons returns the reasons the system stopped a campaign in
// the last two years. Kakao only answers for campaigns whose system config
// is ADMIN_STOP or EXTERNAL_SERVICE_STOP.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/campaign
func (s *CampaignService) SystemStopReasons(ctx context.Context, adAccountID, campaignID int64) ([]SystemStopReason, error) {
	if adAccountID == 0 || campaignID == 0 {
		return nil, errors.New("moment: ad account ID and campaign ID are required")
	}
	var rs []SystemStopReason
	if err := s.c.do(ctx, call{method: http.MethodGet, path: idPath("/v4/campaigns/%d/systemConfigHistories", campaignID), adAccountID: adAccountID}, &rs); err != nil {
		return nil, err
	}
	return rs, nil
}
