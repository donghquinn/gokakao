package keywordad

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	kakao "github.com/donghquinn/gokakao"
)

// Keyword is a keyword (키워드) in an ad group.
type Keyword struct {
	AdGroupID kakao.Int64 `json:"adGroupId"`
	ID        kakao.Int64 `json:"id"`
	Text      string      `json:"text"`
	Config    Config      `json:"config"`
	// LandingInfo is the keyword's own landing URL; all fields are empty
	// when it uses the creative's.
	LandingInfo LandingInfo `json:"landingInfo"`
	BidStrategy BidStrategy `json:"bidStrategy"`
	// Status lists the delivery statuses; see [Status].
	Status       []Status     `json:"status"`
	ReviewStatus ReviewStatus `json:"reviewStatus"`
}

// KeywordListOptions are the optional parameters of [KeywordService.List].
type KeywordListOptions struct {
	// Configs filters by status; Kakao lists ON and OFF keywords when
	// empty.
	Configs []Config
}

// List returns the keywords of an ad group.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/keyword
func (s *KeywordService) List(ctx context.Context, adAccountID, adGroupID int64, opts *KeywordListOptions) ([]Keyword, error) {
	if adGroupID == 0 {
		return nil, errors.New("keywordad: ad group ID is required")
	}
	q := url.Values{"adGroupId": {strconv.FormatInt(adGroupID, 10)}}
	if opts != nil {
		for _, cfg := range opts.Configs {
			q.Add("config", string(cfg))
		}
	}
	var ks []Keyword
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v1/keywords", adAccountID: adAccountID, query: q}, &ks); err != nil {
		return nil, err
	}
	return ks, nil
}

// Get returns a keyword.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/keyword
func (s *KeywordService) Get(ctx context.Context, adAccountID, keywordID int64) (*Keyword, error) {
	var k Keyword
	if err := s.c.do(ctx, call{method: http.MethodGet, path: idPath("/v1/keywords/%d", keywordID), adAccountID: adAccountID}, &k); err != nil {
		return nil, err
	}
	return &k, nil
}

// NewKeywords describes keywords to create with one bid strategy.
type NewKeywords struct {
	AdGroupID   int64       `json:"adGroupId"`
	Texts       []string    `json:"texts"`
	BidStrategy BidStrategy `json:"bidStrategy"`
}

// Create creates keywords in an ad group and returns them. Keywords already
// in the ad group are left out of the result. Kakao allows one request per
// 5 seconds.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/keyword
func (s *KeywordService) Create(ctx context.Context, adAccountID int64, k NewKeywords) ([]Keyword, error) {
	if k.AdGroupID == 0 || len(k.Texts) == 0 {
		return nil, errors.New("keywordad: ad group ID and keyword texts are required")
	}
	if err := validateKeywordBidStrategy(k.BidStrategy); err != nil {
		return nil, err
	}
	var ks []Keyword
	if err := s.c.do(ctx, call{method: http.MethodPost, path: "/v1/keywords", adAccountID: adAccountID, body: k}, &ks); err != nil {
		return nil, err
	}
	return ks, nil
}

// NewIndividualKeyword is one keyword of [NewIndividualKeywords].
type NewIndividualKeyword struct {
	Text        string       `json:"text"`
	BidStrategy BidStrategy  `json:"bidStrategy"`
	LandingInfo *LandingInfo `json:"landingInfo,omitempty"`
}

// NewIndividualKeywords describes keywords to create, each with its own
// bid strategy and landing URL.
type NewIndividualKeywords struct {
	AdGroupID int64                  `json:"adGroupId"`
	Keywords  []NewIndividualKeyword `json:"individualKeywords"`
}

// CreateIndividual creates keywords with per-keyword settings and returns
// them. Keywords already in the ad group are left out of the result. Kakao
// allows one request per 5 seconds.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/keyword
func (s *KeywordService) CreateIndividual(ctx context.Context, adAccountID int64, k NewIndividualKeywords) ([]Keyword, error) {
	if k.AdGroupID == 0 || len(k.Keywords) == 0 {
		return nil, errors.New("keywordad: ad group ID and keywords are required")
	}
	for _, kw := range k.Keywords {
		if kw.Text == "" {
			return nil, errors.New("keywordad: keyword text is required")
		}
		if err := validateKeywordBidStrategy(kw.BidStrategy); err != nil {
			return nil, err
		}
	}
	var ks []Keyword
	if err := s.c.do(ctx, call{method: http.MethodPost, path: "/v1/keywords/individual", adAccountID: adAccountID, body: k}, &ks); err != nil {
		return nil, err
	}
	return ks, nil
}

