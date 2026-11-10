package moment

import (
	"context"
	"errors"
	"net/http"
)

// List returns the ad account's ad views. Use an ad view's ID as the
// landing of a TALK_BIZ_BOARD creative.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/ad-view
func (s *AdViewService) List(ctx context.Context, adAccountID int64) ([]AdView, error) {
	if adAccountID == 0 {
		return nil, errors.New("moment: ad account ID is required")
	}
	var vs []AdView
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v4/adViews", adAccountID: adAccountID}, &vs); err != nil {
		return nil, err
	}
	return vs, nil
}

// Get returns an ad view.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/ad-view
func (s *AdViewService) Get(ctx context.Context, adAccountID, adViewID int64) (*AdView, error) {
	if adAccountID == 0 || adViewID == 0 {
		return nil, errors.New("moment: ad account ID and ad view ID are required")
	}
	var v AdView
	if err := s.c.do(ctx, call{method: http.MethodGet, path: idPath("/v4/adViews/%d", adViewID), adAccountID: adAccountID}, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// Delete deletes an ad view. Kakao refuses while a creative lands on it.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/ad-view
func (s *AdViewService) Delete(ctx context.Context, adAccountID, adViewID int64) error {
	if adAccountID == 0 || adViewID == 0 {
		return errors.New("moment: ad account ID and ad view ID are required")
	}
	return s.c.do(ctx, call{method: http.MethodDelete, path: idPath("/v4/adViews/%d", adViewID), adAccountID: adAccountID}, nil)
}
