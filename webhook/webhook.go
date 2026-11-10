// Package webhook receives the webhooks Kakao sends to a service: user
// unlinks, account status changes (Security Event Tokens), KakaoTalk
// Channel add/block events, KakaoTalk Share deliveries and push delivery
// failures.
//
// Each webhook has a Parse function that authenticates and decodes a
// request, and a Handler that also writes the response Kakao requires.
// Kakao expects a response within 3 seconds, so callbacks should hand
// slow work off to a queue or goroutine.
//
// Docs: https://developers.kakao.com/docs/ko/getting-started/callback
package webhook

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// maxBodySize limits webhook request bodies.
const maxBodySize = 64 << 10

// ErrUnauthorized is returned when a request does not carry the app's
// primary admin key in its Authorization header.
var ErrUnauthorized = errors.New("webhook: invalid Authorization header")

// authorize checks the "Authorization: KakaoAK <primary admin key>" header
// in constant time.
func authorize(r *http.Request, primaryAdminKey string) error {
	if primaryAdminKey == "" {
		return errors.New("webhook: empty primary admin key")
	}
	want := "KakaoAK " + primaryAdminKey
	if subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte(want)) != 1 {
		return ErrUnauthorized
	}
	return nil
}

// status returns the HTTP status for a parse error.
func status(err error) int {
	if errors.Is(err, ErrUnauthorized) {
		return http.StatusUnauthorized
	}
	return http.StatusBadRequest
}

// readBody reads up to maxBodySize bytes of the request body.
func readBody(r *http.Request) ([]byte, error) {
	b, err := io.ReadAll(io.LimitReader(r.Body, maxBodySize+1))
	if err != nil {
		return nil, fmt.Errorf("webhook: read body: %w", err)
	}
	if len(b) > maxBodySize {
		return nil, errors.New("webhook: body too large")
	}
	return b, nil
}

// --- Unlink ---

// UnlinkEvent reports that a user unlinked from the app outside the
// service, e.g. by deleting their Kakao Account. It is not sent when the
// service itself calls the unlink API.
type UnlinkEvent struct {
	AppID  string
	UserID int64
	// ReferrerType is how the user unlinked: "ACCOUNT_DELETE",
	// "FORCED_ACCOUNT_DELETE", "UNLINK_FROM_ADMIN", "UNLINK_FROM_APPS" or
	// "INCOMPLETE_SIGN_UP".
	ReferrerType string
	// GroupUserToken is set for apps in an app group.
	GroupUserToken string
}

// ParseUnlink authenticates and decodes an unlink webhook (연결 해제 웹훅),
// sent as a GET query or a POST form.
//
// Docs: https://developers.kakao.com/docs/ko/kakaologin/callback
func ParseUnlink(r *http.Request, primaryAdminKey string) (*UnlinkEvent, error) {
	if err := authorize(r, primaryAdminKey); err != nil {
		return nil, err
	}
	r.Body = http.MaxBytesReader(nil, r.Body, maxBodySize)
	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("webhook: parse unlink: %w", err)
	}
	userID, err := strconv.ParseInt(r.Form.Get("user_id"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("webhook: parse unlink user_id: %w", err)
	}
	return &UnlinkEvent{
		AppID:          r.Form.Get("app_id"),
		UserID:         userID,
		ReferrerType:   r.Form.Get("referrer_type"),
		GroupUserToken: r.Form.Get("group_user_token"),
	}, nil
}

// UnlinkHandler returns a handler for the unlink webhook. It calls fn for
// each authenticated event and replies 200 OK, as Kakao requires even when
// the user is unknown. Requests without the primary admin key get 401.
func UnlinkHandler(primaryAdminKey string, fn func(r *http.Request, e *UnlinkEvent)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e, err := ParseUnlink(r, primaryAdminKey)
		if err != nil {
			http.Error(w, err.Error(), status(err))
			return
		}
		fn(r, e)
		w.WriteHeader(http.StatusOK)
	})
}

// --- KakaoTalk Channel ---

// ChannelEvent reports that a user added or blocked a KakaoTalk Channel
// linked to the app.
type ChannelEvent struct {
	// Event is "added" or "blocked".
	Event string `json:"event"`
	// ID identifies the user; see IDType.
	ID string `json:"id"`
	// IDType is "app_user_id" (a user ID, when the user consented to
	// plusfriends) or "open_id".
	IDType          string    `json:"id_type"`
	ChannelPublicID string    `json:"channel_public_id"`
	ChannelUUID     string    `json:"channel_uuid"`
	UpdatedAt       time.Time `json:"updated_at"`
	// ResourceID is the X-Kakao-Resource-ID header, unique per webhook.
	ResourceID string `json:"-"`
}

// ParseChannel authenticates and decodes a KakaoTalk Channel webhook.
//
// Docs: https://developers.kakao.com/docs/ko/kakaotalk-channel/callback
func ParseChannel(r *http.Request, primaryAdminKey string) (*ChannelEvent, error) {
	if err := authorize(r, primaryAdminKey); err != nil {
		return nil, err
	}
	b, err := readBody(r)
	if err != nil {
		return nil, err
	}
	var e ChannelEvent
	if err := json.Unmarshal(b, &e); err != nil {
		return nil, fmt.Errorf("webhook: parse channel event: %w", err)
	}
	e.ResourceID = r.Header.Get("X-Kakao-Resource-ID")
	return &e, nil
}

