package moment

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	kakao "github.com/donghquinn/gokakao"
)

// Ad groups (광고그룹) sit under a campaign. Display, message (카카오톡 채널 X
// 도달) and personalized message (개인화 메시지 X 도달) ad groups share the
// same endpoints but take different request bodies, so each has its own
// create and update method.

// AdGroupBidStrategy is how an ad group bids (입찰 방식).
type AdGroupBidStrategy string

// Ad group bid strategies.
const (
	BidStrategyManual  AdGroupBidStrategy = "MANUAL"  // 수동
	BidStrategyAutoBid AdGroupBidStrategy = "AUTOBID" // 자동입찰
	// BidStrategyOptimization is only returned, never requested.
	BidStrategyOptimization AdGroupBidStrategy = "OPTIMIZATION" // 목적 최적화
)

// OptimizationStatuses is the optimization status of an ad group. The docs
// type it as an array of [OptimizationStatus]; a single string is also
// accepted.
type OptimizationStatuses []OptimizationStatus

// UnmarshalJSON implements [json.Unmarshaler].
func (s *OptimizationStatuses) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && b[0] == '"' {
		var v OptimizationStatus
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}
		*s = nil
		if v != "" {
			*s = OptimizationStatuses{v}
		}
		return nil
	}
	var v []OptimizationStatus
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*s = v
	return nil
}

// AdGroupSummary is an ad group in [AdGroupService.List].
type AdGroupSummary struct {
	ID     kakao.Int64 `json:"id"`
	Name   string      `json:"name"`
	Config Config      `json:"config"`
	// UserConfig is the deprecated predecessor of Config, kept for
	// history.
	UserConfig   Config       `json:"userConfig"`
	SystemConfig SystemConfig `json:"systemConfig"`
}

// AdGroupDetail is an ad group (광고그룹) as returned by Get, create and
// update. Which fields are set depends on the kind of ad group: display ad
// groups have pacing, bidding and a daily budget; message ad groups have
// MessageSendingInfo, ProfileID and a total budget; personalized message
// ad groups only have the basic fields. Amounts are in KRW.
type AdGroupDetail struct {
	ID                kakao.Int64        `json:"id"`
	Name              string             `json:"name"`
	Config            Config             `json:"config"`
	Pacing            Pacing             `json:"pacing"`
	PricingType       PricingType        `json:"pricingType"`
	BidAmount         kakao.Float        `json:"bidAmount"`
	BidStrategy       AdGroupBidStrategy `json:"bidStrategy"`
	BidStrategyTarget *BidStrategyTarget `json:"bidStrategyTarget"`
	// StatusDescription describes the current delivery state, e.g. 운영중.
	StatusDescription string          `json:"statusDescription"`
	Status            []AdGroupStatus `json:"status"`
	// OptimizationStatus is set only when the ad group has one.
	OptimizationStatus OptimizationStatuses `json:"optimizationStatus"`
	DeviceTypes        []DeviceType         `json:"deviceTypes"`
	// AdServingCategories are the network sub-placements.
	AdServingCategories []NetworkCategoryCode `json:"adServingCategories"`
	// SectionCategories are KakaoTalk sections, for bizboard, display,
	// video and product catalog campaigns only.
	SectionCategories []TalkSectionCode `json:"sectionCategories"`
	Placements        []Placement       `json:"placements"`
	Targeting         *AdGroupTargeting `json:"targeting"`
	// Schedule includes BeginTime and EndTime for message ad groups.
	Schedule               *MessageSchedule    `json:"schedule"`
	Campaign               *CampaignDetail     `json:"campaign"`
	MessageSendingInfo     *MessageSendingInfo `json:"messageSendingInfo"`
	ProfileID              string              `json:"profileId"` // 카카오톡 채널 프로필 ID
	UseWifiOnly            bool                `json:"useWifiOnly"`
	CreativeCount          kakao.Int64         `json:"creativeCount"`
	SystemConfig           SystemConfig        `json:"systemConfig"`
	AllAvailableDeviceType bool                `json:"allAvailableDeviceType"`
	AllAvailablePlacement  bool                `json:"allAvailablePlacement"`
	Adult                  bool                `json:"adult"` // 성인 타게팅
	// TotalBudget is the period budget; for message ad groups it is the
	// purchase amount (구매 금액) and TotalBudgetWithVAT includes VAT.
	TotalBudget           kakao.Float `json:"totalBudget"`
	TotalBudgetWithVAT    kakao.Float `json:"totalBudgetWithVAT"`
	DailyBudgetAmount     kakao.Float `json:"dailyBudgetAmount"`
	DailyBudgetAmountOver bool        `json:"isDailyBudgetAmountOver"`
	ValidPeriod           bool        `json:"isValidPeriod"` // 집행기간 유효 여부
	CreatedDate           kakao.Time  `json:"createdDate"`
	LastModifiedDate      kakao.Time  `json:"lastModifiedDate"`
}

// campaignID returns the ID of g's campaign, or 0.
func (g *AdGroupDetail) campaignID() int64 {
	if g.Campaign == nil {
		return 0
	}
	return int64(g.Campaign.ID)
}

