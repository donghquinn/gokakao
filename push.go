package kakao

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// PushService calls the Push Notification v2 API. It uses the admin key.
// APNs and FCM credentials must be registered in the app settings.
//
// Docs: https://developers.kakao.com/docs/ko/push/rest-api
type PushService struct {
	c *Client
}

// PushType is a push provider.
type PushType string

// Push types.
const (
	PushAPNs PushType = "apns"
	PushFCM  PushType = "fcm"
)

// Device is a device to register for push notifications.
type Device struct {
	// UUID is the service's user ID, 1 to 2^63-1.
	UUID int64
	// DeviceID distinguishes the user's devices (up to 64 chars).
	DeviceID  string
	PushType  PushType
	PushToken string
}

// Register registers a device's push token. It returns the number of days
// until the registration expires, or -1 for never; registrations are not
// renewed automatically.
//
// Docs: https://developers.kakao.com/docs/ko/push/rest-api
func (s *PushService) Register(ctx context.Context, d Device) (int, error) {
	if d.UUID <= 0 || d.DeviceID == "" || d.PushType == "" || d.PushToken == "" {
		return 0, errors.New("kakao: UUID, device ID, push type and push token are required")
	}
	form := url.Values{
		"uuid":       {strconv.FormatInt(d.UUID, 10)},
		"device_id":  {d.DeviceID},
		"push_type":  {string(d.PushType)},
		"push_token": {d.PushToken},
	}
	var days int
	err := s.c.do(ctx, &request{method: http.MethodPost, host: HostAPI, path: "/v2/push/register", auth: authAdminKey, form: form}, &days)
	return days, err
}

// PushToken is a registered push token.
type PushToken struct {
	UUID      string   `json:"uuid"`
	DeviceID  string   `json:"device_id"`
	PushType  PushType `json:"push_type"`
	PushToken string   `json:"push_token"`
	CreatedAt Time     `json:"created_at"`
	UpdatedAt Time     `json:"updated_at"`
}

