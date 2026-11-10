package moment

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	kakao "github.com/donghquinn/gokakao"
)

// TrackerRole is the ad account's permission on a Pixel & SDK.
type TrackerRole string

// Tracker roles.
const (
	TrackerMaster  TrackerRole = "MASTER"  // 마스터
	TrackerMember  TrackerRole = "MEMBER"  // 멤버
	TrackerRequest TrackerRole = "REQUEST" // 멤버 권한 요청 중; available trackers only
)

// TrackerEventStatus is whether a Pixel & SDK is collecting events.
type TrackerEventStatus string

// Tracker event statuses.
const (
	TrackerPreCollection TrackerEventStatus = "PRE_COLLECTION" // 수집전
	TrackerCollection    TrackerEventStatus = "COLLECTION"     // 수집중
)

// TrackRight is a Pixel & SDK linked to an ad account.
type TrackRight struct {
	// TrackID is the Pixel & SDK ID.
	TrackID   string `json:"trackId"`
	TrackName string `json:"trackName"`
	// Population is the estimated number of users reachable with the
	// collected events (타겟 모수), refreshed daily.
	Population  kakao.Int64        `json:"population"`
	EventStatus TrackerEventStatus `json:"eventStatus"`
	CreatedDate kakao.Time         `json:"createdDate"`
	// LastEventDate is when an event was last received; zero if none.
	LastEventDate kakao.Time  `json:"lastEventDate"`
	Role          TrackerRole `json:"role"`
}

// List returns the Pixel & SDKs linked to an ad account. Kakao allows one
// call per 5 seconds per user.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/pixel-and-sdk
func (s *PixelService) List(ctx context.Context, adAccountID int64) ([]TrackRight, error) {
	if adAccountID == 0 {
		return nil, errors.New("moment: ad account ID is required")
	}
	var res struct {
		TrackRights []TrackRight `json:"trackRights"`
	}
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v4/trackers/rights", adAccountID: adAccountID}, &res); err != nil {
		return nil, err
	}
	return res.TrackRights, nil
}

// AvailableTracker is a Pixel & SDK that can be linked to an ad account.
type AvailableTracker struct {
	// TrackID is the Pixel & SDK ID to pass to [PixelService.Link].
	TrackID   string      `json:"trackId"`
	Name      string      `json:"name"`
	CreatedAt kakao.Time  `json:"createdAt"`
	UpdatedAt kakao.Time  `json:"updatedAt"`
	Role      TrackerRole `json:"role"`
}

// Available returns the Pixel & SDKs that can be linked to an ad account.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/pixel-and-sdk
func (s *PixelService) Available(ctx context.Context, adAccountID int64) ([]AvailableTracker, error) {
	if adAccountID == 0 {
		return nil, errors.New("moment: ad account ID is required")
	}
	var ts []AvailableTracker
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v4/trackers/rightAvailables", adAccountID: adAccountID}, &ts); err != nil {
		return nil, err
	}
	return ts, nil
}

// Link links a Pixel & SDK from [PixelService.Available] to an ad
// account. An ad account can link up to 100.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/pixel-and-sdk
func (s *PixelService) Link(ctx context.Context, adAccountID int64, trackID string) error {
	if adAccountID == 0 || trackID == "" {
		return errors.New("moment: ad account ID and track ID are required")
	}
	body := struct {
		TrackID     string `json:"trackId"`
		AdAccountID int64  `json:"adAccountId"`
	}{trackID, adAccountID}
	return s.c.do(ctx, call{method: http.MethodPost, path: "/v4/trackers", adAccountID: adAccountID, body: body}, nil)
}

// Unlink unlinks a Pixel & SDK from an ad account.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/pixel-and-sdk
func (s *PixelService) Unlink(ctx context.Context, adAccountID int64, trackID string) error {
	if adAccountID == 0 || trackID == "" {
		return errors.New("moment: ad account ID and track ID are required")
	}
	return s.c.do(ctx, call{method: http.MethodDelete, path: "/v4/trackers/" + url.PathEscape(trackID), adAccountID: adAccountID}, nil)
}