// DisplayChanges returns g's current settings as [DisplayAdGroupChanges],
// ready to be modified and passed to [AdGroupService.UpdateDisplay].
func (g *AdGroupDetail) DisplayChanges() DisplayAdGroupChanges {
	ch := DisplayAdGroupChanges{
		CampaignID:             g.campaignID(),
		Name:                   g.Name,
		Placements:             g.Placements,
		AdServingCategories:    g.AdServingCategories,
		SectionCategories:      g.SectionCategories,
		AllAvailableDeviceType: g.AllAvailableDeviceType,
		AllAvailablePlacement:  g.AllAvailablePlacement,
		DeviceTypes:            g.DeviceTypes,
		Targeting:              g.Targeting.Spec(),
		Adult:                  g.Adult,
		DailyBudgetAmount:      float64(g.DailyBudgetAmount),
		BidStrategy:            g.BidStrategy,
		PricingType:            g.PricingType,
		BidAmount:              float64(g.BidAmount),
		Pacing:                 g.Pacing,
	}
	if g.BidStrategyTarget != nil {
		t := *g.BidStrategyTarget
		ch.BidStrategyTarget = &t
	}
	if s := g.Schedule; s != nil {
		ch.Schedule = Schedule{
			BeginDate:     s.BeginDate,
			EndDate:       s.EndDate,
			LateNight:     s.LateNight,
			DetailTime:    s.DetailTime,
			MondayTime:    s.MondayTime,
			TuesdayTime:   s.TuesdayTime,
			WednesdayTime: s.WednesdayTime,
			ThursdayTime:  s.ThursdayTime,
			FridayTime:    s.FridayTime,
			SaturdayTime:  s.SaturdayTime,
			SundayTime:    s.SundayTime,
		}
	}
	return ch
}

// MessageChanges returns g's current settings as [MessageAdGroupChanges],
// ready to be modified and passed to [AdGroupService.UpdateMessage].
func (g *AdGroupDetail) MessageChanges() MessageAdGroupChanges {
	ch := MessageAdGroupChanges{
		CampaignID:             g.campaignID(),
		Name:                   g.Name,
		AllAvailableDeviceType: g.AllAvailableDeviceType,
		DeviceTypes:            g.DeviceTypes,
		Targeting:              g.Targeting.Spec(),
		DailyBudgetAmount:      float64(g.DailyBudgetAmount),
	}
	if g.MessageSendingInfo != nil {
		ch.MessageSendingInfo = *g.MessageSendingInfo
		ch.MessageSendingInfo.StartedAt = ""
		ch.MessageSendingInfo.FinishedAt = ""
	}
	if g.Schedule != nil {
		ch.Schedule = *g.Schedule
	}
	return ch
}

// AdGroupTargeting is the targeting of an ad group as returned by the ad
// group APIs. Use [AdGroupTargeting.Spec] to turn it back into a request.
type AdGroupTargeting struct {
	// ID and Name identify a saved audience; unset for NORMAL.
	ID           kakao.Int64   `json:"id"`
	Name         string        `json:"name"`
	Type         AudienceType  `json:"type"`
	AgeType      SelectionType `json:"ageType"`
	Ages         []string      `json:"ages"`
	GenderType   SelectionType `json:"genderType"`
	Genders      []Gender      `json:"genders"`
	LocationType LocationType  `json:"locationType"`
	// Locations, Depth2Locations and Depth3Locations are 시/도, 시/군/구
	// and 동/읍/면; Deprecated marks removed districts.
	Locations       []Location       `json:"locations"`
	Depth2Locations []Depth2Location `json:"depth2Locations"`
	Depth3Locations []Depth3Location `json:"depth3Locations"`
	// UfoInterests and UfoBusinessTypes are category IDs from
	// [TargetingService.Categories]; the Deprecated lists name the ones
	// that are no longer offered.
	UfoInterests               []string                    `json:"ufoInterests"`
	UfoDeprecatedInterests     []string                    `json:"ufoDeprecatedInterests"`
	UfoBusinessTypes           []string                    `json:"ufoBusinessTypes"`
	UfoDeprecatedBusinessTypes []string                    `json:"ufoDeprecatedBusinessTypes"`
	CustomerDataTargetings     []CustomerDataTargeting     `json:"customerDataTargetings"`
	CustomerFileTargetings     []CustomerFileTargeting     `json:"customerFileTargetings"`
	TrackerTargetings          []TrackerTargeting          `json:"trackerTargetings"`
	CohortTargetings           []CohortTargeting           `json:"cohortTargetings"`
	TalkChannelTargetings      []TalkChannelTargeting      `json:"talkChannelTargetings"`
	SyncAppTargetings          []SyncAppTargeting          `json:"syncAppTargetings"`
	TalkChannelGroupTargetings []TalkChannelGroupTargeting `json:"talkChannelGroupTargetings"`
}

