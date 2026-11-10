package keywordad

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	kakao "github.com/donghquinn/gokakao"
)

// BusinessInfo is a registered business (사업자 정보).
type BusinessInfo struct {
	// ID is the company ID to use as
	// [NewAdAccount.AdvertiserCompanyID].
	ID                         kakao.Int64 `json:"id"`
	BusinessRegistrationNumber string      `json:"businessRegistrationNumber"`
	Name                       string      `json:"name"`
	// RepresentationName is the representative's name (대표자).
	RepresentationName string `json:"representationName"`
	// BusinessCategory (업태) and BusinessItem (종목) classify the
	// business.
	BusinessCategory string  `json:"businessCategory"`
	BusinessItem     string  `json:"businessItem"`
	Address          Address `json:"address"`
}

// Get returns the business with a business registration number
// (사업자등록번호), e.g. "120-81-47521".
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/business-info
func (s *BusinessInfoService) Get(ctx context.Context, adAccountID int64, businessRegistrationNumber string) (*BusinessInfo, error) {
	if businessRegistrationNumber == "" {
		return nil, errors.New("keywordad: business registration number is required")
	}
	q := url.Values{"bizNum": {businessRegistrationNumber}}
	var b BusinessInfo
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v1/companies", adAccountID: adAccountID, query: q}, &b); err != nil {
		return nil, err
	}
	return &b, nil
}
