package keywordad

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	kakao "github.com/donghquinn/gokakao"
)

// AdAccount is an ad account in a list.
type AdAccount struct {
	ID         kakao.Int64 `json:"id"`
	Name       string      `json:"name"`
	MemberType MemberType  `json:"memberType"`
	Config     Config      `json:"config"`
}

// AdAccountListOptions are the optional parameters of
// [AdAccountService.List].
type AdAccountListOptions struct {
	PageOptions
	// Configs filters by status; Kakao lists ON and OFF accounts when
	// empty.
	Configs []Config
}

// List returns a page of the user's ad accounts.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/ad-account
func (s *AdAccountService) List(ctx context.Context, opts *AdAccountListOptions) (*Page[AdAccount], error) {
	q := url.Values{}
	if opts != nil {
		for _, cfg := range opts.Configs {
			q.Add("config", string(cfg))
		}
		opts.query(q)
	}
	var page Page[AdAccount]
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v1/adAccounts/pages", query: q}, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// AdAccountType is the kind of an ad account.
type AdAccountType string

// Ad account types.
const (
	AdAccountBusiness   AdAccountType = "BUSINESS"
	AdAccountIndividual AdAccountType = "INDIVIDUAL"
)

// AdAccountDetail is an ad account.
type AdAccountDetail struct {
	ID     kakao.Int64   `json:"id"`
	Name   string        `json:"name"`
	Type   AdAccountType `json:"type"`
	Config Config        `json:"config"`
	// AdminStop reports a stop by Kakao administrators.
	AdminStop    bool `json:"adminStop"`
	OutOfBalance bool `json:"outOfBalance"`
	// StatusDescription combines Config, AdminStop and OutOfBalance into
	// the delivery status, e.g. "운영 가능".
	StatusDescription string      `json:"statusDescription"`
	BizWalletID       kakao.Int64 `json:"bizWalletId"`
	// Advertiser is set for business ad accounts.
	Advertiser *Company `json:"advertiser"`
}

// Get returns an ad account.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/ad-account
func (s *AdAccountService) Get(ctx context.Context, adAccountID int64) (*AdAccountDetail, error) {
	var a AdAccountDetail
	err := s.c.do(ctx, call{method: http.MethodGet, path: idPath("/v1/adAccounts/%d", adAccountID), adAccountID: adAccountID}, &a)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// Balance is an ad account's balance in KRW.
type Balance struct {
	ID          kakao.Int64 `json:"id"`
	BizWalletID kakao.Int64 `json:"bizWalletId"`
	// Cash is paid cash and FreeCash is promotional credit.
	Cash     float64 `json:"cash"`
	FreeCash float64 `json:"freeCash"`
}

// Balance returns an ad account's current balance.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/ad-account
func (s *AdAccountService) Balance(ctx context.Context, adAccountID int64) (*Balance, error) {
	var b Balance
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v1/adAccounts/balance", adAccountID: adAccountID}, &b); err != nil {
		return nil, err
	}
	return &b, nil
}

// BizRight is the company holding an ad account's business rights (영업권).
type BizRight struct {
	ID                  kakao.Int64 `json:"id"`
	BizRightCompanyName string      `json:"bizRightCompanyName"`
}

// BizRight returns the holder of an ad account's business rights.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/ad-account
func (s *AdAccountService) BizRight(ctx context.Context, adAccountID int64) (*BizRight, error) {
	var b BizRight
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v1/adAccounts/bizRight", adAccountID: adAccountID}, &b); err != nil {
		return nil, err
	}
	return &b, nil
}

// ChannelProfile is one of the user's KakaoTalk Channel profiles.
type ChannelProfile struct {
	// UUID is the channel's search ID, e.g. "@kakao". Use it to create a
	// biz channel.
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

// ChannelProfiles returns the user's KakaoTalk Channel profiles.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/ad-account
func (s *AdAccountService) ChannelProfiles(ctx context.Context) ([]ChannelProfile, error) {
	var ps []ChannelProfile
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v1/adAccounts/channel/profiles"}, &ps); err != nil {
		return nil, err
	}
	return ps, nil
}

// NewAdAccount describes a business ad account to create.
type NewAdAccount struct {
	Name string `json:"name"`
	// AdvertiserCompanyID is the advertiser's ID from
	// [BusinessInfoService].
	AdvertiserCompanyID int64 `json:"advertiserCompanyId"`
}

// Create creates a business ad account.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/ad-account
func (s *AdAccountService) Create(ctx context.Context, a NewAdAccount) (*AdAccountDetail, error) {
	if a.Name == "" || a.AdvertiserCompanyID == 0 {
		return nil, errors.New("keywordad: name and advertiser company ID are required")
	}
	var res AdAccountDetail
	if err := s.c.do(ctx, call{method: http.MethodPost, path: "/v1/adAccounts", body: a}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// SetConfig turns an ad account ON or OFF.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/ad-account
func (s *AdAccountService) SetConfig(ctx context.Context, adAccountID int64, cfg Config) error {
	body := struct {
		Config Config `json:"config"`
	}{cfg}
	return s.c.do(ctx, call{method: http.MethodPut, path: "/v1/adAccounts", adAccountID: adAccountID, body: body}, nil)
}
