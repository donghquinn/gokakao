package keywordad

import (
	"bytes"
	"errors"
	"mime"
	"mime/multipart"
	"net/http"
	"slices"
	"testing"

	kakao "github.com/donghquinn/gokakao"
)

func TestBizChannelService_List(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v1/bizChannels/websites", 200, `[{"id":"2222222221","name":"카카오1","websiteUrl":{"rspvUrl":"http://www.daum.net","pcUrl":null,"mobileUrl":null},"ufoBusinessTypeId":445},{"id":"2222222222","name":"카카오2","websiteUrl":{"rspvUrl":"http://www.kakao.com","pcUrl":null,"mobileUrl":null},"ufoBusinessTypeId":"445"}]`)

	cs, err := c.BizChannels.List(t.Context(), 1111111111)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111111111")
	if req.Method != http.MethodGet || req.Path != "/openapi/v1/bizChannels/websites" || len(req.Query) != 0 {
		t.Errorf("request = %s %s?%v", req.Method, req.Path, req.Query)
	}
	want := BizChannel{ID: 2222222222, Name: "카카오2", WebsiteURL: WebsiteURL{RspvURL: "http://www.kakao.com"}, BusinessTypeID: 445}
	if len(cs) != 2 || cs[0].BusinessTypeID != 445 || cs[1] != want {
		t.Errorf("channels = %+v", cs)
	}

	if _, err := c.BizChannels.List(t.Context(), 0); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestBizChannelService_Get(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v1/bizChannels/websites/2222222221", 200, `{"id":"2222222221","name":"카카오1","websiteUrl":{"rspvUrl":"https://www.daum.net","mobileUrl":null,"pcUrl":null},"reviewStatus":"APPROVED","hasAccessAccount":false,"talkChannels":[{"uuid":"@kakao1","name":"카카오채널1"},{"uuid":"@kakao2","name":"카카오채널2"}],"ufoBusinessTypeId":"445"}`)

	ch, err := c.BizChannels.Get(t.Context(), 1111111111, 2222222221)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111111111")
	if req.Method != http.MethodGet || req.Path != "/openapi/v1/bizChannels/websites/2222222221" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	if ch.ID != 2222222221 || ch.Name != "카카오1" || ch.WebsiteURL.RspvURL != "https://www.daum.net" ||
		ch.ReviewStatus != ReviewApproved || ch.HasAccessAccount || ch.BusinessTypeID != 445 ||
		!slices.Equal(ch.TalkChannels, []TalkChannel{{UUID: "@kakao1", Name: "카카오채널1"}, {UUID: "@kakao2", Name: "카카오채널2"}}) {
		t.Errorf("channel = %+v", ch)
	}

	for _, ids := range [][2]int64{{0, 1}, {1, 0}} {
		if _, err := c.BizChannels.Get(t.Context(), ids[0], ids[1]); err == nil {
			t.Errorf("Get(%d, %d): want error", ids[0], ids[1])
		}
	}
}

func TestBizChannelService_Create(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v1/bizChannels/websites", 200, `{"id":"2222222221","name":"카카오","websiteUrl":{"rspvUrl":"https://www.daum.net"},"reviewStatus":"APPROVED","hasAccessAccount":true,"talkChannels":[{"uuid":"@카카오채널1","name":"카카오채널1"},{"uuid":"@카카오채널2","name":"카카오채널2"}],"ufoBusinessTypeId":445}`)

	ch, err := c.BizChannels.Create(t.Context(), 1111111111, NewBizChannel{
		Name:           "카카오",
		BusinessTypeID: 445,
		WebsiteURL:     WebsiteURL{RspvURL: "https://www.daum.net"},
		AccessAccount:  &AccessAccount{ID: "kakao", Password: "k1234"},
		TalkChannels:   []string{"@카카오채널1", "@카카오채널2"},
		ReviewDocuments: []ReviewDocumentURL{
			{Type: ReviewDocumentBizChannel, RequiredReviewDocumentID: 93, FileURL: "http://www.kakao.com/test.jpg"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111111111")
	wantBody := `{"name":"카카오","ufoBusinessTypeId":445,"websiteUrl":{"rspvUrl":"https://www.daum.net"},"accessAccount":{"id":"kakao","password":"k1234"},"talkChannels":["@카카오채널1","@카카오채널2"],"reviewDocuments":[{"type":"BIZ_CHANNEL","requiredReviewDocumentId":93,"fileUrl":"http://www.kakao.com/test.jpg"}]}`
	if req.Method != http.MethodPost || req.Path != "/openapi/v1/bizChannels/websites" || string(req.Body) != wantBody ||
		req.Header.Get("Content-Type") != "application/json" {
		t.Errorf("request = %s %s %s", req.Method, req.Path, req.Body)
	}
	if ch.ID != 2222222221 || !ch.HasAccessAccount || ch.ReviewStatus != ReviewApproved || len(ch.TalkChannels) != 2 ||
		ch.TalkChannels[0].UUID != "@카카오채널1" || ch.BusinessTypeID != 445 {
		t.Errorf("channel = %+v", ch)
	}

	// Optional fields are omitted.
	if _, err := c.BizChannels.Create(t.Context(), 1, NewBizChannel{Name: "n", BusinessTypeID: 1, WebsiteURL: WebsiteURL{PCURL: "https://a", MobileURL: "https://m"}}); err != nil {
		t.Fatal(err)
	}
	if got := string(srv.LastRequest().Body); got != `{"name":"n","ufoBusinessTypeId":1,"websiteUrl":{"pcUrl":"https://a","mobileUrl":"https://m"}}` {
		t.Errorf("minimal body = %s", got)
	}

	n := len(srv.Requests())
	for _, nc := range []NewBizChannel{
		{BusinessTypeID: 1, WebsiteURL: WebsiteURL{RspvURL: "https://a"}},
		{Name: "n", WebsiteURL: WebsiteURL{RspvURL: "https://a"}},
		{Name: "n", BusinessTypeID: 1},
	} {
		if _, err := c.BizChannels.Create(t.Context(), 1, nc); err == nil {
			t.Errorf("Create(%+v): want error", nc)
		}
	}
	if _, err := c.BizChannels.Create(t.Context(), 0, NewBizChannel{Name: "n", BusinessTypeID: 1, WebsiteURL: WebsiteURL{RspvURL: "https://a"}}); err == nil {
		t.Error("missing ad account: want error")
	}
	if len(srv.Requests()) != n {
		t.Error("invalid input was sent")
	}
}

func TestBizChannelService_TalkChannels(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PUT /openapi/v1/bizChannels/websites/2222222221/talkChannels", 200, ``)

	if err := c.BizChannels.AddTalkChannels(t.Context(), 1111111111, 2222222221, []string{"@카카오채널1", "@카카오채널2"}); err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111111111")
	if req.Method != http.MethodPut || string(req.Body) != `{"registration":"ADD","talkChannels":["@카카오채널1","@카카오채널2"]}` {
		t.Errorf("add request = %s %s", req.Method, req.Body)
	}

	if err := c.BizChannels.RemoveTalkChannels(t.Context(), 1111111111, 2222222221, []string{"@kakao"}); err != nil {
		t.Fatal(err)
	}
	req = assertRequest(t, srv, "1111111111")
	if string(req.Body) != `{"registration":"DEL","talkChannels":["@kakao"]}` {
		t.Errorf("remove body = %s", req.Body)
	}

	if err := c.BizChannels.AddTalkChannels(t.Context(), 1, 2, nil); err == nil {
		t.Error("no channels: want error")
	}
	if err := c.BizChannels.RemoveTalkChannels(t.Context(), 1, 0, []string{"@kakao"}); err == nil {
		t.Error("missing biz channel: want error")
	}
}

func TestBizChannelService_Delete(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("DELETE /openapi/v1/bizChannels/2222222221", 200, ``)

	if err := c.BizChannels.Delete(t.Context(), 1111111111, 2222222221); err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111111111")
	if req.Method != http.MethodDelete || req.Path != "/openapi/v1/bizChannels/2222222221" || len(req.Body) != 0 {
		t.Errorf("request = %s %s %s", req.Method, req.Path, req.Body)
	}

	if err := c.BizChannels.Delete(t.Context(), 0, 2222222221); err == nil {
		t.Error("missing ad account: want error")
	}
}

// TestBizChannelCampaignDocument_apiErrors checks that the biz channel,
// campaign, business info and document services return API errors.
func TestBizChannelCampaignDocument_apiErrors(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	const errJSON = `{"code":-820,"msg":"KakaoKeywordAdException","extras":{"detailCode":103005,"detailMsg":"없는 비즈채널입니다."}}`
	for _, p := range []string{
		"GET /openapi/v1/bizChannels/websites", "GET /openapi/v1/bizChannels/websites/2", "POST /openapi/v1/bizChannels/websites",
		"GET /openapi/v1/campaigns", "GET /openapi/v1/campaigns/2", "POST /openapi/v1/campaigns", "PUT /openapi/v1/campaigns/2",
		"GET /openapi/v1/companies", "GET /openapi/v1/documents",
	} {
		srv.HandleJSON(p, 400, errJSON)
	}
	ctx := t.Context()
	calls := map[string]func() error{
		"BizChannels.List": func() error { _, err := c.BizChannels.List(ctx, 1); return err },
		"BizChannels.Get":  func() error { _, err := c.BizChannels.Get(ctx, 1, 2); return err },
		"BizChannels.Create": func() error {
			_, err := c.BizChannels.Create(ctx, 1, NewBizChannel{Name: "n", BusinessTypeID: 1, WebsiteURL: WebsiteURL{RspvURL: "https://a"}})
			return err
		},
		"Campaigns.List": func() error { _, err := c.Campaigns.List(ctx, 1, nil); return err },
		"Campaigns.Get":  func() error { _, err := c.Campaigns.Get(ctx, 1, 2); return err },
		"Campaigns.Create": func() error {
			_, err := c.Campaigns.Create(ctx, 1, NewCampaign{BizChannelID: 2, Name: "c"})
			return err
		},
		"Campaigns.Update": func() error { _, err := c.Campaigns.Update(ctx, 1, 2, CampaignChanges{Name: "c"}); return err },
		"BusinessInfo.Get": func() error { _, err := c.BusinessInfo.Get(ctx, 1, "120-81-47521"); return err },
		"Documents.List":   func() error { _, err := c.Documents.List(ctx, 1, 2); return err },
	}
	for name, f := range calls {
		var e *kakao.Error
		if err := f(); !errors.As(err, &e) || e.Extras == nil || e.Extras.DetailCode != 103005 {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}

func TestBizChannelService_Create_files(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v1/bizChannels/websites", 200, `{"id":"2222222221","name":"카카오"}`)

	ch, err := c.BizChannels.Create(t.Context(), 1111111111, NewBizChannel{
		Name:           "카카오",
		BusinessTypeID: 445,
		WebsiteURL:     WebsiteURL{RspvURL: "https://wwws.daum.net"},
		Files: []ReviewDocumentFile{
			{Type: ReviewDocumentBizChannel, RequiredReviewDocumentID: 8, Filename: "document.jpg", Data: []byte("doc")},
			{Type: ReviewDocumentCreative, RequiredReviewDocumentID: 93, Filename: "content_document.jpg", Data: []byte("doc2")},
		},
	})
	if err != nil || ch.ID != 2222222221 {
		t.Fatalf("Create = %+v, %v", ch, err)
	}
	req := assertRequest(t, srv, "1111111111")
	mt, params, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil || mt != "multipart/form-data" {
		t.Fatalf("Content-Type = %q", req.Header.Get("Content-Type"))
	}
	form, err := multipart.NewReader(bytes.NewReader(req.Body), params["boundary"]).ReadForm(1 << 20)
	if err != nil {
		t.Fatal(err)
	}
	if got := form.Value["websiteBizChannelDto"]; len(got) != 1 || got[0] != `{"name":"카카오","ufoBusinessTypeId":445,"websiteUrl":{"rspvUrl":"https://wwws.daum.net"}}` {
		t.Errorf("dto = %q", got)
	}
	files := form.File["files"]
	if len(files) != 2 || files[0].Filename != "document.jpg" ||
		files[0].Header.Get("Required-Review-Document-Id") != "8" || files[0].Header.Get("Review-Document-Type") != "BIZ_CHANNEL" ||
		files[1].Header.Get("Required-Review-Document-Id") != "93" || files[1].Header.Get("Review-Document-Type") != "CREATIVE" {
		t.Errorf("files = %+v", files)
	}

	_, err = c.BizChannels.Create(t.Context(), 1, NewBizChannel{
		Name: "x", BusinessTypeID: 1, WebsiteURL: WebsiteURL{RspvURL: "https://x"},
		Files: []ReviewDocumentFile{{Filename: "a"}},
	})
	if err == nil {
		t.Error("incomplete file: want error")
	}
}
