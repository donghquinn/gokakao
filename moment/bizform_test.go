package moment

import (
	"errors"
	"net/http"
	"testing"
	"time"

	kakao "github.com/donghquinn/gokakao"
)

func TestBizFormService_Results(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/adAccounts/bizFormPlus/report", 200, `{
  "code": 200,
  "message": "Success",
  "data": [
    {
      "applyId": 526,
      "submittedAt": "2026-03-20T08:27:15.848222+09:00",
      "expiresAt": "2026-06-18T08:27:15.848222+09:00",
      "email": "test1@example.com",
      "birthDate": "${BIRTH_DATE}",
      "gender": "MALE",
      "name": "${NAME}",
      "phoneNumber": "${PHONE_NUMBER}",
      "address": "${ADDRESS}",
      "answers": [
        {"question": "단일 선택형 제목", "answer": "3"},
        {"question": "복수 선택형 제목", "answer": "1, 3, 없음"},
        {"question": "서술형 제목", "answer": "테스트 데이터 생성을 위한 긴 텍스트 응답입니다. 실제 사용자의 응답을 시뮬레이션합니다."},
        {"question": "날짜형 제목", "answer": "2024-04-11"},
        {"question": "숫자형 제목", "answer": "37"}
      ],
      "optionalAgreements": [{"title": "기타 동의문 제목", "agree": false}],
      "channelAddStatus": "Y",
      "inflowSource": "모먼트",
      "kakaoResponseUse": "Y",
      "creativeId": 193057
    }
  ],
  "hasNext": true,
  "nextCursor": 526
}`)

	page, err := c.BizForms.Results(t.Context(), 1111, "MA-1111", &BizFormResultOptions{CursorID: 500, Size: 100})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodGet || req.Query.Get("formId") != "MA-1111" || req.Query.Get("cursorId") != "500" || req.Query.Get("size") != "100" {
		t.Errorf("request = %s ?%v", req.Method, req.Query)
	}
	if !page.HasNext || page.NextCursor != 526 || len(page.Data) != 1 {
		t.Fatalf("page = %+v", page)
	}
	r := page.Data[0]
	kst := time.FixedZone("", 9*60*60)
	if r.ApplyID != 526 || !r.SubmittedAt.Equal(time.Date(2026, 3, 20, 8, 27, 15, 848222000, kst)) ||
		!r.ExpiresAt.Equal(time.Date(2026, 6, 18, 8, 27, 15, 848222000, kst)) ||
		r.Email != "test1@example.com" || r.Gender != BizFormMale || r.Name != "${NAME}" ||
		len(r.Answers) != 5 || r.Answers[1] != (BizFormAnswer{Question: "복수 선택형 제목", Answer: "1, 3, 없음"}) ||
		len(r.OptionalAgreements) != 1 || r.OptionalAgreements[0] != (BizFormAgreement{Title: "기타 동의문 제목"}) ||
		r.ChannelAddStatus != "Y" || r.InflowSource != "모먼트" || !bool(r.KakaoResponseUse) || r.CreativeID != 193057 {
		t.Errorf("result = %+v", r)
	}

	if _, err := c.BizForms.Results(t.Context(), 1111, "MA-1111", nil); err != nil {
		t.Fatal(err)
	}
	if q := srv.LastRequest().Query; len(q) != 1 {
		t.Errorf("nil options query = %v", q)
	}
	if _, err := c.BizForms.Results(t.Context(), 1111, "", nil); err == nil {
		t.Error("missing form: want error")
	}
	if _, err := c.BizForms.Results(t.Context(), 0, "MA-1", nil); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestBizFormService_Results_error(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("/", 400, `{"code":-813,"msg":"KakaoMomentException"}`)
	_, err := c.BizForms.Results(t.Context(), 1, "MA-1", nil)
	var e *kakao.Error
	if !errors.As(err, &e) || e.Code != -813 {
		t.Errorf("err = %v", err)
	}
}
