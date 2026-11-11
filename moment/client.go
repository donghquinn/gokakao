// Package moment is a client for the Kakao Moment (카카오모먼트) API.
//
// It authenticates with a business token from Business Auth
// (kakao.Client.Business) and reuses a [kakao.Client] for transport, errors
// and retries. Errors are [*kakao.Error]; Kakao Moment details are in its
// Extras field.
//
//	kc, _ := kakao.NewClient(kakao.WithRESTKey(restKey))
//	m := moment.New(kc, businessToken)
//	accounts, err := m.AdAccounts.List(ctx, nil)
//
// The app needs permission for the Kakao Moment API.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/common
package moment

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	kakao "github.com/donghquinn/gokakao"
)

// Client calls the Kakao Moment API on behalf of one business token. It is
// cheap to create and safe for concurrent use.
type Client struct {
	kc    *kakao.Client
	token string

	AdAccounts           *AdAccountService
	Campaigns            *CampaignService
	AdGroups             *AdGroupService
	Targeting            *TargetingService
	Creatives            *CreativeService
	BizboardLandings     *BizboardLandingService
	Reports              *ReportService
	Messages             *MessageService
	PersonalizedMessages *PersonalizedMessageService
	MessageAds           *MessageAdService
	AdViews              *AdViewService
	BizForms             *BizFormService
	Pixels               *PixelService
	Audiences            *AudienceService
	ReactionTargets      *ReactionTargetService
	CustomerData         *CustomerDataService
	CustomerFiles        *CustomerFileService
	FriendGroups         *FriendGroupService
}

// New returns a Client that sends requests through kc with businessToken.
func New(kc *kakao.Client, businessToken string) *Client {
	c := &Client{kc: kc, token: businessToken}
	c.AdAccounts = &AdAccountService{c: c}
	c.Campaigns = &CampaignService{c: c}
	c.AdGroups = &AdGroupService{c: c}
	c.Targeting = &TargetingService{c: c}
	c.Creatives = &CreativeService{c: c}
	c.BizboardLandings = &BizboardLandingService{c: c}
	c.Reports = &ReportService{c: c}
	c.Messages = &MessageService{c: c}
	c.PersonalizedMessages = &PersonalizedMessageService{c: c}
	c.MessageAds = &MessageAdService{c: c}
	c.AdViews = &AdViewService{c: c}
	c.BizForms = &BizFormService{c: c}
	c.Pixels = &PixelService{c: c}
	c.Audiences = &AudienceService{c: c}
	c.ReactionTargets = &ReactionTargetService{c: c}
	c.CustomerData = &CustomerDataService{c: c}
	c.CustomerFiles = &CustomerFileService{c: c}
	c.FriendGroups = &FriendGroupService{c: c}
	return c
}

// Services. Each is documented in the file that implements it.
type (
	// AdAccountService manages ad accounts (광고계정).
	AdAccountService struct{ c *Client }
	// CampaignService manages campaigns (캠페인).
	CampaignService struct{ c *Client }
	// AdGroupService manages ad groups (광고그룹).
	AdGroupService struct{ c *Client }
	// TargetingService reads targeting options for ad groups (광고그룹 타겟
	// 조회, 맞춤타겟 조회).
	TargetingService struct{ c *Client }
	// CreativeService manages creatives (소재).
	CreativeService struct{ c *Client }
	// BizboardLandingService lists bizboard landing types (비즈보드 랜딩 유형).
	BizboardLandingService struct{ c *Client }
	// ReportService reads performance reports (보고서).
	ReportService struct{ c *Client }
	// MessageService manages messages (메시지 관리).
	MessageService struct{ c *Client }
	// PersonalizedMessageService manages personalized messages (개인화 메시지
	// 관리).
	PersonalizedMessageService struct{ c *Client }
	// MessageAdService operates and manages message ads (메시지광고 운영,
	// 관리).
	MessageAdService struct{ c *Client }
	// AdViewService manages ad views (애드뷰).
	AdViewService struct{ c *Client }
	// BizFormService manages business forms (비즈니스폼+).
	BizFormService struct{ c *Client }
	// PixelService manages Pixel & SDK links (픽셀 & SDK 연동 관리).
	PixelService struct{ c *Client }
	// AudienceService manages audiences (오디언스).
	AudienceService struct{ c *Client }
	// ReactionTargetService manages ad reaction targets (광고반응타겟).
	ReactionTargetService struct{ c *Client }
	// CustomerDataService manages customer data (고객데이터).
	CustomerDataService struct{ c *Client }
	// CustomerFileService manages customer files (고객파일).
	CustomerFileService struct{ c *Client }
	// FriendGroupService manages friend groups (친구그룹).
	FriendGroupService struct{ c *Client }
)

// call describes one Keyword Ad API request.
type call struct {
	method string
	path   string // below /openapi, e.g. "/v4/adAccounts/pages"
	// adAccountID is sent as the adAccountId header when non-zero.
	adAccountID int64
	query       url.Values
	// body is sent as JSON; multipart as multipart/form-data.
	body      any
	multipart *kakao.Multipart
	header    http.Header
}

// do sends r and decodes the JSON response into out, which may be nil.
func (c *Client) do(ctx context.Context, r call, out any) error {
	if c.token == "" {
		return fmt.Errorf("%w: business token", kakao.ErrMissingCredential)
	}
	h := http.Header{}
	for k, v := range r.header {
		h[k] = v
	}
	if r.adAccountID != 0 {
		h.Set("adAccountId", strconv.FormatInt(r.adAccountID, 10))
	}
	return c.kc.Do(ctx, &kakao.Request{
		Method:      r.method,
		Host:        kakao.HostMoment,
		Path:        "/openapi" + r.path,
		Query:       r.query,
		JSON:        r.body,
		Multipart:   r.multipart,
		Header:      h,
		BearerToken: c.token,
	}, out)
}

// Page is a page of a paged list.
type Page[T any] struct {
	Content       []T `json:"content"`
	Number        int `json:"number"`
	Size          int `json:"size"`
	TotalElements int `json:"totalElements"`
	TotalPages    int `json:"totalPages"`
	// Page is the 0-based page number.
	Page int `json:"page"`
}

// PageOptions selects a page of a paged list.
type PageOptions struct {
	// Page is the 0-based page number.
	Page int
	// Size is the page size; Kakao's default is 10.
	Size int
}

func (o *PageOptions) query(q url.Values) url.Values {
	if q == nil {
		q = url.Values{}
	}
	if o != nil {
		if o.Page > 0 {
			q.Set("page", strconv.Itoa(o.Page))
		}
		if o.Size > 0 {
			q.Set("size", strconv.Itoa(o.Size))
		}
	}
	return q
}

func idPath(format string, id int64) string {
	return fmt.Sprintf(format, id)
}
