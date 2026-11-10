package moment

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"

	kakao "github.com/donghquinn/gokakao"
)

// CreativeFormat is the format of a display creative (소재 유형). Message
// creatives use [MessageType] values in the same field.
type CreativeFormat string

// CreativeFormat values.
const (
	CreativeFormatImageBanner CreativeFormat = "IMAGE_BANNER" // 이미지 배너
	CreativeFormatImageNative CreativeFormat = "IMAGE_NATIVE" // 이미지 네이티브
	CreativeFormatVideoNative CreativeFormat = "VIDEO_NATIVE" // 동영상 네이티브
)

// FrequencyCapType is how often a display creative is shown to a user
// (게재빈도).
type FrequencyCapType string

// FrequencyCapType values.
const (
	FrequencyCapAuto   FrequencyCapType = "AUTO"    // 자동 설정
	FrequencyCapDayImp FrequencyCapType = "DAY_IMP" // 상세 설정
)

// CreativeSummary is a creative in a list.
type CreativeSummary struct {
	// ID is the original creative ID (원본 소재 번호) used when the
	// creative runs.
	ID           kakao.Int64  `json:"id"`
	Name         string       `json:"name"`
	Config       Config       `json:"config"`
	SystemConfig SystemConfig `json:"systemConfig"`
}

// CreativeListOptions are the optional parameters of
// [CreativeService.List].
type CreativeListOptions struct {
	// Config filters by status; Kakao's default is ON and OFF.
	Config []Config
}

// List returns the creatives of an ad group.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/creatives
func (s *CreativeService) List(ctx context.Context, adAccountID, adGroupID int64, opts *CreativeListOptions) ([]CreativeSummary, error) {
	if adGroupID == 0 {
		return nil, errors.New("moment: ad group ID is required")
	}
	q := url.Values{"adGroupId": {strconv.FormatInt(adGroupID, 10)}}
	if opts != nil {
		for _, c := range opts.Config {
			q.Add("config", string(c))
		}
	}
	var res struct {
		Content []CreativeSummary `json:"content"`
	}
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v4/creatives", adAccountID: adAccountID, query: q}, &res); err != nil {
		return nil, err
	}
	return res.Content, nil
}

// CreativeVideo is the video of a display creative.
//
// It differs from [Video], whose Duration is a string: creative responses
// send the duration as a number of seconds.
type CreativeVideo struct {
	URL      string      `json:"url"`
	FileName string      `json:"fileName"`
	Size     kakao.Int64 `json:"size"`     // bytes
	Duration kakao.Float `json:"duration"` // seconds
}