// Spec returns t as a request targeting. A nil t returns the zero spec.
func (t *AdGroupTargeting) Spec() AdGroupTargetingSpec {
	if t == nil {
		return AdGroupTargetingSpec{}
	}
	s := AdGroupTargetingSpec{
		ID:               int64(t.ID),
		Type:             t.Type,
		AgeType:          t.AgeType,
		Ages:             t.Ages,
		GenderType:       t.GenderType,
		Genders:          t.Genders,
		UfoInterests:     t.UfoInterests,
		UfoBusinessTypes: t.UfoBusinessTypes,
		LocationType:     t.LocationType,
		Locations:        t.Locations,
		Depth2Locations:  t.Depth2Locations,
		Depth3Locations:  t.Depth3Locations,
	}
	for _, v := range t.CustomerDataTargetings {
		s.CustomerDataTargetings = append(s.CustomerDataTargetings, CustomerDataTarget{CustomerDataID: int64(v.ID), InclusionType: v.InclusionType})
	}
	for _, v := range t.CustomerFileTargetings {
		s.CustomerFileTargetings = append(s.CustomerFileTargetings, CustomerFileTarget{CustomerFileID: int64(v.ID), InclusionType: v.InclusionType})
	}
	for _, v := range t.TrackerTargetings {
		s.TrackerTargetings = append(s.TrackerTargetings, TrackerTarget{
			TrackID: v.TrackID, InclusionType: v.InclusionType, EventCode: v.EventCode,
			TrackRuleID: v.TrackRuleID, TrackRuleName: v.TrackRuleName, Term: v.Term,
		})
	}
	for _, v := range t.CohortTargetings {
		id := v.CohortID
		if id == 0 {
			id = v.ID
		}
		s.CohortTargetings = append(s.CohortTargetings, CohortTarget{CohortID: int64(id), InclusionType: v.InclusionType})
	}
	for _, v := range t.TalkChannelTargetings {
		s.TalkChannelTargetings = append(s.TalkChannelTargetings, TalkChannelTarget{ProfileID: strconv.FormatInt(int64(v.ProfileID), 10), InclusionType: v.InclusionType})
	}
	s.SyncAppTargetings = t.SyncAppTargetings
	for _, v := range t.TalkChannelGroupTargetings {
		s.TalkChannelGroupTargetings = append(s.TalkChannelGroupTargetings, TalkChannelGroupTarget{
			TalkChannelGroupFileID: int64(v.TalkChannelGroupFileID), Name: v.Name,
			InclusionType: v.InclusionType, FileType: v.FileType, GroupKey: v.GroupKey,
		})
	}
	return s
}

// AdGroupTargetingSpec is the targeting of an ad group to create or update.
// With Type DISPLAY or MESSAGE (a saved audience), set only ID and Type and
// leave the rest empty. Customer files and friend groups are being retired;
// use customer data for new targets.
type AdGroupTargetingSpec struct {
	// ID is the audience to use, with Type DISPLAY or MESSAGE.
	ID int64 `json:"id,omitempty"`
	// Type is NORMAL by default; display ad groups using an audience need
	// DISPLAY.
	Type AudienceType `json:"type,omitempty"`
	// AgeType must be NOT_ALL for adult ad groups.
	AgeType SelectionType `json:"ageType,omitempty"`
	// Ages is required for NOT_ALL and not allowed for ALL: "15" (15~19),
	// "20", "25", ... "65" in five-year bands, "70" (70 이상). Adult ad
	// groups need "20" through "70".
	Ages       []string      `json:"ages,omitempty"`
	GenderType SelectionType `json:"genderType,omitempty"`
	// Genders is required for NOT_ALL and not allowed for ALL.
	Genders []Gender `json:"genders,omitempty"`
	// UfoInterests and UfoBusinessTypes are category IDs from
	// [TargetingService.Categories].
	UfoInterests     []string `json:"ufoInterests,omitempty"`
	UfoBusinessTypes []string `json:"ufoBusinessTypes,omitempty"`
	// LocationType is domestic and overseas when empty. ALL requires every
	// province code and U (알 수 없음) in Locations, without Z (해외); AREA
	// requires at least one location of any depth.
	LocationType    LocationType     `json:"locationType,omitempty"`
	Locations       []Location       `json:"locations,omitempty"`
	Depth2Locations []Depth2Location `json:"depth2Locations,omitempty"`
	Depth3Locations []Depth3Location `json:"depth3Locations,omitempty"`
	// The custom targets come from the TargetingService availability
	// lists.
	CustomerDataTargetings []CustomerDataTarget `json:"customerDataTargetings,omitempty"`
	CustomerFileTargetings []CustomerFileTarget `json:"customerFileTargetings,omitempty"`
	TrackerTargetings      []TrackerTarget      `json:"trackerTargetings,omitempty"`
	CohortTargetings       []CohortTarget       `json:"cohortTargetings,omitempty"`
	TalkChannelTargetings  []TalkChannelTarget  `json:"talkChannelTargetings,omitempty"`
	SyncAppTargetings      []SyncAppTargeting   `json:"syncAppTargetings,omitempty"`
	// TalkChannelGroupTargetings is only for KakaoTalk Channel reach
	// campaigns.
	TalkChannelGroupTargetings []TalkChannelGroupTarget `json:"talkChannelGroupTargetings,omitempty"`
}

// CustomerDataTarget targets customer data (고객데이터) from
// [TargetingService.CustomerData].
type CustomerDataTarget struct {
	CustomerDataID int64         `json:"customerDataId"`
	InclusionType  InclusionType `json:"inclusionType"`
}

// CustomerFileTarget targets a customer file (고객파일, being retired) from
// [TargetingService.CustomerFiles].
type CustomerFileTarget struct {
	CustomerFileID int64         `json:"customerFileId"`
	InclusionType  InclusionType `json:"inclusionType"`
}

