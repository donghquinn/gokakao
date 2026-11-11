package keywordad

import "testing"

const creativeDetailJSON = `{
  "adAccountId": "1111111111",
  "bizChannelId": "2222222221",
  "id": "66666666661",
  "name": "소재이름",
  "title": "타이틀",
  "description": "설명문구",
  "landingInfo": {"rspvLandingUrl": "https://www.daum.net", "pcLandingUrl": null, "mobileLandingUrl": null},
  "license": {"licenseNumber": "1234567890", "reviewDocumentId": "9999999991"},
  "assets": {
    "thumbnail": {"imageId": "8888888881", "text": "썸네일이미지설명"},
    "multiThumbnails": [
      {"imageId": "8888888882", "text": "멀티썸네일1", "landingInfo": {"rspvLandingUrl": "https://www.daum.net", "pcLandingUrl": null, "mobileLandingUrl": null}},
      {"imageId": "8888888883", "text": "멀티썸네일2", "landingInfo": {"rspvLandingUrl": "https://www.daum.net", "pcLandingUrl": null, "mobileLandingUrl": null}}
    ],
    "additionalTitle": "추가제목",
    "additionalLinks": [
      {"text": "부가링크1", "landingInfo": {"rspvLandingUrl": "https://www.daum.net", "pcLandingUrl": null, "mobileLandingUrl": null}}
    ],
    "subject": {"subject": "DISCOUNT", "text": "할인말머리", "landingInfo": {"rspvLandingUrl": "https://www.daum.net", "pcLandingUrl": null, "mobileLandingUrl": null}},
    "priceTables": [
      {"text": "가격테이블1", "price": 1000, "range": "STATIC", "landingInfo": {"rspvLandingUrl": "https://www.daum.net", "pcLandingUrl": null, "mobileLandingUrl": null}},
      {"text": "가격테이블2", "price": 2000, "range": "FROM", "landingInfo": {"rspvLandingUrl": "https://www.daum.net", "pcLandingUrl": null, "mobileLandingUrl": null}}
    ],
    "phoneNumber": "010-1234-5678",
    "calculate": {"landingInfo": {"rspvLandingUrl": "https://www.daum.net", "pcLandingUrl": null, "mobileLandingUrl": null}},
    "talkChannelId": "@kakao1"
  }
}`

