package moment

import (
	"encoding/json"
	"net/http"
	"slices"
	"testing"
)

func TestMessageAdService_List(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/message/v2/message-ads", 200, `{
  "content": [
    {
      "messageAdId": "msg-ad-1196357486962585601",
      "messageAdType": "BROADCAST",
      "name": "메시지광고_테스트_와이드_이미지",
      "status": "READY",
      "sendingReservation": {
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
          "customs": []
        },
        "date": "2024-01-15 16:37"
      },
      "metrics": null,
      "createdDate": "2024-01-15 16:37:32",
      "lastModifiedDate": "2024-01-15 16:37:32",
      "message": {
        "type": "WIDE_MESSAGE",
        "mainTitle": "홍보문구",
        "mainImageUrl": "https://t1.kakaocdn.net/b2/creative/56493/72750c9573241f3409d6f8e44cd66ac8.jpg"
      }
    }
  ],
  "pageable": {
    "pageNumber": 0,
    "pageSize": 10,
    "sort": {
      "empty": false,
      "sorted": true,
      "unsorted": false
    },
    "offset": 0,
    "paged": true,
    "unpaged": false
  },
  "last": true,
  "totalElements": 1,
  "totalPages": 1,
  "first": true,
  "size": 10,
  "number": 0,
  "sort": {
    "empty": false,
    "sorted": true,
    "unsorted": false
  },
  "numberOfElements": 1,
  "empty": false
}`)

	p, err := c.MessageAds.List(t.Context(), 1111, "_ZQxd", &MessageAdListOptions{
		Page: 2, Limit: 10, From: "2024-01-01", To: "2024-01-31",
		MessageAdIDs: []string{"msg-ad-1"}, Statuses: []MessageAdStatus{MessageAdReady}, MessageTypes: []MessageType{MessageTypeWide},
		Start: "2024-01-01", End: "2024-01-15", SearchKeyword: "메시지",
	})
	if err != nil {
		t.Fatal(err)
	}
	req := messageadopsRequest(t, srv, http.MethodPost, "/openapi/message/v2/message-ads")
	if q := req.Query; q.Get("page") != "2" || q.Get("limit") != "10" || q.Get("from") != "2024-01-01" || q.Get("to") != "2024-01-31" {
		t.Errorf("query = %v", q)
	}
	messageadopsJSONEqual(t, req.Body, `{"filter":{"messageAdType":"BROADCAST","messageAdIds":["msg-ad-1"],"statuses":["READY"],
		"messageTypes":["WIDE_MESSAGE"],"start":"2024-01-01","end":"2024-01-15","searchKeyword":"메시지"}}`)
	if p.TotalElements != 1 || p.TotalPages != 1 || p.Size != 10 || len(p.Content) != 1 {
		t.Fatalf("page = %+v", p)
	}
	a := p.Content[0]
	if a.MessageAdID != "msg-ad-1196357486962585601" || a.MessageAdType != MessageAdTypeBroadcast || a.Name != "메시지광고_테스트_와이드_이미지" ||
		a.Status != MessageAdReady || a.SendingReservation == nil || a.SendingReservation.Date != "2024-01-15 16:37" ||
		a.SendingReservation.Targeting.Type != MessageTargetingDirect || a.Metrics != nil || a.CreatedDate.Minute() != 37 ||
		a.LastModifiedDate.Second() != 32 || a.Message == nil || a.Message.Type != MessageTypeWide || a.Message.MainTitle != "홍보문구" ||
		a.Message.MainImageURL == "" || a.Messages != nil {
		t.Errorf("summary = %+v", a)
	}

	if _, err := c.MessageAds.List(t.Context(), 1111, "_ZQxd", nil); err != nil {
		t.Fatal(err)
	}
	if req := srv.LastRequest(); len(req.Query) != 0 || string(req.Body) != `{"filter":{"messageAdType":"BROADCAST"}}` {
		t.Errorf("nil options = %v %s", req.Query, req.Body)
	}
	if _, err := c.MessageAds.List(t.Context(), 1111, "_ZQxd", &MessageAdListOptions{Start: "2024-01-01"}); err == nil {
		t.Error("start without end: want error")
	}
	if _, err := c.MessageAds.List(t.Context(), 1111, "", nil); err == nil {
		t.Error("missing channel: want error")
	}
}

func TestMessageAdService_Get(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/message/v2/message-ads/msg-ad-1195183093423038464", 200, `{
  "messageAdId": "msg-ad-1195183093423038464",
  "messageAdType": "BROADCAST",
  "name": "메시지광고_테스트_와이드_이미지",
  "message": {
    "type": "WIDE_MESSAGE",
    "items": [
      {
        "imageUrl": "https://t1.kakaocdn.net/b2/creative/56493/72750c9573241f3409d6f8e44cd66ac8.jpg",
        "landing": {
          "pcLandingUrl": "https://daum.net",
          "mobileLandingUrl": "https://daum.net"
        }
      }
    ],
    "buttons": [
      {
        "title": "버튼 1 버튼명",
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
    "shareFlag": true,
    "description": "홍보문구"
  },
  "sendingReservation": {},
  "status": "DRAFT",
  "ageVerification": false,
  "createdDate": "2024-01-12 10:50:54",
  "lastModifiedDate": "2024-01-12 10:50:54"
}`)

	d, err := c.MessageAds.Get(t.Context(), 1111, "_ZQxd", "msg-ad-1195183093423038464")
	if err != nil {
		t.Fatal(err)
	}
	messageadopsRequest(t, srv, http.MethodGet, "/openapi/message/v2/message-ads/msg-ad-1195183093423038464")
	m := d.Message.Message
	if d.MessageAdID != "msg-ad-1195183093423038464" || d.MessageAdType != MessageAdTypeBroadcast || d.Status != MessageAdDraft ||
		d.AgeVerification || d.CreatedDate.Day() != 12 || d.SendingReservation == nil || d.SendingReservation.Date != "" ||
		d.Message.Multi != nil || m == nil || m.Type != MessageTypeWide || m.Description != "홍보문구" || len(m.Items) != 1 ||
		len(m.Items[0].Landing) != 1 || m.Items[0].Landing[0].PCLandingURL != "https://daum.net" || len(m.Buttons) != 1 ||
		m.CouponBook == nil || !m.AdFlag || !m.ShareFlag {
		t.Errorf("detail = %+v, message = %+v", d, m)
	}

	// An optimized message ad.
	srv.HandleJSON("GET /openapi/message/v2/message-ads/msg-ad-example", 200, `{
  "messageAdId": "msg-ad-example",
  "messageAdType": "MULTI_BROADCAST",
  "name": "최적화 메시지",
  "message": {
    "messageAdType": "MULTI_BROADCAST",
    "ageVerification": false,
    "messages": [
      {"type": "BASIC_TEXT_MESSAGE", "description": "a", "adFlag": true, "shareFlag": false, "items": [], "buttons": []},
      {"type": "BASIC_TEXT_MESSAGE", "description": "b", "adFlag": true, "shareFlag": false, "items": [], "buttons": []}
    ]
  },
  "sendingReservation": {
    "targeting": {"type": "OPTIMIZATION_WHOLE", "optimizationTarget": {"type": "AI_CUSTOM", "sendRatios": []}},
    "date": "2026-10-16 10:00",
    "startedDate": "2026-10-16 10:00"
  },
  "status": "SENDING",
  "ageVerification": false
}`)
	d, err = c.MessageAds.Get(t.Context(), 1111, "_ZQxd", "msg-ad-example")
	if err != nil {
		t.Fatal(err)
	}
	if d.MessageAdType != MessageAdTypeMulti || d.Message.Message != nil || d.Message.Multi == nil || len(d.Message.Multi.Messages) != 2 ||
		d.Message.Multi.Messages[1].Description != "b" || d.SendingReservation.StartedDate != "2026-10-16 10:00" ||
		d.SendingReservation.Targeting.OptimizationTarget.Type != OptimizationAICustom || d.Status != MessageAdSending {
		t.Errorf("optimized detail = %+v", d)
	}

	if _, err := c.MessageAds.Get(t.Context(), 1111, "_ZQxd", ""); err == nil {
		t.Error("missing message ad: want error")
	}
}

func TestMessageAdService_Delete(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("DELETE /openapi/message/v2/message-ads/msg-ad-1", 200, ``)

	if err := c.MessageAds.Delete(t.Context(), 1111, "_ZQxd", "msg-ad-1"); err != nil {
		t.Fatal(err)
	}
	messageadopsRequest(t, srv, http.MethodDelete, "/openapi/message/v2/message-ads/msg-ad-1")
	if err := c.MessageAds.Delete(t.Context(), 0, "_ZQxd", "msg-ad-1"); err == nil {
		t.Error("missing ad account: want error")
	}
}

const messageadmgmtMessageJSON = `{
  "messageAdId": "msg-ad-1195179794616258561",
  "name": "메시지광고_테스트_베이직_텍스트",
  "type": "BASIC_TEXT_MESSAGE",
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
  "shareFlag": true,
  "description": "홍보문구"
}`

func TestMessageAdService_GetMessage(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/message/v2/message-ads/msg-ad-1195179794616258561/message", 200, messageadmgmtMessageJSON)

	m, err := c.MessageAds.GetMessage(t.Context(), 1111, "_ZQxd", "msg-ad-1195179794616258561")
	if err != nil {
		t.Fatal(err)
	}
	messageadopsRequest(t, srv, http.MethodGet, "/openapi/message/v2/message-ads/msg-ad-1195179794616258561/message")
	if m.Multi != nil || m.Message == nil || m.Message.MessageAdID != "msg-ad-1195179794616258561" || m.Message.Description != "홍보문구" {
		t.Fatalf("message = %+v", m)
	}
	messageadopsAssertBasicText(t, m.Message)

	// An optimized message ad.
	srv.HandleJSON("GET /openapi/message/v2/message-ads/msg-ad-example/message", 200, messageadopsOptimizedJSON("최적화 메시지"))
	m, err = c.MessageAds.GetMessage(t.Context(), 1111, "_ZQxd", "msg-ad-example")
	if err != nil {
		t.Fatal(err)
	}
	if m.Message != nil || m.Multi == nil {
		t.Fatalf("optimized message = %+v", m)
	}
	messageadopsAssertOptimized(t, m.Multi, "최적화 메시지")

	if _, err := c.MessageAds.GetMessage(t.Context(), 1111, "", "m"); err == nil {
		t.Error("missing channel: want error")
	}
}

func TestMessageAdService_Copy(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/message/v2/message-ads/msg-ad-1195179794616258561/copy", 200, messageadmgmtMessageJSON)

	m, err := c.MessageAds.Copy(t.Context(), 1111, "_ZQxd", "msg-ad-1195179794616258561")
	if err != nil {
		t.Fatal(err)
	}
	req := messageadopsRequest(t, srv, http.MethodPost, "/openapi/message/v2/message-ads/msg-ad-1195179794616258561/copy")
	if len(req.Body) != 0 {
		t.Errorf("body = %s", req.Body)
	}
	if m.Message == nil {
		t.Fatalf("message = %+v", m)
	}
	messageadopsAssertBasicText(t, m.Message)

	if _, err := c.MessageAds.Copy(t.Context(), 1111, "_ZQxd", ""); err == nil {
		t.Error("missing message ad: want error")
	}
}

func TestMessageAdService_Report(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/message/v2/message-ads/reports", 200, `{
  "code": 200,
  "message": "Success",
  "data": [
    {
      "start": "2023-10-10",
      "end": "2023-10-10",
      "dimensions": {
        "message_ad_id": "msg-ad-1"
      },
      "metrics": {
        "msg_click": 0,
        "msg_send": 0,
        "cost": 0.0,
        "msg_send_fail": 0,
        "msg_open": 1
      }
    }
  ]
}`)

	rows, err := c.MessageAds.Report(t.Context(), 1111, "_ZQxd", MessageAdReportQuery{
		Dimension: MessageAdDimensionMessageAd, DatePreset: DatePresetLast30Day,
		MessageAdIDs:  []string{"msg-ad-1164368618878062593"},
		MetricsGroups: []MetricsGroup{MetricsMessage, MetricsMessageAddition, MetricsMessageClick, MetricsPixelSDKConversion},
	})
	if err != nil {
		t.Fatal(err)
	}
	req := messageadopsRequest(t, srv, http.MethodPost, "/openapi/message/v2/message-ads/reports")
	messageadopsJSONEqual(t, req.Body, `{"dimension":"MESSAGE_AD","datePreset":"LAST_30DAY","messageAdIds":["msg-ad-1164368618878062593"],
		"metricsGroup":["MESSAGE","MESSAGE_ADDITION","MESSAGE_CLICK","PIXEL_SDK_CONVERSION"]}`)
	if len(rows) != 1 || rows[0].Start != "2023-10-10" || rows[0].End != "2023-10-10" || rows[0].Dimensions["message_ad_id"] != "msg-ad-1" ||
		rows[0].Metrics["msg_open"] != 1 || rows[0].Metrics["cost"] != 0 || len(rows[0].Metrics) != 5 {
		t.Errorf("rows = %+v", rows)
	}

	if _, err := c.MessageAds.Report(t.Context(), 1111, "_ZQxd", MessageAdReportQuery{
		Dimension: MessageAdDimensionHour, Start: "2023-10-01", End: "2023-10-10", MetricsGroups: []MetricsGroup{MetricsMessageCarouselView},
	}); err != nil {
		t.Fatal(err)
	}
	if got := string(srv.LastRequest().Body); got != `{"start":"2023-10-01","end":"2023-10-10","dimension":"HOUR","metricsGroup":["MESSAGE_CAROUSEL_VIEW"]}` {
		t.Errorf("body = %s", got)
	}

	ids := make([]string, 101)
	for _, q := range []MessageAdReportQuery{
		{MetricsGroups: []MetricsGroup{MetricsMessage}},
		{Dimension: MessageAdDimensionDeviceType},
		{Dimension: MessageAdDimensionMessageType, MetricsGroups: []MetricsGroup{MetricsMessage}, End: "2023-10-10"},
		{Dimension: MessageAdDimensionMessageType, MetricsGroups: []MetricsGroup{MetricsMessage}, MessageAdIDs: ids},
	} {
		if _, err := c.MessageAds.Report(t.Context(), 1111, "_ZQxd", q); err == nil {
			t.Errorf("Report(%+v): want error", q)
		}
	}
	if _, err := c.MessageAds.Report(t.Context(), 0, "_ZQxd", MessageAdReportQuery{}); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestMessageAdService_TalkChannelVideos(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/message/v2/message-ads/talk-channel-videos", 200, `{
  "items": [
    {
      "url": "https://pf.kakao.com/_Xxju/vod/svbjfjfy6nip184ci0zq1maf3",
      "title": "테스트 영상",
      "talkChannelProfileId": "_Xxju",
      "vid": "svbjfjfy6nip184ci0zq1maf3",
      "mediaId": 408,
      "status": "PUBLIC",
      "duration": 15,
      "width": 1920,
      "height": 1080,
      "playCount": 0,
      "autoThumbnail": {
        "url": "https://thumb.kakaocdn.net/dna/kamp-sbox/.../thumb.jpg",
        "width": 0,
        "height": 0,
        "fileSize": 0
      }
    }
  ],
  "totalElements": 1
}`)

	p, err := c.MessageAds.TalkChannelVideos(t.Context(), &PageOptions{Page: 1, Size: 20})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "")
	if req.Method != http.MethodGet || req.Path != "/openapi/message/v2/message-ads/talk-channel-videos" ||
		req.Query.Get("page") != "1" || req.Query.Get("size") != "20" || req.Header.Get("channel-profile-id") != "" {
		t.Errorf("request = %s %s?%v", req.Method, req.Path, req.Query)
	}
	if p.TotalElements != 1 || len(p.Items) != 1 {
		t.Fatalf("page = %+v", p)
	}
	v := p.Items[0]
	if v.Title != "테스트 영상" || v.TalkChannelProfileID != "_Xxju" || v.VID != "svbjfjfy6nip184ci0zq1maf3" || v.MediaID != 408 ||
		v.Status != "PUBLIC" || v.Duration != 15 || v.Width != 1920 || v.Height != 1080 || v.AutoThumbnail == nil ||
		v.AutoThumbnail.URL == "" || v.UploadThumbnail != nil {
		t.Errorf("video = %+v", v)
	}

	if _, err := c.MessageAds.TalkChannelVideos(t.Context(), nil); err != nil {
		t.Fatal(err)
	}
	if q := srv.LastRequest().Query; len(q) != 0 {
		t.Errorf("nil options query = %v", q)
	}
}

// TestMessageAdSummary_Optimized checks that MessageAdSummary decodes optimized
// summaries into Messages.
func TestMessageAdSummary_Optimized(t *testing.T) {
	t.Parallel()
	var s MessageAdSummary
	in := `{"messageAdType":"MULTI_BROADCAST","messages":[{"type":"WIDE_MESSAGE","mainTitle":"a"},{"type":"SPOTLIGHT_MESSAGE","mainTitle":"b"}],
		"metrics":{"cost":10,"msg_send":2}}`
	if err := json.Unmarshal([]byte(in), &s); err != nil {
		t.Fatal(err)
	}
	types := []MessageType{s.Messages[0].Type, s.Messages[1].Type}
	if s.Message != nil || !slices.Equal(types, []MessageType{MessageTypeWide, MessageTypeSpotlight}) || s.Metrics["cost"] != 10 {
		t.Errorf("summary = %+v", s)
	}
}
