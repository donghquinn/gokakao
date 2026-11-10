package moment

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strings"

	kakao "github.com/donghquinn/gokakao"
)

// Message ads (메시지광고) are sent from a KakaoTalk Channel. Every message
// ad method takes the ad account ID and the channel profile ID (카카오톡
// 채널 프로필 ID, e.g. "_ZQxd"), sent as the adAccountId and
// channel-profile-id headers. Message ad IDs are opaque strings such as
// "msg-ad-1402922220759433217".

// messageAdCall returns a call to a /message/v2/message-ads path with the
// ad account and channel profile headers.
func messageAdCall(method, path string, adAccountID int64, channelProfileID string) call {
	return call{
		method:      method,
		path:        "/message/v2/message-ads" + path,
		adAccountID: adAccountID,
		header:      http.Header{"Channel-Profile-Id": {channelProfileID}},
	}
}

// messageAdPath returns "/<escaped messageAdID><suffix>".
func messageAdPath(messageAdID, suffix string) string {
	return "/" + url.PathEscape(messageAdID) + suffix
}

func checkMessageAdHeaders(adAccountID int64, channelProfileID string) error {
	if adAccountID == 0 || channelProfileID == "" {
		return errors.New("moment: ad account ID and channel profile ID are required")
	}
	return nil
}

func checkMessageAdIDs(adAccountID int64, channelProfileID, messageAdID string) error {
	if adAccountID == 0 || channelProfileID == "" || messageAdID == "" {
		return errors.New("moment: ad account ID, channel profile ID and message ad ID are required")
	}
	return nil
}

// MessageAdMessage is the message of a message ad: Message for a regular
// message ad, Multi for an optimized one (최적화 메시지광고). Exactly one is
// set.
type MessageAdMessage struct {
	Message *ResponseMessage
	Multi   *ResponseMultiMessage
}

// UnmarshalJSON implements [json.Unmarshaler]. A body with a "messages"
// array or a messageAdType of MULTI_BROADCAST decodes into Multi.
//
// The docs type Item.landing as Landing[] but the examples send a single
// object; both are accepted.
func (m *MessageAdMessage) UnmarshalJSON(b []byte) error {
	*m = MessageAdMessage{}
	if bytes.Equal(bytes.TrimSpace(b), []byte("null")) {
		return nil
	}
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(b, &probe); err != nil {
		return err
	}
	if probe == nil {
		return nil
	}
	var adType MessageAdType
	if raw, ok := probe["messageAdType"]; ok {
		_ = json.Unmarshal(raw, &adType)
	}
	if raw, ok := probe["messages"]; (ok && !bytes.Equal(raw, []byte("null"))) || adType == MessageAdTypeMulti {
		var msgs []map[string]json.RawMessage
		if ok {
			if err := json.Unmarshal(raw, &msgs); err != nil {
				return err
			}
		}
		for _, msg := range msgs {
			if err := messageAdWrapLandings(msg); err != nil {
				return err
			}
		}
		if msgs != nil {
			nb, err := json.Marshal(msgs)
			if err != nil {
				return err
			}
			probe["messages"] = nb
		}
		nb, err := json.Marshal(probe)
		if err != nil {
			return err
		}
		m.Multi = new(ResponseMultiMessage)
		return json.Unmarshal(nb, m.Multi)
	}
	if err := messageAdWrapLandings(probe); err != nil {
		return err
	}
	nb, err := json.Marshal(probe)
	if err != nil {
		return err
	}
	m.Message = new(ResponseMessage)
	return json.Unmarshal(nb, m.Message)
}

// messageAdWrapLandings rewrites msg.items[].landing objects as one-element
// arrays so they decode into Item.Landing.
func messageAdWrapLandings(msg map[string]json.RawMessage) error {
	raw, ok := msg["items"]
	if !ok || bytes.Equal(raw, []byte("null")) {
		return nil
	}
	var items []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		return err
	}
	changed := false
	for _, it := range items {
		if l, ok := it["landing"]; ok {
			if l = bytes.TrimSpace(l); len(l) > 0 && l[0] == '{' {
				it["landing"] = append(append([]byte{'['}, l...), ']')
				changed = true
			}
		}
	}
	if !changed {
		return nil
	}
	nb, err := json.Marshal(items)
	if err != nil {
		return err
	}
	msg["items"] = nb
	return nil
}