// TrackerTarget targets a Pixel & SDK event from
// [TargetingService.TrackerEvents].
type TrackerTarget struct {
	TrackID       string        `json:"trackId"`
	InclusionType InclusionType `json:"inclusionType"`
	// EventCode is "*" for all events.
	EventCode     string `json:"eventCode"`
	TrackRuleID   string `json:"trackRuleId"`
	TrackRuleName string `json:"trackRuleName"`
	Term          int    `json:"term"` // target days, 1~180
}

// CohortTarget targets an ad reaction target (광고반응타겟) from
// [TargetingService.Cohorts].
type CohortTarget struct {
	CohortID      int64         `json:"cohortId"`
	InclusionType InclusionType `json:"inclusionType"`
}

// TalkChannelTarget targets the friends of a KakaoTalk Channel from
// [TargetingService.TalkChannels].
type TalkChannelTarget struct {
	ProfileID     string        `json:"profileId"`
	InclusionType InclusionType `json:"inclusionType"`
}

// TalkChannelGroupTarget targets a friend group (친구그룹, being retired)
// from [TargetingService.FriendGroups].
type TalkChannelGroupTarget struct {
	TalkChannelGroupFileID int64         `json:"talkChannelGroupFileId"`
	Name                   string        `json:"name"`
	InclusionType          InclusionType `json:"inclusionType"`
	// FileType is APP_USER_ID, PHONE_NUMBER or MESSAGE_RETARGET.
	FileType CustomerDataFileType `json:"fileType"`
	GroupKey string               `json:"groupKey"`
}

// idRef is a {"id": ...} reference such as the campaign of an ad group.
type idRef struct {
	ID int64 `json:"id"`
}

// AdGroupListOptions are the optional parameters of [AdGroupService.List].
type AdGroupListOptions struct {
	// Configs filters by status; Kakao lists ON and OFF ad groups when
	// empty.
	Configs []Config
}

// List returns the ad groups of a campaign.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/ad-group
func (s *AdGroupService) List(ctx context.Context, adAccountID, campaignID int64, opts *AdGroupListOptions) ([]AdGroupSummary, error) {
	if adAccountID == 0 || campaignID == 0 {
		return nil, errors.New("moment: ad account ID and campaign ID are required")
	}
	q := url.Values{"campaignId": {strconv.FormatInt(campaignID, 10)}}
	if opts != nil {
		for _, cfg := range opts.Configs {
			q.Add("config", string(cfg))
		}
	}
	var res struct {
		Content []AdGroupSummary `json:"content"`
	}
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v4/adGroups", adAccountID: adAccountID, query: q}, &res); err != nil {
		return nil, err
	}
	return res.Content, nil
}

// Get returns an ad group of any kind.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/ad-group
func (s *AdGroupService) Get(ctx context.Context, adAccountID, adGroupID int64) (*AdGroupDetail, error) {
	if err := adGroupCheckIDs(adAccountID, adGroupID); err != nil {
		return nil, err
	}
	var g AdGroupDetail
	if err := s.c.do(ctx, call{method: http.MethodGet, path: idPath("/v4/adGroups/%d", adGroupID), adAccountID: adAccountID}, &g); err != nil {
		return nil, err
	}
	return &g, nil
}

// NewDisplayAdGroup describes a display ad group to create. See the docs
// for which campaign types allow each option and for the bid limits.
type NewDisplayAdGroup struct {
	CampaignID int64 `json:"-"`
	// Name is up to 50 characters; Kakao generates one when empty.
	Name       string      `json:"name,omitempty"`
	Placements []Placement `json:"placements"`
	// AdServingCategories are network sub-placements from
	// [TargetingService.AdServingCategories], required with
	// PlacementNetwork; bizboard campaigns need all of them.
	AdServingCategories []NetworkCategoryCode `json:"adServingCategories,omitempty"`
	// SectionCategories are KakaoTalk sections from
	// [TargetingService.SectionCategories], allowed with PlacementKakaoTalk
	// for bizboard, display, video and product catalog campaigns.
	SectionCategories []TalkSectionCode `json:"sectionCategories,omitempty"`
	// AllAvailableDeviceType and AllAvailablePlacement default to true.
	// AllAvailableDeviceType true needs every device in DeviceTypes.
	AllAvailableDeviceType *bool                `json:"allAvailableDeviceType,omitempty"`
	AllAvailablePlacement  *bool                `json:"allAvailablePlacement,omitempty"`
	DeviceTypes            []DeviceType         `json:"deviceTypes"`
	Targeting              AdGroupTargetingSpec `json:"targeting"`
	// Adult limits delivery to ages 20 and up; it cannot be turned off
	// later.
	Adult bool `json:"adult,omitempty"`
	// DailyBudgetAmount is 10,000 to 500,000,000 KRW in steps of 10.
	DailyBudgetAmount float64            `json:"dailyBudgetAmount"`
	BidStrategy       AdGroupBidStrategy `json:"bidStrategy"`
	PricingType       PricingType        `json:"pricingType"`
	// BidAmount is the manual maximum bid in KRW; 0 for automatic bidding.
	BidAmount         float64            `json:"bidAmount"`
	BidStrategyTarget *BidStrategyTarget `json:"bidStrategyTarget,omitempty"`
	// Pacing is NORMAL or QUICK for manual bidding and NONE for automatic
	// bidding.
	Pacing   Pacing   `json:"pacing"`
	Schedule Schedule `json:"schedule"`
}

