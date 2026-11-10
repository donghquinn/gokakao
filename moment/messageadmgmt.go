package moment

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	kakao "github.com/donghquinn/gokakao"
)

// MessageAdType is the kind of a message ad (메시지광고 타입).
type MessageAdType string

// Message ad types.
const (
	MessageAdTypeBroadcast MessageAdType = "BROADCAST"       // 일반 메시지광고
	MessageAdTypeMulti     MessageAdType = "MULTI_BROADCAST" // 최적화 메시지광고
)

// MessageAdStatus is the state of a message ad (메시지 상태).
type MessageAdStatus string

// Message ad statuses.
const (
	MessageAdDraft     MessageAdStatus = "DRAFT"      // 임시저장
	MessageAdReady     MessageAdStatus = "READY"      // 발송전
	MessageAdSending   MessageAdStatus = "SENDING"    // 발송중
	MessageAdPause     MessageAdStatus = "PAUSE"      // 발송중지
	MessageAdFinished  MessageAdStatus = "FINISHED"   // 발송완료
	MessageAdAdminStop MessageAdStatus = "ADMIN_STOP" // 관리자정지
	MessageAdDeleted   MessageAdStatus = "DEL"        // 삭제
)

// MessageAdDimension groups a message ad report (보고서 조회 기준).
type MessageAdDimension string

// Message ad report dimensions. Each adds a key to the row's Dimensions.
const (
	// MessageAdDimensionMessageAd adds "message_ad_id".
	MessageAdDimensionMessageAd MessageAdDimension = "MESSAGE_AD"
	// MessageAdDimensionMessageType adds "message_type", a [MessageType].
	MessageAdDimensionMessageType MessageAdDimension = "MESSAGE_TYPE"
	// MessageAdDimensionDeviceType adds "device_type": total, PC, Android,
	// iOS or N/A.
	MessageAdDimensionDeviceType MessageAdDimension = "DEVICE_TYPE"
	// MessageAdDimensionHour adds "hour": total or "00" to "23".
	MessageAdDimensionHour MessageAdDimension = "HOUR"
)

// MetricsMessageCarouselView is the carousel view metrics group of message
// ad reports (메시지 노출 지표): msg_carousel_intro_view and
// msg_carousel1_view to msg_carousel6_view.
const MetricsMessageCarouselView MetricsGroup = "MESSAGE_CAROUSEL_VIEW"

// SimpleMessage is a summary of a message (메시지 간편 정보).
type SimpleMessage struct {
	Type MessageType `json:"type"`
	// MainImageURL is the representative image, empty when there is none.
	MainImageURL string `json:"mainImageUrl"`
	// MainTitle is the representative text: the title, or the description
	// when there is no title.
	MainTitle string `json:"mainTitle"`
}

// MessageAdSummary is a message ad in a list.
type MessageAdSummary struct {
	MessageAdID   string        `json:"messageAdId"`
	MessageAdType MessageAdType `json:"messageAdType"`
	// Message is set for a regular message ad.
	Message *SimpleMessage `json:"message"`
	// Messages are set, in order, for an optimized message ad.
	Messages           []SimpleMessage     `json:"messages"`
	Name               string              `json:"name"`
	Status             MessageAdStatus     `json:"status"`
	SendingReservation *SendingReservation `json:"sendingReservation"`
	// Metrics are the summary metrics: cost, msg_send, msg_send_fail,
	// msg_open and msg_click; nil before sending.
	Metrics          Metrics    `json:"metrics"`
	CreatedDate      kakao.Time `json:"createdDate"`
	LastModifiedDate kakao.Time `json:"lastModifiedDate"`
}

// MessageAdListOptions are the optional parameters of
// [MessageAdService.List].
type MessageAdListOptions struct {
	// Page is the 0-based page number.
	Page int
	// Limit is the page size.
	Limit int
	// From and To ("yyyy-MM-dd") are the metrics period; Kakao's default is
	// the last 2 years.
	From, To string
	// MessageAdIDs filters by message ad.
	MessageAdIDs []string
	// Statuses filters by status.
	Statuses []MessageAdStatus
	// MessageTypes filters by message type.
	MessageTypes []MessageType
	// Start and End ("yyyy-MM-dd") filter by send date; set both or
	// neither. End is at most today.
	Start, End string
	// SearchKeyword filters by a substring of the name.
	SearchKeyword string
}

