package moment

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	kakao "github.com/donghquinn/gokakao"
)

// Friend groups (친구그룹) are being retired: register new targets as
// customer data instead. A friend group can only be used by
// KakaoTalk Channel x reach (카카오톡 채널 X 도달) campaigns.

// FriendGroupStatus is the state of a friend group.
type FriendGroupStatus string

// Friend group statuses.
const (
	FriendGroupWaiting  FriendGroupStatus = "WAITING"  // 생성 대기중
	FriendGroupComplete FriendGroupStatus = "COMPLETE" // 생성 완료
	FriendGroupDelete   FriendGroupStatus = "DELETE"   // 삭제 또는 삭제중
	FriendGroupError    FriendGroupStatus = "ERROR"    // 그 외 비정상 (생성 실패)
)

// FriendGroup is a friend group (TalkChannelGroupFile).
type FriendGroup struct {
	ID kakao.Int64 `json:"id"`
	// ProfileID is the KakaoTalk Channel profile ID. The docs type it as a
	// string but the examples send a number; both are accepted.
	ProfileID              string `json:"profileId"`
	TalkChannelProfileName string `json:"talkChannelProfileName"`
	// FileType is APP_USER_ID, PHONE_NUMBER or MESSAGE_RETARGET.
	FileType    CustomerDataFileType `json:"fileType"`
	GroupKey    string               `json:"groupKey"`
	Name        string               `json:"name"`
	FriendCount kakao.Int64          `json:"friendCount"`
	Status      FriendGroupStatus    `json:"talkChannelGroupFileStatus"`
	CreatedDate kakao.Time           `json:"createdDate"`
	// LastModifiedDate is only returned by [FriendGroupService.Rename].
	LastModifiedDate kakao.Time `json:"lastModifiedDate"`
}

// UnmarshalJSON implements [json.Unmarshaler].
func (g *FriendGroup) UnmarshalJSON(b []byte) error {
	type plain FriendGroup
	aux := struct {
		*plain
		ProfileID json.RawMessage `json:"profileId"`
	}{plain: (*plain)(g)}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	id, err := friendGroupString(aux.ProfileID)
	g.ProfileID = id
	return err
}

// friendGroupString decodes a JSON string or number as a string.
func friendGroupString(raw json.RawMessage) (string, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return "", nil
	}
	if raw[0] == '"' {
		var s string
		err := json.Unmarshal(raw, &s)
		return s, err
	}
	var n json.Number
	if err := json.Unmarshal(raw, &n); err != nil {
		return "", err
	}
	return n.String(), nil
}

// List returns the ad account's friend groups.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/friend-group
func (s *FriendGroupService) List(ctx context.Context, adAccountID int64) ([]FriendGroup, error) {
	if adAccountID == 0 {
		return nil, errors.New("moment: ad account ID is required")
	}
	var gs []FriendGroup
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v4/talkChannelGroupFiles", adAccountID: adAccountID}, &gs); err != nil {
		return nil, err
	}
	return gs, nil
}

// NewFriendGroup describes a friend group to register. It is sent as
// multipart/form-data.
type NewFriendGroup struct {
	// Name is up to 50 characters.
	Name string
	// ProfileID is a KakaoTalk Channel profile ID from the channel profile
	// list.
	ProfileID string
	// FileType is CustomerDataFileAppUserID, CustomerDataFilePhoneNumber
	// or CustomerDataFileMessageRetarget.
	FileType CustomerDataFileType
	// Files are required for APP_USER_ID and PHONE_NUMBER: up to 10 CSV
	// files, 200 MB in total. ContentType defaults to text/csv.
	Files []UploadFile
	// AdGroupIDs are required for MESSAGE_RETARGET: undeleted message ad
	// groups of the ad account sent within 30 days.
	AdGroupIDs []int64
}

// FriendGroupUpload is the result of [FriendGroupService.Create].
type FriendGroupUpload struct {
	AdAccountID kakao.Int64          `json:"adAccountId"`
	ID          kakao.Int64          `json:"id"`
	Name        string               `json:"name"`
	FileType    CustomerDataFileType `json:"fileType"`
	// AdGroupIDs are the message ad groups. The docs type this as a single
	// Long; both a number and an array are accepted.
	AdGroupIDs    []kakao.Int64     `json:"adGroupIds"`
	SuccessCount  int               `json:"successCount"`
	FailedCount   int               `json:"failedCount"`
	FailedFileURL string            `json:"failedFileUrl"` // failed identifiers
	Status        FriendGroupStatus `json:"status"`
	CreatedDate   kakao.Time        `json:"createdDate"`
}

// UnmarshalJSON implements [json.Unmarshaler].
func (u *FriendGroupUpload) UnmarshalJSON(b []byte) error {
	type plain FriendGroupUpload
	aux := struct {
		*plain
		AdGroupIDs json.RawMessage `json:"adGroupIds"`
	}{plain: (*plain)(u)}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	raw := bytes.TrimSpace(aux.AdGroupIDs)
	switch {
	case len(raw) == 0 || bytes.Equal(raw, []byte("null")):
		u.AdGroupIDs = nil
	case raw[0] == '[':
		return json.Unmarshal(raw, &u.AdGroupIDs)
	default:
		var id kakao.Int64
		if err := json.Unmarshal(raw, &id); err != nil {
			return err
		}
		u.AdGroupIDs = []kakao.Int64{id}
	}
	return nil
}

