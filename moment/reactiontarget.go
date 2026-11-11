package moment

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	kakao "github.com/donghquinn/gokakao"
)

// AudienceAll is the audience type of an ad reaction target usable in both
// display and message ads (전체).
const AudienceAll AudienceType = "ALL"

// ReactionTargetStatus is the population state of an ad reaction target
// (타겟 모수 상태).
type ReactionTargetStatus string

// Ad reaction target statuses.
const (
	ReactionTargetWaiting        ReactionTargetStatus = "WAITING"         // 준비 중
	ReactionTargetAvailableError ReactionTargetStatus = "AVAILABLE_ERROR" // 모수 추출 에러
	ReactionTargetAvailable      ReactionTargetStatus = "AVAILABLE"       // 모수 추출 완료
	ReactionTargetModifying      ReactionTargetStatus = "MODIFYING"       // 수정 중
	ReactionTargetDelete         ReactionTargetStatus = "DELETE"          // 삭제 또는 삭제중
	ReactionTargetError          ReactionTargetStatus = "ERROR"           // 그 외 에러
)

// ReactionTarget is an ad reaction target (광고반응타겟): users who
// reacted to ads, usable as a cohort targeting of ad groups.
type ReactionTarget struct {
	ID          kakao.Int64 `json:"id"`
	AdAccountID kakao.Int64 `json:"adAccountId"`
	Name        string      `json:"name"`
	// AudienceType is AudienceDisplay, AudienceMessage or AudienceAll.
	AudienceType AudienceType     `json:"audienceType"`
	BaseAds      []ReactionBaseAd `json:"baseAds"`
	// CollectDuration is the number of days of reactions used, counted
	// back from today.
	CollectDuration int                  `json:"collectDuration"`
	CohortStatus    ReactionTargetStatus `json:"cohortStatus"`
	// PopulationScore is the estimated reach (타겟 모수), valid when
	// CohortStatus is AVAILABLE. It is also read from the misspelled
	// "popultaionScore" of Kakao's field table.
	PopulationScore      kakao.Int64 `json:"populationScore"`
	PopulationUpdateDate kakao.Time  `json:"populationUpdateDate"`
	CreatedDate          kakao.Time  `json:"createdDate"`
	LastModifiedDate     kakao.Time  `json:"lastModifiedDate"`
}