// List returns a page of the channel's message ads, regular and
// optimized.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/msg-ad-mgmt
func (s *MessageAdService) List(ctx context.Context, adAccountID int64, channelProfileID string, opts *MessageAdListOptions) (*Page[MessageAdSummary], error) {
	if err := checkMessageAdHeaders(adAccountID, channelProfileID); err != nil {
		return nil, err
	}
	type filter struct {
		MessageAdType MessageAdType     `json:"messageAdType"`
		MessageAdIDs  []string          `json:"messageAdIds,omitempty"`
		Statuses      []MessageAdStatus `json:"statuses,omitempty"`
		MessageTypes  []MessageType     `json:"messageTypes,omitempty"`
		Start         string            `json:"start,omitempty"`
		End           string            `json:"end,omitempty"`
		SearchKeyword string            `json:"searchKeyword,omitempty"`
	}
	// The docs fix messageAdType to BROADCAST; optimized ads are still
	// listed.
	f := filter{MessageAdType: MessageAdTypeBroadcast}
	q := url.Values{}
	if opts != nil {
		if (opts.Start == "") != (opts.End == "") {
			return nil, errors.New("moment: set both start and end or neither")
		}
		if opts.Page > 0 {
			q.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.Limit > 0 {
			q.Set("limit", strconv.Itoa(opts.Limit))
		}
		if opts.From != "" {
			q.Set("from", opts.From)
		}
		if opts.To != "" {
			q.Set("to", opts.To)
		}
		f.MessageAdIDs, f.Statuses, f.MessageTypes = opts.MessageAdIDs, opts.Statuses, opts.MessageTypes
		f.Start, f.End, f.SearchKeyword = opts.Start, opts.End, opts.SearchKeyword
	}
	c := messageAdCall(http.MethodPost, "", adAccountID, channelProfileID)
	c.query = q
	c.body = struct {
		Filter filter `json:"filter"`
	}{f}
	var p Page[MessageAdSummary]
	if err := s.c.do(ctx, c, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// MessageAdDetail is a message ad with its message and reservation.
type MessageAdDetail struct {
	MessageAdID   string        `json:"messageAdId"`
	MessageAdType MessageAdType `json:"messageAdType"`
	Name          string        `json:"name"`
	// Message omits messageAdId and name; use the fields above.
	Message MessageAdMessage `json:"message"`
	// SendingReservation is empty before a send is reserved.
	SendingReservation *SendingReservation `json:"sendingReservation"`
	Status             MessageAdStatus     `json:"status"`
	AgeVerification    bool                `json:"ageVerification"`
	CreatedDate        kakao.Time          `json:"createdDate"`
	LastModifiedDate   kakao.Time          `json:"lastModifiedDate"`
}

// Get returns a message ad with its message and sending reservation.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/msg-ad-mgmt
func (s *MessageAdService) Get(ctx context.Context, adAccountID int64, channelProfileID, messageAdID string) (*MessageAdDetail, error) {
	if err := checkMessageAdIDs(adAccountID, channelProfileID, messageAdID); err != nil {
		return nil, err
	}
	var d MessageAdDetail
	if err := s.c.do(ctx, messageAdCall(http.MethodGet, messageAdPath(messageAdID, ""), adAccountID, channelProfileID), &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// Delete deletes a message ad. It cannot be read or restored afterwards.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/msg-ad-mgmt
func (s *MessageAdService) Delete(ctx context.Context, adAccountID int64, channelProfileID, messageAdID string) error {
	if err := checkMessageAdIDs(adAccountID, channelProfileID, messageAdID); err != nil {
		return err
	}
	return s.c.do(ctx, messageAdCall(http.MethodDelete, messageAdPath(messageAdID, ""), adAccountID, channelProfileID), nil)
}

// GetMessage returns the message of a message ad: Message for a regular
// one, Multi for an optimized one.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/msg-ad-mgmt
func (s *MessageAdService) GetMessage(ctx context.Context, adAccountID int64, channelProfileID, messageAdID string) (*MessageAdMessage, error) {
	if err := checkMessageAdIDs(adAccountID, channelProfileID, messageAdID); err != nil {
		return nil, err
	}
	return s.doMessage(ctx, messageAdCall(http.MethodGet, messageAdPath(messageAdID, "/message"), adAccountID, channelProfileID))
}

// Copy copies the message of a message ad into a new message ad and
// returns it: Message for a regular one, Multi for an optimized one.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/msg-ad-mgmt
func (s *MessageAdService) Copy(ctx context.Context, adAccountID int64, channelProfileID, messageAdID string) (*MessageAdMessage, error) {
	if err := checkMessageAdIDs(adAccountID, channelProfileID, messageAdID); err != nil {
		return nil, err
	}
	return s.doMessage(ctx, messageAdCall(http.MethodPost, messageAdPath(messageAdID, "/copy"), adAccountID, channelProfileID))
}

// MessageAdReportQuery selects a message ad report.
type MessageAdReportQuery struct {
	// MessageAdIDs limits the report to up to 100 message ads.
	MessageAdIDs []string `json:"messageAdIds,omitempty"`
	// DatePreset is the period when Start and End are empty; Kakao's
	// default is TODAY.
	DatePreset DatePreset `json:"datePreset,omitempty"`
	// Start and End ("yyyy-MM-dd", at most today) are the period; set both
	// or neither. They take precedence over DatePreset.
	Start string `json:"start,omitempty"`
	End   string `json:"end,omitempty"`
	// Dimension groups the rows; required.
	Dimension MessageAdDimension `json:"dimension"`
	// MetricsGroups are the metrics to return: MESSAGE, MESSAGE_ADDITION,
	// PIXEL_SDK_CONVERSION, MESSAGE_CLICK or MESSAGE_CAROUSEL_VIEW;
	// required.
	MetricsGroups []MetricsGroup `json:"metricsGroup"`
}

// Report returns a message ad report. A day's figures may change until
// 08:00 the next day.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/msg-ad-mgmt
func (s *MessageAdService) Report(ctx context.Context, adAccountID int64, channelProfileID string, q MessageAdReportQuery) ([]ReportRow, error) {
	if err := checkMessageAdHeaders(adAccountID, channelProfileID); err != nil {
		return nil, err
	}
	if q.Dimension == "" || len(q.MetricsGroups) == 0 {
		return nil, errors.New("moment: dimension and metrics groups are required")
	}
	if (q.Start == "") != (q.End == "") {
		return nil, errors.New("moment: set both start and end or neither")
	}
	if len(q.MessageAdIDs) > 100 {
		return nil, errors.New("moment: at most 100 message ad IDs are allowed")
	}
	c := messageAdCall(http.MethodPost, "/reports", adAccountID, channelProfileID)
	c.body = q
	var res struct {
		Data []ReportRow `json:"data"`
	}
	if err := s.c.do(ctx, c, &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}

// TalkChannelVideoThumbnail is a thumbnail of a KakaoTalk Channel video.
type TalkChannelVideoThumbnail struct {
	URL      string      `json:"url"`
	FileName string      `json:"fileName"`
	Width    int         `json:"width"`
	Height   int         `json:"height"`
	FileSize kakao.Int64 `json:"fileSize"` // bytes
	// MimeType is documented as "mineType".
	MimeType string `json:"mineType"`
}

// TalkChannelVideo is a KakaoTalk Channel video (채널 영상) for message
// items; use MediaID as the item's video.mediaId.
type TalkChannelVideo struct {
	URL                  string                     `json:"url"`
	Title                string                     `json:"title"`
	TalkChannelProfileID string                     `json:"talkChannelProfileId"`
	VID                  string                     `json:"vid"`
	MediaID              kakao.Int64                `json:"mediaId"`
	Status               string                     `json:"status"`   // e.g. PUBLIC
	Duration             int                        `json:"duration"` // seconds
	Width                int                        `json:"width"`
	Height               int                        `json:"height"`
	PlayCount            int                        `json:"playCount"`
	AutoThumbnail        *TalkChannelVideoThumbnail `json:"autoThumbnail"`
	UploadThumbnail      *TalkChannelVideoThumbnail `json:"uploadThumbnail"`
}

// TalkChannelVideoPage is a page of channel videos.
type TalkChannelVideoPage struct {
	Items         []TalkChannelVideo `json:"items"`
	TotalElements int                `json:"totalElements"`
}

// TalkChannelVideos returns a page of the KakaoTalk Channel videos usable
// in message ads. The channel is determined by the business token, so no
// ad account or channel profile is sent. Kakao's default page size is 20,
// at most 1000.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/msg-ad-mgmt
func (s *MessageAdService) TalkChannelVideos(ctx context.Context, opts *PageOptions) (*TalkChannelVideoPage, error) {
	var p TalkChannelVideoPage
	c := call{method: http.MethodGet, path: "/message/v2/message-ads/talk-channel-videos", query: opts.query(nil)}
	if err := s.c.do(ctx, c, &p); err != nil {
		return nil, err
	}
	return &p, nil
}
