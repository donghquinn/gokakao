package moment

import (
	"encoding/json"
	"net/http"
	"slices"
	"testing"

	"github.com/donghquinn/gokakao/internal/kakaotest"
)

// messageadopsRequest checks the auth, adAccountId and channel-profile-id
// headers, method and path of the last request.
func messageadopsRequest(t *testing.T, srv *kakaotest.Server, method, path string) kakaotest.Request {
	t.Helper()
	req := assertRequest(t, srv, "1111")
	if got := req.Header.Get("channel-profile-id"); got != "_ZQxd" {
		t.Errorf("channel-profile-id = %q", got)
	}
	if req.Method != method || req.Path != path {
		t.Errorf("request = %s %s, want %s %s", req.Method, req.Path, method, path)
	}
	return req
}

// messageadopsJSONEqual checks that got is the same JSON value as want.
func messageadopsJSONEqual(t *testing.T, got []byte, want string) {
	t.Helper()
	var g, w any
	if err := json.Unmarshal(got, &g); err != nil {
		t.Fatalf("body %s: %v", got, err)
	}
	if err := json.Unmarshal([]byte(want), &w); err != nil {
		t.Fatal(err)
	}
	gb, _ := json.Marshal(g)
	wb, _ := json.Marshal(w)
	if string(gb) != string(wb) {
		t.Errorf("body = %s\nwant   %s", gb, wb)
	}
}

const messageadopsBasicTextJSON = `{
  "messageAdId": "msg-ad-1402922220759433217",
  "name": "메시지광고_테스트_베이직_텍스트",
  "type": "BASIC_TEXT_MESSAGE",
  "description": "홍보문구",
  "items": [
    {
      "imageUrl": "https://partner.com/img/message/001.jpg"
    }
  ],
  "buttons": [
    {
      "title": "버튼 1 버튼명",
      "pcLandingUrl": "https://daum.net/1",
      "mobileLandingUrl": "https://daum.net/1"
    },
    {
      "title": "버튼 2 버튼명",
      "pcLandingUrl": "https://daum.net/1",
      "mobileLandingUrl": "https://daum.net/1"
    }
  ],
  "couponBook": {
    "title": "쿠폰 상세 설명",
    "pcLandingUrl": "https://daum.net",
    "mobileLandingUrl": "https://daum.net",
    "couponBookTitle": "쿠폰 타이틀",
    "couponBookTitleType": "UPGRADE"
  },
  "ageVerification": false,
  "adFlag": true,
  "shareFlag": true
}`

func messageadopsBasicText() Message {
	return Message{
		Type: MessageTypeBasicText, Description: "홍보문구", ShareFlag: true, AdFlag: true,
		Items: []Item{{ImageURL: "https://partner.com/img/message/001.jpg"}},
		Buttons: []Button{
			{Title: "버튼 1 버튼명", PCLandingURL: "https://daum.net/1", MobileLandingURL: "https://daum.net/1"},
			{Title: "버튼 2 버튼명", PCLandingURL: "https://daum.net/1", MobileLandingURL: "https://daum.net/1"},
		},
		CouponBook: &CouponBook{
			CouponBookTitleType: CouponBookUpgrade, CouponBookTitle: "쿠폰 타이틀", Title: "쿠폰 상세 설명",
			PCLandingURL: "https://daum.net", MobileLandingURL: "https://daum.net",
		},
	}
}

const messageadopsBasicTextRequest = `{
  "name": "메시지광고_테스트_베이직_텍스트",
  "ageVerification": false,
  "message": {
    "type": "BASIC_TEXT_MESSAGE",
    "description": "홍보문구",
    "shareFlag": true,
    "adFlag": true,
    "items": [{"imageUrl": "https://partner.com/img/message/001.jpg"}],
    "buttons": [
      {"title": "버튼 1 버튼명", "pcLandingUrl": "https://daum.net/1", "mobileLandingUrl": "https://daum.net/1"},
      {"title": "버튼 2 버튼명", "pcLandingUrl": "https://daum.net/1", "mobileLandingUrl": "https://daum.net/1"}
    ],
    "couponBook": {
      "couponBookTitleType": "UPGRADE",
      "couponBookTitle": "쿠폰 타이틀",
      "title": "쿠폰 상세 설명",
      "pcLandingUrl": "https://daum.net",
      "mobileLandingUrl": "https://daum.net"
    }
  }
}`

func messageadopsAssertBasicText(t *testing.T, m *ResponseMessage) {
	t.Helper()
	if m.Name != "메시지광고_테스트_베이직_텍스트" || m.Type != MessageTypeBasicText ||
		len(m.Items) != 1 || m.Items[0].ImageURL == "" || len(m.Buttons) != 2 || m.Buttons[1].Title != "버튼 2 버튼명" ||
		m.CouponBook == nil || m.CouponBook.CouponBookTitleType != CouponBookUpgrade || m.AgeVerification || !m.ShareFlag {
		t.Errorf("message = %+v", m)
	}
}

