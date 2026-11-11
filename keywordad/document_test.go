package keywordad

import (
	"bytes"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"testing"
)

func TestDocumentService_List(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v1/documents", 200, `{"bizChannelId":"2222222221","bizChannelName":"비즈채널_변경이력1","documents":[{"id":"9999999991","fileName":"test.jpg","requiredDocumentId":92,"requiredDocumentName":"의료광고심의필증","type":"CREATIVE","url":"https://www.kakao.com/document1.jpg"},{"id":"9999999991","fileName":"test.jpg","requiredDocumentId":93,"requiredDocumentName":"기타서류","type":"CREATIVE","url":"https://www.kakao.com/document2.jpg"}]}`)

	d, err := c.Documents.List(t.Context(), 1111111111, 2222222221)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111111111")
	if req.Method != http.MethodGet || req.Path != "/openapi/v1/documents" || req.Query.Get("bizChannelId") != "2222222221" {
		t.Errorf("request = %s %s?%v", req.Method, req.Path, req.Query)
	}
	want := ReviewDocument{ID: 9999999991, FileName: "test.jpg", RequiredDocumentID: 93, RequiredDocumentName: "기타서류", Type: ReviewDocumentCreative, URL: "https://www.kakao.com/document2.jpg"}
	if d.BizChannelID != 2222222221 || d.BizChannelName != "비즈채널_변경이력1" || len(d.Documents) != 2 || d.Documents[1] != want {
		t.Errorf("documents = %+v", d)
	}

	if _, err := c.Documents.List(t.Context(), 1111111111, 0); err == nil {
		t.Error("missing biz channel: want error")
	}
}

func TestDocumentService_Create(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v1/documents", 200, ``)

	err := c.Documents.Create(t.Context(), 1111111111, NewDocument{BizChannelID: 2222222221, Type: ReviewDocumentCreative, RequiredDocumentID: 93, FileURL: "https://www.image.com/document1.png"})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111111111")
	if req.Method != http.MethodPost || req.Path != "/openapi/v1/documents" || req.Header.Get("Content-Type") != "application/json" ||
		string(req.Body) != `{"bizChannelId":2222222221,"type":"CREATIVE","requiredDocumentId":93,"fileUrl":"https://www.image.com/document1.png"}` {
		t.Errorf("request = %s %s %s", req.Method, req.Path, req.Body)
	}

	for _, f := range []DocumentFile{
		{Filename: "doc.pdf", ContentType: "application/pdf", Data: []byte("%PDF")},
		{Data: []byte("x")},
	} {
		err = c.Documents.Create(t.Context(), 1111111111, NewDocument{BizChannelID: 2222222221, Type: ReviewDocumentBizChannel, RequiredDocumentID: 1, File: &f})
		if err != nil {
			t.Fatal(err)
		}
		req = assertRequest(t, srv, "1111111111")
		mt, params, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
		if err != nil || mt != "multipart/form-data" {
			t.Fatalf("Content-Type = %q", req.Header.Get("Content-Type"))
		}
		form, err := multipart.NewReader(bytes.NewReader(req.Body), params["boundary"]).ReadForm(1 << 20)
		if err != nil {
			t.Fatal(err)
		}
		if form.Value["bizChannelId"][0] != "2222222221" || form.Value["type"][0] != "BIZ_CHANNEL" || form.Value["requiredDocumentId"][0] != "1" {
			t.Errorf("fields = %v", form.Value)
		}
		fhs := form.File["file"]
		if len(fhs) != 1 {
			t.Fatalf("files = %v", form.File)
		}
		wantName := f.Filename
		if wantName == "" {
			wantName = "document"
		}
		fh := fhs[0]
		file, err := fh.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, _ := io.ReadAll(file)
		_ = file.Close()
		if fh.Filename != wantName || !bytes.Equal(data, f.Data) {
			t.Errorf("file = %q %q", fh.Filename, data)
		}
		if f.ContentType != "" && fh.Header.Get("Content-Type") != f.ContentType {
			t.Errorf("file Content-Type = %q", fh.Header.Get("Content-Type"))
		}
	}

	n := len(srv.Requests())
	for _, d := range []NewDocument{
		{Type: ReviewDocumentCreative, RequiredDocumentID: 93, FileURL: "https://a"},
		{BizChannelID: 2, RequiredDocumentID: 93, FileURL: "https://a"},
		{BizChannelID: 2, Type: ReviewDocumentCreative, FileURL: "https://a"},
		{BizChannelID: 2, Type: ReviewDocumentCreative, RequiredDocumentID: 93},
		{BizChannelID: 2, Type: ReviewDocumentCreative, RequiredDocumentID: 93, FileURL: "https://a", File: &DocumentFile{}},
	} {
		if err := c.Documents.Create(t.Context(), 1, d); err == nil {
			t.Errorf("Create(%+v): want error", d)
		}
	}
	if err := c.Documents.Create(t.Context(), 0, NewDocument{BizChannelID: 2, Type: ReviewDocumentCreative, RequiredDocumentID: 93, FileURL: "https://a"}); err == nil {
		t.Error("missing ad account: want error")
	}
	if len(srv.Requests()) != n {
		t.Error("invalid input was sent")
	}
}
