package keywordad

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	kakao "github.com/donghquinn/gokakao"
)

// CreativeLink links a creative to an ad group (소재연결).
type CreativeLink struct {
	CreativeLinkID kakao.Int64 `json:"creativeLinkId"`
	AdGroupID      kakao.Int64 `json:"adGroupId"`
	CreativeID     kakao.Int64 `json:"creativeId"`
	Config         Config      `json:"config"`
	// Status lists why the linked creative can or cannot run.
	Status []Status `json:"status"`
}

// CreativeLinkListOptions are the optional parameters of
// [CreativeLinkService.ListByAdGroup] and
// [CreativeLinkService.ListByCreative].
type CreativeLinkListOptions struct {
	// Configs filters by status; Kakao lists ON and OFF links when empty.
	Configs []Config
}

func (s *CreativeLinkService) list(ctx context.Context, adAccountID int64, key string, id int64, opts *CreativeLinkListOptions) ([]CreativeLink, error) {
	q := url.Values{key: {strconv.FormatInt(id, 10)}}
	if opts != nil {
		for _, cfg := range opts.Configs {
			q.Add("config", string(cfg))
		}
	}
	var ls []CreativeLink
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v1/creativeLinks", adAccountID: adAccountID, query: q}, &ls); err != nil {
		return nil, err
	}
	return ls, nil
}

// ListByAdGroup returns the creatives linked to an ad group.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/creative-link
func (s *CreativeLinkService) ListByAdGroup(ctx context.Context, adAccountID, adGroupID int64, opts *CreativeLinkListOptions) ([]CreativeLink, error) {
	if adGroupID == 0 {
		return nil, errors.New("keywordad: ad group ID is required")
	}
	return s.list(ctx, adAccountID, "adGroupId", adGroupID, opts)
}

// ListByCreative returns the ad groups a creative is linked to.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/creative-link
func (s *CreativeLinkService) ListByCreative(ctx context.Context, adAccountID, creativeID int64, opts *CreativeLinkListOptions) ([]CreativeLink, error) {
	if creativeID == 0 {
		return nil, errors.New("keywordad: creative ID is required")
	}
	return s.list(ctx, adAccountID, "creativeId", creativeID, opts)
}

// Get returns a creative link.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/creative-link
func (s *CreativeLinkService) Get(ctx context.Context, adAccountID, creativeLinkID int64) (*CreativeLink, error) {
	var l CreativeLink
	if err := s.c.do(ctx, call{method: http.MethodGet, path: idPath("/v1/creativeLinks/%d", creativeLinkID), adAccountID: adAccountID}, &l); err != nil {
		return nil, err
	}
	return &l, nil
}

// Link links a creative to an ad group. Kakao allows one call per 5
// seconds.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/creative-link
func (s *CreativeLinkService) Link(ctx context.Context, adAccountID, creativeID, adGroupID int64) (*CreativeLink, error) {
	if creativeID == 0 || adGroupID == 0 {
		return nil, errors.New("keywordad: creative ID and ad group ID are required")
	}
	body := struct {
		CreativeID int64 `json:"creativeId"`
		AdGroupID  int64 `json:"adGroupId"`
	}{creativeID, adGroupID}
	var l CreativeLink
	if err := s.c.do(ctx, call{method: http.MethodPost, path: "/v1/creativeLinks", adAccountID: adAccountID, body: body}, &l); err != nil {
		return nil, err
	}
	return &l, nil
}

// SetConfig turns a creative link ON or OFF. Only links that are ON or OFF
// can be changed. Kakao allows one call per second.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/creative-link
func (s *CreativeLinkService) SetConfig(ctx context.Context, adAccountID, creativeLinkID int64, cfg Config) error {
	body := struct {
		Config Config `json:"config"`
	}{cfg}
	return s.c.do(ctx, call{method: http.MethodPatch, path: idPath("/v1/creativeLinks/%d/onOff", creativeLinkID), adAccountID: adAccountID, body: body}, nil)
}

// Unlink deletes a creative link.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/creative-link
func (s *CreativeLinkService) Unlink(ctx context.Context, adAccountID, creativeLinkID int64) error {
	return s.c.do(ctx, call{method: http.MethodDelete, path: idPath("/v1/creativeLinks/%d", creativeLinkID), adAccountID: adAccountID}, nil)
}
