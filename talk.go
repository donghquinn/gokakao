package kakao

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"strconv"

	"github.com/donghquinn/gokakao/template"
)

// TalkService calls the KakaoTalk Message and KakaoTalk Social APIs. Every
// call except [TalkService.Profiles] acts as the user with their access
// token.
//
// Docs: https://developers.kakao.com/docs/ko/kakaotalk-message/rest-api
// and https://developers.kakao.com/docs/ko/kakaotalk-social/rest-api
type TalkService struct {
	c *Client
}

// maxReceivers is the most friends a message can be sent to at once.
const maxReceivers = 5

// CustomTemplate is a message template made in the app's message template
// tool (사용자 정의 템플릿).
type CustomTemplate struct {
	// ID is the template ID.
	ID int64
	// Args fills the template's ${KEY} arguments. Arguments left out are
	// shown literally in the message.
	Args map[string]string
}

// ScrapOptions are the optional parameters of the scrap message sends.
type ScrapOptions struct {
	// Template, if set, lays out the scraped page with a custom template.
	Template *CustomTemplate
}

// SendToMe sends a default template message to the user's own chat
// ("나와의 채팅"). It needs the talk_message consent item.
//
// Docs: https://developers.kakao.com/docs/ko/kakaotalk-message/rest-api
func (s *TalkService) SendToMe(ctx context.Context, accessToken string, t template.Template) error {
	form, err := templateForm(t)
	if err != nil {
		return err
	}
	return s.sendToMe(ctx, accessToken, "/v2/api/talk/memo/default/send", form)
}

// SendCustomToMe sends a custom template message to the user's own chat.
//
// Docs: https://developers.kakao.com/docs/ko/kakaotalk-message/rest-api
func (s *TalkService) SendCustomToMe(ctx context.Context, accessToken string, t CustomTemplate) error {
	form, err := customTemplateForm(&t)
	if err != nil {
		return err
	}
	return s.sendToMe(ctx, accessToken, "/v2/api/talk/memo/send", form)
}

// SendScrapToMe sends a message built from a web page to the user's own
// chat. The page's domain must be registered in the app settings.
//
// Docs: https://developers.kakao.com/docs/ko/kakaotalk-message/rest-api
func (s *TalkService) SendScrapToMe(ctx context.Context, accessToken, requestURL string, opts *ScrapOptions) error {
	form, err := scrapForm(requestURL, opts)
	if err != nil {
		return err
	}
	return s.sendToMe(ctx, accessToken, "/v2/api/talk/memo/scrap/send", form)
}

func (s *TalkService) sendToMe(ctx context.Context, accessToken, path string, form url.Values) error {
	var res struct {
		ResultCode int `json:"result_code"`
	}
	err := s.c.doAs(ctx, AccessToken(accessToken), &request{method: http.MethodPost, host: HostAPI, path: path, form: form}, &res)
	if err != nil {
		return err
	}
	if res.ResultCode != 0 {
		return fmt.Errorf("kakao: send to me: result_code %d", res.ResultCode)
	}
	return nil
}

// SendResult reports which friends a message reached.
type SendResult struct {
	SuccessfulReceiverUUIDs []string `json:"successful_receiver_uuids"`
	// FailureInfo groups failed receivers by error. It is only set when
	// sending to more than one friend; a failure for a single receiver is
	// returned as an [*Error] instead.
	FailureInfo []SendFailure `json:"failure_info"`
}

// SendFailure is a group of receivers that failed with the same error.
type SendFailure struct {
	// Code is the Kakao error code, e.g. -532 (daily quota exceeded).
	Code          int      `json:"code"`
	Msg           string   `json:"msg"`
	ReceiverUUIDs []string `json:"receiver_uuids"`
}

// SendToFriends sends a default template message to up to 5 friends,
// identified by the UUIDs from [TalkService.Friends]. It needs the
// talk_message consent item and permission to message friends.
//
// Partial failures are reported in [SendResult.FailureInfo], not as an
// error.
//
// Docs: https://developers.kakao.com/docs/ko/kakaotalk-message/rest-api
func (s *TalkService) SendToFriends(ctx context.Context, accessToken string, receiverUUIDs []string, t template.Template) (*SendResult, error) {
	form, err := templateForm(t)
	if err != nil {
		return nil, err
	}
	return s.sendToFriends(ctx, accessToken, "/v1/api/talk/friends/message/default/send", receiverUUIDs, form)
}

