package moment

import (
	"net/http"
	"testing"
)

func TestMessageService_Coupons(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/messages/coupons", 200, `[
  {
    "id": 12345, "title": "첫 번째 테스트 쿠폰", "content": "첫 번째 테스트 쿠폰입니다.", "statusCode": "done",
    "permalink": "http://xxx.kakao.com/_xxx/coupons/_xxx",
    "entryImage": {"type": "image", "url": "http://xxx.kakao.com/img_xl.jpg", "thumbnail": null, "previewImage": null, "playUrl": null, "name": null},
    "createdAt": "2021-01-11T15:43:12"
  },
  {
    "id": 12346, "title": "두번째 테스트 쿠폰", "content": "두번째 테스트 쿠폰입니다.", "statusCode": "done",
    "permalink": "http://xxx.kakao.com/_xxx/coupons/_xxx",
    "entryImage": {"type": "image", "url": "http://xxx.kakao.com/img_xl.jpg", "thumbnail": null, "previewImage": null, "playUrl": null, "name": null},
    "createdAt": "2021-01-11T15:43:12"
  }
]`)

	list, err := c.Messages.Coupons(t.Context(), 5, "_xbHxd")
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "5")
	if req.Method != http.MethodGet || req.Query.Get("profileId") != "_xbHxd" {
		t.Errorf("request = %s %v", req.Method, req.Query)
	}
	if len(list) != 2 || list[0].ID != 12345 || list[0].Title != "첫 번째 테스트 쿠폰" || list[0].StatusCode != "done" ||
		list[0].EntryImage == nil || list[0].EntryImage.URL != "http://xxx.kakao.com/img_xl.jpg" || list[1].CreatedAt != "2021-01-11T15:43:12" {
		t.Errorf("coupons = %+v", list)
	}
	if _, err := c.Messages.Coupons(t.Context(), 5, ""); err == nil {
		t.Error("missing profile: want error")
	}
}

func TestMessageService_Coupon(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/messages/coupons/12345", 200, `{
  "id": 12345,
  "title": "테스트 쿠폰입니다.",
  "content": "테스트입니다.",
  "statusCode": "done",
  "permalink": "http://xxx.kakao.com/xxx/coupons/xxx",
  "entryImage": null,
  "createdAt": "2021-02-26T16:12:31"
}`)

	cp, err := c.Messages.Coupon(t.Context(), 5, 12345, "_xbHxd")
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "5")
	if req.Query.Get("profileId") != "_xbHxd" {
		t.Errorf("query = %v", req.Query)
	}
	if cp.ID != 12345 || cp.Content != "테스트입니다." || cp.Permalink != "http://xxx.kakao.com/xxx/coupons/xxx" || cp.EntryImage != nil {
		t.Errorf("coupon = %+v", cp)
	}
	if _, err := c.Messages.Coupon(t.Context(), 5, 12345, ""); err == nil {
		t.Error("missing profile: want error")
	}
}

func TestMessageService_AdViews(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/messages/adViews", 200, `[
  {"id": 1234, "name": "애드뷰1", "type": "COMPACT", "templateType": "IMAGE"},
  {"id": 5678, "name": "애드뷰2", "type": "COMPACT", "templateType": "IMAGE"}
]`)

	list, err := c.Messages.AdViews(t.Context(), 5)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "5")
	if len(req.Query) != 0 {
		t.Errorf("query = %v", req.Query)
	}
	if len(list) != 2 || list[1].ID != 5678 || list[1].Name != "애드뷰2" || list[1].Type != AdViewCompact || list[1].TemplateType != AdViewTemplateImage {
		t.Errorf("ad views = %+v", list)
	}
}

func TestMessageService_Posts(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/messages/posts/{profileID}", 200, `[
  {"id": 1234, "title": "소식", "status": "published", "permalink": "http://kakao.com/_xxxx/1", "createdDate": "2021-01-01T00:00:00", "publishedDate": "2021-01-01T00:00:00"},
  {"id": 1235, "title": "소식 2", "status": "published", "permalink": "http://kakao.com/_xxxx/2", "createdDate": "2021-01-02T00:00:00", "publishedDate": "2021-01-02T00:00:00"}
]`)

	list, err := c.Messages.Posts(t.Context(), 5, "_xbHxd")
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "5")
	if req.Path != "/openapi/v4/messages/posts/_xbHxd" {
		t.Errorf("path = %s", req.Path)
	}
	if len(list) != 2 || list[0].ID != 1234 || list[0].Status != "published" || list[1].Title != "소식 2" || list[1].PublishedDate != "2021-01-02T00:00:00" {
		t.Errorf("posts = %+v", list)
	}
	if _, err := c.Messages.Posts(t.Context(), 5, ""); err == nil {
		t.Error("missing profile: want error")
	}
}
