package keywordad

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	kakao "github.com/donghquinn/gokakao"
)

// AdGroup is an ad group (광고그룹). Amounts are in KRW and bid weights
// are percentages, e.g. 150 for 150%.
type AdGroup struct {
	CampaignID kakao.Int64 `json:"campaignId"`
	ID         kakao.Int64 `json:"id"`
	Name       string      `json:"name"`
	Config     Config      `json:"config"`
	// Status lists the delivery statuses; see [Status].
	Status              []Status     `json:"status"`
	Placements          []Placement  `json:"placements"`
	DeviceTypes         []DeviceType `json:"deviceTypes"`
	UseKeywordExtension bool         `json:"useKeywordExtension"`
	// ExceptionKeywords are excluded from keyword extension.
	ExceptionKeywords     []string `json:"exceptionKeywords"`
	DailyBudgetAmount     float64  `json:"dailyBudgetAmount"`
	DailyBudgetAmountOver bool     `json:"dailyBudgetAmountOver"`
	// BidAmount is the default bid for keywords that bid by ad group.
	BidAmount                 float64 `json:"bidAmount"`
	MobileBidWeight           float64 `json:"mobileBidWeight"`
	PCBidWeight               float64 `json:"pcBidWeight"`
	KeywordExtensionBidWeight float64 `json:"keywordExtensionBidWeight"`
	// ContentBidAmount is the bid on content placements; 0 when unset.
	ContentBidAmount float64  `json:"contentBidAmount"`
	Schedule         Schedule `json:"schedule"`
}

// Changes returns g's current settings as [AdGroupChanges], ready to be
// modified and passed to [AdGroupService.Update].
func (g *AdGroup) Changes() AdGroupChanges {
	ch := AdGroupChanges{
		Name:                      g.Name,
		Placements:                g.Placements,
		DeviceTypes:               g.DeviceTypes,
		UseKeywordExtension:       g.UseKeywordExtension,
		ExceptionKeywords:         g.ExceptionKeywords,
		BidAmount:                 int(g.BidAmount),
		DailyBudgetAmount:         int(g.DailyBudgetAmount),
		MobileBidWeight:           int(g.MobileBidWeight),
		PCBidWeight:               int(g.PCBidWeight),
		KeywordExtensionBidWeight: int(g.KeywordExtensionBidWeight),
		Schedule:                  g.Schedule,
	}
	if g.ContentBidAmount != 0 {
		v := int(g.ContentBidAmount)
		ch.ContentBidAmount = &v
	}
	return ch
}

// AdGroupListOptions are the optional parameters of [AdGroupService.List].
type AdGroupListOptions struct {
	// Configs filters by status; Kakao lists ON and OFF ad groups when
	// empty.
	Configs []Config
}

// List returns the ad groups of a campaign.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/ad-group
func (s *AdGroupService) List(ctx context.Context, adAccountID, campaignID int64, opts *AdGroupListOptions) ([]AdGroup, error) {
	if campaignID == 0 {
		return nil, errors.New("keywordad: campaign ID is required")
	}
	q := url.Values{"campaignId": {strconv.FormatInt(campaignID, 10)}}
	if opts != nil {
		for _, cfg := range opts.Configs {
			q.Add("config", string(cfg))
		}
	}
	var gs []AdGroup
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v1/adGroups", adAccountID: adAccountID, query: q}, &gs); err != nil {
		return nil, err
	}
	return gs, nil
}

// Get returns an ad group.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/ad-group
func (s *AdGroupService) Get(ctx context.Context, adAccountID, adGroupID int64) (*AdGroup, error) {
	var g AdGroup
	if err := s.c.do(ctx, call{method: http.MethodGet, path: idPath("/v1/adGroups/%d", adGroupID), adAccountID: adAccountID}, &g); err != nil {
		return nil, err
	}
	return &g, nil
}

// NewAdGroup describes an ad group to create. Zero optional fields take
// Kakao's defaults.
type NewAdGroup struct {
	CampaignID int64  `json:"campaignId"`
	Name       string `json:"name"`
	// Placements and DeviceTypes default to all.
	Placements  []Placement  `json:"placements,omitempty"`
	DeviceTypes []DeviceType `json:"deviceTypes,omitempty"`
	// UseKeywordExtension defaults to true.
	UseKeywordExtension *bool    `json:"useKeywordExtension,omitempty"`
	ExceptionKeywords   []string `json:"exceptionKeywords,omitempty"`
	// BidAmount is the default bid in KRW; Kakao's default is 300.
	BidAmount int `json:"bidAmount,omitempty"`
	// DailyBudgetAmount is in KRW; Kakao's default is 100000.
	DailyBudgetAmount int `json:"dailyBudgetAmount,omitempty"`
	// Bid weights are percentages; Kakao's default is 100.
	MobileBidWeight           int `json:"mobileBidWeight,omitempty"`
	PCBidWeight               int `json:"pcBidWeight,omitempty"`
	KeywordExtensionBidWeight int `json:"keywordExtensionBidWeight,omitempty"`
	// ContentBidAmount is the bid on content placements in KRW; unset by
	// default.
	ContentBidAmount int       `json:"contentBidAmount,omitempty"`
	Schedule         *Schedule `json:"schedule,omitempty"`
}