// SendCustomToFriends sends a custom template message to up to 5 friends.
//
// Docs: https://developers.kakao.com/docs/ko/kakaotalk-message/rest-api
func (s *TalkService) SendCustomToFriends(ctx context.Context, accessToken string, receiverUUIDs []string, t CustomTemplate) (*SendResult, error) {
	form, err := customTemplateForm(&t)
	if err != nil {
		return nil, err
	}
	return s.sendToFriends(ctx, accessToken, "/v1/api/talk/friends/message/send", receiverUUIDs, form)
}

// SendScrapToFriends sends a message built from a web page to up to 5
// friends.
//
// Docs: https://developers.kakao.com/docs/ko/kakaotalk-message/rest-api
func (s *TalkService) SendScrapToFriends(ctx context.Context, accessToken string, receiverUUIDs []string, requestURL string, opts *ScrapOptions) (*SendResult, error) {
	form, err := scrapForm(requestURL, opts)
	if err != nil {
		return nil, err
	}
	return s.sendToFriends(ctx, accessToken, "/v1/api/talk/friends/message/scrap/send", receiverUUIDs, form)
}

func (s *TalkService) sendToFriends(ctx context.Context, accessToken, path string, receiverUUIDs []string, form url.Values) (*SendResult, error) {
	if len(receiverUUIDs) == 0 || len(receiverUUIDs) > maxReceivers {
		return nil, fmt.Errorf("kakao: need 1 to %d receiver UUIDs, got %d", maxReceivers, len(receiverUUIDs))
	}
	if err := setJSON(form, "receiver_uuids", receiverUUIDs); err != nil {
		return nil, err
	}
	var res SendResult
	err := s.c.doAs(ctx, AccessToken(accessToken), &request{method: http.MethodPost, host: HostAPI, path: path, form: form}, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func templateForm(t template.Template) (url.Values, error) {
	if t == nil {
		return nil, errors.New("kakao: nil template")
	}
	b, err := json.Marshal(t)
	if err != nil {
		return nil, fmt.Errorf("kakao: encode template: %w", err)
	}
	return url.Values{"template_object": {string(b)}}, nil
}

func customTemplateForm(t *CustomTemplate) (url.Values, error) {
	if t.ID == 0 {
		return nil, errors.New("kakao: empty template ID")
	}
	form := url.Values{"template_id": {strconv.FormatInt(t.ID, 10)}}
	if err := setJSON(form, "template_args", t.Args); err != nil {
		return nil, err
	}
	return form, nil
}

func scrapForm(requestURL string, opts *ScrapOptions) (url.Values, error) {
	if requestURL == "" {
		return nil, errors.New("kakao: empty scrap URL")
	}
	form := url.Values{}
	if opts != nil && opts.Template != nil {
		var err error
		if form, err = customTemplateForm(opts.Template); err != nil {
			return nil, err
		}
	}
	form.Set("request_url", requestURL)
	return form, nil
}

// TalkProfile is the user's KakaoTalk profile.
type TalkProfile struct {
	// ID is the user ID.
	ID              int64  `json:"id"`
	Nickname        string `json:"nickName"`
	ProfileImageURL string `json:"profileImageURL"` // 640x640
	ThumbnailURL    string `json:"thumbnailURL"`    // 110x110
}

// Profile returns the user's KakaoTalk profile. Fields are empty unless
// the user consented to them.
//
// Docs: https://developers.kakao.com/docs/ko/kakaotalk-social/rest-api
func (s *TalkService) Profile(ctx context.Context, accessToken string) (*TalkProfile, error) {
	var p TalkProfile
	err := s.c.doAs(ctx, AccessToken(accessToken), &request{method: http.MethodGet, host: HostAPI, path: "/v1/api/talk/profile"}, &p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// FriendOrder chooses what the friend list is sorted by.
type FriendOrder string

// Friend orders.
const (
	FriendOrderFavorite FriendOrder = "favorite" // favorites first (default)
	FriendOrderNickname FriendOrder = "nickname"
)

// FriendsOptions are the optional parameters of [TalkService.Friends].
type FriendsOptions struct {
	Offset int
	// Limit is the page size, up to 100 (default 10).
	Limit       int
	Descending  bool
	FriendOrder FriendOrder
}

// FriendsPage is a page of friends.
type FriendsPage struct {
	Elements      []Friend `json:"elements"`
	TotalCount    int      `json:"total_count"`
	FavoriteCount int      `json:"favorite_count"`
	BeforeURL     string   `json:"before_url"`
	// AfterURL is empty on the last page.
	AfterURL string `json:"after_url"`
}

// Friend is a KakaoTalk friend who also uses the app.
type Friend struct {
	// ID is the friend's user ID.
	ID int64 `json:"id"`
	// UUID identifies the friend when sending messages.
	UUID                  string `json:"uuid"`
	Favorite              bool   `json:"favorite"`
	ProfileNickname       string `json:"profile_nickname"`
	ProfileThumbnailImage string `json:"profile_thumbnail_image"`
}

// Friends returns a page of the user's KakaoTalk friends who use the app.
// It needs the friends consent item and permission.
//
// Docs: https://developers.kakao.com/docs/ko/kakaotalk-social/rest-api
func (s *TalkService) Friends(ctx context.Context, accessToken string, opts *FriendsOptions) (*FriendsPage, error) {
	q := url.Values{}
	if opts != nil {
		setInt(q, "offset", opts.Offset)
		setInt(q, "limit", opts.Limit)
		if opts.Descending {
			q.Set("order", "desc")
		}
		setString(q, "friend_order", string(opts.FriendOrder))
	}
	var page FriendsPage
	err := s.c.doAs(ctx, AccessToken(accessToken), &request{method: http.MethodGet, host: HostAPI, path: "/v1/api/talk/friends", query: q}, &page)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// AllFriends iterates over all of the user's friends who use the app,
// fetching pages as needed. Iteration stops at the first error, which is
// yielded.
func (s *TalkService) AllFriends(ctx context.Context, accessToken string, opts *FriendsOptions) iter.Seq2[Friend, error] {
	return func(yield func(Friend, error) bool) {
		o := FriendsOptions{Limit: 100}
		if opts != nil {
			o = *opts
		}
		for {
			page, err := s.Friends(ctx, accessToken, &o)
			if err != nil {
				yield(Friend{}, err)
				return
			}
			for _, f := range page.Elements {
				if !yield(f, nil) {
					return
				}
			}
			if page.AfterURL == "" || len(page.Elements) == 0 {
				return
			}
			o.Offset += len(page.Elements)
		}
	}
}

// TalkProfiles is the response of [TalkService.Profiles].
type TalkProfiles struct {
	// Viewer is the viewer's profile, when a viewer was given.
	Viewer *TalkProfileInfo `json:"viewer"`
	// Targets are the profiles as the viewer sees them, including
	// multi-profiles for friends of the viewer.
	Targets []TalkProfileInfo `json:"targets"`
}

// TalkProfileInfo is a KakaoTalk profile.
type TalkProfileInfo struct {
	UserID         int64  `json:"user_id"`
	Nickname       string `json:"nickname"`
	ProfileImage   string `json:"profile_image"`
	ThumbnailImage string `json:"thumbnail_image"`
}

// ProfilesOptions are the optional parameters of [TalkService.Profiles].
type ProfilesOptions struct {
	// ViewerID returns the profiles as this user sees them, including
	// multi-profiles, and the viewer's own profile.
	ViewerID int64
}

// Profiles returns the KakaoTalk profiles of several users with the admin
// key. The call fails if any target is not linked to the app.
//
// Docs: https://developers.kakao.com/docs/ko/kakaotalk-social/rest-api
func (s *TalkService) Profiles(ctx context.Context, userIDs []int64, opts *ProfilesOptions) (*TalkProfiles, error) {
	if len(userIDs) == 0 {
		return nil, errors.New("kakao: no user IDs")
	}
	q := url.Values{"target_id_type": {"user_id"}}
	if err := setJSON(q, "target_ids", userIDs); err != nil {
		return nil, err
	}
	if opts != nil && opts.ViewerID != 0 {
		q.Set("viewer_id", strconv.FormatInt(opts.ViewerID, 10))
		q.Set("viewer_id_type", "user_id")
	}
	var res TalkProfiles
	err := s.c.do(ctx, &request{method: http.MethodGet, host: HostAPI, path: "/v2/api/talk/profiles", auth: authAdminKey, query: q}, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