// Creative is a creative (소재). Display creatives (IMAGE_BANNER,
// IMAGE_NATIVE, VIDEO_NATIVE) set the display fields; message and
// personalized message creatives set MessageElement. Unused fields may be
// null.
type Creative struct {
	// ID is the original creative ID (원본 소재 번호) used when the
	// creative runs.
	ID kakao.Int64 `json:"id"`
	// CreativeID is the creative revision ID (소재 번호). It differs from ID
	// after an approved display creative is changed in a way that needs a
	// new review; message creatives always have CreativeID == ID.
	CreativeID kakao.Int64 `json:"creativeId"`
	Name       string      `json:"name"`
	AdGroupID  kakao.Int64 `json:"adGroupId"`
	// Format is a CreativeFormat for display creatives and a MessageType
	// value for message creatives.
	Format         CreativeFormat       `json:"format"`
	Config         Config               `json:"config"`
	SystemConfig   SystemConfig         `json:"systemConfig"`
	ReviewStatus   CreativeReviewStatus `json:"reviewStatus"` // display only
	CreativeStatus CreativeStatus       `json:"creativeStatus"`
	// StatusDescription describes the delivery status, e.g. 심사중 or
	// 발송 대기.
	StatusDescription string `json:"statusDescription"`

	// Landing URLs (랜딩 URL) for PC, mobile and responsive (반응형) web.
	PCLandingURL     string `json:"pcLandingUrl"`
	MobileLandingURL string `json:"mobileLandingUrl"`
	RspvLandingURL   string `json:"rspvLandingUrl"`
	// LandingURL appears in some responses instead of the URLs above.
	LandingURL  string       `json:"landingUrl"`
	LandingInfo *LandingInfo `json:"landingInfo"`

	FrequencyCapType FrequencyCapType `json:"frequencyCapType"`
	// FrequencyCapTime is the frequency window in hours, 1~24.
	FrequencyCapTime kakao.Int64 `json:"frequencyCapTime"`
	// FrequencyCap is the number of impressions per window, 1~5.
	FrequencyCap kakao.Int64 `json:"frequencyCap"`

	Image   *Image `json:"image"`
	AltText string `json:"altText"` // IMAGE_BANNER
	// AILabeled reports the generative AI mark (AI 활용 표기).
	AILabeled      bool           `json:"isAiLabeled"`
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	ActionButton   ActionButton   `json:"actionButton"`
	ProfileName    string         `json:"profileName"`
	ProfileImage   *Image         `json:"profileImage"`
	Video          *CreativeVideo `json:"video"`
	ThumbnailImage *Image         `json:"thumbnailImage"` // 맞춤 썸네일
	// VideoSkippableType is SECONDS_5 or SECONDS_15 for in-stream video.
	VideoSkippableType string `json:"videoSkippableType"`
	// AssetGroups are the slides (슬라이드) of a native creative.
	AssetGroups    []AssetGroup     `json:"assetGroups"`
	HasExpandable  bool             `json:"hasExpandable"`
	RejectedReason []RejectedReason `json:"rejectedReason"`
	// Opinion and OpinionProof are the review opinion and its supporting
	// files (의견, 증빙). They may be null until the review system syncs.
	Opinion      string        `json:"opinion"`
	OpinionProof []OpinionFile `json:"opinionProof"`

	// MessageElement is the message of a message creative.
	MessageElement *MessageCreativeElement `json:"messageElement"`
	// AgeVerification reports an age-verified message (연령인증 메시지).
	AgeVerification bool `json:"ageVerification"`

	CreatedDate      kakao.Time `json:"createdDate"`
	LastModifiedDate kakao.Time `json:"lastModifiedDate"`
}

// Get returns a creative of any type.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/creatives
func (s *CreativeService) Get(ctx context.Context, adAccountID, creativeID int64) (*Creative, error) {
	var cr Creative
	if err := s.c.do(ctx, call{method: http.MethodGet, path: idPath("/v4/creatives/%d", creativeID), adAccountID: adAccountID}, &cr); err != nil {
		return nil, err
	}
	return &cr, nil
}

// SetConfig turns a display creative on or off. config must be ConfigOn or
// ConfigOff. Creatives under KakaoTalk Channel campaigns cannot be changed.
// Kakao allows one call per second per user and ad account.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/creatives
func (s *CreativeService) SetConfig(ctx context.Context, adAccountID, creativeID int64, config Config) error {
	if creativeID == 0 {
		return errors.New("moment: creative ID is required")
	}
	if config != ConfigOn && config != ConfigOff {
		return errors.New("moment: config must be ON or OFF")
	}
	body := struct {
		ID     int64  `json:"id"`
		Config Config `json:"config"`
	}{creativeID, config}
	return s.c.do(ctx, call{method: http.MethodPut, path: "/v4/creatives/onOff", adAccountID: adAccountID, body: body}, nil)
}

// Delete deletes a creative. Deleting gives up operating the creative; its
// data is kept.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/creatives
func (s *CreativeService) Delete(ctx context.Context, adAccountID, creativeID int64) error {
	return s.c.do(ctx, call{method: http.MethodDelete, path: idPath("/v4/creatives/%d", creativeID), adAccountID: adAccountID}, nil)
}

