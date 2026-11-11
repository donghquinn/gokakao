package moment

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"

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
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/ad-account
func (s *AdAccountService) List(ctx context.Context, opts *AdAccountListOptions) (*Page[AdAccount], error) {
	q := url.Values{}
	if opts != nil {
		if len(opts.Configs) > 0 {
			cs := make([]string, len(opts.Configs))
			for i, c := range opts.Configs {
				cs[i] = string(c)
			}
			q.Set("config", strings.Join(cs, ","))
		}
		opts.query(q)
	}
	var page Page[AdAccount]
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v4/adAccounts/pages", query: q}, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// AdAccountType is the kind of an ad account (광고계정 유형).
type AdAccountType string

// Ad account types.
const (
	AdAccountBusiness   AdAccountType = "BUSINESS"   // 사업자 광고계정
	AdAccountIndividual AdAccountType = "INDIVIDUAL" // 개인 광고계정
)

// AdAccountDetail is an ad account.
type AdAccountDetail struct {
	ID   kakao.Int64   `json:"id"`
	Name string        `json:"name"`
	Type AdAccountType `json:"type"`
	// Advertiser is set for business ad accounts.
	Advertiser *Company `json:"advertiser"`
	Config     Config   `json:"config"`
	// AdminStop reports a stop by Kakao administrators.
	AdminStop    bool `json:"isAdminStop"`
	OutOfBalance bool `json:"isOutOfBalance"`
	// StatusDescription combines Config, AdminStop and OutOfBalance into
	// the delivery status, e.g. "운영중" or "잔액부족".
	StatusDescription string `json:"statusDescription"`
	// BizWalletID is the linked biz wallet; [AdAccountService.Create]
	// does not return it.
	BizWalletID kakao.Int64 `json:"bizWalletId"`
}

// Get returns an ad account. Kakao allows one call per second per user and
// app.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/ad-account
func (s *AdAccountService) Get(ctx context.Context, adAccountID int64) (*AdAccountDetail, error) {
	if adAccountID == 0 {
		return nil, errors.New("moment: ad account ID is required")
	}
	var a AdAccountDetail
	if err := s.c.do(ctx, call{method: http.MethodGet, path: idPath("/v4/adAccounts/%d", adAccountID), adAccountID: adAccountID}, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

// AdAccountBalance is an ad account's balance (실시간 잔액) in KRW.
type AdAccountBalance struct {
	ID          kakao.Int64 `json:"id"`
	BizWalletID kakao.Int64 `json:"bizWalletId"`
	// Cash is paid cash (유상캐시) and FreeCash free cash (무상캐시).
	Cash     int64 `json:"cash"`
	FreeCash int64 `json:"freeCash"`
}

// Balance returns an ad account's real-time balance.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/ad-account
func (s *AdAccountService) Balance(ctx context.Context, adAccountID int64) (*AdAccountBalance, error) {
	if adAccountID == 0 {
		return nil, errors.New("moment: ad account ID is required")
	}
	var b AdAccountBalance
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v4/adAccounts/balance", adAccountID: adAccountID}, &b); err != nil {
		return nil, err
	}
	return &b, nil
}

// BizRight is the company holding an ad account's business right (영업권).
type BizRight struct {
	// ID is the ad account ID.
	ID                  kakao.Int64 `json:"id"`
	BizRightCompanyName string      `json:"bizRightCompanyName"`
}

// BizRight returns an ad account's business right.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/ad-account
func (s *AdAccountService) BizRight(ctx context.Context, adAccountID int64) (*BizRight, error) {
	if adAccountID == 0 {
		return nil, errors.New("moment: ad account ID is required")
	}
	var b BizRight
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v4/adAccounts/bizRight", adAccountID: adAccountID}, &b); err != nil {
		return nil, err
	}
	return &b, nil
}

// ApproveBizRight approves an ad account's business right (영업권 자동
// 승인) and returns it.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/ad-account
func (s *AdAccountService) ApproveBizRight(ctx context.Context, adAccountID int64) (*BizRight, error) {
	if adAccountID == 0 {
		return nil, errors.New("moment: ad account ID is required")
	}
	var b BizRight
	if err := s.c.do(ctx, call{method: http.MethodPost, path: idPath("/v4/adAccounts/%d/bizRight", adAccountID), adAccountID: adAccountID}, &b); err != nil {
		return nil, err
	}
	return &b, nil
}

// ChannelProfile is a KakaoTalk Channel profile of the user.
type ChannelProfile struct {
	// ID is the profile ID, e.g. "_ZQxd", used as [Objective] Value when
	// creating a campaign.
	ID   string `json:"id"`
	Name string `json:"name"`
	// SearchID is the search ID.
	SearchID string `json:"searchId"`
}

// ChannelProfiles returns the user's KakaoTalk Channel profiles.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/ad-account
func (s *AdAccountService) ChannelProfiles(ctx context.Context, adAccountID int64) ([]ChannelProfile, error) {
	if adAccountID == 0 {
		return nil, errors.New("moment: ad account ID is required")
	}
	var ps []ChannelProfile
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v4/adAccounts/channel/profiles", adAccountID: adAccountID}, &ps); err != nil {
		return nil, err
	}
	return ps, nil
}