func (g *NewDisplayAdGroup) validate() error {
	if g.CampaignID == 0 {
		return errors.New("moment: campaign ID is required")
	}
	if len(g.Placements) == 0 || len(g.DeviceTypes) == 0 {
		return errors.New("moment: placements and device types are required")
	}
	if g.DailyBudgetAmount <= 0 || g.BidStrategy == "" || g.PricingType == "" || g.Pacing == "" {
		return errors.New("moment: daily budget, bid strategy, pricing type and pacing are required")
	}
	if g.Schedule.BeginDate == "" {
		return errors.New("moment: schedule begin date is required")
	}
	return nil
}

// CreateDisplay creates a display ad group. Kakao allows one request per
// 5 seconds per user and ad account.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/ad-group
func (s *AdGroupService) CreateDisplay(ctx context.Context, adAccountID int64, g NewDisplayAdGroup) (*AdGroupDetail, error) {
	if adAccountID == 0 {
		return nil, errors.New("moment: ad account ID is required")
	}
	if err := g.validate(); err != nil {
		return nil, err
	}
	body := struct {
		Campaign idRef `json:"campaign"`
		NewDisplayAdGroup
		Type AdGroupType `json:"type"`
	}{idRef{g.CampaignID}, g, AdGroupTypeDisplay}
	return s.save(ctx, http.MethodPost, adAccountID, body)
}

// DisplayAdGroupChanges is the full new state of a display ad group.
//
// Kakao replaces the whole ad group: a field left empty is reset rather
// than kept. Start from [AdGroupDetail.DisplayChanges] of the current ad
// group and modify what should change.
type DisplayAdGroupChanges struct {
	CampaignID int64 `json:"-"`
	// Name keeps the current name when empty.
	Name                   string                `json:"name,omitempty"`
	Placements             []Placement           `json:"placements"`
	AdServingCategories    []NetworkCategoryCode `json:"adServingCategories,omitempty"`
	SectionCategories      []TalkSectionCode     `json:"sectionCategories,omitempty"`
	AllAvailableDeviceType bool                  `json:"allAvailableDeviceType"`
	AllAvailablePlacement  bool                  `json:"allAvailablePlacement"`
	DeviceTypes            []DeviceType          `json:"deviceTypes"`
	Targeting              AdGroupTargetingSpec  `json:"targeting"`
	// Adult cannot be changed back to false once set.
	Adult             bool               `json:"adult"`
	DailyBudgetAmount float64            `json:"dailyBudgetAmount"`
	BidStrategy       AdGroupBidStrategy `json:"bidStrategy"`
	PricingType       PricingType        `json:"pricingType"`
	BidAmount         float64            `json:"bidAmount"`
	BidStrategyTarget *BidStrategyTarget `json:"bidStrategyTarget,omitempty"`
	// Pacing can only be set for manual bidding; bizboard conversion
	// campaigns need NONE.
	Pacing   Pacing   `json:"pacing"`
	Schedule Schedule `json:"schedule"`
}

// UpdateDisplay replaces a display ad group's settings; see
// [DisplayAdGroupChanges]. Kakao allows one request per 5 seconds per user
// and ad account.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/ad-group
func (s *AdGroupService) UpdateDisplay(ctx context.Context, adAccountID, adGroupID int64, ch DisplayAdGroupChanges) (*AdGroupDetail, error) {
	if err := adGroupCheckIDs(adAccountID, adGroupID); err != nil {
		return nil, err
	}
	n := NewDisplayAdGroup{
		CampaignID: ch.CampaignID, Placements: ch.Placements, DeviceTypes: ch.DeviceTypes,
		DailyBudgetAmount: ch.DailyBudgetAmount, BidStrategy: ch.BidStrategy, PricingType: ch.PricingType,
		Pacing: ch.Pacing, Schedule: ch.Schedule,
	}
	if err := n.validate(); err != nil {
		return nil, err
	}
	body := struct {
		ID       int64 `json:"id"`
		Campaign idRef `json:"campaign"`
		DisplayAdGroupChanges
		Type AdGroupType `json:"type"`
	}{adGroupID, idRef{ch.CampaignID}, ch, AdGroupTypeDisplay}
	return s.save(ctx, http.MethodPut, adAccountID, body)
}

// UpdateDailyBudget sets a display ad group's daily budget: 10,000 to
// 500,000,000 KRW in steps of 10, and no more than the campaign's daily
// budget. Kakao allows one request per second per user and ad account.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/ad-group
func (s *AdGroupService) UpdateDailyBudget(ctx context.Context, adAccountID, adGroupID int64, amount float64) error {
	if err := adGroupCheckIDs(adAccountID, adGroupID); err != nil {
		return err
	}
	if amount < 10000 || amount > 500000000 {
		return errors.New("moment: daily budget must be between 10,000 and 500,000,000")
	}
	body := struct {
		ID                int64   `json:"id"`
		DailyBudgetAmount float64 `json:"dailyBudgetAmount"`
	}{adGroupID, amount}
	return s.c.do(ctx, call{method: http.MethodPut, path: "/v4/adGroups/dailyBudgetAmount", adAccountID: adAccountID, body: body}, nil)
}

