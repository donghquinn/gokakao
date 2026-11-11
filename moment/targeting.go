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

// The targeting lookups return the values to use in an ad group's
// targeting ([AdGroupTargetingSpec]) and placements. Region, placement and
// custom target category lookups are not tied to an ad account and take
// no adAccountID.

// Depth1Locations returns the provinces and metropolitan cities (시/도)
// for Locations of a targeting.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/targeting
func (s *TargetingService) Depth1Locations(ctx context.Context) ([]Location, error) {
	var ls []Location
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v4/targetings/location/depth1"}, &ls); err != nil {
		return nil, err
	}
	return ls, nil
}

// Depth2Locations searches the cities, counties and districts (시/군/구)
// for Depth2Locations of a targeting.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/targeting
func (s *TargetingService) Depth2Locations(ctx context.Context, keyword string) ([]Depth2Location, error) {
	if keyword == "" {
		return nil, errors.New("moment: keyword is required")
	}
	var ls []Depth2Location
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v4/targetings/location/depth2", query: url.Values{"keyword": {keyword}}}, &ls); err != nil {
		return nil, err
	}
	return ls, nil
}

// Depth3Locations searches the towns and neighborhoods (동/읍/면) for
// Depth3Locations of a targeting.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/targeting
func (s *TargetingService) Depth3Locations(ctx context.Context, keyword string) ([]Depth3Location, error) {
	if keyword == "" {
		return nil, errors.New("moment: keyword is required")
	}
	var ls []Depth3Location
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v4/targetings/location/depth3", query: url.Values{"keyword": {keyword}}}, &ls); err != nil {
		return nil, err
	}
	return ls, nil
}

// AdServingCategory is a network sub-placement (네트워크 하위 게재지면).
type AdServingCategory struct {
	// Code goes in an ad group's AdServingCategories.
	Code NetworkCategoryCode `json:"code"`
	Name string              `json:"name"`
}

// AdServingCategories returns the network sub-placements. Network
// placements are not available to display conversion campaigns.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/targeting
func (s *TargetingService) AdServingCategories(ctx context.Context) ([]AdServingCategory, error) {
	var cs []AdServingCategory
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v4/targetings/placement/adServingCategories"}, &cs); err != nil {
		return nil, err
	}
	return cs, nil
}

// SectionCategory is a KakaoTalk section (섹션 카테고리).
type SectionCategory struct {
	// Code goes in an ad group's SectionCategories.
	Code TalkSectionCode `json:"code"`
	Name string          `json:"name"`
}

// SectionCategories returns the KakaoTalk sections, selectable for
// bizboard, display, video and product catalog campaigns placed on
// KakaoTalk.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/targeting
func (s *TargetingService) SectionCategories(ctx context.Context) ([]SectionCategory, error) {
	var cs []SectionCategory
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v4/targetings/placement/sectionCategories"}, &cs); err != nil {
		return nil, err
	}
	return cs, nil
}

// TargetableCohortStatus is the audience state of an ad reaction target
// (타겟 모수 상태).
type TargetableCohortStatus string

// Targetable cohort statuses.
const (
	TargetableCohortWaiting        TargetableCohortStatus = "WAITING"         // 준비 중
	TargetableCohortAvailableError TargetableCohortStatus = "AVAILABLE_ERROR" // 모수 추출 에러
	TargetableCohortAvailable      TargetableCohortStatus = "AVAILABLE"       // 모수 추출 완료
	TargetableCohortSeedNotEnough  TargetableCohortStatus = "SEED_NOT_ENOUGH" // 모수부족
	TargetableCohortDelete         TargetableCohortStatus = "DELETE"          // 삭제 또는 삭제중
	TargetableCohortError          TargetableCohortStatus = "ERROR"           // 그 외 에러
)

