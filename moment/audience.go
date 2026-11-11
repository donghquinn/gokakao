package moment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	kakao "github.com/donghquinn/gokakao"
)

// AudienceStatus is the state of a saved audience (오디언스 상태).
type AudienceStatus string

// Audience statuses.
const (
	AudienceStatusStandby       AudienceStatus = "STANDBY"         // 준비중
	AudienceStatusAvailable     AudienceStatus = "AVAILABLE"       // 사용가능
	AudienceStatusSeedNotEnough AudienceStatus = "SEED_NOT_ENOUGH" // 모수 부족
	AudienceStatusError         AudienceStatus = "ERROR"           // 사용불가
)

// Audience is a saved audience (오디언스) that ad groups can use as their
// targeting. [AudienceService.List] fills only the summary fields: ID,
// Name, AudienceStatus, Type, CreatedDate and LastModifiedDate.
type Audience struct {
	ID          kakao.Int64 `json:"id"`
	AdAccountID kakao.Int64 `json:"adAccountId"`
	Name        string      `json:"name"`
	// Type is AudienceDisplay or AudienceMessage.
	Type           AudienceType   `json:"type"`
	AudienceStatus AudienceStatus `json:"audienceStatus"`
	// PopulationScore is the estimated audience size (예상 모수).
	PopulationScore  kakao.Int64 `json:"populationScore"`
	CreatedDate      kakao.Time  `json:"createdDate"`
	LastModifiedDate kakao.Time  `json:"lastModifiedDate"`

	AgeType SelectionType `json:"ageType"`
	// Ages are age band starts such as "15" (15~19) through "65" (65~69).
	Ages       []string      `json:"ages"`
	GenderType SelectionType `json:"genderType"`
	Genders    []Gender      `json:"genders"`
	// LocationType is not documented in responses; it is set only if Kakao
	// sends it.
	LocationType    LocationType     `json:"locationType"`
	Locations       []Location       `json:"locations"`
	Depth2Locations []Depth2Location `json:"depth2Locations"`
	Depth3Locations []Depth3Location `json:"depth3Locations"`
	// UfoInterests and UfoBusinessTypes are interest and business type
	// keys, e.g. "I0202" and "P1021"; see TargetingService for the
	// categories. The Deprecated variants list keys Kakao no longer
	// supports.
	UfoInterests               []string                    `json:"ufoInterests"`
	UfoDeprecatedInterests     []string                    `json:"ufoDeprecatedInterests"`
	UfoBusinessTypes           []string                    `json:"ufoBusinessTypes"`
	UfoDeprecatedBusinessTypes []string                    `json:"ufoDeprecatedBusinessTypes"`
	TalkChannelTargetings      []TalkChannelTargeting      `json:"talkChannelTargetings"`
	SyncAppTargetings          []SyncAppTargeting          `json:"syncAppTargetings"`
	CustomerDataTargetings     []CustomerDataTargeting     `json:"customerDataTargetings"`
	CustomerFileTargetings     []CustomerFileTargeting     `json:"customerFileTargetings"`
	CohortTargetings           []CohortTargeting           `json:"cohortTargetings"`
	TrackerTargetings          []TrackerTargeting          `json:"trackerTargetings"`
	TalkChannelGroupTargetings []TalkChannelGroupTargeting `json:"talkChannelGroupTargetings"`
}