// ChannelHandler returns a handler for the KakaoTalk Channel webhook. It
// calls fn for each authenticated event and replies 200 OK.
func ChannelHandler(primaryAdminKey string, fn func(r *http.Request, e *ChannelEvent)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e, err := ParseChannel(r, primaryAdminKey)
		if err != nil {
			http.Error(w, err.Error(), status(err))
			return
		}
		fn(r, e)
		w.WriteHeader(http.StatusOK)
	})
}

// --- KakaoTalk Share ---

// ShareEvent reports that a KakaoTalk Share message sent with
// serverCallbackArgs was delivered.
type ShareEvent struct {
	// ChatType is "MemoChat", "DirectChat", "MultiChat", "OpenDirectChat"
	// or "OpenMultiChat".
	ChatType string
	// HashChatID identifies the chat room, consistently per service.
	HashChatID string
	// TemplateID is the custom template used, or 0.
	TemplateID int64
	// IsSingleChatroom and ChatMemberCountRange are only sent to apps with
	// paid KakaoTalk Share enabled.
	IsSingleChatroom     *bool
	ChatMemberCountRange string
	// Args holds the serverCallbackArgs set when sharing. Values are
	// strings for GET requests and JSON values for POST requests.
	Args map[string]any
	// ResourceID is the X-Kakao-Resource-ID header, unique per webhook.
	ResourceID string
}

var shareFields = map[string]bool{
	"CHAT_TYPE": true, "HASH_CHAT_ID": true, "TEMPLATE_ID": true,
	"IS_SINGLE_CHATROOM": true, "CHAT_MEMBER_COUNT_RANGE": true,
}

// ParseShare authenticates and decodes a KakaoTalk Share webhook, sent as a
// GET query or a POST JSON body.
//
// Docs: https://developers.kakao.com/docs/ko/kakaotalk-share/callback
func ParseShare(r *http.Request, primaryAdminKey string) (*ShareEvent, error) {
	if err := authorize(r, primaryAdminKey); err != nil {
		return nil, err
	}
	fields := map[string]any{}
	if r.Method == http.MethodGet {
		for k, v := range r.URL.Query() {
			fields[k] = v[0]
		}
	} else {
		b, err := readBody(r)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(b, &fields); err != nil {
			return nil, fmt.Errorf("webhook: parse share event: %w", err)
		}
	}

	e := &ShareEvent{Args: map[string]any{}, ResourceID: r.Header.Get("X-Kakao-Resource-ID")}
	e.ChatType, _ = fields["CHAT_TYPE"].(string)
	e.HashChatID, _ = fields["HASH_CHAT_ID"].(string)
	e.ChatMemberCountRange, _ = fields["CHAT_MEMBER_COUNT_RANGE"].(string)
	switch v := fields["TEMPLATE_ID"].(type) {
	case float64:
		e.TemplateID = int64(v)
	case string:
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("webhook: parse share TEMPLATE_ID: %w", err)
		}
		e.TemplateID = id
	}
	switch v := fields["IS_SINGLE_CHATROOM"].(type) {
	case bool:
		e.IsSingleChatroom = &v
	case string:
		b, err := strconv.ParseBool(v)
		if err != nil {
			return nil, fmt.Errorf("webhook: parse share IS_SINGLE_CHATROOM: %w", err)
		}
		e.IsSingleChatroom = &b
	}
	for k, v := range fields {
		if !shareFields[k] {
			e.Args[k] = v
		}
	}
	return e, nil
}

// ShareHandler returns a handler for the KakaoTalk Share webhook. It calls
// fn for each authenticated event and replies 200 OK.
func ShareHandler(primaryAdminKey string, fn func(r *http.Request, e *ShareEvent)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e, err := ParseShare(r, primaryAdminKey)
		if err != nil {
			http.Error(w, err.Error(), status(err))
			return
		}
		fn(r, e)
		w.WriteHeader(http.StatusOK)
	})
}

// --- Push delivery failures ---

// PushFailure reports a push notification that APNs or FCM rejected. Kakao
// posts it to the return_url of the push message.
type PushFailure struct {
	// UUID is the user the push was sent to.
	UUID      string
	PushToken string
	Date      time.Time
	// NewPushToken is a replacement token FCM may provide with bypass
	// sends.
	NewPushToken string
}

// ParsePushFailure decodes a push failure feedback request. Kakao does not
// authenticate these requests, so put a hard-to-guess secret in the
// return_url and check it before trusting the feedback.
//
// Docs: https://developers.kakao.com/docs/ko/push/rest-api
func ParsePushFailure(r *http.Request) (*PushFailure, error) {
	r.Body = http.MaxBytesReader(nil, r.Body, maxBodySize)
	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("webhook: parse push failure: %w", err)
	}
	f := &PushFailure{
		UUID:         r.Form.Get("userId"),
		PushToken:    r.Form.Get("pushToken"),
		NewPushToken: r.Form.Get("newPushToken"),
	}
	if d := r.Form.Get("date"); d != "" {
		ms, err := strconv.ParseInt(d, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("webhook: parse push failure date: %w", err)
		}
		f.Date = time.UnixMilli(ms)
	}
	return f, nil
}

// PushFailureHandler returns a handler for push failure feedback. It calls
// fn for each request and replies 200 OK. See [ParsePushFailure] about
// authentication.
func PushFailureHandler(fn func(r *http.Request, f *PushFailure)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := ParsePushFailure(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fn(r, f)
		w.WriteHeader(http.StatusOK)
	})
}