// Create creates an ad group. Kakao allows one request per 5 seconds.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/ad-group
func (s *AdGroupService) Create(ctx context.Context, adAccountID int64, g NewAdGroup) (*AdGroup, error) {
	if g.CampaignID == 0 || g.Name == "" {
		return nil, errors.New("keywordad: campaign ID and name are required")
	}
	var res AdGroup
	if err := s.c.do(ctx, call{method: http.MethodPost, path: "/v1/adGroups", adAccountID: adAccountID, body: g}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// AdGroupChanges is the full new state of an ad group. Kakao replaces
// every field, so start from [AdGroup.Changes] of the current ad group and
// modify what should change.
type AdGroupChanges struct {
	Name                string       `json:"name"`
	Placements          []Placement  `json:"placements"`
	DeviceTypes         []DeviceType `json:"deviceTypes"`
	UseKeywordExtension bool         `json:"useKeywordExtension"`
	// ExceptionKeywords is ignored when UseKeywordExtension is false.
	ExceptionKeywords []string `json:"exceptionKeywords,omitempty"`
	// BidAmount and DailyBudgetAmount are in KRW.
	BidAmount         int `json:"bidAmount"`
	DailyBudgetAmount int `json:"dailyBudgetAmount"`
	// Bid weights are percentages.
	MobileBidWeight           int `json:"mobileBidWeight"`
	PCBidWeight               int `json:"pcBidWeight"`
	KeywordExtensionBidWeight int `json:"keywordExtensionBidWeight"`
	// ContentBidAmount is the bid on content placements in KRW; nil leaves
	// it unset.
	ContentBidAmount *int     `json:"contentBidAmount,omitempty"`
	Schedule         Schedule `json:"schedule"`
}

// Update replaces an ad group's settings. Kakao allows one request per 5
// seconds.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/ad-group
func (s *AdGroupService) Update(ctx context.Context, adAccountID, adGroupID int64, ch AdGroupChanges) (*AdGroup, error) {
	if ch.Name == "" {
		return nil, errors.New("keywordad: ad group name is required")
	}
	var res AdGroup
	if err := s.c.do(ctx, call{method: http.MethodPut, path: idPath("/v1/adGroups/%d", adGroupID), adAccountID: adAccountID, body: ch}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// SetConfig turns an ad group ON or OFF. It only works while the ad group
// is ON or OFF. Kakao allows one request per second.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/ad-group
func (s *AdGroupService) SetConfig(ctx context.Context, adAccountID, adGroupID int64, cfg Config) error {
	body := struct {
		Config Config `json:"config"`
	}{cfg}
	return s.c.do(ctx, call{method: http.MethodPatch, path: idPath("/v1/adGroups/%d/onOff", adGroupID), adAccountID: adAccountID, body: body}, nil)
}

// UpdateBidAmount sets an ad group's default bid in KRW. Kakao allows one
// request per second.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/ad-group
func (s *AdGroupService) UpdateBidAmount(ctx context.Context, adAccountID, adGroupID int64, bidAmount int) error {
	body := struct {
		BidAmount int `json:"bidAmount"`
	}{bidAmount}
	return s.c.do(ctx, call{method: http.MethodPatch, path: idPath("/v1/adGroups/%d/bidAmount", adGroupID), adAccountID: adAccountID, body: body}, nil)
}

// UpdateDailyBudget sets an ad group's daily budget in KRW. Kakao allows
// one request per second.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/ad-group
func (s *AdGroupService) UpdateDailyBudget(ctx context.Context, adAccountID, adGroupID int64, dailyBudgetAmount int) error {
	body := struct {
		DailyBudgetAmount int `json:"dailyBudgetAmount"`
	}{dailyBudgetAmount}
	return s.c.do(ctx, call{method: http.MethodPatch, path: idPath("/v1/adGroups/%d/dailyBudget", adGroupID), adAccountID: adAccountID, body: body}, nil)
}

// BidWeights are an ad group's bid weights in percent, e.g. 150 for 150%.
type BidWeights struct {
	Mobile           int `json:"mobileBidWeight"`
	PC               int `json:"pcBidWeight"`
	KeywordExtension int `json:"keywordExtensionBidWeight"`
}

// UpdateBidWeights sets an ad group's mobile, PC and keyword extension bid
// weights. Kakao allows one request per second.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/ad-group
func (s *AdGroupService) UpdateBidWeights(ctx context.Context, adAccountID, adGroupID int64, w BidWeights) error {
	return s.c.do(ctx, call{method: http.MethodPatch, path: idPath("/v1/adGroups/%d/bidWeight", adGroupID), adAccountID: adAccountID, body: w}, nil)
}

// UpdateContentBidAmount sets an ad group's bid on content placements in
// KRW. A contentBidAmount of 0 unsets it (sent as null). Kakao allows one
// request per second.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/ad-group
func (s *AdGroupService) UpdateContentBidAmount(ctx context.Context, adAccountID, adGroupID int64, contentBidAmount int) error {
	body := struct {
		ContentBidAmount *int `json:"contentBidAmount"`
	}{}
	if contentBidAmount != 0 {
		body.ContentBidAmount = &contentBidAmount
	}
	return s.c.do(ctx, call{method: http.MethodPatch, path: idPath("/v1/adGroups/%d/contentBidAmount", adGroupID), adAccountID: adAccountID, body: body}, nil)
}

// Delete deletes an ad group.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/ad-group
func (s *AdGroupService) Delete(ctx context.Context, adAccountID, adGroupID int64) error {
	return s.c.do(ctx, call{method: http.MethodDelete, path: idPath("/v1/adGroups/%d", adGroupID), adAccountID: adAccountID}, nil)
}
