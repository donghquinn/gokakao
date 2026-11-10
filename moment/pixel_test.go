package moment

import (
	"errors"
	"net/http"
	"testing"
	"time"

	kakao "github.com/donghquinn/gokakao"
)

func TestPixelService_List(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/trackers/rights", 200, `{
  "trackRights": [
    {
      "trackId": "1234567890",
      "trackName": "픽셀",
      "population": 0,
      "eventStatus": "COLLECTION",
      "createdDate": "2019-12-10T08:18:22.000+0000",
      "lastEventDate": null,
      "role": "MASTER"
    },
    {
      "trackId": "1234567891",
      "trackName": "픽셀",
      "population": 0,
      "eventStatus": "PRE_COLLECTION",
      "createdDate": "2020-01-01T00:00:00.000+0000",
      "lastEventDate": "2020-01-01 15:00:00",
      "role": "MEMBER"
    }
  ]
}`)

	ts, err := c.Pixels.List(t.Context(), 12345)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "12345")
	if req.Method != http.MethodGet {
		t.Errorf("method = %s", req.Method)
	}
	if len(ts) != 2 {
		t.Fatalf("track rights = %+v", ts)
	}
	if ts[0].TrackID != "1234567890" || ts[0].TrackName != "픽셀" || ts[0].Population != 0 ||
		ts[0].EventStatus != TrackerCollection || ts[0].Role != TrackerMaster ||
		!ts[0].CreatedDate.Equal(time.Date(2019, 12, 10, 8, 18, 22, 0, time.UTC)) || !ts[0].LastEventDate.IsZero() {
		t.Errorf("track right 0 = %+v", ts[0])
	}
	if ts[1].EventStatus != TrackerPreCollection || ts[1].Role != TrackerMember ||
		!ts[1].LastEventDate.Equal(time.Date(2020, 1, 1, 6, 0, 0, 0, time.UTC)) {
		t.Errorf("track right 1 = %+v", ts[1])
	}

	if _, err := c.Pixels.List(t.Context(), 0); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestPixelService_Available(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/trackers/rightAvailables", 200, `[
  {
    "trackId": "1234567890123456789",
    "name": "권한추가_가능한_픽셀&SDK",
    "createdAt": "2020-01-01T00:00:00.000+0000",
    "updatedAt": "2020-01-01T00:00:00.000+0000",
    "role": "MEMBER"
  },
  {
    "trackId": "9876543210987654321",
    "name": "권한추가_가능한_픽셀&SDK",
    "createdAt": "2020-01-01T00:00:00.000+0000",
    "updatedAt": "2020-01-01T00:00:00.000+0000",
    "role": "REQUEST"
  }
]`)

	ts, err := c.Pixels.Available(t.Context(), 12345)
	if err != nil {
		t.Fatal(err)
	}
	assertRequest(t, srv, "12345")
	if len(ts) != 2 || ts[0].TrackID != "1234567890123456789" || ts[0].Name != "권한추가_가능한_픽셀&SDK" ||
		ts[1].Role != TrackerRequest || ts[1].UpdatedAt.Year() != 2020 || !ts[0].CreatedAt.Equal(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("trackers = %+v", ts)
	}

	if _, err := c.Pixels.Available(t.Context(), 0); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestPixelService_Link(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v4/trackers", 200, ``)

	if err := c.Pixels.Link(t.Context(), 12345, "5678901234"); err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "12345")
	if req.Method != http.MethodPost || string(req.Body) != `{"trackId":"5678901234","adAccountId":12345}` {
		t.Errorf("request = %s %s", req.Method, req.Body)
	}

	if err := c.Pixels.Link(t.Context(), 12345, ""); err == nil {
		t.Error("missing track: want error")
	}
}

func TestPixelService_Unlink(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("DELETE /openapi/v4/trackers/5678901234", 200, ``)

	if err := c.Pixels.Unlink(t.Context(), 12345, "5678901234"); err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "12345")
	if req.Method != http.MethodDelete || req.Path != "/openapi/v4/trackers/5678901234" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}

	if err := c.Pixels.Unlink(t.Context(), 0, "1"); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestPixelService_apiErrors(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("/", 400, `{"code":-813,"msg":"KakaoMomentException"}`)
	ctx := t.Context()
	calls := map[string]func() error{
		"List":      func() error { _, err := c.Pixels.List(ctx, 1); return err },
		"Available": func() error { _, err := c.Pixels.Available(ctx, 1); return err },
	}
	for name, f := range calls {
		var e *kakao.Error
		if err := f(); !errors.As(err, &e) || e.Code != -813 {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}