// SystemStopReason is why a creative was stopped by the system
// (시스템 정지 사유).
type SystemStopReason struct {
	ID               kakao.Int64  `json:"id"`
	SystemConfig     SystemConfig `json:"systemConfig"`
	Reason           string       `json:"reason"`
	CreatedDate      kakao.Time   `json:"createdDate"`
	LastModifiedDate kakao.Time   `json:"lastModifiedDate"`
}

// SystemStopReason returns the latest admin stop reason of a creative. Kakao
// only answers when the creative's SystemConfig is ADMIN_STOP.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/creatives
func (s *CreativeService) SystemStopReason(ctx context.Context, adAccountID, creativeID int64) (*SystemStopReason, error) {
	var r SystemStopReason
	if err := s.c.do(ctx, call{method: http.MethodGet, path: idPath("/v4/creatives/%d/systemConfigHistory", creativeID), adAccountID: adAccountID}, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// SystemStopReasons returns the system stop reasons of a creative from the
// last two years. Kakao only answers when the creative's SystemConfig is
// ADMIN_STOP.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/creatives
func (s *CreativeService) SystemStopReasons(ctx context.Context, adAccountID, creativeID int64) ([]SystemStopReason, error) {
	var r []SystemStopReason
	if err := s.c.do(ctx, call{method: http.MethodGet, path: idPath("/v4/creatives/%d/systemConfigHistories", creativeID), adAccountID: adAccountID}, &r); err != nil {
		return nil, err
	}
	return r, nil
}

// copyCreatives copies creatives into an ad group. Display and message
// creatives share the endpoint.
func (s *CreativeService) copyCreatives(ctx context.Context, adAccountID, adGroupID int64, creativeIDs []int64) ([]Creative, error) {
	if adGroupID == 0 || len(creativeIDs) == 0 {
		return nil, errors.New("moment: ad group ID and creative IDs are required")
	}
	body := struct {
		AdGroupID   int64   `json:"adGroupId"`
		CreativeIDs []int64 `json:"creativeIds"`
	}{adGroupID, creativeIDs}
	var res []Creative
	if err := s.c.do(ctx, call{method: http.MethodPost, path: "/v4/creatives/copy", adAccountID: adAccountID, body: body}, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// UploadFile is a local file sent as a multipart/form-data part.
type UploadFile struct {
	Filename string
	// ContentType is detected from Data when empty.
	ContentType string
	Data        []byte
}

// part returns f as the multipart file field.
func (f *UploadFile) part(field string) (kakao.MultipartFile, error) {
	if f.Filename == "" || len(f.Data) == 0 {
		return kakao.MultipartFile{}, fmt.Errorf("moment: %s: file name and data are required", field)
	}
	ct := f.ContentType
	if ct == "" {
		ct = http.DetectContentType(f.Data)
	}
	return kakao.MultipartFile{Field: field, Filename: f.Filename, ContentType: ct, Data: f.Data}, nil
}

// creativeForm flattens the JSON encoding of v into the dotted form field
// names Kakao uses for multipart creatives, such as
// "messageElement.buttonAssetGroups[0].title".
func creativeForm(v any) (url.Values, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("moment: encode form: %w", err)
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	var m any
	if err := d.Decode(&m); err != nil {
		return nil, fmt.Errorf("moment: encode form: %w", err)
	}
	form := url.Values{}
	creativeFormValue(form, "", m)
	return form, nil
}

func creativeFormValue(form url.Values, name string, v any) {
	switch v := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			key := k
			if name != "" {
				key = name + "." + k
			}
			creativeFormValue(form, key, v[k])
		}
	case []any:
		for i, e := range v {
			creativeFormValue(form, name+"["+strconv.Itoa(i)+"]", e)
		}
	case nil:
	case string:
		form.Set(name, v)
	case json.Number:
		form.Set(name, v.String())
	case bool:
		form.Set(name, strconv.FormatBool(v))
	}
}