// TargetableCohort is an ad reaction target (광고반응타겟) usable in a
// targeting.
type TargetableCohort struct {
	// ID goes in a [CohortTarget] as CohortID.
	ID           kakao.Int64  `json:"id"`
	AudienceType AudienceType `json:"audienceType"` // DISPLAY or MESSAGE
	AdAccountID  kakao.Int64  `json:"adAccountId"`
	Name         string       `json:"name"`
	BaseAds      []BaseAd     `json:"baseAds"`
	// CollectDuration is the number of days of reactions used.
	CollectDuration  int                    `json:"collectDuration"`
	Status           TargetableCohortStatus `json:"status"`
	PopulationScore  kakao.Int64            `json:"populationScore"`
	CreatedDate      kakao.Time             `json:"createdDate"`
	LastModifiedDate kakao.Time             `json:"lastModifiedDate"`
}

// Cohorts returns the ad reaction targets usable in a targeting.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/targeting
func (s *TargetingService) Cohorts(ctx context.Context, adAccountID int64) ([]TargetableCohort, error) {
	var cs []TargetableCohort
	if err := s.list(ctx, adAccountID, "/v4/adGroups/targetings/cohort/availables", nil, &cs); err != nil {
		return nil, err
	}
	return cs, nil
}

// TrackerEventRule is the tracker rule of a [TargetableTrackerEvent].
type TrackerEventRule struct {
	TrackRuleID string `json:"trackRuleId"`
	TrackID     string `json:"trackId"`
	EventCode   string `json:"eventCode"`
	Name        string `json:"name"`
	// RuleSetStr is the rule set definition. The docs type it as a string
	// array but the example sends a JSON string ("[]"), so it is kept raw.
	RuleSetStr json.RawMessage `json:"ruleSetStr"`
}

// TargetableTrackerEvent is a Pixel & SDK event usable in a targeting.
type TargetableTrackerEvent struct {
	// EventCode is "*" for all events.
	EventCode      string `json:"eventCode"`
	EventName      string `json:"eventName"`
	EventExtraName string `json:"eventExtraName"` // 커스텀 이벤트 이름
	TrackRuleID    string `json:"trackRuleId"`
	// TrackRule is nil for all events.
	TrackRule  *TrackerEventRule `json:"trackRule"`
	Population kakao.Int64       `json:"population"`
	// LastEventDate is the last seen time; the docs say
	// "yyyy-MM-dd'T'HH:mm:ss" but the example is "2019.10.16 19:36".
	LastEventDate string `json:"lastEventDate"`
}

// TrackerEvents returns the events of a Pixel & SDK tracker usable in a
// targeting over term days (1~180).
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/targeting
func (s *TargetingService) TrackerEvents(ctx context.Context, adAccountID int64, trackID string, term int) ([]TargetableTrackerEvent, error) {
	if trackID == "" {
		return nil, errors.New("moment: track ID is required")
	}
	if term < 1 || term > 180 {
		return nil, errors.New("moment: term must be between 1 and 180")
	}
	var es []TargetableTrackerEvent
	q := url.Values{"term": {strconv.Itoa(term)}}
	if err := s.list(ctx, adAccountID, "/v4/adGroups/targetings/trackers/eventCreatables/"+url.PathEscape(trackID), q, &es); err != nil {
		return nil, err
	}
	return es, nil
}

// TargetableCustomerData is customer data (고객데이터) usable in a
// targeting.
type TargetableCustomerData struct {
	// ID goes in a [CustomerDataTarget] as CustomerDataID.
	ID          kakao.Int64          `json:"id"`
	AdAccountID kakao.Int64          `json:"adAccountId"`
	Name        string               `json:"name"`
	KuidListKey string               `json:"kuidListKey"` // 등록 Key
	FileType    CustomerDataFileType `json:"fileType"`
	Status      CustomerDataStatus   `json:"status"`
	// PopulationScore is the number of Kakao users found; data is usable
	// within an hour of registration.
	PopulationScore kakao.Int64 `json:"populationScore"`
	Ready           bool        `json:"ready"`
	// CreatedDate and LastModifyRequestDate are local times; the docs say
	// "yyyy-MM-dd'T'HH:mm:ss" but the example is "2020-01-01 00:00".
	CreatedDate           string `json:"createdDate"`
	LastModifyRequestDate string `json:"lastModifyRequestDate"`
}