// KeywordChanges is the full new state of a keyword. Kakao replaces both
// fields.
type KeywordChanges struct {
	// LandingInfo nil removes the keyword's landing URL.
	LandingInfo *LandingInfo `json:"landingInfo,omitempty"`
	BidStrategy BidStrategy  `json:"bidStrategy"`
}

// Update replaces a keyword's landing URL and bid strategy. Kakao allows
// one request per 5 seconds.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/keyword
func (s *KeywordService) Update(ctx context.Context, adAccountID, keywordID int64, ch KeywordChanges) (*Keyword, error) {
	if err := validateKeywordBidStrategy(ch.BidStrategy); err != nil {
		return nil, err
	}
	var k Keyword
	if err := s.c.do(ctx, call{method: http.MethodPut, path: idPath("/v1/keywords/%d", keywordID), adAccountID: adAccountID, body: ch}, &k); err != nil {
		return nil, err
	}
	return &k, nil
}

// SetConfig turns a keyword ON or OFF. It only works while the keyword is
// ON or OFF. Kakao allows one request per second.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/keyword
func (s *KeywordService) SetConfig(ctx context.Context, adAccountID, keywordID int64, cfg Config) error {
	body := struct {
		Config Config `json:"config"`
	}{cfg}
	return s.c.do(ctx, call{method: http.MethodPatch, path: idPath("/v1/keywords/%d/onOff", keywordID), adAccountID: adAccountID, body: body}, nil)
}

// UpdateBidStrategy sets a keyword's bid strategy. Kakao allows one request
// per 5 seconds.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/keyword
func (s *KeywordService) UpdateBidStrategy(ctx context.Context, adAccountID, keywordID int64, b BidStrategy) error {
	if err := validateKeywordBidStrategy(b); err != nil {
		return err
	}
	body := struct {
		BidStrategy BidStrategy `json:"bidStrategy"`
	}{b}
	return s.c.do(ctx, call{method: http.MethodPatch, path: idPath("/v1/keywords/%d/bidStrategy", keywordID), adAccountID: adAccountID, body: body}, nil)
}

// UpdateLandingInfo sets a keyword's landing URL; a zero LandingInfo
// removes it. Kakao allows one request per 5 seconds.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/keyword
func (s *KeywordService) UpdateLandingInfo(ctx context.Context, adAccountID, keywordID int64, l LandingInfo) error {
	body := struct {
		LandingInfo LandingInfo `json:"landingInfo"`
	}{l}
	return s.c.do(ctx, call{method: http.MethodPatch, path: idPath("/v1/keywords/%d/landingInfo", keywordID), adAccountID: adAccountID, body: body}, nil)
}

// KeywordReviewHoldReasons are why a keyword's review was held (심사보류).
type KeywordReviewHoldReasons struct {
	ID      kakao.Int64 `json:"id"`
	Reasons []string    `json:"reasons"`
}

// ReviewHoldReasons returns why a keyword's review was held. Kakao allows
// one request per 10 seconds.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/keyword
func (s *KeywordService) ReviewHoldReasons(ctx context.Context, adAccountID, keywordID int64) (*KeywordReviewHoldReasons, error) {
	var r KeywordReviewHoldReasons
	if err := s.c.do(ctx, call{method: http.MethodGet, path: idPath("/v1/keywords/%d/reviewReason", keywordID), adAccountID: adAccountID}, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// KeywordQuality is a keyword's quality index (품질지수).
type KeywordQuality struct {
	ID      kakao.Int64 `json:"id"`
	Text    string      `json:"text"`
	Quality int         `json:"quality"`
}

// QualityIndex returns a keyword's quality index. Kakao allows one request
// per 10 seconds.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/keyword
func (s *KeywordService) QualityIndex(ctx context.Context, adAccountID, keywordID int64) (*KeywordQuality, error) {
	var q KeywordQuality
	if err := s.c.do(ctx, call{method: http.MethodGet, path: idPath("/v1/keywords/%d/quality", keywordID), adAccountID: adAccountID}, &q); err != nil {
		return nil, err
	}
	return &q, nil
}

// Delete deletes a keyword.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/keyword
func (s *KeywordService) Delete(ctx context.Context, adAccountID, keywordID int64) error {
	return s.c.do(ctx, call{method: http.MethodDelete, path: idPath("/v1/keywords/%d", keywordID), adAccountID: adAccountID}, nil)
}

func validateKeywordBidStrategy(b BidStrategy) error {
	if b.Type == "" {
		return errors.New("keywordad: bid strategy type is required")
	}
	if b.Type == BidManual && b.BidAmount <= 0 {
		return errors.New("keywordad: manual bid strategy requires a bid amount")
	}
	return nil
}
