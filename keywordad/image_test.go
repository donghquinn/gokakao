package keywordad

import (
	"bytes"
	"io"
	"mime"
	"mime/multipart"
	"testing"
	"time"
)

const imageJSON = `{"id": "8888888881", "lastUsedDate": "2021-02-20T12:00:00.702463", "url": "https://www.kakao.com/image1.jpg", "fileName": "image1.jpg", "width": 450, "height": 450, "size": 200000}`

func assertImage(t *testing.T, img *Image) {
	t.Helper()
	want := time.Date(2021, 2, 20, 12, 0, 0, 702463000, time.FixedZone("KST", 9*60*60))
	if img.ID != 8888888881 || img.URL != "https://www.kakao.com/image1.jpg" || img.FileName != "image1.jpg" ||
		img.Width != 450 || img.Height != 450 || img.Size != 200000 || !img.LastUsedDate.Equal(want) {
		t.Errorf("image = %+v", img)
	}
}

func TestImageService_List(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v1/images", 200, `{
  "content": [
    {"id": "8888888882", "url": "http://www.kakao.com/image1.jpg", "fileName": null, "width": 450, "height": 450, "size": 200000, "lastUsedDate": "2021-02-20T12:00:00.702463"},
    {"id": "8888888882", "url": "http://www.kakao.com/image2.jpg", "fileName": null, "width": 450, "height": 450, "size": 200000, "lastUsedDate": "2021-02-20T12:20:00.702463"}
  ],
  "page": {"number": 0, "size": 10, "totalElements": 2, "totalPages": 0},
  "recentlyUsed": false
}`)

	page, err := c.Images.List(t.Context(), 1111111111, &ImageListOptions{RecentlyUsed: true, Page: 2, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111111111")
	if q := req.Query; q.Get("recentlyUsed") != "true" || q.Get("page") != "2" || q.Get("limit") != "10" {
		t.Errorf("query = %v", q)
	}
	if len(page.Content) != 2 || page.Page.TotalElements != 2 || page.RecentlyUsed ||
		page.Content[1].URL != "http://www.kakao.com/image2.jpg" || page.Content[1].LastUsedDate.Minute() != 20 {
		t.Errorf("page = %+v", page)
	}

	if _, err := c.Images.List(t.Context(), 1111111111, nil); err != nil {
		t.Fatal(err)
	}
	if q := srv.LastRequest().Query; len(q) != 0 {
		t.Errorf("nil options query = %v", q)
	}
}

func TestImageService_Get(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v1/images/8888888881", 200, imageJSON)

	img, err := c.Images.Get(t.Context(), 1111111111, 8888888881)
	if err != nil {
		t.Fatal(err)
	}
	assertRequest(t, srv, "1111111111")
	assertImage(t, img)
}

func TestImageService_Upload(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v1/images", 200, imageJSON)
	png := []byte("\x89PNG\r\n\x1a\n fake png")

	img, err := c.Images.Upload(t.Context(), 1111111111, "image1.png", png)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111111111")
	assertImage(t, img)

	mediaType, params, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil || mediaType != "multipart/form-data" {
		t.Fatalf("Content-Type = %q, %v", req.Header.Get("Content-Type"), err)
	}
	r := multipart.NewReader(bytes.NewReader(req.Body), params["boundary"])
	part, err := r.NextPart()
	if err != nil {
		t.Fatal(err)
	}
	data, _ := io.ReadAll(part)
	if part.FormName() != "imageFile" || part.FileName() != "image1.png" || part.Header.Get("Content-Type") != "image/png" || !bytes.Equal(data, png) {
		t.Errorf("part %q %q %q = %q", part.FormName(), part.FileName(), part.Header.Get("Content-Type"), data)
	}
	if _, err := r.NextPart(); err != io.EOF {
		t.Errorf("extra part: %v", err)
	}

	if _, err := c.Images.Upload(t.Context(), 1, "", png); err == nil {
		t.Error("missing file name: want error")
	}
	if _, err := c.Images.Upload(t.Context(), 1, "a.png", nil); err == nil {
		t.Error("missing data: want error")
	}
}

func TestImageService_UploadURL(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v1/images", 200, imageJSON)

	img, err := c.Images.UploadURL(t.Context(), 1111111111, "https://www.image.com/image1.jpg")
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111111111")
	if req.Header.Get("Content-Type") != "application/json" || string(req.Body) != `{"imageFileUrl":"https://www.image.com/image1.jpg"}` {
		t.Errorf("body = %s", req.Body)
	}
	assertImage(t, img)

	if _, err := c.Images.UploadURL(t.Context(), 1, ""); err == nil {
		t.Error("missing URL: want error")
	}
}

func TestImageService_Delete(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("DELETE /openapi/v1/images/8888888881", 200, ``)

	if err := c.Images.Delete(t.Context(), 1111111111, 8888888881); err != nil {
		t.Fatal(err)
	}
	assertRequest(t, srv, "1111111111")
}
