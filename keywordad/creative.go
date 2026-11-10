package keywordad

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	kakao "github.com/donghquinn/gokakao"
)

// PageInfo describes a page of the lists that nest paging details in a
// "page" object, such as creatives and images.
type PageInfo struct {
	// Number is the 0-based page number.
	Number        int `json:"number"`
	Size          int `json:"size"`
	TotalElements int `json:"totalElements"`
	TotalPages    int `json:"totalPages"`
}

// limitQuery sets the page and limit parameters used by the lists that
// return a [PageInfo].
func limitQuery(q url.Values, page, limit int) {
	if page > 0 {
		q.Set("page", strconv.Itoa(page))
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
}

// CreativeSummary is a creative in a list.
type CreativeSummary struct {
	BizChannelID kakao.Int64 `json:"bizChannelId"`
	ID           kakao.Int64 `json:"id"`
	Name         string      `json:"name"`
	// Title and Description are the basic creative's text.
	Title       string `json:"title"`
	Description string `json:"description"`
	// AssetTypes lists the extension assets in use.
	AssetTypes []AssetType `json:"assetTypes"`
	Config     Config      `json:"config"`
}

// CreativePage is a page of creatives.
type CreativePage struct {
	Content []CreativeSummary `json:"content"`
	Page    PageInfo          `json:"page"`
}

// CreativeListOptions are the optional parameters of
// [CreativeService.List].
type CreativeListOptions struct {
	// Page is the 0-based page number.
	Page int
	// Limit is the page size; Kakao's default is 10 and maximum 1000.
	Limit int
	// Config filters by status: ConfigOn or ConfigDel.
	Config Config
}

// List returns a page of a biz channel's creatives.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/creative
func (s *CreativeService) List(ctx context.Context, adAccountID, bizChannelID int64, opts *CreativeListOptions) (*CreativePage, error) {
	if bizChannelID == 0 {
		return nil, errors.New("keywordad: biz channel ID is required")
	}
	q := url.Values{"bizChannelId": {strconv.FormatInt(bizChannelID, 10)}}
	if opts != nil {
		if opts.Config != "" {
			q.Set("config", string(opts.Config))
		}
		limitQuery(q, opts.Page, opts.Limit)
	}
	var page CreativePage
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v1/creatives/basic", adAccountID: adAccountID, query: q}, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Creative is a creative (소재): a basic creative with optional extension
// assets.
type Creative struct {
	AdAccountID  kakao.Int64 `json:"adAccountId"`
	BizChannelID kakao.Int64 `json:"bizChannelId"`
	ID           kakao.Int64 `json:"id"`
	Name         string      `json:"name"`
	Title        string      `json:"title"`
	Description  string      `json:"description"`
	// LandingInfo defaults to the biz channel's URL.
	LandingInfo *LandingInfo `json:"landingInfo"`
	License     *License     `json:"license"`
	Assets      *Assets      `json:"assets"`
}

// Get returns a creative.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/creative
func (s *CreativeService) Get(ctx context.Context, adAccountID, creativeID int64) (*Creative, error) {
	var cr Creative
	if err := s.c.do(ctx, call{method: http.MethodGet, path: idPath("/v1/creatives/basic/%d", creativeID), adAccountID: adAccountID}, &cr); err != nil {
		return nil, err
	}
	return &cr, nil
}

// NewCreative describes a creative to create.
type NewCreative struct {
	BizChannelID int64  `json:"bizChannelId"`
	Name         string `json:"name"`
	// Title and Description are the basic creative's text.
	Title       string `json:"title"`
	Description string `json:"description"`
	// LandingInfo defaults to the biz channel's URL.
	LandingInfo *LandingInfo `json:"landingInfo,omitempty"`
	// License is the medical advertising review number, if required.
	License *License `json:"license,omitempty"`
	Assets  *Assets  `json:"assets,omitempty"`
}

// Create creates a creative. Kakao allows one call per 5 seconds.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/creative
func (s *CreativeService) Create(ctx context.Context, adAccountID int64, cr NewCreative) (*Creative, error) {
	if cr.BizChannelID == 0 || cr.Name == "" || cr.Title == "" || cr.Description == "" {
		return nil, errors.New("keywordad: biz channel ID, name, title and description are required")
	}
	var res Creative
	if err := s.c.do(ctx, call{method: http.MethodPost, path: "/v1/creatives/basic", adAccountID: adAccountID, body: cr}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Delete deletes a creative.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/creative
func (s *CreativeService) Delete(ctx context.Context, adAccountID, creativeID int64) error {
	return s.c.do(ctx, call{method: http.MethodDelete, path: idPath("/v1/creatives/%d", creativeID), adAccountID: adAccountID}, nil)
}