// CustomerData returns the customer data usable in a targeting.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/targeting
func (s *TargetingService) CustomerData(ctx context.Context, adAccountID int64) ([]TargetableCustomerData, error) {
	var ds []TargetableCustomerData
	if err := s.list(ctx, adAccountID, "/v4/adGroups/targetings/customerData/availables", nil, &ds); err != nil {
		return nil, err
	}
	return ds, nil
}

// TargetableCustomerFile is a customer file (고객파일, being retired)
// usable in a targeting.
type TargetableCustomerFile struct {
	// ID goes in a [CustomerFileTarget] as CustomerFileID.
	ID              kakao.Int64        `json:"id"`
	AdAccountID     kakao.Int64        `json:"adAccountId"`
	Name            string             `json:"name"`
	AdidListKey     string             `json:"adidListKey"` // 등록 Key
	Status          CustomerFileStatus `json:"status"`
	PopulationScore kakao.Int64        `json:"populationScore"`
	Ready           bool               `json:"ready"`
	// CreatedDate and LastModifyRequestDate are as in
	// [TargetableCustomerData].
	CreatedDate           string `json:"createdDate"`
	LastModifyRequestDate string `json:"lastModifyRequestDate"`
}

// CustomerFiles returns the customer files usable in a targeting. Customer
// files are being retired; use customer data for new targets.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/targeting
func (s *TargetingService) CustomerFiles(ctx context.Context, adAccountID int64) ([]TargetableCustomerFile, error) {
	var fs []TargetableCustomerFile
	if err := s.list(ctx, adAccountID, "/v4/adGroups/targetings/customerFiles/availables", nil, &fs); err != nil {
		return nil, err
	}
	return fs, nil
}

// TargetableFriendGroupStatus is the state of a friend group file.
type TargetableFriendGroupStatus string

// Targetable friend group statuses.
const (
	TargetableFriendGroupWaiting  TargetableFriendGroupStatus = "WAITING"  // 대기중
	TargetableFriendGroupComplete TargetableFriendGroupStatus = "COMPLETE" // 완료
	TargetableFriendGroupDelete   TargetableFriendGroupStatus = "DELETE"   // 삭제
	TargetableFriendGroupError    TargetableFriendGroupStatus = "ERROR"    // 생성 실패
)

// TargetableFriendGroup is a friend group (친구그룹, being retired) usable
// in a targeting.
type TargetableFriendGroup struct {
	// ID goes in a [TalkChannelGroupTarget] as TalkChannelGroupFileID.
	ID        kakao.Int64 `json:"id"`
	Name      string      `json:"name"`
	ProfileID string      `json:"profileId"`
	GroupKey  string      `json:"groupKey"`
	// FileType is APP_USER_ID, PHONE_NUMBER or MESSAGE_RETARGET.
	FileType             CustomerDataFileType        `json:"fileType"`
	AppID                kakao.Int64                 `json:"appId"`
	Status               TargetableFriendGroupStatus `json:"talkChannelGroupFileStatus"`
	FriendCount          kakao.Int64                 `json:"friendCount"`
	CreatedDate          kakao.Time                  `json:"createdDate"`
	PopulationUpdateDate kakao.Time                  `json:"populationUpdateDate"`
}

// FriendGroups returns the friend groups of a KakaoTalk Channel usable in
// a targeting. profileID is the channel profile ID, the Objective value of
// the campaign. Friend groups are being retired; use customer data for new
// targets.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/targeting
func (s *TargetingService) FriendGroups(ctx context.Context, adAccountID int64, profileID string) ([]TargetableFriendGroup, error) {
	if profileID == "" {
		return nil, errors.New("moment: profile ID is required")
	}
	var gs []TargetableFriendGroup
	if err := s.list(ctx, adAccountID, "/v4/adGroups/targetings/talkChannelGroupFiles/availables", url.Values{"profileId": {profileID}}, &gs); err != nil {
		return nil, err
	}
	return gs, nil
}