// UpdateBidAmount sets the manual maximum bid of a display ad group in
// KRW. It cannot exceed 100,000 or half the daily budget, and only applies
// to display, video and bizboard campaigns other than reach. Kakao allows
// one request per second per user and ad account.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/ad-group
func (s *AdGroupService) UpdateBidAmount(ctx context.Context, adAccountID, adGroupID int64, bidAmount float64) error {
	if err := adGroupCheckIDs(adAccountID, adGroupID); err != nil {
		return err
	}
	if bidAmount <= 0 {
		return errors.New("moment: bid amount must be positive")
	}
	body := struct {
		ID        int64   `json:"id"`
		BidAmount float64 `json:"bidAmount"`
	}{adGroupID, bidAmount}
	return s.c.do(ctx, call{method: http.MethodPut, path: "/v4/adGroups/bidAmount", adAccountID: adAccountID, body: body}, nil)
}

// UpdatePacing sets the pacing of a display ad group to PacingNormal or
// PacingQuick. It only works for manual bidding in display, video and
// bizboard campaigns other than reach. Kakao allows one request per second
// per user and ad account.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/ad-group
func (s *AdGroupService) UpdatePacing(ctx context.Context, adAccountID, adGroupID int64, pacing Pacing) error {
	if err := adGroupCheckIDs(adAccountID, adGroupID); err != nil {
		return err
	}
	if pacing == "" {
		return errors.New("moment: pacing is required")
	}
	body := struct {
		ID     int64  `json:"id"`
		Pacing Pacing `json:"pacing"`
	}{adGroupID, pacing}
	return s.c.do(ctx, call{method: http.MethodPut, path: "/v4/adGroups/pacing", adAccountID: adAccountID, body: body}, nil)
}

// NewMessageAdGroup describes a message ad group to create under a
// KakaoTalk Channel reach campaign. The fixed fields are filled in:
// placements KAKAO_TALK, all available placements false, adult false,
// bid strategy MANUAL, pricing CPMS, pacing NONE and a bid amount equal to
// MessageSendingInfo.Price.
type NewMessageAdGroup struct {
	CampaignID int64 `json:"-"`
	// Name is up to 150 characters; Kakao generates one when empty.
	Name string `json:"name,omitempty"`
	// AllAvailableDeviceType true needs every device in DeviceTypes; it
	// must be true for LongTerm sends.
	AllAvailableDeviceType bool         `json:"allAvailableDeviceType"`
	DeviceTypes            []DeviceType `json:"deviceTypes"`
	// MessageSendingInfo.Status and SyncStatus default to SAVE and READY.
	// Price is 15 KRW without targeting and 20 KRW with it. ContractCount
	// cannot be below [AdGroupService.EstimatedAudienceSize].
	MessageSendingInfo MessageSendingInfo `json:"messageSendingInfo"`
	// Targeting with no options is ageType ALL, genderType ALL.
	Targeting AdGroupTargetingSpec `json:"targeting"`
	// Schedule needs BeginDate and BeginTime ("HH:mm:ss"); LongTerm sends
	// also need EndDate and EndTime.
	Schedule MessageSchedule `json:"schedule"`
}

// messageAdGroupBody is the request body of a message ad group.
type messageAdGroupBody struct {
	ID                     int64                `json:"id,omitempty"`
	Campaign               idRef                `json:"campaign"`
	Name                   string               `json:"name,omitempty"`
	Placements             []Placement          `json:"placements"`
	AllAvailableDeviceType bool                 `json:"allAvailableDeviceType"`
	AllAvailablePlacement  bool                 `json:"allAvailablePlacement"`
	DeviceTypes            []DeviceType         `json:"deviceTypes"`
	MessageSendingInfo     MessageSendingInfo   `json:"messageSendingInfo"`
	Targeting              AdGroupTargetingSpec `json:"targeting"`
	Adult                  bool                 `json:"adult"`
	DailyBudgetAmount      float64              `json:"dailyBudgetAmount,omitempty"`
	BidStrategy            AdGroupBidStrategy   `json:"bidStrategy"`
	PricingType            PricingType          `json:"pricingType"`
	BidAmount              float64              `json:"bidAmount"`
	Pacing                 Pacing               `json:"pacing"`
	Schedule               MessageSchedule      `json:"schedule"`
	Type                   AdGroupType          `json:"type"`
}

func newMessageAdGroupBody(campaignID int64, name string, allDevices bool, devices []DeviceType, info MessageSendingInfo, t AdGroupTargetingSpec, sch MessageSchedule) (*messageAdGroupBody, error) {
	if campaignID == 0 {
		return nil, errors.New("moment: campaign ID is required")
	}
	if len(devices) == 0 {
		return nil, errors.New("moment: device types are required")
	}
	if sch.BeginDate == "" {
		return nil, errors.New("moment: schedule begin date is required")
	}
	if info.Status == "" {
		info.Status = MessageSendSave
	}
	if info.SyncStatus == "" {
		info.SyncStatus = MessageSyncReady
	}
	return &messageAdGroupBody{
		Campaign:               idRef{campaignID},
		Name:                   name,
		Placements:             []Placement{PlacementKakaoTalk},
		AllAvailableDeviceType: allDevices,
		DeviceTypes:            devices,
		MessageSendingInfo:     info,
		Targeting:              t,
		BidStrategy:            BidStrategyManual,
		PricingType:            PricingCPMS,
		BidAmount:              float64(info.Price),
		Pacing:                 PacingNone,
		Schedule:               sch,
		Type:                   AdGroupTypeDirectMessage,
	}, nil
}

