package keywordad

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	kakao "github.com/donghquinn/gokakao"
)

// Campaign is a campaign (캠페인), the top-level unit of ad strategy.
type Campaign struct {
	ID           kakao.Int64 `json:"id"`
	BizChannelID kakao.Int64 `json:"bizChannelId"`
	Name         string      `json:"name"`
	Config       Config      `json:"config"`
	// Status lists the delivery statuses; [StatusOn] means it can run.
	Status []Status `json:"status"`
	// TrackID is the conversion tracking code (track_id), empty when not
	// linked.
	TrackID string `json:"trackId"`
	// TraceURL reports whether tracking URL parameters are appended.
	TraceURL bool `json:"traceUrl"`
	// DailyBudgetAmount is the daily budget in KRW, 0 when unset.
	DailyBudgetAmount float64 `json:"dailyBudgetAmount"`
	// DailyBudgetAmountOver reports whether today's budget is spent.
	DailyBudgetAmountOver bool `json:"dailyBudgetAmountOver"`
}

// Changes returns the [CampaignChanges] that keep c as it is, for editing
// before [CampaignService.Update].
func (c *Campaign) Changes() CampaignChanges {
	ch := CampaignChanges{Name: c.Name, TraceURL: &c.TraceURL}
	if c.TrackID != "" {
		ch.TrackID = &c.TrackID
	}
	if c.DailyBudgetAmount > 0 {
		amount := int64(c.DailyBudgetAmount)
		ch.DailyBudgetAmount = &amount
	}
	return ch
}

// CampaignListOptions are the optional parameters of
// [CampaignService.List].
type CampaignListOptions struct {
	// Configs filters by status; Kakao lists ON and OFF campaigns when
	// empty.
	Configs []Config
}

// List returns the ad account's campaigns.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/campaign
func (s *CampaignService) List(ctx context.Context, adAccountID int64, opts *CampaignListOptions) ([]Campaign, error) {
	if adAccountID == 0 {
		return nil, errors.New("keywordad: ad account ID is required")
	}
	q := url.Values{}
	if opts != nil {
		for _, cfg := range opts.Configs {
			q.Add("config", string(cfg))
		}
	}
	var cs []Campaign
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v1/campaigns", adAccountID: adAccountID, query: q}, &cs); err != nil {
		return nil, err
	}
	return cs, nil
}

// Get returns a campaign.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/campaign
func (s *CampaignService) Get(ctx context.Context, adAccountID, campaignID int64) (*Campaign, error) {
	if adAccountID == 0 || campaignID == 0 {
		return nil, errors.New("keywordad: ad account ID and campaign ID are required")
	}
	var c Campaign
	if err := s.c.do(ctx, call{method: http.MethodGet, path: idPath("/v1/campaigns/%d", campaignID), adAccountID: adAccountID}, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// NewCampaign describes a campaign to create.
type NewCampaign struct {
	BizChannelID int64  `json:"bizChannelId"`
	Name         string `json:"name"`
	// TrackID is the conversion tracking code to link; none by default.
	TrackID string `json:"trackId,omitempty"`
	// TraceURL appends tracking URL parameters; Kakao's default is true.
	TraceURL *bool `json:"traceUrl,omitempty"`
	// DailyBudgetAmount is the daily budget in KRW; unset (no limit) when
	// 0.
	DailyBudgetAmount int64 `json:"dailyBudgetAmount,omitempty"`
}

// Create creates a campaign. It is limited to 1 request per 5 seconds.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/campaign
func (s *CampaignService) Create(ctx context.Context, adAccountID int64, c NewCampaign) (*Campaign, error) {
	if adAccountID == 0 {
		return nil, errors.New("keywordad: ad account ID is required")
	}
	if c.BizChannelID == 0 || c.Name == "" {
		return nil, errors.New("keywordad: biz channel ID and name are required")
	}
	var res Campaign
	if err := s.c.do(ctx, call{method: http.MethodPost, path: "/v1/campaigns", adAccountID: adAccountID, body: c}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// CampaignChanges is the new state of a campaign for
// [CampaignService.Update].
//
// Kakao replaces the campaign: a nil field resets to Kakao's default rather
// than keeping the current value. Start from [Campaign.Changes] to keep
// the fields you do not change.
type CampaignChanges struct {
	Name string `json:"name"`
	// TrackID is the conversion tracking code; nil unlinks it.
	TrackID *string `json:"trackId,omitempty"`
	// TraceURL appends tracking URL parameters; nil means true.
	TraceURL *bool `json:"traceUrl,omitempty"`
	// DailyBudgetAmount is the daily budget in KRW; nil unsets it.
	DailyBudgetAmount *int64 `json:"dailyBudgetAmount,omitempty"`
}

// Update replaces a campaign's settings; see [CampaignChanges]. It is
// limited to 1 request per 5 seconds.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/campaign
func (s *CampaignService) Update(ctx context.Context, adAccountID, campaignID int64, c CampaignChanges) (*Campaign, error) {
	if adAccountID == 0 || campaignID == 0 {
		return nil, errors.New("keywordad: ad account ID and campaign ID are required")
	}
	if c.Name == "" {
		return nil, errors.New("keywordad: name is required")
	}
	var res Campaign
	err := s.c.do(ctx, call{method: http.MethodPut, path: idPath("/v1/campaigns/%d", campaignID), adAccountID: adAccountID, body: c}, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// SetConfig turns a campaign ON or OFF. It only applies to campaigns that
// are ON or OFF, and is limited to 1 request per second.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/campaign
func (s *CampaignService) SetConfig(ctx context.Context, adAccountID, campaignID int64, cfg Config) error {
	if adAccountID == 0 || campaignID == 0 {
		return errors.New("keywordad: ad account ID and campaign ID are required")
	}
	if cfg != ConfigOn && cfg != ConfigOff {
		return errors.New("keywordad: config must be ON or OFF")
	}
	body := struct {
		ID     int64  `json:"id"`
		Config Config `json:"config"`
	}{campaignID, cfg}
	return s.c.do(ctx, call{method: http.MethodPatch, path: idPath("/v1/campaigns/%d/onOff", campaignID), adAccountID: adAccountID, body: body}, nil)
}

// UpdateDailyBudget sets a campaign's daily budget in KRW; 0 unsets it (no
// limit). It is limited to 1 request per second.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/campaign
func (s *CampaignService) UpdateDailyBudget(ctx context.Context, adAccountID, campaignID, amount int64) error {
	if adAccountID == 0 || campaignID == 0 {
		return errors.New("keywordad: ad account ID and campaign ID are required")
	}
	if amount < 0 {
		return errors.New("keywordad: daily budget must not be negative")
	}
	body := struct {
		DailyBudgetAmount *int64 `json:"dailyBudgetAmount"`
	}{}
	if amount > 0 {
		body.DailyBudgetAmount = &amount
	}
	return s.c.do(ctx, call{method: http.MethodPatch, path: idPath("/v1/campaigns/%d/dailyBudget", campaignID), adAccountID: adAccountID, body: body}, nil)
}

// Delete deletes a campaign.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/campaign
func (s *CampaignService) Delete(ctx context.Context, adAccountID, campaignID int64) error {
	if adAccountID == 0 || campaignID == 0 {
		return errors.New("keywordad: ad account ID and campaign ID are required")
	}
	return s.c.do(ctx, call{method: http.MethodDelete, path: idPath("/v1/campaigns/%d", campaignID), adAccountID: adAccountID}, nil)
}