// UnmarshalJSON implements [json.Unmarshaler]. Kakao's examples send
// interest and business type keys as strings while its type reference
// documents {"value": key} objects; both are accepted.
func (a *Audience) UnmarshalJSON(b []byte) error {
	type plain Audience
	var v struct {
		plain
		UfoInterests               []json.RawMessage `json:"ufoInterests"`
		UfoDeprecatedInterests     []json.RawMessage `json:"ufoDeprecatedInterests"`
		UfoBusinessTypes           []json.RawMessage `json:"ufoBusinessTypes"`
		UfoDeprecatedBusinessTypes []json.RawMessage `json:"ufoDeprecatedBusinessTypes"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*a = Audience(v.plain)
	var err error
	if a.UfoInterests, err = audienceKeys(v.UfoInterests); err != nil {
		return err
	}
	if a.UfoDeprecatedInterests, err = audienceKeys(v.UfoDeprecatedInterests); err != nil {
		return err
	}
	if a.UfoBusinessTypes, err = audienceKeys(v.UfoBusinessTypes); err != nil {
		return err
	}
	a.UfoDeprecatedBusinessTypes, err = audienceKeys(v.UfoDeprecatedBusinessTypes)
	return err
}

// audienceKeys decodes category keys sent as strings or {"value": key}.
func audienceKeys(raw []json.RawMessage) ([]string, error) {
	if raw == nil {
		return nil, nil
	}
	keys := make([]string, 0, len(raw))
	for _, r := range raw {
		var s string
		if err := json.Unmarshal(r, &s); err == nil {
			keys = append(keys, s)
			continue
		}
		var o struct {
			Value string `json:"value"`
		}
		if err := json.Unmarshal(r, &o); err != nil {
			return nil, fmt.Errorf("moment: category key: %w", err)
		}
		keys = append(keys, o.Value)
	}
	return keys, nil
}

// Changes returns the [AudienceChanges] that keep a as it is, for editing
// before [AudienceService.Update]. Deprecated interest and business type
// keys are dropped, and LocationType is AREA when a has any location and
// ALL otherwise, unless Kakao sent it.
func (a *Audience) Changes() AudienceChanges {
	lt := a.LocationType
	if lt == "" {
		lt = LocationTypeAll
		if len(a.Locations)+len(a.Depth2Locations)+len(a.Depth3Locations) > 0 {
			lt = LocationTypeArea
		}
	}
	return AudienceChanges{
		Name: a.Name, Type: a.Type,
		AgeType: a.AgeType, Ages: a.Ages,
		GenderType: a.GenderType, Genders: a.Genders,
		LocationType: lt, Locations: a.Locations,
		Depth2Locations: a.Depth2Locations, Depth3Locations: a.Depth3Locations,
		UfoInterests: a.UfoInterests, UfoBusinessTypes: a.UfoBusinessTypes,
		CustomerDataTargetings:     a.CustomerDataTargetings,
		CustomerFileTargetings:     a.CustomerFileTargetings,
		TrackerTargetings:          a.TrackerTargetings,
		CohortTargetings:           a.CohortTargetings,
		TalkChannelTargetings:      a.TalkChannelTargetings,
		SyncAppTargetings:          a.SyncAppTargetings,
		TalkChannelGroupTargetings: a.TalkChannelGroupTargetings,
	}
}

// List returns the ad account's audiences with their summary fields.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/audience
func (s *AudienceService) List(ctx context.Context, adAccountID int64) ([]Audience, error) {
	if adAccountID == 0 {
		return nil, errors.New("moment: ad account ID is required")
	}
	var res struct {
		Content []Audience `json:"content"`
	}
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v4/targetings/audiences", adAccountID: adAccountID}, &res); err != nil {
		return nil, err
	}
	return res.Content, nil
}

// Get returns an audience with its targeting.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/audience
func (s *AudienceService) Get(ctx context.Context, adAccountID, audienceID int64) (*Audience, error) {
	if adAccountID == 0 || audienceID == 0 {
		return nil, errors.New("moment: ad account ID and audience ID are required")
	}
	var a Audience
	if err := s.c.do(ctx, call{method: http.MethodGet, path: idPath("/v4/targetings/%d", audienceID), adAccountID: adAccountID}, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

// NewAudience describes an audience to create.
//
// Which custom targets an audience may use depends on Type: friend groups
// are MESSAGE only, customer files and Pixel & SDK are DISPLAY only, and
// the rest are allowed for both. Build the targeting entries from the
// targetable lists of [TargetingService].
type NewAudience struct {
	Name string `json:"name"`
	// Type is AudienceDisplay or AudienceMessage.
	Type AudienceType `json:"type"`
	// AgeType is required. Ages are required for NOT_ALL and not allowed
	// for ALL; use NOT_ALL with "20" through "65" to target adults.
	AgeType SelectionType `json:"ageType"`
	// Ages are age band starts: "15" (15~19), "20" (20~24), ... "65"
	// (65~69).
	Ages []string `json:"ages,omitempty"`
	// GenderType is required. Genders are required for NOT_ALL and not
	// allowed for ALL.
	GenderType SelectionType `json:"genderType"`
	Genders    []Gender      `json:"genders,omitempty"`
	// LocationType is LocationTypeAll or LocationTypeArea. For AREA, set
	// at least one of Locations (시/도), Depth2Locations (시/군/구) and
	// Depth3Locations (동/읍/면); for ALL, set none. Overseas (Z) is not
	// allowed for Kakao bizboard campaigns.
	LocationType    LocationType     `json:"locationType"`
	Locations       []Location       `json:"locations,omitempty"`
	Depth2Locations []Depth2Location `json:"depth2Locations,omitempty"`
	Depth3Locations []Depth3Location `json:"depth3Locations,omitempty"`
	// UfoInterests and UfoBusinessTypes are interest and business type
	// keys, e.g. "I0202" and "P1021".
	UfoInterests           []string                `json:"ufoInterests,omitempty"`
	UfoBusinessTypes       []string                `json:"ufoBusinessTypes,omitempty"`
	CustomerDataTargetings []CustomerDataTargeting `json:"customerDataTargetings,omitempty"`
	// CustomerFileTargetings are being retired (종료 예정).
	CustomerFileTargetings []CustomerFileTargeting `json:"customerFileTargetings,omitempty"`
	TrackerTargetings      []TrackerTargeting      `json:"trackerTargetings,omitempty"`
	CohortTargetings       []CohortTargeting       `json:"cohortTargetings,omitempty"`
	TalkChannelTargetings  []TalkChannelTargeting  `json:"talkChannelTargetings,omitempty"`
	SyncAppTargetings      []SyncAppTargeting      `json:"syncAppTargetings,omitempty"`
	// TalkChannelGroupTargetings are being retired (종료 예정) and only
	// work in KakaoTalk Channel reach campaigns.
	TalkChannelGroupTargetings []TalkChannelGroupTargeting `json:"talkChannelGroupTargetings,omitempty"`
}

func (a *NewAudience) validate() error {
	if a.Name == "" || a.Type == "" || a.AgeType == "" || a.GenderType == "" || a.LocationType == "" {
		return errors.New("moment: audience name, type, age type, gender type and location type are required")
	}
	if (a.AgeType == SelectionNotAll) != (len(a.Ages) > 0) {
		return errors.New("moment: audience ages are required for age type NOT_ALL and not allowed for ALL")
	}
	if (a.GenderType == SelectionNotAll) != (len(a.Genders) > 0) {
		return errors.New("moment: audience genders are required for gender type NOT_ALL and not allowed for ALL")
	}
	hasLoc := len(a.Locations)+len(a.Depth2Locations)+len(a.Depth3Locations) > 0
	if a.LocationType == LocationTypeArea && !hasLoc {
		return errors.New("moment: audience location type AREA needs at least one location")
	}
	if a.LocationType == LocationTypeAll && hasLoc {
		return errors.New("moment: audience locations are not allowed for location type ALL")
	}
	return nil
}

// Create creates an audience to use in ad groups.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/audience
func (s *AudienceService) Create(ctx context.Context, adAccountID int64, a NewAudience) (*Audience, error) {
	if adAccountID == 0 {
		return nil, errors.New("moment: ad account ID is required")
	}
	if err := a.validate(); err != nil {
		return nil, err
	}
	body := struct {
		AdAccountID int64 `json:"adAccountId"`
		NewAudience
	}{adAccountID, a}
	var res Audience
	if err := s.c.do(ctx, call{method: http.MethodPost, path: "/v4/targetings", adAccountID: adAccountID, body: body}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// AudienceChanges is the new state of an audience for
// [AudienceService.Update]. It has the fields and rules of [NewAudience].
//
// Kakao replaces the audience: an omitted field is cleared rather than
// kept. Start from [Audience.Changes] to keep the fields you do not change.
type AudienceChanges NewAudience

// Update replaces an audience's settings; see [AudienceChanges].
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/audience
func (s *AudienceService) Update(ctx context.Context, adAccountID, audienceID int64, a AudienceChanges) (*Audience, error) {
	if adAccountID == 0 || audienceID == 0 {
		return nil, errors.New("moment: ad account ID and audience ID are required")
	}
	na := NewAudience(a)
	if err := na.validate(); err != nil {
		return nil, err
	}
	body := struct {
		ID          int64 `json:"id"`
		AdAccountID int64 `json:"adAccountId"`
		NewAudience
	}{audienceID, adAccountID, na}
	var res Audience
	if err := s.c.do(ctx, call{method: http.MethodPut, path: "/v4/targetings", adAccountID: adAccountID, body: body}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Delete deletes an audience.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/audience
func (s *AudienceService) Delete(ctx context.Context, adAccountID, audienceID int64) error {
	if adAccountID == 0 || audienceID == 0 {
		return errors.New("moment: ad account ID and audience ID are required")
	}
	return s.c.do(ctx, call{method: http.MethodDelete, path: idPath("/v4/targetings/%d", audienceID), adAccountID: adAccountID}, nil)
}

// BulkDeleteResult is the outcome of deleting several targets at once.
type BulkDeleteResult struct {
	SuccessCount int `json:"successCount"`
	FailCount    int `json:"failCount"`
	// ErrorMessages are the reasons of the failures, in Korean.
	ErrorMessages []string `json:"errorMessages"`
}

// DeleteMany deletes several audiences. Kakao deletes what it can and
// reports the failures in the result.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/audience
func (s *AudienceService) DeleteMany(ctx context.Context, adAccountID int64, audienceIDs []int64) (*BulkDeleteResult, error) {
	return deleteMany(ctx, s.c, adAccountID, "/v4/targetings", "audienceIds", audienceIDs)
}

// deleteMany sends a bulk delete with the IDs comma-joined in param.
func deleteMany(ctx context.Context, c *Client, adAccountID int64, path, param string, ids []int64) (*BulkDeleteResult, error) {
	if adAccountID == 0 || len(ids) == 0 {
		return nil, errors.New("moment: ad account ID and IDs are required")
	}
	parts := make([]string, len(ids))
	for i, id := range ids {
		if id == 0 {
			return nil, errors.New("moment: IDs must be non-zero")
		}
		parts[i] = strconv.FormatInt(id, 10)
	}
	q := url.Values{param: {strings.Join(parts, ",")}}
	var res BulkDeleteResult
	if err := c.do(ctx, call{method: http.MethodDelete, path: path, adAccountID: adAccountID, query: q}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// AudienceAdGroup is an ad group that uses an audience.
type AudienceAdGroup struct {
	CampaignID       kakao.Int64      `json:"campaignId"`
	CampaignName     string           `json:"campaignName"`
	CampaignTypeGoal CampaignTypeGoal `json:"campaignTypeGoal"`
	AdGroupID        kakao.Int64      `json:"adGroupId"`
	AdGroupName      string           `json:"adGroupName"`
	AdGroupStatus    []AdGroupStatus  `json:"adGroupStatus"`
	AdGroupType      AdGroupType      `json:"adGroupType"`
}

// AdGroups returns the ad groups that use an audience (오디언스 사용 현황).
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/audience
func (s *AudienceService) AdGroups(ctx context.Context, adAccountID, audienceID int64) ([]AudienceAdGroup, error) {
	if adAccountID == 0 || audienceID == 0 {
		return nil, errors.New("moment: ad account ID and audience ID are required")
	}
	var res struct {
		Content []AudienceAdGroup `json:"content"`
	}
	if err := s.c.do(ctx, call{method: http.MethodGet, path: idPath("/v4/targetings/%d/adGroups", audienceID), adAccountID: adAccountID}, &res); err != nil {
		return nil, err
	}
	return res.Content, nil
}

// Unlink stops an ad group from using an audience (오디언스 사용 해제).
// The ad group ID is sent both in the body, as the endpoint docs show, and
// as the adGroupId query parameter, as the API reference shows.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/audience
func (s *AudienceService) Unlink(ctx context.Context, adAccountID, audienceID, adGroupID int64) error {
	if adAccountID == 0 || audienceID == 0 || adGroupID == 0 {
		return errors.New("moment: ad account ID, audience ID and ad group ID are required")
	}
	body := struct {
		AdGroupID int64 `json:"adGroupId"`
	}{adGroupID}
	q := url.Values{"adGroupId": {strconv.FormatInt(adGroupID, 10)}}
	return s.c.do(ctx, call{method: http.MethodPut, path: idPath("/v4/targetings/%d/unlink", audienceID), adAccountID: adAccountID, query: q, body: body}, nil)
}
