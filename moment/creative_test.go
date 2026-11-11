package moment

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/donghquinn/gokakao/internal/kakaotest"
)

// creativeFilePart is a file part of a multipart request.
type creativeFilePart struct {
	Filename    string
	ContentType string
	Data        string
}

// creativeParseMultipart parses the multipart/form-data body of req.
func creativeParseMultipart(t *testing.T, req kakaotest.Request) (url.Values, map[string]creativeFilePart) {
	t.Helper()
	mediaType, params, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil || mediaType != "multipart/form-data" {
		t.Fatalf("Content-Type = %q, %v", req.Header.Get("Content-Type"), err)
	}
	fields := url.Values{}
	files := map[string]creativeFilePart{}
	r := multipart.NewReader(bytes.NewReader(req.Body), params["boundary"])
	for {
		p, err := r.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		data, _ := io.ReadAll(p)
		if p.FileName() != "" {
			files[p.FormName()] = creativeFilePart{p.FileName(), p.Header.Get("Content-Type"), string(data)}
		} else {
			fields.Add(p.FormName(), string(data))
		}
	}
	return fields, files
}

// creativeAssertJSON checks that the JSON body of req equals want.
func creativeAssertJSON(t *testing.T, req kakaotest.Request, want string) {
	t.Helper()
	if ct := req.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q", ct)
	}
	var got, w any
	if err := json.Unmarshal(req.Body, &got); err != nil {
		t.Fatalf("body %s: %v", req.Body, err)
	}
	if err := json.Unmarshal([]byte(want), &w); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, w) {
		t.Errorf("body = %s\nwant %s", req.Body, want)
	}
}

var creativeKST = time.FixedZone("KST", 9*60*60)

const creativeErrorJSON = `{"code": -813, "msg": "KakaoMomentException", "extras": {"detailCode": 33003, "detailMsg": "소재가 존재하지 않습니다."}}`