// toMessage returns the request form of a decoded message.
func (m *ResponseMessage) toMessage() Message {
	return Message{
		Type: m.Type, MessageTitle: m.MessageTitle, Description: m.Description,
		Items: m.Items, Buttons: m.Buttons, CouponBook: m.CouponBook,
		IntroCarousel: m.IntroCarousel, Carousels: m.Carousels, Catalog: m.Catalog,
		MoreButton: m.MoreButton, ButtonTitle: m.ButtonTitle, LandingURL: m.LandingURL,
		AdFlag: m.AdFlag, ShareFlag: m.ShareFlag,
	}
}

// Changes returns the [MessageAdChanges] that keep m as it is, for editing
// before [MessageAdService.UpdateMessage].
func (m *ResponseMessage) Changes() MessageAdChanges {
	return MessageAdChanges{Name: m.Name, AgeVerification: m.AgeVerification, Message: m.toMessage()}
}

// Changes returns the [OptimizedMessageAdChanges] that keep m as it is,
// for editing before [MessageAdService.UpdateOptimized].
func (m *ResponseMultiMessage) Changes() OptimizedMessageAdChanges {
	ch := OptimizedMessageAdChanges{Name: m.Name, AgeVerification: m.AgeVerification}
	for _, c := range m.Messages {
		ch.Messages = append(ch.Messages, Message{
			Type: c.Type, MessageTitle: c.MessageTitle, Description: c.Description,
			Items: c.Items, Buttons: c.Buttons, CouponBook: c.CouponBook,
			IntroCarousel: c.IntroCarousel, Carousels: c.Carousels, MoreButton: c.MoreButton,
			ButtonTitle: c.ButtonTitle, LandingURL: c.LandingURL,
			AdFlag: c.AdFlag, ShareFlag: c.ShareFlag,
		})
	}
	return ch
}