// CreateMessage creates a message ad group under a KakaoTalk Channel reach
// campaign. The purchase (VAT included) is charged to the ad account
// balance. Kakao allows one request per 5 seconds per user and ad account.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/ad-group
func (s *AdGroupService) CreateMessage(ctx context.Context, adAccountID int64, g NewMessageAdGroup) (*AdGroupDetail, error) {
	if adAccountID == 0 {
		return nil, errors.New("moment: ad account ID is required")
	}
	body, err := newMessageAdGroupBody(g.CampaignID, g.Name, g.AllAvailableDeviceType, g.DeviceTypes, g.MessageSendingInfo, g.Targeting, g.Schedule)
	if err != nil {
		return nil, err
	}
	return s.save(ctx, http.MethodPost, adAccountID, body)
}

// MessageAdGroupChanges is the full new state of a message ad group. The
// fixed fields are filled in as for [NewMessageAdGroup].
//
// Kakao replaces the whole ad group: a field left empty is reset rather
// than kept. Start from [AdGroupDetail.MessageChanges] of the current ad
// group and modify what should change.
type MessageAdGroupChanges struct {
	CampaignID int64
	// Name keeps the current name when empty.
	Name                   string
	AllAvailableDeviceType bool
	DeviceTypes            []DeviceType
	MessageSendingInfo     MessageSendingInfo
	Targeting              AdGroupTargetingSpec
	// DailyBudgetAmount is listed as required by the docs for this
	// request; it is omitted when 0.
	DailyBudgetAmount float64
	Schedule          MessageSchedule
}

// UpdateMessage replaces a message ad group's settings; see
// [MessageAdGroupChanges]. Kakao allows one request per 5 seconds per user
// and ad account.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/ad-group
func (s *AdGroupService) UpdateMessage(ctx context.Context, adAccountID, adGroupID int64, ch MessageAdGroupChanges) (*AdGroupDetail, error) {
	if err := adGroupCheckIDs(adAccountID, adGroupID); err != nil {
		return nil, err
	}
	body, err := newMessageAdGroupBody(ch.CampaignID, ch.Name, ch.AllAvailableDeviceType, ch.DeviceTypes, ch.MessageSendingInfo, ch.Targeting, ch.Schedule)
	if err != nil {
		return nil, err
	}
	body.ID = adGroupID
	body.DailyBudgetAmount = ch.DailyBudgetAmount
	return s.save(ctx, http.MethodPut, adAccountID, body)
}

// AudienceSizeQuery describes the targeting of a message ad group whose
// audience size to estimate.
type AudienceSizeQuery struct {
	CampaignTypeGoal CampaignTypeGoal `json:"campaignTypeGoal"`
	// Objective is required for KakaoTalk Channel reach campaigns: the
	// channel to send from.
	Objective *Objective `json:"objective,omitempty"`
	// Placements defaults to KAKAO_TALK.
	Placements             []Placement           `json:"placements"`
	AdServingCategories    []NetworkCategoryCode `json:"adServingCategories,omitempty"`
	AllAvailableDeviceType bool                  `json:"allAvailableDeviceType"`
	DeviceTypes            []DeviceType          `json:"deviceTypes"`
	Targeting              AdGroupTargetingSpec  `json:"targeting"`
}

// EstimatedAudienceSize returns the estimated audience (예상 모수) of a
// message ad group targeting, the basis for its purchased send count.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/ad-group
func (s *AdGroupService) EstimatedAudienceSize(ctx context.Context, adAccountID int64, q AudienceSizeQuery) (int64, error) {
	if adAccountID == 0 {
		return 0, errors.New("moment: ad account ID is required")
	}
	if q.CampaignTypeGoal.CampaignType == "" || q.CampaignTypeGoal.Goal == "" {
		return 0, errors.New("moment: campaign type and goal are required")
	}
	if len(q.Placements) == 0 {
		q.Placements = []Placement{PlacementKakaoTalk}
	}
	if q.DeviceTypes == nil {
		q.DeviceTypes = []DeviceType{}
	}
	body := struct {
		AudienceSizeQuery
		AllAvailablePlacement bool `json:"allAvailablePlacement"`
	}{q, false}
	var res struct {
		PopulationScore kakao.Int64 `json:"populationScore"`
	}
	if err := s.c.do(ctx, call{method: http.MethodPost, path: "/v4/targetings/populationScore", adAccountID: adAccountID, body: body}, &res); err != nil {
		return 0, err
	}
	return int64(res.PopulationScore), nil
}

// CancelContract terminates the contract of a message ad group (계약 해지).
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/ad-group
func (s *AdGroupService) CancelContract(ctx context.Context, adAccountID, adGroupID int64) error {
	if err := adGroupCheckIDs(adAccountID, adGroupID); err != nil {
		return err
	}
	return s.c.do(ctx, call{method: http.MethodPut, path: idPath("/v4/adGroups/cancel/%d", adGroupID), adAccountID: adAccountID}, nil)
}

// NewPersonalizedMessageAdGroup describes a personalized message ad group
// to create under a personalized message reach campaign. Sends are billed
// per message (CPMS).
type NewPersonalizedMessageAdGroup struct {
	CampaignID int64
	// Name is up to 150 characters; Kakao generates one when empty.
	Name string
}