// Tracker is a Pixel & SDK the ad account has permission on.
type Tracker struct {
	// ID is the Pixel & SDK ID, used as a campaign's track ID or
	// [Objective] Value.
	ID   string `json:"id"`
	Name string `json:"name"`
}

// TrackerListOptions are the optional parameters of
// [AdAccountService.Trackers].
type TrackerListOptions struct {
	// CampaignType and Goal narrow the list to Pixel & SDKs usable with
	// that campaign type and goal.
	CampaignType CampaignType
	Goal         Goal
}

// Trackers returns the Pixel & SDKs the ad account has permission on.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/ad-account
func (s *AdAccountService) Trackers(ctx context.Context, adAccountID int64, opts *TrackerListOptions) ([]Tracker, error) {
	if adAccountID == 0 {
		return nil, errors.New("moment: ad account ID is required")
	}
	q := url.Values{}
	if opts != nil {
		if opts.CampaignType != "" {
			q.Set("campaignType", string(opts.CampaignType))
		}
		if opts.Goal != "" {
			q.Set("goal", string(opts.Goal))
		}
	}
	var ts []Tracker
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v4/adAccounts/trackers", adAccountID: adAccountID, query: q}, &ts); err != nil {
		return nil, err
	}
	return ts, nil
}

// NewAdAccount describes an ad account to create.
type NewAdAccount struct {
	Name string `json:"name"`
	// AdvertiserCompanyID is the advertiser's ID from
	// [AdAccountService.Company].
	AdvertiserCompanyID int64 `json:"advertiserCompanyId"`
}

// Create creates a business ad account. Only business users can create
// ad accounts. Kakao allows one call per second per user.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/ad-account
func (s *AdAccountService) Create(ctx context.Context, a NewAdAccount) (*AdAccountDetail, error) {
	if a.Name == "" || a.AdvertiserCompanyID == 0 {
		return nil, errors.New("moment: name and advertiser company ID are required")
	}
	var res AdAccountDetail
	if err := s.c.do(ctx, call{method: http.MethodPost, path: "/v4/adAccounts", body: a}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// SetConfig turns an ad account ON or OFF. It only applies to accounts
// that are ON or OFF and not being deleted. Turning an account OFF ends
// or stops the sending of its direct message creatives. Kakao allows one
// call per second per user.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/ad-account
func (s *AdAccountService) SetConfig(ctx context.Context, adAccountID int64, cfg Config) error {
	if adAccountID == 0 {
		return errors.New("moment: ad account ID is required")
	}
	if cfg != ConfigOn && cfg != ConfigOff {
		return errors.New("moment: config must be ON or OFF")
	}
	body := struct {
		Config Config `json:"config"`
	}{cfg}
	return s.c.do(ctx, call{method: http.MethodPut, path: "/v4/adAccounts/onOff", adAccountID: adAccountID, body: body}, nil)
}

// CompanyAddress is a business address.
type CompanyAddress struct {
	ZipCode       string `json:"zipCode"`
	BaseAddress   string `json:"baseAddress"`
	DetailAddress string `json:"detailAddress"`
}

// CompanyDetail is a business (사업자) as returned by
// [AdAccountService.Company].
type CompanyDetail struct {
	// ID is used as [NewAdAccount] AdvertiserCompanyID.
	ID                         kakao.Int64 `json:"id"`
	BusinessRegistrationNumber string      `json:"businessRegistrationNumber"`
	Name                       string      `json:"name"`
	// RepresentationName is the representative (대표자).
	RepresentationName string          `json:"representationName"`
	Address            *CompanyAddress `json:"address"`
	// BusinessCategory (업태) and BusinessItem (종목).
	BusinessCategory string `json:"businessCategory"`
	BusinessItem     string `json:"businessItem"`
}

// Company looks up a business by its business registration number, e.g.
// "120-81-47521".
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/ad-account-mgmt
func (s *AdAccountService) Company(ctx context.Context, bizNum string) (*CompanyDetail, error) {
	if bizNum == "" {
		return nil, errors.New("moment: business registration number is required")
	}
	var c CompanyDetail
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v4/companies", query: url.Values{"bizNum": {bizNum}}}, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
