package moment

import (
	"errors"
	"net/http"
	"testing"

	kakao "github.com/donghquinn/gokakao"
)

func TestBizboardLandingService_BizForms(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/creatives/landing/talkBizForms", 200, `[
  {
    "id": 2617,
    "adAccountId": 1334,
    "bizFormId": 5678,
    "title": "비즈니스폼",
    "flowType": "SURVEY",
    "linkedDate": "2020-01-01 01:00:00",
    "beginDateTime": "2020-01-01 00:00",
    "endDateTime": "2020-01-01 00:00"
  }
]`)

	fs, err := c.BizboardLandings.BizForms(t.Context(), 1334)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1334")
	if req.Method != http.MethodGet {
		t.Errorf("method = %s", req.Method)
	}
	if len(fs) != 1 || fs[0].ID != 2617 || fs[0].AdAccountID != 1334 || fs[0].BizFormID != 5678 ||
		fs[0].Title != "비즈니스폼" || fs[0].FlowType != BizFormSurvey || fs[0].LinkedDate != "2020-01-01 01:00:00" ||
		fs[0].BeginDateTime != "2020-01-01 00:00" || fs[0].EndDateTime != "2020-01-01 00:00" {
		t.Errorf("biz forms = %+v", fs)
	}

	if _, err := c.BizboardLandings.BizForms(t.Context(), 0); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestBizboardLandingService_AdViews(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/creatives/landing/adViews", 200, `[
  {"id": 5678, "name": "첫번째_애드뷰", "type": "COMPACT", "templateType": "IMAGE"},
  {"id": 5679, "name": "두번째_애드뷰", "type": "FULL", "templateType": "SCROLL"}
]`)

	vs, err := c.BizboardLandings.AdViews(t.Context(), 1334)
	if err != nil {
		t.Fatal(err)
	}
	assertRequest(t, srv, "1334")
	if len(vs) != 2 || vs[1] != (AdView{ID: 5679, Name: "두번째_애드뷰", Type: AdViewFull, TemplateType: AdViewTemplateScroll}) {
		t.Errorf("ad views = %+v", vs)
	}

	if _, err := c.BizboardLandings.AdViews(t.Context(), 0); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestBizboardLandingService_ChannelPosts(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/creatives/landing/channelPosts/77", 200, `{
  "items": [
    {
      "id": 1,
      "title": "포스트",
      "status": "published",
      "permalink": "http://kakao.com/_xxxx/1",
      "createdDate": "2020-01-01 00:00:00",
      "publishedDate": "2020-01-01 00:00:00"
    }
  ]
}`)

	ps, err := c.BizboardLandings.ChannelPosts(t.Context(), 1334, 77)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1334")
	if req.Path != "/openapi/v4/creatives/landing/channelPosts/77" {
		t.Errorf("path = %s", req.Path)
	}
	if len(ps) != 1 || ps[0] != (ChannelPost{ID: 1, Title: "포스트", Status: "published", Permalink: "http://kakao.com/_xxxx/1", CreatedDate: "2020-01-01 00:00:00", PublishedDate: "2020-01-01 00:00:00"}) {
		t.Errorf("posts = %+v", ps)
	}

	if _, err := c.BizboardLandings.ChannelPosts(t.Context(), 1334, 0); err == nil {
		t.Error("missing profile: want error")
	}
}

func TestBizboardLandingService_apiErrors(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("/", 400, `{"code":-813,"msg":"KakaoMomentException"}`)
	ctx := t.Context()
	calls := map[string]func() error{
		"BizForms":     func() error { _, err := c.BizboardLandings.BizForms(ctx, 1); return err },
		"AdViews":      func() error { _, err := c.BizboardLandings.AdViews(ctx, 1); return err },
		"ChannelPosts": func() error { _, err := c.BizboardLandings.ChannelPosts(ctx, 1, 2); return err },
	}
	for name, f := range calls {
		var e *kakao.Error
		if err := f(); !errors.As(err, &e) || e.Code != -813 {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}