// TalkChannels returns the KakaoTalk Channels whose friends can be
// targeted. Each ID goes in a [TalkChannelTarget] as ProfileID.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/targeting
func (s *TargetingService) TalkChannels(ctx context.Context, adAccountID int64) ([]ChannelProfile, error) {
	var cs []ChannelProfile
	if err := s.list(ctx, adAccountID, "/v4/adGroups/targetings/talkChannels/availables", nil, &cs); err != nil {
		return nil, err
	}
	return cs, nil
}

// LoginUserChannels returns the KakaoTalk Channels whose Kakao Login users
// (카카오 로그인 이용자) can be targeted. Each ID goes in a
// [SyncAppTargeting] as ProfileID.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/targeting
func (s *TargetingService) LoginUserChannels(ctx context.Context, adAccountID int64) ([]ChannelProfile, error) {
	var cs []ChannelProfile
	if err := s.list(ctx, adAccountID, "/v4/adGroups/targetings/syncProfiles/availables", nil, &cs); err != nil {
		return nil, err
	}
	return cs, nil
}

// CategoryType is a kind of custom target category (카테고리 타입).
type CategoryType string

// Category types.
const (
	CategoryInterest     CategoryType = "INTEREST"      // 관심사
	CategoryBusinessType CategoryType = "BUSINESS_TYPE" // 업종
)

// CategoryTree is the custom target categories of a type.
type CategoryTree struct {
	Type CategoryType `json:"type"`
	// Data is the category tree. IDs have 3 characters at the top level and
	// add 2 digits per level.
	Data []CategoryData `json:"data"`
}

// Categories returns the interest or business type categories, whose IDs
// go in UfoInterests or UfoBusinessTypes of a targeting.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/target-custom
func (s *TargetingService) Categories(ctx context.Context, typ CategoryType) (*CategoryTree, error) {
	if typ == "" {
		return nil, errors.New("moment: category type is required")
	}
	var c CategoryTree
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v4/targetings/data/categories", query: url.Values{"type": {string(typ)}}}, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// MapLocation is a place found for moment targeting.
type MapLocation struct {
	Address   string `json:"address"`
	LabelName string `json:"labelName"` // 기준 위치
	// Geohash goes in the geohashes of an ad group.
	Geohash         string      `json:"geohash"`
	CenterLatitude  kakao.Float `json:"centerLatitude"`
	CenterLongitude kakao.Float `json:"centerLongitude"`
	Latitude        kakao.Float `json:"latitude"`
	Longitude       kakao.Float `json:"longitude"`
	POI             string      `json:"poi"` // 대표 코드
}

// MapLocationResults are the results of [TargetingService.MapLocations].
type MapLocationResults struct {
	Query string        `json:"query"`
	Items []MapLocation `json:"items"`
}

// MapLocations searches places for moment targeting (모먼트 타게팅) by a
// detailed address (road name and building number, or neighborhood and
// lot number) or place name, most accurate first.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/target-custom
func (s *TargetingService) MapLocations(ctx context.Context, keyword string) (*MapLocationResults, error) {
	if keyword == "" {
		return nil, errors.New("moment: keyword is required")
	}
	var m MapLocationResults
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v4/targetings/location/map", query: url.Values{"keyword": {keyword}}}, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// list sends a GET for an ad account's targetable items.
func (s *TargetingService) list(ctx context.Context, adAccountID int64, path string, q url.Values, out any) error {
	if adAccountID == 0 {
		return errors.New("moment: ad account ID is required")
	}
	return s.c.do(ctx, call{method: http.MethodGet, path: path, adAccountID: adAccountID, query: q}, out)
}
