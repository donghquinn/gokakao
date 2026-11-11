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
	// UserID is the service's own user ID (Kakao calls it uuid), 1 to
	// 2^63-1.
	UserID int64
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
	if d.UserID <= 0 || d.DeviceID == "" || d.PushType == "" || d.PushToken == "" {
		return 0, errors.New("kakao: user ID, device ID, push type and push token are required")
	}
	form := url.Values{
		"uuid":       {strconv.FormatInt(d.UserID, 10)},
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
	// UserID is the service's user ID the token is registered for.
	UserID    int64    `json:"uuid,string"`
	DeviceID  string   `json:"device_id"`
	PushType  PushType `json:"push_type"`
	PushToken string   `json:"push_token"`
	CreatedAt Time     `json:"created_at"`
	UpdatedAt Time     `json:"updated_at"`
}

// Tokens returns the push tokens registered for up to 100 users.
//
// Docs: https://developers.kakao.com/docs/ko/push/rest-api
func (s *PushService) Tokens(ctx context.Context, userIDs []int64) ([]PushToken, error) {
	if len(userIDs) == 0 || len(userIDs) > 100 {
		return nil, fmt.Errorf("kakao: need 1 to 100 user IDs, got %d", len(userIDs))
	}
	q := url.Values{}
	if len(userIDs) == 1 {
		q.Set("uuid", strconv.FormatInt(userIDs[0], 10))
	} else if err := setJSON(q, "uuids", formatIDs(userIDs)); err != nil {
		return nil, err
	}
	var tokens []PushToken
	err := s.c.do(ctx, &request{method: http.MethodGet, host: HostAPI, path: "/v2/push/tokens", auth: authAdminKey, query: q}, &tokens)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

// DeregisterOptions limits [PushService.Deregister] to one device. Both
// fields are needed; otherwise all of the user's tokens are removed.
type DeregisterOptions struct {
	DeviceID string
	PushType PushType
}

// Deregister removes a user's push tokens: all of them, or only one
// device's when opts names it.
//
// Docs: https://developers.kakao.com/docs/ko/push/rest-api
func (s *PushService) Deregister(ctx context.Context, userID int64, opts *DeregisterOptions) error {
	if userID <= 0 {
		return errors.New("kakao: invalid user ID")
	}
	form := url.Values{"uuid": {strconv.FormatInt(userID, 10)}}
	if opts != nil && opts.DeviceID != "" && opts.PushType != "" {
		form.Set("device_id", opts.DeviceID)
		form.Set("push_type", string(opts.PushType))
	}
	return s.c.do(ctx, &request{method: http.MethodPost, host: HostAPI, path: "/v2/push/deregister", auth: authAdminKey, form: form}, nil)
}

// APNsEnv selects the APNs server.
type APNsEnv string

// APNs environments.
const (
	APNsProduction APNsEnv = "production" // default
	APNsSandbox    APNsEnv = "sandbox"
)

// FCMPriority is the delivery priority of an FCM message.
type FCMPriority string

// FCM priorities.
const (
	FCMNormal FCMPriority = "normal" // default
	FCMHigh   FCMPriority = "high"   // delivered in Doze mode
)

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
	PushToken         string  `json:"push_token,omitempty"`
	APNsEnv           APNsEnv `json:"apns_env,omitempty"`
	ThreadID          string  `json:"thread-id,omitempty"`
	APNsPushType      string  `json:"apns-push-type,omitempty"`
	APNsPriority      int     `json:"apns-priority,omitempty"`
	TargetContentID   string  `json:"target-content-id,omitempty"`
	InterruptionLevel string  `json:"interruption-level,omitempty"`
	RelevanceScore    float64 `json:"relevance-score,omitempty"`
}

// FCMMessage is the FCM part of a [PushMessage].
type FCMMessage struct {
	Collapse   string      `json:"collapse,omitempty"`
	TimeToLive int         `json:"time_to_live,omitempty"` // seconds
	DryRun     bool        `json:"dry_run,omitempty"`
	Priority   FCMPriority `json:"priority,omitempty"`
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
	// tokens. Exactly one user ID must be given.
	Bypass bool
}

// Send sends msg to every registered device of up to 100 users. Success
// means Kakao accepted the message, not that it was delivered; set a
// ReturnURL to receive failures.
//
// Docs: https://developers.kakao.com/docs/ko/push/rest-api
func (s *PushService) Send(ctx context.Context, userIDs []int64, msg PushMessage, opts *SendOptions) error {
	if len(userIDs) == 0 || len(userIDs) > 100 {
		return fmt.Errorf("kakao: need 1 to 100 user IDs, got %d", len(userIDs))
	}
	bypass := opts != nil && opts.Bypass
	if bypass && len(userIDs) != 1 {
		return errors.New("kakao: bypass sends to exactly one user")
	}
	b, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("kakao: encode push message: %w", err)
	}
	form := url.Values{"push_message": {string(b)}}
	if err := setJSON(form, "uuids", formatIDs(userIDs)); err != nil {
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
