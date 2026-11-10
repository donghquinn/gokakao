package keywordad

import (
	"context"
	"errors"
	"net/http"

	kakao "github.com/donghquinn/gokakao"
)

// BizChannel is a biz channel (비즈채널) in a list: a website promoted by
// the ad account's ads.
type BizChannel struct {
	ID         kakao.Int64 `json:"id"`
	Name       string      `json:"name"`
	WebsiteURL WebsiteURL  `json:"websiteUrl"`
	// BusinessTypeID is the business category (업종) ID, which decides the
	// required review documents (RequiredDocument in the reference).
	BusinessTypeID kakao.Int64 `json:"ufoBusinessTypeId"`
}

// BizChannelDetail is a biz channel.
type BizChannelDetail struct {
	ID         kakao.Int64 `json:"id"`
	Name       string      `json:"name"`
	WebsiteURL WebsiteURL  `json:"websiteUrl"`
	// BusinessTypeID is the business category (업종) ID.
	BusinessTypeID kakao.Int64  `json:"ufoBusinessTypeId"`
	ReviewStatus   ReviewStatus `json:"reviewStatus"`
	// HasAccessAccount reports whether a reviewer login is registered.
	HasAccessAccount bool          `json:"hasAccessAccount"`
	TalkChannels     []TalkChannel `json:"talkChannels"`
}

// List returns the ad account's biz channel websites.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/biz-channel
func (s *BizChannelService) List(ctx context.Context, adAccountID int64) ([]BizChannel, error) {
	if adAccountID == 0 {
		return nil, errors.New("keywordad: ad account ID is required")
	}
	var cs []BizChannel
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v1/bizChannels/websites", adAccountID: adAccountID}, &cs); err != nil {
		return nil, err
	}
	return cs, nil
}

// Get returns a biz channel website.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/biz-channel
func (s *BizChannelService) Get(ctx context.Context, adAccountID, bizChannelID int64) (*BizChannelDetail, error) {
	if adAccountID == 0 || bizChannelID == 0 {
		return nil, errors.New("keywordad: ad account ID and biz channel ID are required")
	}
	var c BizChannelDetail
	err := s.c.do(ctx, call{method: http.MethodGet, path: idPath("/v1/bizChannels/websites/%d", bizChannelID), adAccountID: adAccountID}, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// ReviewDocumentURL is a review document that Kakao downloads from a URL
// (UfoReviewDocument in the reference).
type ReviewDocumentURL struct {
	Type ReviewDocumentType `json:"type"`
	// RequiredReviewDocumentID is the kind of document required for the
	// business category, e.g. 1 (사업자등록증); see RequiredDocument in the
	// reference.
	RequiredReviewDocumentID int64 `json:"requiredReviewDocumentId"`
	// FileURL is an http:// or https:// URL of the file.
	FileURL string `json:"fileUrl"`
}

// NewBizChannel describes a biz channel website to create.
type NewBizChannel struct {
	Name string `json:"name"`
	// BusinessTypeID is the business category (업종) ID; see
	// RequiredDocument in the reference.
	BusinessTypeID int64      `json:"ufoBusinessTypeId"`
	WebsiteURL     WebsiteURL `json:"websiteUrl"`
	// AccessAccount is a login for reviewing a members-only site.
	AccessAccount *AccessAccount `json:"accessAccount,omitempty"`
	// TalkChannels are KakaoTalk Channel search IDs, e.g. "@kakao".
	TalkChannels []string `json:"talkChannels,omitempty"`
	// ReviewDocuments are review documents given by URL. To upload a
	// local file instead, create the biz channel and then use
	// [DocumentService.Create].
	ReviewDocuments []ReviewDocumentURL `json:"reviewDocuments,omitempty"`
}

// Create creates a biz channel website. It is limited to 1 request per 5
// seconds.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/biz-channel
func (s *BizChannelService) Create(ctx context.Context, adAccountID int64, c NewBizChannel) (*BizChannelDetail, error) {
	if adAccountID == 0 {
		return nil, errors.New("keywordad: ad account ID is required")
	}
	if c.Name == "" || c.BusinessTypeID == 0 || c.WebsiteURL == (WebsiteURL{}) {
		return nil, errors.New("keywordad: name, business type ID and website URL are required")
	}
	var res BizChannelDetail
	if err := s.c.do(ctx, call{method: http.MethodPost, path: "/v1/bizChannels/websites", adAccountID: adAccountID, body: c}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// AddTalkChannels registers KakaoTalk Channels, by search ID such as
// "@kakao", to a biz channel. It is limited to 1 request per 5 seconds,
// shared with [BizChannelService.RemoveTalkChannels].
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/biz-channel
func (s *BizChannelService) AddTalkChannels(ctx context.Context, adAccountID, bizChannelID int64, uuids []string) error {
	return s.updateTalkChannels(ctx, adAccountID, bizChannelID, "ADD", uuids)
}

// RemoveTalkChannels unregisters KakaoTalk Channels, by search ID such as
// "@kakao", from a biz channel. It is limited to 1 request per 5 seconds,
// shared with [BizChannelService.AddTalkChannels].
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/biz-channel
func (s *BizChannelService) RemoveTalkChannels(ctx context.Context, adAccountID, bizChannelID int64, uuids []string) error {
	return s.updateTalkChannels(ctx, adAccountID, bizChannelID, "DEL", uuids)
}

func (s *BizChannelService) updateTalkChannels(ctx context.Context, adAccountID, bizChannelID int64, registration string, uuids []string) error {
	if adAccountID == 0 || bizChannelID == 0 {
		return errors.New("keywordad: ad account ID and biz channel ID are required")
	}
	if len(uuids) == 0 {
		return errors.New("keywordad: talk channels are required")
	}
	body := struct {
		Registration string   `json:"registration"`
		TalkChannels []string `json:"talkChannels"`
	}{registration, uuids}
	return s.c.do(ctx, call{
		method: http.MethodPut, path: idPath("/v1/bizChannels/websites/%d/talkChannels", bizChannelID),
		adAccountID: adAccountID, body: body,
	}, nil)
}

// Delete deletes a biz channel.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/biz-channel
func (s *BizChannelService) Delete(ctx context.Context, adAccountID, bizChannelID int64) error {
	if adAccountID == 0 || bizChannelID == 0 {
		return errors.New("keywordad: ad account ID and biz channel ID are required")
	}
	return s.c.do(ctx, call{method: http.MethodDelete, path: idPath("/v1/bizChannels/%d", bizChannelID), adAccountID: adAccountID}, nil)
}