func TestCreativeService_List(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v1/creatives/basic", 200, `{
  "content": [
    {"bizChannelId": "2222222221", "id": "66666666661", "name": "소재1", "title": "타이틀", "description": "설명문구",
     "assetTypes": ["THUMBNAIL", "MULTI_THUMBNAIL", "ADDITIONAL_TITLE", "ADDITIONAL_LINK", "SUBJECT", "PRICE_TABLE", "CALCULATE", "PHONE_NUMBER", "TALK_CHANNEL"], "config": "ON"},
    {"bizChannelId": "2222222221", "id": "66666666662", "name": "소재2", "title": "타이틀", "description": "설명문구", "assetTypes": [], "config": "ON"}
  ],
  "page": {"number": 0, "size": 10, "totalElements": 1, "totalPages": 1}
}`)

	page, err := c.Creatives.List(t.Context(), 1111111111, 2222222221, &CreativeListOptions{Config: ConfigOn, Page: 1, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111111111")
	if q := req.Query; q.Get("bizChannelId") != "2222222221" || q.Get("config") != "ON" || q.Get("page") != "1" || q.Get("limit") != "10" {
		t.Errorf("query = %v", q)
	}
	if len(page.Content) != 2 || page.Page != (PageInfo{Number: 0, Size: 10, TotalElements: 1, TotalPages: 1}) {
		t.Fatalf("page = %+v", page)
	}
	first := page.Content[0]
	if first.ID != 66666666661 || first.BizChannelID != 2222222221 || first.Name != "소재1" || first.Config != ConfigOn ||
		len(first.AssetTypes) != 9 || first.AssetTypes[8] != AssetTalkChannel {
		t.Errorf("first = %+v", first)
	}

	if _, err := c.Creatives.List(t.Context(), 1111111111, 2222222221, nil); err != nil {
		t.Fatal(err)
	}
	if q := srv.LastRequest().Query; len(q) != 1 {
		t.Errorf("nil options query = %v", q)
	}

	if _, err := c.Creatives.List(t.Context(), 1111111111, 0, nil); err == nil {
		t.Error("missing biz channel: want error")
	}
}

func TestCreativeService_Get(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v1/creatives/basic/66666666661", 200, creativeDetailJSON)

	cr, err := c.Creatives.Get(t.Context(), 1111111111, 66666666661)
	if err != nil {
		t.Fatal(err)
	}
	assertRequest(t, srv, "1111111111")
	assertCreativeDetail(t, cr)
}

func assertCreativeDetail(t *testing.T, cr *Creative) {
	t.Helper()
	if cr.AdAccountID != 1111111111 || cr.BizChannelID != 2222222221 || cr.ID != 66666666661 || cr.Title != "타이틀" ||
		cr.LandingInfo == nil || cr.LandingInfo.RspvLandingURL != "https://www.daum.net" ||
		cr.License == nil || cr.License.ReviewDocumentID != "9999999991" {
		t.Errorf("creative = %+v", cr)
	}
	a := cr.Assets
	if a == nil || a.Thumbnail == nil || a.Thumbnail.ImageID != "8888888881" || len(a.MultiThumbnails) != 2 ||
		a.AdditionalTitle != "추가제목" || len(a.AdditionalLinks) != 1 || a.Subject == nil || a.Subject.Subject != SubjectDiscount ||
		len(a.PriceTables) != 2 || a.PriceTables[1].Range != PriceFrom || a.PriceTables[1].Price != 2000 ||
		a.Calculate == nil || a.Calculate.LandingInfo == nil || a.PhoneNumber != "010-1234-5678" || a.TalkChannelID != "@kakao1" {
		t.Errorf("assets = %+v", a)
	}
}

func TestCreativeService_Create(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v1/creatives/basic", 200, creativeDetailJSON)

	cr, err := c.Creatives.Create(t.Context(), 1111111111, NewCreative{
		BizChannelID: 2222222221,
		Name:         "소재이름",
		Title:        "타이틀",
		Description:  "설명문구",
		LandingInfo:  &LandingInfo{RspvLandingURL: "https://www.daum.net"},
		Assets: &Assets{
			Thumbnail:   &ThumbnailAsset{ImageID: "8888888881", Text: "썸네일이미지설명"},
			PriceTables: []PriceTableAsset{{Text: "가격테이블1", Price: 1000, Range: PriceStatic}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111111111")
	if req.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q", req.Header.Get("Content-Type"))
	}
	want := `{"bizChannelId":2222222221,"name":"소재이름","title":"타이틀","description":"설명문구","landingInfo":{"rspvLandingUrl":"https://www.daum.net"},"assets":{"thumbnail":{"imageId":"8888888881","text":"썸네일이미지설명"},"priceTables":[{"text":"가격테이블1","price":1000,"range":"STATIC"}]}}`
	if string(req.Body) != want {
		t.Errorf("body = %s", req.Body)
	}
	assertCreativeDetail(t, cr)

	for _, bad := range []NewCreative{
		{Name: "n", Title: "t", Description: "d"},
		{BizChannelID: 1, Title: "t", Description: "d"},
		{BizChannelID: 1, Name: "n", Description: "d"},
		{BizChannelID: 1, Name: "n", Title: "t"},
	} {
		if _, err := c.Creatives.Create(t.Context(), 1, bad); err == nil {
			t.Errorf("Create(%+v): want error", bad)
		}
	}
}

func TestCreativeService_Delete(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("DELETE /openapi/v1/creatives/66666666661", 200, ``)

	if err := c.Creatives.Delete(t.Context(), 1111111111, 66666666661); err != nil {
		t.Fatal(err)
	}
	assertRequest(t, srv, "1111111111")
}