func TestCreativeService_List(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/creatives", 200, `{
  "content": [
    {"id": 1111, "name": "소재 1", "config": "ON"},
    {"id": 1112, "name": "소재 2", "config": "OFF"}
  ]
}`)

	list, err := c.Creatives.List(t.Context(), 1234, 1234, &CreativeListOptions{Config: []Config{ConfigOn, ConfigOff}})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1234")
	if req.Method != http.MethodGet || req.Path != "/openapi/v4/creatives" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	if q := req.Query; q.Get("adGroupId") != "1234" || !reflect.DeepEqual(q["config"], []string{"ON", "OFF"}) {
		t.Errorf("query = %v", q)
	}
	if len(list) != 2 || list[0].ID != 1111 || list[0].Name != "소재 1" || list[1].Config != ConfigOff {
		t.Errorf("list = %+v", list)
	}

	if _, err := c.Creatives.List(t.Context(), 1234, 99, nil); err != nil {
		t.Fatal(err)
	}
	if q := srv.LastRequest().Query; len(q) != 1 || q.Get("adGroupId") != "99" {
		t.Errorf("nil options query = %v", q)
	}
	if _, err := c.Creatives.List(t.Context(), 1234, 0, nil); err == nil {
		t.Error("missing ad group: want error")
	}
}

func TestCreativeService_Get(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/creatives/1111", 200, `{
  "id": 1111,
  "name": "디스플레이 광고 소재",
  "landingInfo": {"landingType": "AD_VIEW", "communicatorAdViewId": 1},
  "format": "IMAGE_BANNER",
  "landingUrl": "http://www.daum.net",
  "frequencyCapType": "DAY_IMP",
  "frequencyCap": 3,
  "frequencyCapTime": 24,
  "config": "ON",
  "image": {"url": "http://xxx.kakao.co.kr/sample.jpg", "fileName": "sample.jpeg", "width": 640, "height": 100, "size": 50000},
  "reviewStatus": "REJECTED",
  "creativeStatus": "UNAPPROVED",
  "statusDescription": "심사보류",
  "rejectedReason": [
    {"rejectedTitle": "공통1>가격 불일치", "rejectedContent": "등록하신 광고 소재 내 가격 정보가 랜딩페이지 내 실제 판매금액과 일치하지 않아 광고 등록이 보류되었습니다."}
  ],
  "createdDate": "2020-01-01T00:00:00",
  "lastModifiedDate": "2020-01-01T01:00:00",
  "hasExpandable": false
}`)
	srv.HandleJSON("GET /openapi/v4/creatives/9", 400, creativeErrorJSON)

	cr, err := c.Creatives.Get(t.Context(), 77, 1111)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "77")
	if req.Method != http.MethodGet {
		t.Errorf("method = %s", req.Method)
	}
	if cr.ID != 1111 || cr.Name != "디스플레이 광고 소재" || cr.Format != CreativeFormatImageBanner ||
		cr.LandingInfo == nil || cr.LandingInfo.LandingType != LandingTypeAdView || cr.LandingURL != "http://www.daum.net" ||
		cr.FrequencyCapType != FrequencyCapDayImp || cr.FrequencyCap != 3 || cr.FrequencyCapTime != 24 ||
		cr.Config != ConfigOn || cr.ReviewStatus != ReviewRejected || cr.CreativeStatus != CreativeStatusUnapproved ||
		cr.StatusDescription != "심사보류" || cr.HasExpandable {
		t.Errorf("creative = %+v", cr)
	}
	if cr.Image == nil || cr.Image.Width != 640 || cr.Image.Size != 50000 || cr.Image.FileName != "sample.jpeg" {
		t.Errorf("image = %+v", cr.Image)
	}
	if len(cr.RejectedReason) != 1 || cr.RejectedReason[0].RejectedTitle != "공통1>가격 불일치" {
		t.Errorf("rejected = %+v", cr.RejectedReason)
	}
	if !cr.CreatedDate.Equal(time.Date(2020, 1, 1, 0, 0, 0, 0, creativeKST)) || cr.LastModifiedDate.Hour() != 1 {
		t.Errorf("dates = %v %v", cr.CreatedDate, cr.LastModifiedDate)
	}

	if _, err := c.Creatives.Get(t.Context(), 77, 9); err == nil {
		t.Error("error response: want error")
	}
}

func TestCreativeService_SetConfig(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PUT /openapi/v4/creatives/onOff", 200, ``)

	if err := c.Creatives.SetConfig(t.Context(), 5, 1234, ConfigOn); err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "5")
	if req.Method != http.MethodPut {
		t.Errorf("method = %s", req.Method)
	}
	creativeAssertJSON(t, req, `{"id": 1234, "config": "ON"}`)

	if err := c.Creatives.SetConfig(t.Context(), 5, 1234, ConfigDel); err == nil {
		t.Error("DEL: want error")
	}
	if err := c.Creatives.SetConfig(t.Context(), 5, 0, ConfigOff); err == nil {
		t.Error("missing ID: want error")
	}
}

func TestCreativeService_Delete(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("DELETE /openapi/v4/creatives/1234", 200, ``)

	if err := c.Creatives.Delete(t.Context(), 5, 1234); err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "5")
	if req.Method != http.MethodDelete || req.Path != "/openapi/v4/creatives/1234" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
}

func TestCreativeService_SystemStopReason(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/creatives/77/systemConfigHistory", 200, `{
  "id": 1234,
  "systemConfig": "ADMIN_STOP",
  "reason": "해당 소재 랜딩 지원 종료로 관리자 정지되었습니다.",
  "createdDate": "2021-01-01T00:00:00",
  "lastModifiedDate": "2021-01-01T00:00:00"
}`)

	r, err := c.Creatives.LatestSystemStopReason(t.Context(), 5, 77)
	if err != nil {
		t.Fatal(err)
	}
	assertRequest(t, srv, "5")
	if r.ID != 1234 || r.SystemConfig != SystemConfigAdminStop || r.Reason != "해당 소재 랜딩 지원 종료로 관리자 정지되었습니다." ||
		!r.CreatedDate.Equal(time.Date(2021, 1, 1, 0, 0, 0, 0, creativeKST)) {
		t.Errorf("reason = %+v", r)
	}
}

func TestCreativeService_SystemStopReasons(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/creatives/77/systemConfigHistories", 200, `[
  {"id": 1235, "systemConfig": "ADMIN_STOP", "reason": "해당 소재 랜딩 지원 종료로 관리자 정지되었습니다.", "createdDate": "2021-01-01T00:00:00", "lastModifiedDate": "2021-01-01T00:00:00"},
  {"id": 1234, "systemConfig": "ADMIN_STOP", "reason": "해당 소재 랜딩 지원 종료로 관리자 정지되었습니다.", "createdDate": "2021-01-01T00:00:00", "lastModifiedDate": "2021-01-01T00:00:00"}
]`)

	rs, err := c.Creatives.SystemStopReasons(t.Context(), 5, 77)
	if err != nil {
		t.Fatal(err)
	}
	assertRequest(t, srv, "5")
	if len(rs) != 2 || rs[0].ID != 1235 || rs[1].ID != 1234 || rs[1].SystemConfig != SystemConfigAdminStop || rs[1].LastModifiedDate.Year() != 2021 {
		t.Errorf("reasons = %+v", rs)
	}
}

func TestCreativeForm(t *testing.T) {
	t.Parallel()
	v := struct {
		A string         `json:"a"`
		N int            `json:"n"`
		B bool           `json:"b"`
		S []any          `json:"s"`
		M map[string]any `json:"m"`
		Z *int           `json:"z"`
	}{"x", 3, true, []any{map[string]any{"t": "u"}, 1.5}, map[string]any{"k": "v"}, nil}
	got, err := creativeForm(v)
	if err != nil {
		t.Fatal(err)
	}
	want := url.Values{"a": {"x"}, "n": {"3"}, "b": {"true"}, "s[0].t": {"u"}, "s[1]": {"1.5"}, "m.k": {"v"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("form = %v", got)
	}
	if _, err := creativeForm(func() {}); err == nil {
		t.Error("unencodable value: want error")
	}
}

func TestUploadFile_part(t *testing.T) {
	t.Parallel()
	p, err := (&UploadFile{Filename: "a.png", Data: []byte("\x89PNG\r\n\x1a\n")}).part("f")
	if err != nil || p.Field != "f" || p.ContentType != "image/png" {
		t.Errorf("part = %+v, %v", p, err)
	}
	p, _ = (&UploadFile{Filename: "a", ContentType: "x/y", Data: []byte("1")}).part("f")
	if p.ContentType != "x/y" {
		t.Errorf("content type = %q", p.ContentType)
	}
	if _, err := (&UploadFile{Filename: "a"}).part("f"); err == nil {
		t.Error("empty data: want error")
	}
}
