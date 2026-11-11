package moment

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	kakao "github.com/donghquinn/gokakao"
)

// BizFormGender is the gender in a business form response.
type BizFormGender string

// Business form genders.
const (
	BizFormMale   BizFormGender = "MALE"   // 남성
	BizFormFemale BizFormGender = "FEMALE" // 여성
)

// BizFormAnswer is the answer to one survey question.
type BizFormAnswer struct {
	Question string `json:"question"`
	// Answer is the answer as text, e.g. "1, 3, 없음" for several choices.
	Answer string `json:"answer"`
}

// BizFormAgreement is an optional agreement (선택 동의문) and whether the
// user agreed.
type BizFormAgreement struct {
	Title string `json:"title"`
	Agree bool   `json:"agree"`
}

// BizFormResult is one response to a business form (비즈니스폼+ 응답).
// Personal fields are empty when the form does not collect them.
type BizFormResult struct {
	ApplyID     kakao.Int64 `json:"applyId"`
	SubmittedAt kakao.Time  `json:"submittedAt"`
	// ExpiresAt is 90 days after SubmittedAt.
	ExpiresAt kakao.Time `json:"expiresAt"`
	Email     string     `json:"email"`
	// BirthDate is "yyyy-MM-dd".
	BirthDate          string             `json:"birthDate"`
	Gender             BizFormGender      `json:"gender"`
	Name               string             `json:"name"`
	PhoneNumber        string             `json:"phoneNumber"`
	Address            string             `json:"address"`
	Answers            []BizFormAnswer    `json:"answers"`
	OptionalAgreements []BizFormAgreement `json:"optionalAgreements"`
	// ChannelAddStatus is e.g. "-", "already added", "Y" or "N".
	ChannelAddStatus string `json:"channelAddStatus"`
	// InflowSource is e.g. "모먼트" or "기타".
	InflowSource string `json:"inflowSource"`
	// KakaoResponseUse is the consent to use the response to improve
	// ads; false when the form does not ask.
	KakaoResponseUse kakao.YesNo `json:"kakaoResponseUse"`
	CreativeID       kakao.Int64 `json:"creativeId"`
}

// BizFormResultPage is a page of business form responses.
type BizFormResultPage struct {
	Data    []BizFormResult `json:"data"`
	HasNext bool            `json:"hasNext"`
	// NextCursor is the CursorID of the next page.
	NextCursor kakao.Int64 `json:"nextCursor"`
}

// BizFormResultOptions are the optional parameters of
// [BizFormService.Results].
type BizFormResultOptions struct {
	// CursorID is the NextCursor of the previous page; 0 for the first.
	CursorID int64
	// Size is the page size, at most 1,000; Kakao's default is 100.
	Size int
}

// Results returns a page of the responses to a business form created with
// the ad account. formID is e.g. "MA-1111". It needs the
// moment_bizform_result_read business consent, the same business
// registration number on the app and the ad account, and permission to
// download the form's reports. Kakao allows one call per 5 seconds.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/bizformplus
func (s *BizFormService) Results(ctx context.Context, adAccountID int64, formID string, opts *BizFormResultOptions) (*BizFormResultPage, error) {
	if adAccountID == 0 || formID == "" {
		return nil, errors.New("moment: ad account ID and form ID are required")
	}
	q := url.Values{"formId": {formID}}
	if opts != nil {
		if opts.CursorID != 0 {
			q.Set("cursorId", strconv.FormatInt(opts.CursorID, 10))
		}
		if opts.Size > 0 {
			q.Set("size", strconv.Itoa(opts.Size))
		}
	}
	var page BizFormResultPage
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v4/adAccounts/bizFormPlus/report", adAccountID: adAccountID, query: q}, &page); err != nil {
		return nil, err
	}
	return &page, nil
}