// Create registers a friend group from identifier files or from the
// recipients of message ad groups. An ad account has at most 50 friend
// groups. It is limited to 1 request per 5 seconds per user and ad
// account.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/friend-group
func (s *FriendGroupService) Create(ctx context.Context, adAccountID int64, g NewFriendGroup) (*FriendGroupUpload, error) {
	if adAccountID == 0 {
		return nil, errors.New("moment: ad account ID is required")
	}
	if g.Name == "" || g.ProfileID == "" || g.FileType == "" {
		return nil, errors.New("moment: name, profile ID and file type are required")
	}
	switch g.FileType {
	case CustomerDataFileAppUserID, CustomerDataFilePhoneNumber:
		if len(g.Files) == 0 {
			return nil, errors.New("moment: files are required for this file type")
		}
	case CustomerDataFileMessageRetarget:
		if len(g.AdGroupIDs) == 0 {
			return nil, errors.New("moment: ad group IDs are required for MESSAGE_RETARGET")
		}
	default:
		return nil, errors.New("moment: file type must be APP_USER_ID, PHONE_NUMBER or MESSAGE_RETARGET")
	}
	m := &kakao.Multipart{Fields: url.Values{
		"name":      {g.Name},
		"profileId": {g.ProfileID},
		"fileType":  {string(g.FileType)},
	}}
	for _, id := range g.AdGroupIDs {
		m.Fields.Add("adGroupIds", strconv.FormatInt(id, 10))
	}
	for _, f := range g.Files {
		if f.ContentType == "" {
			f.ContentType = "text/csv"
		}
		part, err := f.part("files")
		if err != nil {
			return nil, err
		}
		m.Files = append(m.Files, part)
	}
	var res FriendGroupUpload
	if err := s.c.do(ctx, call{method: http.MethodPost, path: "/v4/talkChannelGroupFiles", adAccountID: adAccountID, multipart: m}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Rename changes a friend group's name, up to 50 characters.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/friend-group
func (s *FriendGroupService) Rename(ctx context.Context, adAccountID, friendGroupID int64, name string) (*FriendGroup, error) {
	if adAccountID == 0 || friendGroupID == 0 {
		return nil, errors.New("moment: ad account ID and friend group ID are required")
	}
	if name == "" {
		return nil, errors.New("moment: name is required")
	}
	body := struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}{friendGroupID, name}
	var g FriendGroup
	if err := s.c.do(ctx, call{method: http.MethodPut, path: "/v4/talkChannelGroupFiles/name", adAccountID: adAccountID, body: body}, &g); err != nil {
		return nil, err
	}
	return &g, nil
}

// Delete deletes a friend group.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/friend-group
func (s *FriendGroupService) Delete(ctx context.Context, adAccountID, friendGroupID int64) error {
	if adAccountID == 0 || friendGroupID == 0 {
		return errors.New("moment: ad account ID and friend group ID are required")
	}
	return s.c.do(ctx, call{method: http.MethodDelete, path: idPath("/v4/talkChannelGroupFiles/%d", friendGroupID), adAccountID: adAccountID}, nil)
}

// DeleteMany deletes several friend groups.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/friend-group
func (s *FriendGroupService) DeleteMany(ctx context.Context, adAccountID int64, friendGroupIDs []int64) error {
	if adAccountID == 0 {
		return errors.New("moment: ad account ID is required")
	}
	if len(friendGroupIDs) == 0 {
		return errors.New("moment: friend group IDs are required")
	}
	ids := make([]string, len(friendGroupIDs))
	for i, id := range friendGroupIDs {
		ids[i] = strconv.FormatInt(id, 10)
	}
	q := url.Values{"ids": {strings.Join(ids, ",")}}
	return s.c.do(ctx, call{method: http.MethodDelete, path: "/v4/talkChannelGroupFiles", adAccountID: adAccountID, query: q}, nil)
}

// FriendGroupUsage is an ad using a friend group (AdgroupAndCampaign):
// Campaign and AdGroup for a display ad, or MessageAd for a message.
type FriendGroupUsage struct {
	Campaign  *Campaign  `json:"campaign"`
	AdGroup   *AdGroup   `json:"adGroup"`
	MessageAd *MessageAd `json:"messageAd"`
}

// Usages returns the ads using a friend group.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/friend-group
func (s *FriendGroupService) Usages(ctx context.Context, adAccountID, friendGroupID int64) ([]FriendGroupUsage, error) {
	if adAccountID == 0 || friendGroupID == 0 {
		return nil, errors.New("moment: ad account ID and friend group ID are required")
	}
	var us []FriendGroupUsage
	if err := s.c.do(ctx, call{method: http.MethodGet, path: idPath("/v4/talkChannelGroupFiles/usages/%d", friendGroupID), adAccountID: adAccountID}, &us); err != nil {
		return nil, err
	}
	return us, nil
}