func TestMessageAdService_Create(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/message/v2/message-ads/message", 200, messageadopsBasicTextJSON)

	m, err := c.MessageAds.Create(t.Context(), 1111, "_ZQxd", NewMessageAd{Name: "메시지광고_테스트_베이직_텍스트", Message: messageadopsBasicText()})
	if err != nil {
		t.Fatal(err)
	}
	req := messageadopsRequest(t, srv, http.MethodPost, "/openapi/message/v2/message-ads/message")
	if req.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q", req.Header.Get("Content-Type"))
	}
	messageadopsJSONEqual(t, req.Body, messageadopsBasicTextRequest)
	messageadopsAssertBasicText(t, m)
	if m.MessageAdID != "msg-ad-1402922220759433217" || m.Description != "홍보문구" || !m.AdFlag {
		t.Errorf("message = %+v", m)
	}

	if _, err := c.MessageAds.Create(t.Context(), 1111, "_ZQxd", NewMessageAd{}); err == nil {
		t.Error("missing type: want error")
	}
	if _, err := c.MessageAds.Create(t.Context(), 1111, "", NewMessageAd{Message: messageadopsBasicText()}); err == nil {
		t.Error("missing channel: want error")
	}
	if _, err := c.MessageAds.Create(t.Context(), 0, "_ZQxd", NewMessageAd{Message: messageadopsBasicText()}); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestMessageAdService_UpdateMessage(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PATCH /openapi/message/v2/message-ads/msg-ad-1195179794616258561/message", 200, `{
  "messageAdId": "msg-ad-1195179794616258561",
  "name": "메시지광고_테스트_베이직_텍스트",
  "type": "BASIC_TEXT_MESSAGE",
  "description": "홍보문구 변경",
  "items": [
    {
      "imageUrl": "https://t1.kakaocdn.net/b2/creative/56493/3912ceec1584f2ec1ccf8fad73145254.jpg"
    }
  ],
  "buttons": [
    {
      "title": "버튼 1 버튼명",
      "pcLandingUrl": "https://daum.net/1",
      "mobileLandingUrl": "https://daum.net/1"
    },
    {
      "title": "버튼 2 버튼명",
      "pcLandingUrl": "https://daum.net/1",
      "mobileLandingUrl": "https://daum.net/1"
    }
  ],
  "couponBook": {
    "title": "쿠폰 상세 설명",
    "pcLandingUrl": "https://daum.net",
    "mobileLandingUrl": "https://daum.net",
    "couponBookTitle": "쿠폰 타이틀",
    "couponBookTitleType": "UPGRADE"
  },
  "ageVerification": false,
  "adFlag": false,
  "shareFlag": true
}`)

	msg := messageadopsBasicText()
	msg.Description, msg.AdFlag = "홍보문구 변경", false
	m, err := c.MessageAds.UpdateMessage(t.Context(), 1111, "_ZQxd", "msg-ad-1195179794616258561", MessageAdChanges{Name: "메시지광고_테스트_베이직_텍스트", Message: msg})
	if err != nil {
		t.Fatal(err)
	}
	req := messageadopsRequest(t, srv, http.MethodPatch, "/openapi/message/v2/message-ads/msg-ad-1195179794616258561/message")
	var body struct {
		Name            string  `json:"name"`
		AgeVerification *bool   `json:"ageVerification"`
		Message         Message `json:"message"`
	}
	if err := json.Unmarshal(req.Body, &body); err != nil || body.AgeVerification == nil || body.Message.Description != "홍보문구 변경" || body.Message.AdFlag {
		t.Errorf("body = %s, %v", req.Body, err)
	}
	messageadopsAssertBasicText(t, m)
	if m.Description != "홍보문구 변경" || m.AdFlag {
		t.Errorf("message = %+v", m)
	}

	// Changes keeps the current content.
	if _, err := c.MessageAds.UpdateMessage(t.Context(), 1111, "_ZQxd", m.MessageAdID, m.Changes()); err != nil {
		t.Fatal(err)
	}
	var got MessageAdChanges
	if err := json.Unmarshal(srv.LastRequest().Body, &got); err != nil || got.Name != m.Name || got.Message.Description != "홍보문구 변경" ||
		got.Message.Type != MessageTypeBasicText || len(got.Message.Buttons) != 2 || got.Message.CouponBook == nil {
		t.Errorf("Changes body = %s, %v", srv.LastRequest().Body, err)
	}

	if _, err := c.MessageAds.UpdateMessage(t.Context(), 1111, "_ZQxd", "m", MessageAdChanges{}); err == nil {
		t.Error("missing type: want error")
	}
	if _, err := c.MessageAds.UpdateMessage(t.Context(), 1111, "_ZQxd", "", MessageAdChanges{Message: msg}); err == nil {
		t.Error("missing message ad: want error")
	}
}

func TestMessageAdService_SendTest(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/message/v2/message-ads/msg-ad-1/send-test", 200, ``)

	if err := c.MessageAds.SendTest(t.Context(), 1111, "_ZQxd", "msg-ad-1", []string{"014-0042-4549"}); err != nil {
		t.Fatal(err)
	}
	req := messageadopsRequest(t, srv, http.MethodPost, "/openapi/message/v2/message-ads/msg-ad-1/send-test")
	if string(req.Body) != `{"phoneNumbers":["014-0042-4549"]}` {
		t.Errorf("body = %s", req.Body)
	}
	if err := c.MessageAds.SendTest(t.Context(), 1111, "_ZQxd", "msg-ad-1", nil); err == nil {
		t.Error("no phone numbers: want error")
	}
	if err := c.MessageAds.SendTest(t.Context(), 1111, "", "msg-ad-1", []string{"x"}); err == nil {
		t.Error("missing channel: want error")
	}
}

const messageadopsReservationJSON = `{
  "messageAdId": "msg-ad-1389380653606703104",
  "targeting": {
    "type": "DIRECT",
    "devices": [],
    "businessTypes": [],
    "interests": [],
    "demographics": {
      "genders": [],
      "ages": [],
      "locationType": "ALL",
      "depth1Locations": [],
      "depth2Locations": [],
      "depth3Locations": []
    },
    "friendPeriod": {
      "startDate": "2023-01-02",
      "endDate": "2025-06-27"
    },
    "customs": [
      {
        "type": "PLUSFRIENDGROUP",
        "referenceId": "3302",
        "inclusionType": "INCLUDE",
        "name": "123"
      },
      {
        "type": "SYNCAPP",
        "referenceId": "192594",
        "inclusionType": "EXCLUDE",
        "name": "테스트 채널 1"
      },
      {
        "type": "COHORT",
        "referenceId": "101407",
        "inclusionType": "INCLUDE",
        "name": "반응고객"
      },
      {
        "type": "PLUSFRIEND",
        "referenceId": "54746",
        "inclusionType": "EXCLUDE",
        "name": "테스트 채널2"
      },
      {
        "type": "CUSTOMERDATA",
        "referenceId": "101407",
        "inclusionType": "INCLUDE",
        "name": "고객데이터이름"
      }
    ]
  },
  "price": 20,
  "contractCount": 10,
  "totalBudget": 200,
  "totalBudgetWithVAT": 220.0,
  "date": "2025-07-01 08:07",
  "trackId": "1498114409013397398",
  "sendRate": 0,
  "pushAlarm": true,
  "immediateSend": true
}`

func messageadopsAssertReservation(t *testing.T, r *SendingReservation) {
	t.Helper()
	tg := r.Targeting
	if r.MessageAdID != "msg-ad-1389380653606703104" || tg.Type != MessageTargetingDirect || tg.Demographics == nil ||
		tg.Demographics.LocationType != LocationTypeAll || tg.FriendPeriod == nil || tg.FriendPeriod.EndDate != "2025-06-27" ||
		len(tg.Customs) != 5 || tg.Customs[0].Type != CustomTargetingPlusFriendGroup || tg.Customs[0].ReferenceID != "3302" ||
		tg.Customs[1].InclusionType != InclusionExclude || r.Price != 20 || r.ContractCount != 10 || r.TotalBudget != 200 ||
		r.TotalBudgetWithVAT != 220 || r.Date != "2025-07-01 08:07" || r.TrackID != "1498114409013397398" || r.SendRate != 0 ||
		!r.PushAlarm || !r.ImmediateSend {
		t.Errorf("reservation = %+v", r)
	}
}

func messageadopsReservationChanges() SendingReservationChanges {
	push := true
	return SendingReservationChanges{
		Targeting: MessageTargeting{
			Type:         MessageTargetingDirect,
			FriendPeriod: &FriendPeriod{StartDate: "2023-01-02", EndDate: "2025-06-27"},
			Customs: []CustomTargeting{
				{Type: CustomTargetingSyncApp, ReferenceID: "192594", InclusionType: InclusionExclude},
				{Type: CustomTargetingCustomerData, ReferenceID: "101407", InclusionType: InclusionInclude, Name: "고객데이터이름"},
			},
		},
		ImmediateSend: true, PushAlarm: &push, TrackID: "1498114409013397398",
	}
}

const messageadopsReservationRequest = `{
  "targeting": {
    "type": "DIRECT",
    "friendPeriod": {"startDate": "2023-01-02", "endDate": "2025-06-27"},
    "customs": [
      {"type": "SYNCAPP", "referenceId": "192594", "inclusionType": "EXCLUDE"},
      {"type": "CUSTOMERDATA", "referenceId": "101407", "inclusionType": "INCLUDE", "name": "고객데이터이름"}
    ]
  },
  "immediateSend": true,
  "pushAlarm": true,
  "trackId": "1498114409013397398"
}`

func TestMessageAdService_Reserve(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/message/v2/message-ads/msg-ad-1389380653606703104/sending-reservation", 200, messageadopsReservationJSON)

	r, err := c.MessageAds.Reserve(t.Context(), 1111, "_ZQxd", "msg-ad-1389380653606703104", messageadopsReservationChanges())
	if err != nil {
		t.Fatal(err)
	}
	req := messageadopsRequest(t, srv, http.MethodPost, "/openapi/message/v2/message-ads/msg-ad-1389380653606703104/sending-reservation")
	messageadopsJSONEqual(t, req.Body, messageadopsReservationRequest)
	messageadopsAssertReservation(t, r)

	// The non-target example response.
	srv.HandleJSON("POST /openapi/message/v2/message-ads/30/sending-reservation", 200, `{
  "messageAdId": "30",
  "targeting": {
    "type": "DIRECT",
    "devices": ["IOS", "ANDROID"],
    "demographics": {
      "genders": [],
      "ages": [],
      "locationType": "ALL",
      "depth1Locations": [],
      "depth2Locations": [],
      "depth3Locations": []
    }
  },
  "price": 15,
  "contractCount": 10,
  "totalBudget": 150,
  "totalBudgetWithVAT": 165.0,
  "date": "2023-10-11 17:00"
}`)
	r, err = c.MessageAds.Reserve(t.Context(), 1111, "_ZQxd", "30", SendingReservationChanges{
		Targeting: MessageTargeting{Type: MessageTargetingDirect, Devices: []DeviceType{DeviceIOS, DeviceAndroid}},
		Date:      "2023-10-11 17:00", SendRate: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := string(srv.LastRequest().Body); got != `{"targeting":{"type":"DIRECT","devices":["IOS","ANDROID"]},"date":"2023-10-11 17:00","sendRate":100}` {
		t.Errorf("body = %s", got)
	}
	if r.Price != 15 || r.TotalBudgetWithVAT != 165 || !slices.Equal(r.Targeting.Devices, []DeviceType{DeviceIOS, DeviceAndroid}) || r.PushAlarm {
		t.Errorf("reservation = %+v", r)
	}

	for _, rc := range []SendingReservationChanges{
		{Date: "2023-10-11 17:00"},
		{Targeting: MessageTargeting{Type: MessageTargetingWhole}},
		{Targeting: MessageTargeting{Type: MessageTargetingWhole}, ImmediateSend: true, Date: "2023-10-11 17:00"},
	} {
		if _, err := c.MessageAds.Reserve(t.Context(), 1111, "_ZQxd", "30", rc); err == nil {
			t.Errorf("Reserve(%+v): want error", rc)
		}
	}
	if _, err := c.MessageAds.Reserve(t.Context(), 1111, "_ZQxd", "", messageadopsReservationChanges()); err == nil {
		t.Error("missing message ad: want error")
	}
}

func TestMessageAdService_GetReservation(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/message/v2/message-ads/msg-ad-1389380653606703104/sending-reservation", 200, messageadopsReservationJSON)

	r, err := c.MessageAds.GetReservation(t.Context(), 1111, "_ZQxd", "msg-ad-1389380653606703104")
	if err != nil {
		t.Fatal(err)
	}
	messageadopsRequest(t, srv, http.MethodGet, "/openapi/message/v2/message-ads/msg-ad-1389380653606703104/sending-reservation")
	messageadopsAssertReservation(t, r)

	if _, err := c.MessageAds.GetReservation(t.Context(), 0, "_ZQxd", "m"); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestMessageAdService_UpdateReservation(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PATCH /openapi/message/v2/message-ads/msg-ad-1389380653606703104/sending-reservation", 200, messageadopsReservationJSON)

	r, err := c.MessageAds.UpdateReservation(t.Context(), 1111, "_ZQxd", "msg-ad-1389380653606703104", messageadopsReservationChanges())
	if err != nil {
		t.Fatal(err)
	}
	req := messageadopsRequest(t, srv, http.MethodPatch, "/openapi/message/v2/message-ads/msg-ad-1389380653606703104/sending-reservation")
	messageadopsJSONEqual(t, req.Body, messageadopsReservationRequest)
	messageadopsAssertReservation(t, r)

	// Changes keeps the current setup; immediate sends drop the date.
	if _, err := c.MessageAds.UpdateReservation(t.Context(), 1111, "_ZQxd", r.MessageAdID, r.Changes()); err != nil {
		t.Fatal(err)
	}
	var got SendingReservationChanges
	if err := json.Unmarshal(srv.LastRequest().Body, &got); err != nil || got.Date != "" || !got.ImmediateSend ||
		got.PushAlarm == nil || !*got.PushAlarm || got.TrackID != r.TrackID || len(got.Targeting.Customs) != 5 {
		t.Errorf("Changes body = %s, %v", srv.LastRequest().Body, err)
	}
	scheduled := &SendingReservation{Targeting: MessageTargeting{Type: MessageTargetingWhole, OptimizationTarget: &TargetingOptimization{}}, Date: "2025-07-01 08:07"}
	if ch := scheduled.Changes(); ch.Date != "2025-07-01 08:07" || ch.ImmediateSend || ch.PushAlarm == nil || *ch.PushAlarm || ch.Targeting.OptimizationTarget != nil {
		t.Errorf("Changes of scheduled = %+v", ch)
	}

	if _, err := c.MessageAds.UpdateReservation(t.Context(), 1111, "_ZQxd", "m", SendingReservationChanges{}); err == nil {
		t.Error("missing targeting: want error")
	}
	if _, err := c.MessageAds.UpdateReservation(t.Context(), 1111, "_ZQxd", "", messageadopsReservationChanges()); err == nil {
		t.Error("missing message ad: want error")
	}
}

func TestMessageAdService_Locations(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/message/v2/message-ads/sending-reservation/location", 200, `{
  "id": "A",
  "name": "강원특별자치도",
  "children": [
    {
      "id": "A7000",
      "name": "강릉시",
      "children": [
        {
          "id": "A7000A001",
          "name": "강남동"
        }
      ]
    }
  ]
}`)

	ls, err := c.MessageAds.Locations(t.Context(), 1111, "_ZQxd", nil)
	if err != nil {
		t.Fatal(err)
	}
	req := messageadopsRequest(t, srv, http.MethodGet, "/openapi/message/v2/message-ads/sending-reservation/location")
	if len(req.Query) != 0 {
		t.Errorf("nil options query = %v", req.Query)
	}
	if len(ls) != 1 || ls[0].ID != "A" || ls[0].Name != "강원특별자치도" || len(ls[0].Children) != 1 ||
		ls[0].Children[0].ID != "A7000" || len(ls[0].Children[0].Children) != 1 || ls[0].Children[0].Children[0].Name != "강남동" {
		t.Errorf("locations = %+v", ls)
	}

	c2, srv2 := newTestClient(t)
	srv2.HandleJSON("GET /openapi/message/v2/message-ads/sending-reservation/location", 200, `[
  {
    "id": "A7000A001",
    "name": "강원특별자치도 강릉시 강남동",
    "deprecated": true
  },
  {
    "id": "Q20000025",
    "name": "세종특별자치시 세종시 해밀동",
    "deprecated": false
  }
]`)
	ls, err = c2.MessageAds.Locations(t.Context(), 1111, "_ZQxd", &MessageLocationOptions{Codes: []string{"A7000A001", "Q20000025"}})
	if err != nil {
		t.Fatal(err)
	}
	if got := srv2.LastRequest().Query.Get("codes"); got != "A7000A001,Q20000025" {
		t.Errorf("codes = %q", got)
	}
	if len(ls) != 2 || !ls[0].Deprecated || ls[1].Deprecated || ls[1].ID != "Q20000025" {
		t.Errorf("locations = %+v", ls)
	}

	c3, srv3 := newTestClient(t)
	srv3.HandleJSON("GET /openapi/message/v2/message-ads/sending-reservation/location", 200, `null`)
	if ls, err := c3.MessageAds.Locations(t.Context(), 1111, "_ZQxd", &MessageLocationOptions{}); err != nil || ls != nil {
		t.Errorf("null = %v, %v", ls, err)
	}
	if _, err := c3.MessageAds.Locations(t.Context(), 1111, "", nil); err == nil {
		t.Error("missing channel: want error")
	}
}

func TestMessageAdService_TargetingPrice(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/message/v2/message-ads/msg-ad-1164046062704201728/sending-reservation/targeting-price", 200, `{
  "messageAdId": "msg-ad-1164046062704201728",
  "price": 15,
  "contractCount": 2,
  "totalBudget": 30,
  "totalBudgetWithVAT": 33.0
}`)

	p, err := c.MessageAds.TargetingPrice(t.Context(), 1111, "_ZQxd", "msg-ad-1164046062704201728", MessageTargeting{Type: MessageTargetingWhole})
	if err != nil {
		t.Fatal(err)
	}
	req := messageadopsRequest(t, srv, http.MethodPost, "/openapi/message/v2/message-ads/msg-ad-1164046062704201728/sending-reservation/targeting-price")
	if string(req.Body) != `{"targeting":{"type":"WHOLE"}}` {
		t.Errorf("body = %s", req.Body)
	}
	if p.MessageAdID != "msg-ad-1164046062704201728" || p.Price != 15 || p.ContractCount != 2 || p.TotalBudget != 30 || p.TotalBudgetWithVAT != 33 {
		t.Errorf("price = %+v", p)
	}
	if _, err := c.MessageAds.TargetingPrice(t.Context(), 1111, "_ZQxd", "m", MessageTargeting{}); err == nil {
		t.Error("missing targeting type: want error")
	}
	if _, err := c.MessageAds.TargetingPrice(t.Context(), 1111, "_ZQxd", "", MessageTargeting{Type: MessageTargetingWhole}); err == nil {
		t.Error("missing message ad: want error")
	}
}

func messageadopsOptimizedJSON(name string) string {
	return `{
  "messageAdId": "msg-ad-example",
  "name": "` + name + `",
  "messageAdType": "MULTI_BROADCAST",
  "ageVerification": false,
  "messages": [
    {
      "type": "BASIC_TEXT_MESSAGE",
      "description": "이번 주 새로운 상품을 확인해 보세요.",
      "adFlag": true,
      "shareFlag": false,
      "items": [],
      "buttons": []
    },
    {
      "type": "BASIC_TEXT_MESSAGE",
      "description": "지금 준비한 특별한 혜택을 만나 보세요.",
      "adFlag": true,
      "shareFlag": false,
      "items": [],
      "buttons": []
    }
  ]
}`
}

func messageadopsOptimizedMessages() []Message {
	return []Message{
		{Type: MessageTypeBasicText, Description: "이번 주 새로운 상품을 확인해 보세요.", AdFlag: true},
		{Type: MessageTypeBasicText, Description: "지금 준비한 특별한 혜택을 만나 보세요.", AdFlag: true},
	}
}

func messageadopsOptimizedRequest(name string) string {
	return `{
  "name": "` + name + `",
  "ageVerification": false,
  "messages": [
    {"type": "BASIC_TEXT_MESSAGE", "description": "이번 주 새로운 상품을 확인해 보세요.", "adFlag": true, "shareFlag": false},
    {"type": "BASIC_TEXT_MESSAGE", "description": "지금 준비한 특별한 혜택을 만나 보세요.", "adFlag": true, "shareFlag": false}
  ]
}`
}

func messageadopsAssertOptimized(t *testing.T, m *ResponseMultiMessage, name string) {
	t.Helper()
	if m.MessageAdID != "msg-ad-example" || m.Name != name || m.MessageAdType != "MULTI_BROADCAST" || m.AgeVerification ||
		len(m.Messages) != 2 || m.Messages[1].Description != "지금 준비한 특별한 혜택을 만나 보세요." || !m.Messages[0].AdFlag {
		t.Errorf("message = %+v", m)
	}
}

func TestMessageAdService_CreateOptimized(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/message/v2/message-ads/optimization/message", 200, messageadopsOptimizedJSON("최적화 메시지"))

	m, err := c.MessageAds.CreateOptimized(t.Context(), 1111, "_ZQxd", NewOptimizedMessageAd{Name: "최적화 메시지", Messages: messageadopsOptimizedMessages()})
	if err != nil {
		t.Fatal(err)
	}
	req := messageadopsRequest(t, srv, http.MethodPost, "/openapi/message/v2/message-ads/optimization/message")
	messageadopsJSONEqual(t, req.Body, messageadopsOptimizedRequest("최적화 메시지"))
	messageadopsAssertOptimized(t, m, "최적화 메시지")

	for _, msgs := range [][]Message{nil, messageadopsOptimizedMessages()[:1], make([]Message, 6), {{Type: MessageTypeWide}, {}}} {
		if _, err := c.MessageAds.CreateOptimized(t.Context(), 1111, "_ZQxd", NewOptimizedMessageAd{Messages: msgs}); err == nil {
			t.Errorf("messages %+v: want error", msgs)
		}
	}
	if _, err := c.MessageAds.CreateOptimized(t.Context(), 0, "_ZQxd", NewOptimizedMessageAd{Messages: messageadopsOptimizedMessages()}); err == nil {
		t.Error("missing ad account: want error")
	}

	// A regular message in the response is rejected.
	c2, srv2 := newTestClient(t)
	srv2.HandleJSON("POST /openapi/message/v2/message-ads/optimization/message", 200, messageadopsBasicTextJSON)
	if _, err := c2.MessageAds.CreateOptimized(t.Context(), 1111, "_ZQxd", NewOptimizedMessageAd{Messages: messageadopsOptimizedMessages()}); err == nil {
		t.Error("regular response: want error")
	}
	srv2.HandleJSON("POST /openapi/message/v2/message-ads/message", 200, messageadopsOptimizedJSON("x"))
	if _, err := c2.MessageAds.Create(t.Context(), 1111, "_ZQxd", NewMessageAd{Message: messageadopsBasicText()}); err == nil {
		t.Error("optimized response: want error")
	}
}

func TestMessageAdService_UpdateOptimized(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PATCH /openapi/message/v2/message-ads/optimization/msg-ad-example/message", 200, messageadopsOptimizedJSON("최적화 메시지 수정"))

	m, err := c.MessageAds.UpdateOptimized(t.Context(), 1111, "_ZQxd", "msg-ad-example", OptimizedMessageAdChanges{Name: "최적화 메시지 수정", Messages: messageadopsOptimizedMessages()})
	if err != nil {
		t.Fatal(err)
	}
	req := messageadopsRequest(t, srv, http.MethodPatch, "/openapi/message/v2/message-ads/optimization/msg-ad-example/message")
	messageadopsJSONEqual(t, req.Body, messageadopsOptimizedRequest("최적화 메시지 수정"))
	messageadopsAssertOptimized(t, m, "최적화 메시지 수정")

	// Changes keeps the current content.
	if _, err := c.MessageAds.UpdateOptimized(t.Context(), 1111, "_ZQxd", m.MessageAdID, m.Changes()); err != nil {
		t.Fatal(err)
	}
	messageadopsJSONEqual(t, srv.LastRequest().Body, messageadopsOptimizedRequest("최적화 메시지 수정"))

	if _, err := c.MessageAds.UpdateOptimized(t.Context(), 1111, "_ZQxd", "m", OptimizedMessageAdChanges{}); err == nil {
		t.Error("no messages: want error")
	}
	if _, err := c.MessageAds.UpdateOptimized(t.Context(), 1111, "_ZQxd", "", OptimizedMessageAdChanges{Messages: messageadopsOptimizedMessages()}); err == nil {
		t.Error("missing message ad: want error")
	}
}

func TestMessageAdService_SendOptimizedTest(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/message/v2/message-ads/optimization/msg-ad-example/send-test", 200, ``)

	if err := c.MessageAds.SendOptimizedTest(t.Context(), 1111, "_ZQxd", "msg-ad-example", 1, []string{"010-0000-0000"}); err != nil {
		t.Fatal(err)
	}
	req := messageadopsRequest(t, srv, http.MethodPost, "/openapi/message/v2/message-ads/optimization/msg-ad-example/send-test")
	if string(req.Body) != `{"messageIndex":1,"phoneNumbers":["010-0000-0000"]}` {
		t.Errorf("body = %s", req.Body)
	}
	if err := c.MessageAds.SendOptimizedTest(t.Context(), 1111, "_ZQxd", "m", -1, []string{"x"}); err == nil {
		t.Error("negative index: want error")
	}
	if err := c.MessageAds.SendOptimizedTest(t.Context(), 1111, "_ZQxd", "m", 0, []string{""}); err == nil {
		t.Error("empty phone number: want error")
	}
	if err := c.MessageAds.SendOptimizedTest(t.Context(), 1111, "_ZQxd", "", 0, []string{"x"}); err == nil {
		t.Error("missing message ad: want error")
	}
}

const messageadopsOptimizedReservationJSON = `{
  "messageAdId": "msg-ad-example",
  "targeting": {
    "type": "OPTIMIZATION_DIRECT",
    "optimizationTarget": {
      "type": "RATIO",
      "sendRatios": [60, 40]
    },
    "devices": ["ANDROID"],
    "businessTypes": [],
    "interests": [],
    "customs": []
  },
  "price": 20,
  "contractCount": 10000,
  "totalBudget": 200000,
  "totalBudgetWithVAT": 220000.0,
  "date": "2026-10-16 11:00",
  "sendRate": 0,
  "pushAlarm": true,
  "immediateSend": false
}`

func TestMessageAdService_ReserveOptimized(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/message/v2/message-ads/optimization/msg-ad-example/sending-reservation", 200, `{
  "messageAdId": "msg-ad-example",
  "targeting": {
    "type": "OPTIMIZATION_WHOLE",
    "optimizationTarget": {
      "type": "AI_CUSTOM",
      "sendRatios": []
    },
    "devices": [],
    "businessTypes": [],
    "interests": [],
    "customs": []
  },
  "price": 20,
  "contractCount": 10000,
  "totalBudget": 200000,
  "totalBudgetWithVAT": 220000.0,
  "date": "2026-10-16 10:00",
  "sendRate": 0,
  "pushAlarm": true,
  "immediateSend": false
}`)

	push := true
	r, err := c.MessageAds.ReserveOptimized(t.Context(), 1111, "_ZQxd", "msg-ad-example", OptimizedSendingReservationChanges{
		Targeting: OptimizationTargeting{Type: MessageTargetingOptimizationWhole, OptimizationTarget: OptimizationTarget{Type: OptimizationAICustom}},
		Date:      "2026-10-16 10:00", PushAlarm: &push,
	})
	if err != nil {
		t.Fatal(err)
	}
	req := messageadopsRequest(t, srv, http.MethodPost, "/openapi/message/v2/message-ads/optimization/msg-ad-example/sending-reservation")
	if string(req.Body) != `{"targeting":{"type":"OPTIMIZATION_WHOLE","optimizationTarget":{"type":"AI_CUSTOM"}},"date":"2026-10-16 10:00","pushAlarm":true}` {
		t.Errorf("body = %s", req.Body)
	}
	if r.Targeting.Type != MessageTargetingOptimizationWhole || r.Targeting.OptimizationTarget == nil ||
		r.Targeting.OptimizationTarget.Type != OptimizationAICustom || r.ContractCount != 10000 || r.TotalBudgetWithVAT != 220000 ||
		r.Date != "2026-10-16 10:00" || r.ImmediateSend {
		t.Errorf("reservation = %+v", r)
	}

	// Immediate send.
	if _, err := c.MessageAds.ReserveOptimized(t.Context(), 1111, "_ZQxd", "msg-ad-example", OptimizedSendingReservationChanges{
		Targeting:     OptimizationTargeting{Type: MessageTargetingOptimizationWhole, OptimizationTarget: OptimizationTarget{Type: OptimizationAICustom}},
		ImmediateSend: true,
	}); err != nil {
		t.Fatal(err)
	}
	if got := string(srv.LastRequest().Body); got != `{"targeting":{"type":"OPTIMIZATION_WHOLE","optimizationTarget":{"type":"AI_CUSTOM"}},"immediateSend":true}` {
		t.Errorf("immediate body = %s", got)
	}

	ok := OptimizationTargeting{Type: MessageTargetingOptimizationWhole, OptimizationTarget: OptimizationTarget{Type: OptimizationAICustom}}
	for _, rc := range []OptimizedSendingReservationChanges{
		{Targeting: OptimizationTargeting{Type: MessageTargetingWhole, OptimizationTarget: OptimizationTarget{Type: OptimizationAICustom}}, ImmediateSend: true},
		{Targeting: OptimizationTargeting{Type: MessageTargetingOptimizationDirect}, ImmediateSend: true},
		{Targeting: ok},
		{Targeting: ok, ImmediateSend: true, Date: "2026-10-16 10:00"},
	} {
		if _, err := c.MessageAds.ReserveOptimized(t.Context(), 1111, "_ZQxd", "msg-ad-example", rc); err == nil {
			t.Errorf("ReserveOptimized(%+v): want error", rc)
		}
	}
	if _, err := c.MessageAds.ReserveOptimized(t.Context(), 1111, "_ZQxd", "", OptimizedSendingReservationChanges{Targeting: ok, ImmediateSend: true}); err == nil {
		t.Error("missing message ad: want error")
	}
}

func TestMessageAdService_UpdateOptimizedReservation(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PATCH /openapi/message/v2/message-ads/optimization/msg-ad-example/sending-reservation", 200, messageadopsOptimizedReservationJSON)

	push := true
	r, err := c.MessageAds.UpdateOptimizedReservation(t.Context(), 1111, "_ZQxd", "msg-ad-example", OptimizedSendingReservationChanges{
		Targeting: OptimizationTargeting{
			Type:               MessageTargetingOptimizationDirect,
			OptimizationTarget: OptimizationTarget{Type: OptimizationRatio, SendRatios: []int{60, 40}},
			Devices:            []DeviceType{DeviceAndroid},
		},
		Date: "2026-10-16 11:00", PushAlarm: &push,
	})
	if err != nil {
		t.Fatal(err)
	}
	req := messageadopsRequest(t, srv, http.MethodPatch, "/openapi/message/v2/message-ads/optimization/msg-ad-example/sending-reservation")
	const want = `{"targeting":{"type":"OPTIMIZATION_DIRECT","optimizationTarget":{"type":"RATIO","sendRatios":[60,40]},"devices":["ANDROID"]},"date":"2026-10-16 11:00","pushAlarm":true}`
	if string(req.Body) != want {
		t.Errorf("body = %s", req.Body)
	}
	if r.Targeting.Type != MessageTargetingOptimizationDirect || !slices.Equal(r.Targeting.OptimizationTarget.SendRatios, []int{60, 40}) ||
		!slices.Equal(r.Targeting.Devices, []DeviceType{DeviceAndroid}) || r.Date != "2026-10-16 11:00" {
		t.Errorf("reservation = %+v", r)
	}

	// OptimizedChanges keeps the current setup.
	if _, err := c.MessageAds.UpdateOptimizedReservation(t.Context(), 1111, "_ZQxd", r.MessageAdID, r.OptimizedChanges()); err != nil {
		t.Fatal(err)
	}
	if got := string(srv.LastRequest().Body); got != want {
		t.Errorf("OptimizedChanges body = %s", got)
	}
	immediate := (&SendingReservation{ImmediateSend: true, Date: "2026-10-16 11:00"}).OptimizedChanges()
	if immediate.Date != "" || !immediate.ImmediateSend || immediate.Targeting.OptimizationTarget.Type != "" {
		t.Errorf("OptimizedChanges of immediate = %+v", immediate)
	}
}

func TestMessageAdService_OptimizedTargetingPrice(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/message/v2/message-ads/optimization/msg-ad-example/sending-reservation/targeting-price", 200, `{
  "messageAdId": "msg-ad-example",
  "price": 20,
  "contractCount": 10000,
  "channelFriendCount": 15000,
  "totalBudget": 200000,
  "totalBudgetWithVAT": 220000.0
}`)

	p, err := c.MessageAds.OptimizedTargetingPrice(t.Context(), 1111, "_ZQxd", "msg-ad-example", OptimizationTargeting{
		Type:               MessageTargetingOptimizationDirect,
		OptimizationTarget: OptimizationTarget{Type: OptimizationRatio, SendRatios: []int{60, 40}},
		Devices:            []DeviceType{DeviceAndroid},
	})
	if err != nil {
		t.Fatal(err)
	}
	req := messageadopsRequest(t, srv, http.MethodPost, "/openapi/message/v2/message-ads/optimization/msg-ad-example/sending-reservation/targeting-price")
	if string(req.Body) != `{"targeting":{"type":"OPTIMIZATION_DIRECT","optimizationTarget":{"type":"RATIO","sendRatios":[60,40]},"devices":["ANDROID"]}}` {
		t.Errorf("body = %s", req.Body)
	}
	if p.MessageAdID != "msg-ad-example" || p.Price != 20 || p.ContractCount != 10000 || p.ChannelFriendCount != 15000 ||
		p.TotalBudget != 200000 || p.TotalBudgetWithVAT != 220000 {
		t.Errorf("price = %+v", p)
	}
	if _, err := c.MessageAds.OptimizedTargetingPrice(t.Context(), 1111, "_ZQxd", "msg-ad-example", OptimizationTargeting{Type: MessageTargetingDirect}); err == nil {
		t.Error("non-optimization targeting: want error")
	}
	if _, err := c.MessageAds.OptimizedTargetingPrice(t.Context(), 1111, "_ZQxd", "", OptimizationTargeting{}); err == nil {
		t.Error("missing message ad: want error")
	}
}

func TestMessageAdService_Finish(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PUT /openapi/message/v2/message-ads/msg-ad-1/finish", 200, ``)

	if err := c.MessageAds.Finish(t.Context(), 1111, "_ZQxd", "msg-ad-1"); err != nil {
		t.Fatal(err)
	}
	messageadopsRequest(t, srv, http.MethodPut, "/openapi/message/v2/message-ads/msg-ad-1/finish")
	if err := c.MessageAds.Finish(t.Context(), 1111, "_ZQxd", ""); err == nil {
		t.Error("missing message ad: want error")
	}
}

func TestMessageAdMessage_UnmarshalJSON(t *testing.T) {
	t.Parallel()
	var m MessageAdMessage
	if err := json.Unmarshal([]byte(`null`), &m); err != nil || m.Message != nil || m.Multi != nil {
		t.Errorf("null = %+v, %v", m, err)
	}
	// Item landing as an object or an array.
	for _, in := range []string{
		`{"type":"WIDE_MESSAGE","items":[{"imageUrl":"a","landing":{"mobileLandingUrl":"https://m"}}]}`,
		`{"type":"WIDE_MESSAGE","items":[{"imageUrl":"a","landing":[{"mobileLandingUrl":"https://m"}]}]}`,
	} {
		if err := json.Unmarshal([]byte(in), &m); err != nil || m.Message == nil || len(m.Message.Items) != 1 ||
			len(m.Message.Items[0].Landing) != 1 || m.Message.Items[0].Landing[0].MobileLandingURL != "https://m" {
			t.Errorf("%s = %+v, %v", in, m.Message, err)
		}
	}
	// messageAdType alone selects Multi.
	if err := json.Unmarshal([]byte(`{"messageAdType":"MULTI_BROADCAST","name":"n"}`), &m); err != nil || m.Multi == nil || m.Message != nil || m.Multi.Name != "n" {
		t.Errorf("multi = %+v, %v", m, err)
	}
	if err := json.Unmarshal([]byte(`{"messages":[{"type":"WIDE_MESSAGE","items":[{"landing":{"pcLandingUrl":"p"}}]}]}`), &m); err != nil ||
		m.Multi == nil || m.Multi.Messages[0].Items[0].Landing[0].PCLandingURL != "p" {
		t.Errorf("multi landing = %+v, %v", m, err)
	}
	for _, bad := range []string{`[]`, `{"items":"x"}`, `{"messages":"x"}`, `{"messages":[{"items":"x"}]}`} {
		if err := json.Unmarshal([]byte(bad), &m); err == nil {
			t.Errorf("%s: want error", bad)
		}
	}
}