// UnmarshalJSON implements [json.Unmarshaler].
func (r *ReactionTarget) UnmarshalJSON(b []byte) error {
	type plain ReactionTarget
	var v struct {
		plain
		Misspelled kakao.Int64 `json:"popultaionScore"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*r = ReactionTarget(v.plain)
	if r.PopulationScore == 0 {
		r.PopulationScore = v.Misspelled
	}
	return nil
}

// ReactionBaseAd is a reaction an ad reaction target is built from. See
// the docs for the allowed Operation and indicator combinations.
type ReactionBaseAd struct {
	BaseAd
	// ThirdIndicator is the third reaction of a MINUS operation, e.g.
	// OPEN - CLICK - PLAY.
	ThirdIndicator Indicator `json:"thirdIndicator"`
}

// NewBaseAd returns the [NewReactionBaseAd] that keeps a as it is.
func (a *ReactionBaseAd) NewBaseAd() NewReactionBaseAd {
	n := NewReactionBaseAd{
		Operation:       a.Operation,
		FirstIndicator:  a.FirstIndicator,
		SecondIndicator: a.SecondIndicator,
		ThirdIndicator:  a.ThirdIndicator,
	}
	if a.Campaign != nil {
		n.CampaignID = int64(a.Campaign.ID)
		n.CampaignType = a.Campaign.CampaignTypeGoal.CampaignType
	}
	if a.AdGroup != nil {
		n.AdGroupID = int64(a.AdGroup.ID)
	}
	if a.MessageAd != nil {
		n.MessageAdID = a.MessageAd.ID
	}
	return n
}

// NewReactionBaseAd is a reaction to build an ad reaction target from, for
// [NewReactionTarget] and [ReactionTargetChanges].
type NewReactionBaseAd struct {
	// CampaignID and AdGroupID select the ad group; see
	// [ReactionTargetService.Creatables].
	CampaignID int64
	// CampaignType is optional; Kakao's example sends it.
	CampaignType CampaignType
	AdGroupID    int64
	// MessageAdID selects a message (e.g. "msg-ad-...") instead of an ad
	// group, as Kakao's responses show; it is undocumented for requests.
	MessageAdID string
	// Operation and FirstIndicator are required; SecondIndicator and
	// ThirdIndicator combine further reactions, e.g. MINUS CLICK PLAY
	// (클릭-재생).
	Operation       Operation
	FirstIndicator  Indicator
	SecondIndicator Indicator
	ThirdIndicator  Indicator
}

// MarshalJSON implements [json.Marshaler], nesting the IDs as
// {"campaign": {"id": ...}, "adGroup": {"id": ...}}.
func (a NewReactionBaseAd) MarshalJSON() ([]byte, error) {
	type typeGoal struct {
		CampaignType CampaignType `json:"campaignType"`
	}
	type campaignRef struct {
		ID               int64     `json:"id"`
		CampaignTypeGoal *typeGoal `json:"campaignTypeGoal,omitempty"`
	}
	type adGroupRef struct {
		ID int64 `json:"id"`
	}
	type messageAdRef struct {
		ID string `json:"id"`
	}
	v := struct {
		Campaign        *campaignRef  `json:"campaign,omitempty"`
		AdGroup         *adGroupRef   `json:"adGroup,omitempty"`
		MessageAd       *messageAdRef `json:"messageAd,omitempty"`
		Operation       Operation     `json:"operation"`
		FirstIndicator  Indicator     `json:"firstIndicator"`
		SecondIndicator Indicator     `json:"secondIndicator,omitempty"`
		ThirdIndicator  Indicator     `json:"thirdIndicator,omitempty"`
	}{Operation: a.Operation, FirstIndicator: a.FirstIndicator, SecondIndicator: a.SecondIndicator, ThirdIndicator: a.ThirdIndicator}
	if a.CampaignID != 0 {
		v.Campaign = &campaignRef{ID: a.CampaignID}
		if a.CampaignType != "" {
			v.Campaign.CampaignTypeGoal = &typeGoal{a.CampaignType}
		}
	}
	if a.AdGroupID != 0 {
		v.AdGroup = &adGroupRef{a.AdGroupID}
	}
	if a.MessageAdID != "" {
		v.MessageAd = &messageAdRef{a.MessageAdID}
	}
	return json.Marshal(v)
}

func validateBaseAds(ads []NewReactionBaseAd) error {
	if len(ads) == 0 {
		return errors.New("moment: ad reaction target base ads are required")
	}
	for _, a := range ads {
		if (a.CampaignID == 0 || a.AdGroupID == 0) && a.MessageAdID == "" {
			return errors.New("moment: each base ad needs a campaign and ad group ID or a message ad ID")
		}
		if a.Operation == "" || a.FirstIndicator == "" {
			return errors.New("moment: each base ad needs an operation and first indicator")
		}
	}
	return nil
}

// List returns the ad account's ad reaction targets.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/target-reaction-mgmt
func (s *ReactionTargetService) List(ctx context.Context, adAccountID int64) ([]ReactionTarget, error) {
	if adAccountID == 0 {
		return nil, errors.New("moment: ad account ID is required")
	}
	var res []ReactionTarget
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v4/targetings/cohort/list", adAccountID: adAccountID}, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// Get returns an ad reaction target.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/target-reaction-mgmt
func (s *ReactionTargetService) Get(ctx context.Context, adAccountID, reactionTargetID int64) (*ReactionTarget, error) {
	if adAccountID == 0 || reactionTargetID == 0 {
		return nil, errors.New("moment: ad account ID and ad reaction target ID are required")
	}
	var r ReactionTarget
	if err := s.c.do(ctx, call{method: http.MethodGet, path: idPath("/v4/targetings/cohort/%d", reactionTargetID), adAccountID: adAccountID}, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// ReactionTargetCreatablesQuery selects the campaigns for
// [ReactionTargetService.Creatables]. All fields are required.
type ReactionTargetCreatablesQuery struct {
	CampaignType CampaignType
	Goal         Goal
	// SearchKeyword matches campaign names.
	SearchKeyword string
}

// CreatableCampaign is a campaign with the ad groups an ad reaction target
// can be built from.
type CreatableCampaign struct {
	Campaign Campaign
	AdGroups []AdGroup
}

// UnmarshalJSON implements [json.Unmarshaler]. It reads "adGroup" as an
// array, as Kakao's example sends it, or as a single object, as its field
// table documents.
func (c *CreatableCampaign) UnmarshalJSON(b []byte) error {
	var v struct {
		Campaign Campaign        `json:"campaign"`
		AdGroup  json.RawMessage `json:"adGroup"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*c = CreatableCampaign{Campaign: v.Campaign}
	raw := bytes.TrimSpace(v.AdGroup)
	switch {
	case len(raw) == 0 || bytes.Equal(raw, []byte("null")):
	case raw[0] == '[':
		return json.Unmarshal(raw, &c.AdGroups)
	default:
		var g AdGroup
		if err := json.Unmarshal(raw, &g); err != nil {
			return err
		}
		c.AdGroups = []AdGroup{g}
	}
	return nil
}

// Creatables returns the campaigns of a type and goal, and their ad groups,
// that an ad reaction target can be built from.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/target-reaction-mgmt
func (s *ReactionTargetService) Creatables(ctx context.Context, adAccountID int64, q ReactionTargetCreatablesQuery) ([]CreatableCampaign, error) {
	if adAccountID == 0 {
		return nil, errors.New("moment: ad account ID is required")
	}
	if q.CampaignType == "" || q.Goal == "" || q.SearchKeyword == "" {
		return nil, errors.New("moment: campaign type, goal and search keyword are required")
	}
	v := url.Values{
		"campaignType":  {string(q.CampaignType)},
		"goal":          {string(q.Goal)},
		"searchKeyword": {q.SearchKeyword},
	}
	var res []CreatableCampaign
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v4/targetings/cohort/creatables", adAccountID: adAccountID, query: v}, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// NewReactionTarget describes an ad reaction target to create.
type NewReactionTarget struct {
	// Name is up to 50 characters.
	Name string `json:"name"`
	// AudienceType is optional; Kakao's example sends it.
	AudienceType AudienceType        `json:"audienceType,omitempty"`
	BaseAds      []NewReactionBaseAd `json:"baseAds"`
}

// Create creates an ad reaction target. An account holds up to 50, and
// Kakao extracts the population within an hour. Kakao allows one call per
// second.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/target-reaction-mgmt
func (s *ReactionTargetService) Create(ctx context.Context, adAccountID int64, r NewReactionTarget) (*ReactionTarget, error) {
	if adAccountID == 0 {
		return nil, errors.New("moment: ad account ID is required")
	}
	if r.Name == "" {
		return nil, errors.New("moment: ad reaction target name is required")
	}
	if err := validateBaseAds(r.BaseAds); err != nil {
		return nil, err
	}
	var res ReactionTarget
	if err := s.c.do(ctx, call{method: http.MethodPost, path: "/v4/targetings/cohort", adAccountID: adAccountID, body: r}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Rename changes the name of an ad reaction target, up to 50 characters
// and unique in the account. Deleted targets cannot be renamed. The API
// reference allows one call per second; the endpoint docs say one per 10
// seconds per user.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/target-reaction-mgmt
func (s *ReactionTargetService) Rename(ctx context.Context, adAccountID, reactionTargetID int64, name string) (*ReactionTarget, error) {
	if adAccountID == 0 || reactionTargetID == 0 || name == "" {
		return nil, errors.New("moment: ad account ID, ad reaction target ID and name are required")
	}
	body := struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}{reactionTargetID, name}
	var res ReactionTarget
	if err := s.c.do(ctx, call{method: http.MethodPut, path: "/v4/targetings/cohort/name", adAccountID: adAccountID, body: body}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// ReactionTargetChanges is the new data of an ad reaction target for
// [ReactionTargetService.Update].
//
// Kakao replaces the base ads: include the ones to keep. Start from
// [ReactionTarget.Changes].
type ReactionTargetChanges struct {
	BaseAds []NewReactionBaseAd `json:"baseAds"`
}

// Changes returns the [ReactionTargetChanges] that keep r as it is, for
// editing before [ReactionTargetService.Update].
func (r *ReactionTarget) Changes() ReactionTargetChanges {
	ads := make([]NewReactionBaseAd, len(r.BaseAds))
	for i := range r.BaseAds {
		ads[i] = r.BaseAds[i].NewBaseAd()
	}
	return ReactionTargetChanges{BaseAds: ads}
}

// Update replaces the base ads of an ad reaction target; see
// [ReactionTargetChanges]. The name cannot be changed here; use
// [ReactionTargetService.Rename]. Until Kakao finishes, within an hour, ad
// groups keep targeting the previous users.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/target-reaction-mgmt
func (s *ReactionTargetService) Update(ctx context.Context, adAccountID, reactionTargetID int64, r ReactionTargetChanges) (*ReactionTarget, error) {
	if adAccountID == 0 || reactionTargetID == 0 {
		return nil, errors.New("moment: ad account ID and ad reaction target ID are required")
	}
	if err := validateBaseAds(r.BaseAds); err != nil {
		return nil, err
	}
	body := struct {
		ID int64 `json:"id"`
		ReactionTargetChanges
	}{reactionTargetID, r}
	var res ReactionTarget
	if err := s.c.do(ctx, call{method: http.MethodPut, path: "/v4/targetings/cohort", adAccountID: adAccountID, body: body}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Delete deletes an ad reaction target. Deleted or used targets cannot be
// deleted. Kakao allows one call per second.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/target-reaction-mgmt
func (s *ReactionTargetService) Delete(ctx context.Context, adAccountID, reactionTargetID int64) error {
	if adAccountID == 0 || reactionTargetID == 0 {
		return errors.New("moment: ad account ID and ad reaction target ID are required")
	}
	return s.c.do(ctx, call{method: http.MethodDelete, path: idPath("/v4/targetings/cohort/%d", reactionTargetID), adAccountID: adAccountID}, nil)
}

// DeleteMany deletes several ad reaction targets under the conditions of
// [ReactionTargetService.Delete], reporting failures in the result. Kakao
// allows one call per second.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/target-reaction-mgmt
func (s *ReactionTargetService) DeleteMany(ctx context.Context, adAccountID int64, reactionTargetIDs []int64) (*BulkDeleteResult, error) {
	return deleteMany(ctx, s.c, adAccountID, "/v4/targetings/cohort", "cohortIds", reactionTargetIDs)
}

// Usages returns the ad groups, with their campaigns, and messages that use
// an ad reaction target (사용 현황).
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/target-reaction-mgmt
func (s *ReactionTargetService) Usages(ctx context.Context, adAccountID, reactionTargetID int64) ([]AdGroupAndCampaign, error) {
	if adAccountID == 0 || reactionTargetID == 0 {
		return nil, errors.New("moment: ad account ID and ad reaction target ID are required")
	}
	var res []AdGroupAndCampaign
	if err := s.c.do(ctx, call{method: http.MethodGet, path: idPath("/v4/targetings/cohort/usages/%d", reactionTargetID), adAccountID: adAccountID}, &res); err != nil {
		return nil, err
	}
	return res, nil
}