// CreatePersonalizedMessage creates a personalized message ad group. Kakao
// allows one request per 5 seconds per user and ad account.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/ad-group
func (s *AdGroupService) CreatePersonalizedMessage(ctx context.Context, adAccountID int64, g NewPersonalizedMessageAdGroup) (*AdGroupDetail, error) {
	if adAccountID == 0 || g.CampaignID == 0 {
		return nil, errors.New("moment: ad account ID and campaign ID are required")
	}
	body := struct {
		Campaign idRef  `json:"campaign"`
		Name     string `json:"name,omitempty"`
	}{idRef{g.CampaignID}, g.Name}
	return s.save(ctx, http.MethodPost, adAccountID, body)
}

// PersonalizedMessageAdGroupChanges is the new state of a personalized
// message ad group. Only the name can change.
type PersonalizedMessageAdGroupChanges struct {
	CampaignID int64
	// Name is up to 150 characters; empty keeps the current name.
	Name string
}

// UpdatePersonalizedMessage renames a personalized message ad group. Kakao
// allows one request per 5 seconds per user and ad account.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/ad-group
func (s *AdGroupService) UpdatePersonalizedMessage(ctx context.Context, adAccountID, adGroupID int64, ch PersonalizedMessageAdGroupChanges) (*AdGroupDetail, error) {
	if err := adGroupCheckIDs(adAccountID, adGroupID); err != nil {
		return nil, err
	}
	if ch.CampaignID == 0 {
		return nil, errors.New("moment: campaign ID is required")
	}
	body := struct {
		ID       int64  `json:"id"`
		Campaign idRef  `json:"campaign"`
		Name     string `json:"name,omitempty"`
	}{adGroupID, idRef{ch.CampaignID}, ch.Name}
	return s.save(ctx, http.MethodPut, adAccountID, body)
}

// SetConfig turns an ad group ON or OFF. It only works while the ad group
// is ON or OFF, for display, video and bizboard campaigns other than reach
// and for KakaoTalk Channel campaigns; turning a message ad group OFF stops
// sending and cancels scheduled messages. Kakao allows one request per
// second per user, ad account and ad group.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/ad-group
func (s *AdGroupService) SetConfig(ctx context.Context, adAccountID, adGroupID int64, cfg Config) error {
	if err := adGroupCheckIDs(adAccountID, adGroupID); err != nil {
		return err
	}
	if cfg != ConfigOn && cfg != ConfigOff {
		return errors.New("moment: config must be ON or OFF")
	}
	body := struct {
		ID     int64  `json:"id"`
		Config Config `json:"config"`
	}{adGroupID, cfg}
	return s.c.do(ctx, call{method: http.MethodPut, path: "/v4/adGroups/onOff", adAccountID: adAccountID, body: body}, nil)
}

// Delete deletes an ad group and its creatives. Ad groups of bizboard
// reach and Daum shopping campaigns cannot be deleted. Past reports stay
// available.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/ad-group
func (s *AdGroupService) Delete(ctx context.Context, adAccountID, adGroupID int64) error {
	if err := adGroupCheckIDs(adAccountID, adGroupID); err != nil {
		return err
	}
	return s.c.do(ctx, call{method: http.MethodDelete, path: idPath("/v4/adGroups/%d", adGroupID), adAccountID: adAccountID}, nil)
}

// LatestSystemStopReason returns the most recent system stop reason of an ad
// group. Kakao only has one while the ad group's SystemConfig is
// ADMIN_STOP or EXTERNAL_SERVICE_STOP.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/ad-group
func (s *AdGroupService) LatestSystemStopReason(ctx context.Context, adAccountID, adGroupID int64) (*SystemStopReason, error) {
	if err := adGroupCheckIDs(adAccountID, adGroupID); err != nil {
		return nil, err
	}
	var r SystemStopReason
	if err := s.c.do(ctx, call{method: http.MethodGet, path: idPath("/v4/adGroups/%d/latestSystemConfigHistory", adGroupID), adAccountID: adAccountID}, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// SystemStopReasons returns the system stop reasons of an ad group from the last
// two years, newest first.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/ad-group
func (s *AdGroupService) SystemStopReasons(ctx context.Context, adAccountID, adGroupID int64) ([]SystemStopReason, error) {
	if err := adGroupCheckIDs(adAccountID, adGroupID); err != nil {
		return nil, err
	}
	var rs []SystemStopReason
	if err := s.c.do(ctx, call{method: http.MethodGet, path: idPath("/v4/adGroups/%d/systemConfigHistories", adGroupID), adAccountID: adAccountID}, &rs); err != nil {
		return nil, err
	}
	return rs, nil
}

// save sends an ad group create or update body and decodes the result.
func (s *AdGroupService) save(ctx context.Context, method string, adAccountID int64, body any) (*AdGroupDetail, error) {
	var g AdGroupDetail
	if err := s.c.do(ctx, call{method: method, path: "/v4/adGroups", adAccountID: adAccountID, body: body}, &g); err != nil {
		return nil, err
	}
	return &g, nil
}

func adGroupCheckIDs(adAccountID, adGroupID int64) error {
	if adAccountID == 0 || adGroupID == 0 {
		return errors.New("moment: ad account ID and ad group ID are required")
	}
	return nil
}
