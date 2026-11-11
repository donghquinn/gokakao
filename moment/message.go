package moment

import (
	"context"
	"errors"
	"net/http"
	"net/url"
)

// Coupons returns the coupons of a KakaoTalk Channel that message creative
// buttons can land on (CHANNEL_COUPON).
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/message
func (s *MessageService) Coupons(ctx context.Context, adAccountID int64, profileID string) ([]ChannelCoupon, error) {
	if profileID == "" {
		return nil, errors.New("moment: profile ID is required")
	}
	var res []ChannelCoupon
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v4/messages/coupons", adAccountID: adAccountID, query: url.Values{"profileId": {profileID}}}, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// Coupon returns a coupon of a KakaoTalk Channel.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/message
func (s *MessageService) Coupon(ctx context.Context, adAccountID, couponID int64, profileID string) (*ChannelCoupon, error) {
	if profileID == "" {
		return nil, errors.New("moment: profile ID is required")
	}
	var res ChannelCoupon
	if err := s.c.do(ctx, call{method: http.MethodGet, path: idPath("/v4/messages/coupons/%d", couponID), adAccountID: adAccountID, query: url.Values{"profileId": {profileID}}}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// AdViews returns the ad views that message creative buttons can land on
// (AD_VIEW).
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/message
func (s *MessageService) AdViews(ctx context.Context, adAccountID int64) ([]AdView, error) {
	var res []AdView
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v4/messages/adViews", adAccountID: adAccountID}, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// Posts returns the posts of a KakaoTalk Channel that message creative
// buttons can land on (CHANNEL_POST).
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/message
func (s *MessageService) Posts(ctx context.Context, adAccountID int64, profileID string) ([]ChannelPost, error) {
	if profileID == "" {
		return nil, errors.New("moment: profile ID is required")
	}
	var res []ChannelPost
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v4/messages/posts/" + url.PathEscape(profileID), adAccountID: adAccountID}, &res); err != nil {
		return nil, err
	}
	return res, nil
}
