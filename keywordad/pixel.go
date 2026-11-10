package keywordad

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// DateTime is a timestamp in a Keyword Ad response. Unlike [kakao.Time] it
// also accepts the API's offsets without a colon
// ("2022-02-20T00:00:00.000+0000") and timestamps without an offset
// ("2021-02-20T12:00:00.702463"), which are read as Korea Standard Time.
// An empty string or null decodes to the zero time.
type DateTime struct {
	time.Time
}

// kst is Korea Standard Time, which has no daylight saving time.
var kst = time.FixedZone("KST", 9*60*60)

// dateTimeLayouts are tried in order; the last has no offset.
var dateTimeLayouts = []string{
	time.RFC3339Nano,
	"2006-01-02T15:04:05.999999999Z0700",
	"2006-01-02T15:04:05.999999999",
}

// UnmarshalJSON implements [json.Unmarshaler].
func (t *DateTime) UnmarshalJSON(b []byte) error {
	t.Time = time.Time{}
	if bytes.Equal(b, []byte("null")) {
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("keywordad: time: %w", err)
	}
	if s == "" {
		return nil
	}
	for i, layout := range dateTimeLayouts {
		loc := time.UTC
		if i == len(dateTimeLayouts)-1 {
			loc = kst
		}
		if v, err := time.ParseInLocation(layout, s, loc); err == nil {
			t.Time = v
			return nil
		}
	}
	return fmt.Errorf("keywordad: time: cannot parse %q", s)
}

// TrackerRole is the ad account's permission on a Pixel & SDK.
type TrackerRole string

// Tracker roles.
const (
	TrackerMaster  TrackerRole = "MASTER"
	TrackerMember  TrackerRole = "MEMBER"
	TrackerRequest TrackerRole = "REQUEST" // member permission requested
)

// TrackRight is a Pixel & SDK linked to an ad account.
type TrackRight struct {
	// TrackID is the Pixel & SDK ID.
	TrackID   string `json:"trackId"`
	TrackName string `json:"trackName"`
	// Population is the estimated number of users reachable with the
	// events collected over the last 120 days (타겟 모수).
	Population  int64       `json:"population"`
	EventStatus EventStatus `json:"eventStatus"`
	CreatedDate DateTime    `json:"createdDate"`
	// LastEventDate is when an event was last received; zero if none.
	LastEventDate DateTime    `json:"lastEventDate"`
	Role          TrackerRole `json:"role"`
}

// List returns the Pixel & SDKs linked to an ad account. Kakao allows one
// call per 5 seconds.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/pixel-and-sdk
func (s *PixelService) List(ctx context.Context, adAccountID int64) ([]TrackRight, error) {
	var res struct {
		TrackRights []TrackRight `json:"trackRights"`
	}
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v2/trackers/rights", adAccountID: adAccountID}, &res); err != nil {
		return nil, err
	}
	return res.TrackRights, nil
}

// AvailableTracker is a Pixel & SDK that can be linked to an ad account.
type AvailableTracker struct {
	// TrackID is the Pixel & SDK ID to pass to [PixelService.Link].
	TrackID   string      `json:"trackId"`
	Name      string      `json:"name"`
	CreatedAt DateTime    `json:"createdAt"`
	UpdatedAt DateTime    `json:"updatedAt"`
	Role      TrackerRole `json:"role"`
}

// UnmarshalJSON implements [json.Unmarshaler]. It reads the name from
// "name", as Kakao's example sends it, or "trackName", as Kakao's field
// table documents it.
func (a *AvailableTracker) UnmarshalJSON(b []byte) error {
	type plain AvailableTracker
	var v struct {
		plain
		TrackName string `json:"trackName"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*a = AvailableTracker(v.plain)
	if a.Name == "" {
		a.Name = v.TrackName
	}
	return nil
}

// Available returns the Pixel & SDKs the user can link to an ad account.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/pixel-and-sdk
func (s *PixelService) Available(ctx context.Context, adAccountID int64) ([]AvailableTracker, error) {
	var raw json.RawMessage
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v2/trackers/rightAvailables", adAccountID: adAccountID}, &raw); err != nil {
		return nil, err
	}
	// Kakao's example is a bare array; its field table wraps the list in
	// rightAvailables. Accept both.
	var ts []AvailableTracker
	if raw = bytes.TrimSpace(raw); len(raw) > 0 && raw[0] == '{' {
		var res struct {
			RightAvailables []AvailableTracker `json:"rightAvailables"`
		}
		if err := json.Unmarshal(raw, &res); err != nil {
			return nil, fmt.Errorf("keywordad: decode available trackers: %w", err)
		}
		return res.RightAvailables, nil
	}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &ts); err != nil {
			return nil, fmt.Errorf("keywordad: decode available trackers: %w", err)
		}
	}
	return ts, nil
}

// Link links a Pixel & SDK from [PixelService.Available] to an ad account.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/pixel-and-sdk
func (s *PixelService) Link(ctx context.Context, adAccountID int64, trackID string) error {
	if trackID == "" {
		return errors.New("keywordad: track ID is required")
	}
	body := struct {
		TrackID string `json:"trackId"`
	}{trackID}
	return s.c.do(ctx, call{method: http.MethodPost, path: "/v2/trackers", adAccountID: adAccountID, body: body}, nil)
}

// Unlink unlinks a Pixel & SDK from an ad account.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/pixel-and-sdk
func (s *PixelService) Unlink(ctx context.Context, adAccountID int64, trackID string) error {
	if trackID == "" {
		return errors.New("keywordad: track ID is required")
	}
	return s.c.do(ctx, call{method: http.MethodDelete, path: "/v2/trackers/" + url.PathEscape(trackID), adAccountID: adAccountID}, nil)
}