func (s *MessageAdService) doMessage(ctx context.Context, c call) (*MessageAdMessage, error) {
	var m MessageAdMessage
	if err := s.c.do(ctx, c, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *MessageAdService) doSingle(ctx context.Context, c call) (*ResponseMessage, error) {
	m, err := s.doMessage(ctx, c)
	if err != nil {
		return nil, err
	}
	if m.Message == nil {
		return nil, errors.New("moment: response is not a regular message ad")
	}
	return m.Message, nil
}

func (s *MessageAdService) doMulti(ctx context.Context, c call) (*ResponseMultiMessage, error) {
	m, err := s.doMessage(ctx, c)
	if err != nil {
		return nil, err
	}
	if m.Multi == nil {
		return nil, errors.New("moment: response is not an optimized message ad")
	}
	return m.Multi, nil
}

// NewMessageAd describes a message ad to save (메시지광고 저장).
type NewMessageAd struct {
	// Name is up to 150 characters; Kakao generates one when empty.
	Name string `json:"name,omitempty"`
	// AgeVerification marks a message that requires age verification
	// (연령인증); false by default.
	AgeVerification bool `json:"ageVerification"`
	// Message is the content. Which fields apply depends on its Type;
	// PREMIUM_VIDEO_MESSAGE cannot be created through the API.
	Message Message `json:"message"`
}

// Create saves a message ad as a draft; reserve it with
// [MessageAdService.Reserve] to send it. It is limited to 1 request per
// second per user and ad account.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/msg-ad-ops
func (s *MessageAdService) Create(ctx context.Context, adAccountID int64, channelProfileID string, ad NewMessageAd) (*ResponseMessage, error) {
	if err := checkMessageAdHeaders(adAccountID, channelProfileID); err != nil {
		return nil, err
	}
	if ad.Message.Type == "" {
		return nil, errors.New("moment: message type is required")
	}
	c := messageAdCall(http.MethodPost, "/message", adAccountID, channelProfileID)
	c.body = ad
	return s.doSingle(ctx, c)
}

// MessageAdChanges is the new content of a message ad for
// [MessageAdService.UpdateMessage].
//
// Kakao replaces the message: start from [ResponseMessage.Changes] to keep
// the fields you do not change. The message type cannot change, and from 5
// minutes before sending only Name is applied.
type MessageAdChanges struct {
	// Name is up to 150 characters.
	Name string `json:"name,omitempty"`
	// AgeVerification is always sent; once a send is reserved it must stay
	// the same.
	AgeVerification bool    `json:"ageVerification"`
	Message         Message `json:"message"`
}

// UpdateMessage replaces the content of a message ad; see
// [MessageAdChanges].
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/msg-ad-ops
func (s *MessageAdService) UpdateMessage(ctx context.Context, adAccountID int64, channelProfileID, messageAdID string, ch MessageAdChanges) (*ResponseMessage, error) {
	if err := checkMessageAdIDs(adAccountID, channelProfileID, messageAdID); err != nil {
		return nil, err
	}
	if ch.Message.Type == "" {
		return nil, errors.New("moment: message type is required")
	}
	c := messageAdCall(http.MethodPatch, messageAdPath(messageAdID, "/message"), adAccountID, channelProfileID)
	c.body = ch
	return s.doSingle(ctx, c)
}

// SendTest sends a message ad as a test to phone numbers ("010-1234-5678")
// that are friends of the channel. "[테스트 발송]" is added to the copy. It
// is limited to 1 request per 30 seconds per user and ad account.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/msg-ad-ops
func (s *MessageAdService) SendTest(ctx context.Context, adAccountID int64, channelProfileID, messageAdID string, phoneNumbers []string) error {
	if err := checkMessageAdIDs(adAccountID, channelProfileID, messageAdID); err != nil {
		return err
	}
	if len(phoneNumbers) == 0 {
		return errors.New("moment: phone numbers are required")
	}
	c := messageAdCall(http.MethodPost, messageAdPath(messageAdID, "/send-test"), adAccountID, channelProfileID)
	c.body = struct {
		PhoneNumbers []string `json:"phoneNumbers"`
	}{phoneNumbers}
	return s.c.do(ctx, c, nil)
}

// SendingReservation is the sending setup of a message ad (발송 예약),
// regular or optimized.
type SendingReservation struct {
	MessageAdID string `json:"messageAdId"`
	// Targeting has OptimizationTarget set for an optimized message ad.
	Targeting MessageTargeting `json:"targeting"`
	// Price is the price per message in KRW (발송 단가), e.g. 15 without
	// and 20 with targeting.
	Price         kakao.Int64 `json:"price"`
	ContractCount kakao.Int64 `json:"contractCount"` // 구매발송수
	TotalBudget   kakao.Int64 `json:"totalBudget"`   // 구매 금액 (KRW)
	// TotalBudgetWithVAT is TotalBudget including VAT, e.g. 165.0.
	TotalBudgetWithVAT kakao.Float `json:"totalBudgetWithVAT"`
	// Date is when sending starts, "yyyy-MM-dd HH:mm".
	Date string `json:"date"`
	// StartedDate is when sending actually started, "yyyy-MM-dd HH:mm";
	// optimized message ads only, once sending has started.
	StartedDate   string `json:"startedDate"`
	ImmediateSend bool   `json:"immediateSend"`
	PushAlarm     bool   `json:"pushAlarm"`
	TrackID       string `json:"trackId"` // Pixel & SDK ID for conversions
	// SendRate is the spread per cycle (분산발송): 0, 100, 500, 1000, 1500
	// or 2000.
	SendRate int `json:"sendRate"`
}

// SendingReservationChanges is the sending setup of a regular message ad,
// for [MessageAdService.Reserve] and [MessageAdService.UpdateReservation].
//
// An update replaces the setup: omitted fields reset to Kakao's defaults.
// Start from [SendingReservation.Changes] to keep the current values.
type SendingReservationChanges struct {
	// Targeting selects the recipients; Type is required.
	Targeting MessageTargeting `json:"targeting"`
	// Date is when sending starts, "yyyy-MM-dd HH:mm" between 08:00 and
	// 20:50, at least 5 minutes ahead in 1-minute steps. Required unless
	// ImmediateSend, and not allowed with it.
	Date          string `json:"date,omitempty"`
	ImmediateSend bool   `json:"immediateSend,omitempty"`
	// PushAlarm sends a push notification; nil means Kakao's default
	// (true).
	PushAlarm *bool `json:"pushAlarm,omitempty"`
	// TrackID is a Pixel & SDK ID for conversion metrics; none when empty.
	TrackID string `json:"trackId,omitempty"`
	// SendRate spreads the send per cycle: 0 (none), 100, 500, 1000, 1500
	// or 2000.
	SendRate int `json:"sendRate,omitempty"`
}

// Changes returns the [SendingReservationChanges] that keep r as it is,
// for editing before [MessageAdService.UpdateReservation].
func (r *SendingReservation) Changes() SendingReservationChanges {
	push := r.PushAlarm
	ch := SendingReservationChanges{
		Targeting: r.Targeting, ImmediateSend: r.ImmediateSend, PushAlarm: &push,
		TrackID: r.TrackID, SendRate: r.SendRate,
	}
	ch.Targeting.OptimizationTarget = nil
	ch.Targeting.Interests, ch.Targeting.BusinessTypes = nil, nil // read-only
	if !r.ImmediateSend {
		ch.Date = r.Date
	}
	return ch
}

// OptimizedChanges returns the [OptimizedSendingReservationChanges] that
// keep r as it is, for editing before
// [MessageAdService.UpdateOptimizedReservation].
func (r *SendingReservation) OptimizedChanges() OptimizedSendingReservationChanges {
	push := r.PushAlarm
	t := r.Targeting
	ch := OptimizedSendingReservationChanges{
		Targeting: OptimizationTargeting{
			Type: t.Type, Devices: t.Devices, Demographics: t.Demographics,
			Customs: t.Customs, FriendPeriod: t.FriendPeriod,
			Interests: t.Interests, BusinessTypes: t.BusinessTypes,
		},
		ImmediateSend: r.ImmediateSend, PushAlarm: &push, TrackID: r.TrackID, SendRate: r.SendRate,
	}
	if o := t.OptimizationTarget; o != nil {
		ch.Targeting.OptimizationTarget = OptimizationTarget{Type: o.Type, SendRatios: o.SendRatios}
	}
	if !r.ImmediateSend {
		ch.Date = r.Date
	}
	return ch
}

func checkReservationDate(date string, immediate bool) error {
	if immediate && date != "" {
		return errors.New("moment: date is not allowed with immediate send")
	}
	if !immediate && date == "" {
		return errors.New("moment: date is required unless sending immediately")
	}
	return nil
}

func (s *MessageAdService) reserve(ctx context.Context, method string, adAccountID int64, channelProfileID, path string, r SendingReservationChanges) (*SendingReservation, error) {
	if r.Targeting.Type == "" {
		return nil, errors.New("moment: targeting type is required")
	}
	if err := checkReservationDate(r.Date, r.ImmediateSend); err != nil {
		return nil, err
	}
	c := messageAdCall(method, path, adAccountID, channelProfileID)
	c.body = r
	var res SendingReservation
	if err := s.c.do(ctx, c, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Reserve sets when and to whom a message ad is sent (메시지광고 발송
// 예약). Sends unsent by 20:55 go out after 08:00 the next day.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/msg-ad-ops
func (s *MessageAdService) Reserve(ctx context.Context, adAccountID int64, channelProfileID, messageAdID string, r SendingReservationChanges) (*SendingReservation, error) {
	if err := checkMessageAdIDs(adAccountID, channelProfileID, messageAdID); err != nil {
		return nil, err
	}
	return s.reserve(ctx, http.MethodPost, adAccountID, channelProfileID, messageAdPath(messageAdID, "/sending-reservation"), r)
}

// GetReservation returns the sending reservation of a message ad.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/msg-ad-ops
func (s *MessageAdService) GetReservation(ctx context.Context, adAccountID int64, channelProfileID, messageAdID string) (*SendingReservation, error) {
	if err := checkMessageAdIDs(adAccountID, channelProfileID, messageAdID); err != nil {
		return nil, err
	}
	var res SendingReservation
	c := messageAdCall(http.MethodGet, messageAdPath(messageAdID, "/sending-reservation"), adAccountID, channelProfileID)
	if err := s.c.do(ctx, c, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// UpdateReservation replaces the sending reservation of a message ad; see
// [SendingReservationChanges]. It is allowed until 5 minutes before
// sending.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/msg-ad-ops
func (s *MessageAdService) UpdateReservation(ctx context.Context, adAccountID int64, channelProfileID, messageAdID string, r SendingReservationChanges) (*SendingReservation, error) {
	if err := checkMessageAdIDs(adAccountID, channelProfileID, messageAdID); err != nil {
		return nil, err
	}
	return s.reserve(ctx, http.MethodPatch, adAccountID, channelProfileID, messageAdPath(messageAdID, "/sending-reservation"), r)
}

// MessageLocation is a region for message ad targeting (타게팅 지역).
type MessageLocation struct {
	// ID is the region code, e.g. "A" (시/도), "A7000" or "A7000A001".
	ID   string `json:"id"`
	Name string `json:"name"`
	// Children are the sub-regions; not returned when looking up codes.
	Children []MessageLocation `json:"children"`
	// Deprecated reports a removed region; only returned when looking up
	// codes.
	Deprecated bool `json:"deprecated"`
}

// MessageLocationOptions are the optional parameters of
// [MessageAdService.Locations].
type MessageLocationOptions struct {
	// Codes looks up these region codes, e.g. "A7000A001", instead of
	// listing the region tree.
	Codes []string
}

// Locations returns the regions for message ad targeting. Without codes
// it returns the region tree; with codes, the named regions with full
// names and no children. The full list is in the administrative district
// CSV linked from the docs.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/msg-ad-ops
func (s *MessageAdService) Locations(ctx context.Context, adAccountID int64, channelProfileID string, opts *MessageLocationOptions) ([]MessageLocation, error) {
	if err := checkMessageAdHeaders(adAccountID, channelProfileID); err != nil {
		return nil, err
	}
	c := messageAdCall(http.MethodGet, "/sending-reservation/location", adAccountID, channelProfileID)
	if opts != nil && len(opts.Codes) > 0 {
		c.query = url.Values{"codes": {strings.Join(opts.Codes, ",")}}
	}
	// The docs show a single object without codes and an array with them.
	var raw json.RawMessage
	if err := s.c.do(ctx, c, &raw); err != nil {
		return nil, err
	}
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return nil, nil
	}
	if raw[0] == '{' {
		var l MessageLocation
		if err := json.Unmarshal(raw, &l); err != nil {
			return nil, err
		}
		return []MessageLocation{l}, nil
	}
	var ls []MessageLocation
	if err := json.Unmarshal(raw, &ls); err != nil {
		return nil, err
	}
	return ls, nil
}

// TargetingPrice is the audience size and price of a message ad targeting
// (타게팅 모수 및 단가).
type TargetingPrice struct {
	MessageAdID   string      `json:"messageAdId"`
	Price         kakao.Int64 `json:"price"`         // 발송 단가 (KRW)
	ContractCount kakao.Int64 `json:"contractCount"` // 구매발송수
	// ChannelFriendCount is the channel's friend count; optimized message
	// ads only.
	ChannelFriendCount kakao.Int64 `json:"channelFriendCount"`
	TotalBudget        kakao.Int64 `json:"totalBudget"`
	TotalBudgetWithVAT kakao.Float `json:"totalBudgetWithVAT"`
}

// TargetingPrice returns the audience size and price for a targeting. It
// only uses the given targeting and does not change the message ad or its
// reservation.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/msg-ad-ops
func (s *MessageAdService) TargetingPrice(ctx context.Context, adAccountID int64, channelProfileID, messageAdID string, targeting MessageTargeting) (*TargetingPrice, error) {
	if err := checkMessageAdIDs(adAccountID, channelProfileID, messageAdID); err != nil {
		return nil, err
	}
	if targeting.Type == "" {
		return nil, errors.New("moment: targeting type is required")
	}
	c := messageAdCall(http.MethodPost, messageAdPath(messageAdID, "/sending-reservation/targeting-price"), adAccountID, channelProfileID)
	c.body = struct {
		Targeting MessageTargeting `json:"targeting"`
	}{targeting}
	var res TargetingPrice
	if err := s.c.do(ctx, c, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// NewOptimizedMessageAd describes an optimized message ad to save
// (최적화 메시지광고 저장).
type NewOptimizedMessageAd struct {
	// Name is up to 150 characters; Kakao generates one when empty.
	Name string `json:"name,omitempty"`
	// AgeVerification applies to all messages; false by default.
	AgeVerification bool `json:"ageVerification"`
	// Messages are the 2 to 5 messages to optimize between.
	Messages []Message `json:"messages"`
}

func checkOptimizedMessages(msgs []Message) error {
	if len(msgs) < 2 || len(msgs) > 5 {
		return errors.New("moment: 2 to 5 messages are required")
	}
	for _, m := range msgs {
		if m.Type == "" {
			return errors.New("moment: message type is required")
		}
	}
	return nil
}

// CreateOptimized saves an optimized message ad as a draft; reserve it with
// [MessageAdService.ReserveOptimized] to send it.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/msg-ad-ops
func (s *MessageAdService) CreateOptimized(ctx context.Context, adAccountID int64, channelProfileID string, ad NewOptimizedMessageAd) (*ResponseMultiMessage, error) {
	if err := checkMessageAdHeaders(adAccountID, channelProfileID); err != nil {
		return nil, err
	}
	if err := checkOptimizedMessages(ad.Messages); err != nil {
		return nil, err
	}
	c := messageAdCall(http.MethodPost, "/optimization/message", adAccountID, channelProfileID)
	c.body = ad
	return s.doMulti(ctx, c)
}

// OptimizedMessageAdChanges is the new content of an optimized message ad
// for [MessageAdService.UpdateOptimized].
//
// Kakao replaces the messages: start from [ResponseMultiMessage.Changes]
// to keep the ones you do not change.
type OptimizedMessageAdChanges struct {
	// Name keeps the current name when empty.
	Name string `json:"name,omitempty"`
	// AgeVerification is always sent; it cannot change once a send is
	// reserved.
	AgeVerification bool `json:"ageVerification"`
	// Messages replace the current ones position by position, so the count
	// must stay the same; their types cannot change once a send is
	// reserved.
	Messages []Message `json:"messages"`
}

// UpdateOptimized replaces the name and messages of an optimized message
// ad in the draft, ready or paused state; see [OptimizedMessageAdChanges].
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/msg-ad-ops
func (s *MessageAdService) UpdateOptimized(ctx context.Context, adAccountID int64, channelProfileID, messageAdID string, ch OptimizedMessageAdChanges) (*ResponseMultiMessage, error) {
	if err := checkMessageAdIDs(adAccountID, channelProfileID, messageAdID); err != nil {
		return nil, err
	}
	if err := checkOptimizedMessages(ch.Messages); err != nil {
		return nil, err
	}
	c := messageAdCall(http.MethodPatch, "/optimization"+messageAdPath(messageAdID, "/message"), adAccountID, channelProfileID)
	c.body = ch
	return s.doMulti(ctx, c)
}

// SendOptimizedTest sends one message of an optimized message ad, at the
// 0-based messageIndex, as a test to phone numbers ("010-0000-0000"). A
// successful request does not guarantee delivery. It is limited to 1
// request per 30 seconds per user and ad account.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/msg-ad-ops
func (s *MessageAdService) SendOptimizedTest(ctx context.Context, adAccountID int64, channelProfileID, messageAdID string, messageIndex int, phoneNumbers []string) error {
	if err := checkMessageAdIDs(adAccountID, channelProfileID, messageAdID); err != nil {
		return err
	}
	if messageIndex < 0 {
		return errors.New("moment: message index must not be negative")
	}
	if len(phoneNumbers) == 0 || slices.Contains(phoneNumbers, "") {
		return errors.New("moment: phone numbers are required")
	}
	c := messageAdCall(http.MethodPost, "/optimization"+messageAdPath(messageAdID, "/send-test"), adAccountID, channelProfileID)
	c.body = struct {
		MessageIndex int      `json:"messageIndex"`
		PhoneNumbers []string `json:"phoneNumbers"`
	}{messageIndex, phoneNumbers}
	return s.c.do(ctx, c, nil)
}

// OptimizedSendingReservationChanges is the sending setup of an optimized
// message ad, for [MessageAdService.ReserveOptimized] and
// [MessageAdService.UpdateOptimizedReservation].
//
// An update replaces the setup, including every targeting option: start
// from [SendingReservation.OptimizedChanges] to keep the current values.
type OptimizedSendingReservationChanges struct {
	// Targeting.Type is OPTIMIZATION_WHOLE or OPTIMIZATION_DIRECT. For
	// RATIO, SendRatios follow the order of the ad's messages.
	Targeting OptimizationTargeting `json:"targeting"`
	// Date is when sending starts, "yyyy-MM-dd HH:mm" between 08:00 and
	// 20:50, from 5 minutes to 30 days ahead. Required unless
	// ImmediateSend, and not allowed with it.
	Date string `json:"date,omitempty"`
	// ImmediateSend starts sending 5 minutes from now.
	ImmediateSend bool `json:"immediateSend,omitempty"`
	// PushAlarm sends a push notification; nil means Kakao's default
	// (true).
	PushAlarm *bool  `json:"pushAlarm,omitempty"`
	TrackID   string `json:"trackId,omitempty"` // Pixel & SDK ID
	// SendRate spreads the send per cycle: 0 (none), 100, 500, 1000, 1500
	// or 2000.
	SendRate int `json:"sendRate,omitempty"`
}

func checkOptimizationTargeting(t OptimizationTargeting) error {
	if t.Type != MessageTargetingOptimizationWhole && t.Type != MessageTargetingOptimizationDirect {
		return errors.New("moment: targeting type must be OPTIMIZATION_WHOLE or OPTIMIZATION_DIRECT")
	}
	if t.OptimizationTarget.Type == "" {
		return errors.New("moment: optimization type is required")
	}
	return nil
}

func (s *MessageAdService) reserveOptimized(ctx context.Context, method string, adAccountID int64, channelProfileID, messageAdID string, r OptimizedSendingReservationChanges) (*SendingReservation, error) {
	if err := checkMessageAdIDs(adAccountID, channelProfileID, messageAdID); err != nil {
		return nil, err
	}
	if err := checkOptimizationTargeting(r.Targeting); err != nil {
		return nil, err
	}
	if err := checkReservationDate(r.Date, r.ImmediateSend); err != nil {
		return nil, err
	}
	c := messageAdCall(method, "/optimization"+messageAdPath(messageAdID, "/sending-reservation"), adAccountID, channelProfileID)
	c.body = r
	var res SendingReservation
	if err := s.c.do(ctx, c, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// ReserveOptimized sets when and to whom an optimized message ad is sent,
// and how its messages are split.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/msg-ad-ops
func (s *MessageAdService) ReserveOptimized(ctx context.Context, adAccountID int64, channelProfileID, messageAdID string, r OptimizedSendingReservationChanges) (*SendingReservation, error) {
	return s.reserveOptimized(ctx, http.MethodPost, adAccountID, channelProfileID, messageAdID, r)
}

// UpdateOptimizedReservation replaces the sending reservation of an
// optimized message ad; see [OptimizedSendingReservationChanges]. It is
// allowed until 5 minutes before sending, and not once sending started.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/msg-ad-ops
func (s *MessageAdService) UpdateOptimizedReservation(ctx context.Context, adAccountID int64, channelProfileID, messageAdID string, r OptimizedSendingReservationChanges) (*SendingReservation, error) {
	return s.reserveOptimized(ctx, http.MethodPatch, adAccountID, channelProfileID, messageAdID, r)
}

// OptimizedTargetingPrice returns the audience size, channel friend count
// and price for an optimized message ad targeting. It does not change the
// reservation; the result is checked again when reserving.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/msg-ad-ops
func (s *MessageAdService) OptimizedTargetingPrice(ctx context.Context, adAccountID int64, channelProfileID, messageAdID string, targeting OptimizationTargeting) (*TargetingPrice, error) {
	if err := checkMessageAdIDs(adAccountID, channelProfileID, messageAdID); err != nil {
		return nil, err
	}
	if err := checkOptimizationTargeting(targeting); err != nil {
		return nil, err
	}
	c := messageAdCall(http.MethodPost, "/optimization"+messageAdPath(messageAdID, "/sending-reservation/targeting-price"), adAccountID, channelProfileID)
	c.body = struct {
		Targeting OptimizationTargeting `json:"targeting"`
	}{targeting}
	var res TargetingPrice
	if err := s.c.do(ctx, c, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Finish ends a message ad (메시지광고 종료): sending stops and a reserved
// send is canceled.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/msg-ad-ops
func (s *MessageAdService) Finish(ctx context.Context, adAccountID int64, channelProfileID, messageAdID string) error {
	if err := checkMessageAdIDs(adAccountID, channelProfileID, messageAdID); err != nil {
		return err
	}
	return s.c.do(ctx, messageAdCall(http.MethodPut, messageAdPath(messageAdID, "/finish"), adAccountID, channelProfileID), nil)
}
