package moment

import (
	"context"
	"errors"
	"net/http"
)

// BizFormLanding is a business form usable as a creative landing.
//
// The type reference names the period fields beginDate and endDate
// ([BusinessForm]) but this API's example sends beginDateTime and
// endDateTime; both are decoded.
type BizFormLanding struct {
	BusinessForm
	// BeginDateTime and EndDateTime are "yyyy-MM-dd HH:mm".
	BeginDateTime string `json:"beginDateTime"`
	EndDateTime   string `json:"endDateTime"`
}

// BizForms returns the business forms usable as the BIZ_FORM landing of a
// creative. Kakao ends BIZ_FORM landings on 2026-12-31.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/bizboard-landing
func (s *BizboardLandingService) BizForms(ctx context.Context, adAccountID int64) ([]BizFormLanding, error) {
	if adAccountID == 0 {
		return nil, errors.New("moment: ad account ID is required")
	}
	var fs []BizFormLanding
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v4/creatives/landing/talkBizForms", adAccountID: adAccountID}, &fs); err != nil {
		return nil, err
	}
	return fs, nil
}

// AdViews returns the ad views usable as the AD_VIEW landing of a
// creative.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/bizboard-landing
func (s *BizboardLandingService) AdViews(ctx context.Context, adAccountID int64) ([]AdView, error) {
	if adAccountID == 0 {
		return nil, errors.New("moment: ad account ID is required")
	}
	var vs []AdView
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v4/creatives/landing/adViews", adAccountID: adAccountID}, &vs); err != nil {
		return nil, err
	}
	return vs, nil
}

// ChannelPosts returns the posts of a KakaoTalk Channel profile usable as
// the CHANNEL_POST landing of a creative.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/bizboard-landing
func (s *BizboardLandingService) ChannelPosts(ctx context.Context, adAccountID, profileID int64) ([]ChannelPost, error) {
	if adAccountID == 0 || profileID == 0 {
		return nil, errors.New("moment: ad account ID and profile ID are required")
	}
	var res struct {
		Items []ChannelPost `json:"items"`
	}
	if err := s.c.do(ctx, call{method: http.MethodGet, path: idPath("/v4/creatives/landing/channelPosts/%d", profileID), adAccountID: adAccountID}, &res); err != nil {
		return nil, err
	}
	return res.Items, nil
}