// Tokens returns the push tokens registered for up to 100 users.
//
// Docs: https://developers.kakao.com/docs/ko/push/rest-api
func (s *PushService) Tokens(ctx context.Context, uuids ...int64) ([]PushToken, error) {
	if len(uuids) == 0 || len(uuids) > 100 {
		return nil, fmt.Errorf("kakao: need 1 to 100 UUIDs, got %d", len(uuids))
	}
	q := url.Values{}
	if len(uuids) == 1 {
		q.Set("uuid", strconv.FormatInt(uuids[0], 10))
	} else if err := setJSON(q, "uuids", formatIDs(uuids)); err != nil {
		return nil, err
	}
	var tokens []PushToken
	err := s.c.do(ctx, &request{method: http.MethodGet, host: HostAPI, path: "/v2/push/tokens", auth: authAdminKey, query: q}, &tokens)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

// Deregister removes a user's push tokens: those of one device when
// deviceID and pushType are given, otherwise all of them.
//
// Docs: https://developers.kakao.com/docs/ko/push/rest-api
func (s *PushService) Deregister(ctx context.Context, uuid int64, deviceID string, pushType PushType) error {
	if uuid <= 0 {
		return errors.New("kakao: invalid UUID")
	}
	form := url.Values{"uuid": {strconv.FormatInt(uuid, 10)}}
	if deviceID != "" && pushType != "" {
		form.Set("device_id", deviceID)
		form.Set("push_type", string(pushType))
	}
	return s.c.do(ctx, &request{method: http.MethodPost, host: HostAPI, path: "/v2/push/deregister", auth: authAdminKey, form: form}, nil)
}

// PushMessage is a push notification for APNs and/or FCM. Leave a field
// nil to skip that provider.
type PushMessage struct {
	APNs *APNsMessage `json:"for_apns,omitempty"`
	FCM  *FCMMessage  `json:"for_fcm,omitempty"`
}

// APNsMessage is the APNs part of a [PushMessage]. Its payload must not
// exceed 4 KB (5 KB for VoIP).
type APNsMessage struct {
	// Topic is "voip" for VoIP notifications.
	Topic            string `json:"topic,omitempty"`
	MutableContent   int    `json:"mutable-content,omitempty"`
	Expiration       int    `json:"expiration,omitempty"` // seconds
	Collapse         string `json:"collapse,omitempty"`
	Badge            *int   `json:"badge,omitempty"`
	Sound            string `json:"sound,omitempty"`
	PushAlert        *bool  `json:"push_alert,omitempty"`
	ContentAvailable int    `json:"content-available,omitempty"`
	Category         string `json:"category,omitempty"`
	// Message is the alert: a string, or an object such as
	// {"loc-key": "...", "loc-args": [...]}.
	Message     any            `json:"message,omitempty"`
	CustomField map[string]any `json:"custom_field,omitempty"`
	// ReturnURL receives delivery failures; see webhook.ParsePushFailure.
	ReturnURL string `json:"return_url,omitempty"`
	// PushToken is the token to use with bypass.
	PushToken string `json:"push_token,omitempty"`
	// APNsEnv is "sandbox" or "production" (default).
	APNsEnv           string  `json:"apns_env,omitempty"`
	ThreadID          string  `json:"thread-id,omitempty"`
	APNsPushType      string  `json:"apns-push-type,omitempty"`
	APNsPriority      int     `json:"apns-priority,omitempty"`
	TargetContentID   string  `json:"target-content-id,omitempty"`
	InterruptionLevel string  `json:"interruption-level,omitempty"`
	RelevanceScore    float64 `json:"relevance-score,omitempty"`
}

// FCMMessage is the FCM part of a [PushMessage].
type FCMMessage struct {
	Collapse   string `json:"collapse,omitempty"`
	TimeToLive int    `json:"time_to_live,omitempty"` // seconds
	DryRun     bool   `json:"dry_run,omitempty"`
	// Priority is "high" or "normal" (default).
	Priority string `json:"priority,omitempty"`
	// CustomField is limited to 4 KB.
	CustomField map[string]any `json:"custom_field,omitempty"`
	// Notification holds FCM notification keys such as title and body.
	Notification map[string]any `json:"notification,omitempty"`
	ReturnURL    string         `json:"return_url,omitempty"`
	PushToken    string         `json:"push_token,omitempty"`
}

// SendOptions are the optional parameters of [PushService.Send].
type SendOptions struct {
	// Bypass sends to the PushToken in the message instead of registered
	// tokens. Exactly one UUID must be given.
	Bypass bool
}

// Send sends msg to every registered device of up to 100 users. Success
// means Kakao accepted the message, not that it was delivered; set a
// ReturnURL to receive failures.
//
// Docs: https://developers.kakao.com/docs/ko/push/rest-api
func (s *PushService) Send(ctx context.Context, uuids []int64, msg PushMessage, opts *SendOptions) error {
	if len(uuids) == 0 || len(uuids) > 100 {
		return fmt.Errorf("kakao: need 1 to 100 UUIDs, got %d", len(uuids))
	}
	bypass := opts != nil && opts.Bypass
	if bypass && len(uuids) != 1 {
		return errors.New("kakao: bypass sends to exactly one UUID")
	}
	b, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("kakao: encode push message: %w", err)
	}
	form := url.Values{"push_message": {string(b)}}
	if err := setJSON(form, "uuids", formatIDs(uuids)); err != nil {
		return err
	}
	if bypass {
		form.Set("bypass", "true")
	}
	return s.c.do(ctx, &request{method: http.MethodPost, host: HostAPI, path: "/v2/push/send", auth: authAdminKey, form: form}, nil)
}

func formatIDs(ids []int64) []string {
	s := make([]string, len(ids))
	for i, id := range ids {
		s[i] = strconv.FormatInt(id, 10)
	}
	return s
}
