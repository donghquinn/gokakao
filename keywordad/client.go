// Package keywordad is a client for the Kakao Keyword Ad (카카오 키워드광고)
// API.
//
// It authenticates with a business token from Business Auth
// (kakao.Client.Business) and reuses a [kakao.Client] for transport, errors
// and retries. Errors are [*kakao.Error]; Keyword Ad details are in its
// Extras field.
//
//	kc, _ := kakao.NewClient(kakao.WithRESTKey(restKey))
//	ka := keywordad.New(kc, businessToken)
//	accounts, err := ka.AdAccounts.List(ctx, nil)
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/common
package keywordad

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	kakao "github.com/donghquinn/gokakao"
)

// Client calls the Keyword Ad API on behalf of one business token. It is
// cheap to create and safe for concurrent use.
type Client struct {
	kc    *kakao.Client
	token string

	AdAccounts    *AdAccountService
	BizChannels   *BizChannelService
	Campaigns     *CampaignService
	AdGroups      *AdGroupService
	Keywords      *KeywordService
	Creatives     *CreativeService
	CreativeLinks *CreativeLinkService
	Images        *ImageService
	Documents     *DocumentService
	Pixels        *PixelService
	BusinessInfo  *BusinessInfoService
	Reports       *ReportService
}

// New returns a Client that sends requests through kc with businessToken.
func New(kc *kakao.Client, businessToken string) *Client {
	c := &Client{kc: kc, token: businessToken}
	c.AdAccounts = &AdAccountService{c: c}
	c.BizChannels = &BizChannelService{c: c}
	c.Campaigns = &CampaignService{c: c}
	c.AdGroups = &AdGroupService{c: c}
	c.Keywords = &KeywordService{c: c}
	c.Creatives = &CreativeService{c: c}
	c.CreativeLinks = &CreativeLinkService{c: c}
	c.Images = &ImageService{c: c}
	c.Documents = &DocumentService{c: c}
	c.Pixels = &PixelService{c: c}
	c.BusinessInfo = &BusinessInfoService{c: c}
	c.Reports = &ReportService{c: c}
	return c
}

// Services. Each is documented in the file that implements it.
type (
	// AdAccountService manages ad accounts (광고계정).
	AdAccountService struct{ c *Client }
	// BizChannelService manages biz channels (비즈채널).
	BizChannelService struct{ c *Client }
	// CampaignService manages campaigns (캠페인).
	CampaignService struct{ c *Client }
	// AdGroupService manages ad groups (광고그룹).
	AdGroupService struct{ c *Client }
	// KeywordService manages keywords (키워드).
	KeywordService struct{ c *Client }
	// CreativeService manages creatives (소재).
	CreativeService struct{ c *Client }
	// CreativeLinkService links creatives to ad groups (소재연결).
	CreativeLinkService struct{ c *Client }
	// ImageService manages images (이미지 관리).
	ImageService struct{ c *Client }
	// DocumentService manages review documents (심사서류 관리).
	DocumentService struct{ c *Client }
	// PixelService manages Pixel & SDK links (픽셀 & SDK 연동 관리).
	PixelService struct{ c *Client }
	// BusinessInfoService reads business information (사업자 정보).
	BusinessInfoService struct{ c *Client }
	// ReportService reads performance reports (보고서).
	ReportService struct{ c *Client }
)

// call describes one Keyword Ad API request.
type call struct {
	method string
	path   string // below /openapi, e.g. "/v1/adAccounts/pages"
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
		Host:        kakao.HostKeywordAd,
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
