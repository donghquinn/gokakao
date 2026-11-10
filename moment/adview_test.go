package moment

import (
	"errors"
	"net/http"
	"testing"

	kakao "github.com/donghquinn/gokakao"
)

func TestAdViewService_List(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/adViews", 200, `[{"id": 1, "name": "첫번째_애드뷰", "type": "COMPACT", "templateType": "IMAGE"}]`)

	vs, err := c.AdViews.List(t.Context(), 1111)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodGet {
		t.Errorf("method = %s", req.Method)
	}
	if len(vs) != 1 || vs[0] != (AdView{ID: 1, Name: "첫번째_애드뷰", Type: AdViewCompact, TemplateType: AdViewTemplateImage}) {
		t.Errorf("ad views = %+v", vs)
	}

	if _, err := c.AdViews.List(t.Context(), 0); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestAdViewService_Get(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/adViews/1", 200, `{"id": 1, "name": "첫번째_애드뷰", "type": "COMPACT", "templateType": "IMAGE"}`)

	v, err := c.AdViews.Get(t.Context(), 1111, 1)
	if err != nil {
		t.Fatal(err)
	}
	assertRequest(t, srv, "1111")
	if *v != (AdView{ID: 1, Name: "첫번째_애드뷰", Type: AdViewCompact, TemplateType: AdViewTemplateImage}) {
		t.Errorf("ad view = %+v", v)
	}

	if _, err := c.AdViews.Get(t.Context(), 1111, 0); err == nil {
		t.Error("missing ad view: want error")
	}
}

func TestAdViewService_Delete(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("DELETE /openapi/v4/adViews/1", 200, ``)
	srv.HandleJSON("DELETE /openapi/v4/adViews/2", 400, `{"code":-813,"msg":"KakaoMomentException","extras":{"detailCode":39020,"detailMsg":"이 애드뷰를 랜딩으로 사용중인 소재가 있어서 삭제할 수 없습니다."}}`)

	if err := c.AdViews.Delete(t.Context(), 1111, 1); err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodDelete || req.Path != "/openapi/v4/adViews/1" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}

	var e *kakao.Error
	if err := c.AdViews.Delete(t.Context(), 1111, 2); !errors.As(err, &e) || e.Code != -813 {
		t.Errorf("in use: err = %v", err)
	}
	if err := c.AdViews.Delete(t.Context(), 0, 1); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestAdViewService_apiErrors(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("/", 400, `{"code":-813,"msg":"KakaoMomentException"}`)
	ctx := t.Context()
	calls := map[string]func() error{
		"List": func() error { _, err := c.AdViews.List(ctx, 1); return err },
		"Get":  func() error { _, err := c.AdViews.Get(ctx, 1, 2); return err },
	}
	for name, f := range calls {
		var e *kakao.Error
		if err := f(); !errors.As(err, &e) || e.Code != -813 {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}
