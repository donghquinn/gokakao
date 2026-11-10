package keywordad

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestPixelService_List(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v2/trackers/rights", 200, `{
  "trackRights": [
    {"trackId": "200000000000001", "trackName": "픽셀1", "population": 0, "eventStatus": "COLLECTION",
     "createdDate": "2022-02-20T00:00:00.000+0000", "lastEventDate": "2022-02-20T20:00:00.000+0000", "role": "MASTER"},
    {"trackId": "200000000000002", "trackName": "픽셀2", "population": 0, "eventStatus": "PRE_COLLECTION",
     "createdDate": "2022-02-20T00:00:00.000+0000", "lastEventDate": null, "role": "MEMBER"}
  ]
}`)

	rs, err := c.Pixels.List(t.Context(), 1111111111)
	if err != nil {
		t.Fatal(err)
	}
	assertRequest(t, srv, "1111111111")
	if len(rs) != 2 || rs[0].TrackID != "200000000000001" || rs[0].TrackName != "픽셀1" || rs[0].EventStatus != EventCollecting ||
		rs[0].Role != TrackerMaster || !rs[0].LastEventDate.Equal(time.Date(2022, 2, 20, 20, 0, 0, 0, time.UTC)) ||
		rs[1].EventStatus != EventPreCollection || !rs[1].LastEventDate.IsZero() || rs[1].CreatedDate.IsZero() {
		t.Errorf("rights = %+v", rs)
	}
}

func TestPixelService_Available(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v2/trackers/rightAvailables", 200, `[
  {"trackId": "200000000000001", "name": "픽셀1", "createdAt": "2022-02-20T00:00:00.000+0000", "updatedAt": "2022-02-20T20:00:00.000+0000", "role": "MEMBER"},
  {"trackId": "200000000000002", "name": "픽셀2", "createdAt": "2022-02-20T00:00:00.000+0000", "updatedAt": null, "role": "REQUEST"}
]`)

	ts, err := c.Pixels.Available(t.Context(), 1111111111)
	if err != nil {
		t.Fatal(err)
	}
	assertRequest(t, srv, "1111111111")
	if len(ts) != 2 || ts[0].TrackID != "200000000000001" || ts[0].Name != "픽셀1" || ts[0].Role != TrackerMember ||
		ts[0].UpdatedAt.Hour() != 20 || ts[1].Role != TrackerRequest || !ts[1].UpdatedAt.IsZero() {
		t.Errorf("available = %+v", ts)
	}
}

func TestPixelService_Available_wrapped(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v2/trackers/rightAvailables", 200, `{"RightAvailables": [{"trackId": "1", "trackName": "픽셀1", "role": "MASTER"}]}`)

	ts, err := c.Pixels.Available(t.Context(), 1111111111)
	if err != nil {
		t.Fatal(err)
	}
	if len(ts) != 1 || ts[0].TrackID != "1" || ts[0].Name != "픽셀1" || ts[0].Role != TrackerMaster {
		t.Errorf("available = %+v", ts)
	}
}

func TestPixelService_Available_bad(t *testing.T) {
	t.Parallel()
	for _, body := range []string{`{"rightAvailables": 1}`, `[1]`, `[{"createdAt": "yesterday"}]`} {
		c, srv := newTestClient(t)
		srv.HandleJSON("GET /openapi/v2/trackers/rightAvailables", 200, body)
		if _, err := c.Pixels.Available(t.Context(), 1); err == nil {
			t.Errorf("%s: want error", body)
		}
	}
}

func TestPixelService_writes(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v2/trackers", 200, ``)
	srv.HandleJSON("DELETE /openapi/v2/trackers/200000000000001", 200, ``)
	ctx := t.Context()

	if err := c.Pixels.Link(ctx, 1111111111, "200000000000001"); err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111111111")
	if string(req.Body) != `{"trackId":"200000000000001"}` {
		t.Errorf("body = %s", req.Body)
	}

	if err := c.Pixels.Unlink(ctx, 1111111111, "200000000000001"); err != nil {
		t.Fatal(err)
	}
	if req := assertRequest(t, srv, "1111111111"); req.Method != http.MethodDelete {
		t.Errorf("method = %s", req.Method)
	}

	if err := c.Pixels.Link(ctx, 1, ""); err == nil {
		t.Error("Link: want error")
	}
	if err := c.Pixels.Unlink(ctx, 1, ""); err == nil {
		t.Error("Unlink: want error")
	}
}

func TestDateTime_UnmarshalJSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want time.Time
	}{
		{`"2022-02-20T00:00:00.000+0000"`, time.Date(2022, 2, 20, 0, 0, 0, 0, time.UTC)},
		{`"2022-02-20T09:00:00+09:00"`, time.Date(2022, 2, 20, 0, 0, 0, 0, time.UTC)},
		{`"2021-02-20T12:00:00.702463"`, time.Date(2021, 2, 20, 3, 0, 0, 702463000, time.UTC)},
		{`""`, time.Time{}},
		{`null`, time.Time{}},
	}
	for _, tt := range tests {
		var d DateTime
		if err := json.Unmarshal([]byte(tt.in), &d); err != nil || !d.Equal(tt.want) {
			t.Errorf("%s = %v, %v; want %v", tt.in, d.Time, err, tt.want)
		}
	}
	for _, in := range []string{`1`, `"2022-02-20"`} {
		var d DateTime
		if err := json.Unmarshal([]byte(in), &d); err == nil {
			t.Errorf("%s: want error", in)
		}
	}
}
