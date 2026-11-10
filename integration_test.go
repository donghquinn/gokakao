//go:build integration

package kakao

import (
	"os"
	"testing"
)

// Live tests against the real Kakao API. Run with:
//
//	KAKAO_REST_KEY=... go test -tags=integration -run Integration -count=1 .
func integrationClient(t *testing.T) *Client {
	t.Helper()
	key := os.Getenv("KAKAO_REST_KEY")
	if key == "" {
		t.Skip("KAKAO_REST_KEY not set")
	}
	c, err := NewClient(WithRESTKey(key))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestIntegration_Search(t *testing.T) {
	c := integrationClient(t)
	ctx := t.Context()
	if res, err := c.Search.Web(ctx, "카카오", &SearchOptions{Size: 1}); err != nil || len(res.Documents) == 0 {
		t.Errorf("Web: %v, %+v", err, res)
	}
	if res, err := c.Search.Book(ctx, "미움받을 용기", &BookSearchOptions{Target: BookTargetTitle, Size: 1}); err != nil || len(res.Documents) == 0 {
		t.Errorf("Book: %v, %+v", err, res)
	}
}

func TestIntegration_Map(t *testing.T) {
	c := integrationClient(t)
	ctx := t.Context()
	pangyo := Point{X: 127.1086228, Y: 37.4012191}

	if res, err := c.Map.SearchAddress(ctx, "경기 성남시 분당구 판교역로 166", nil); err != nil || len(res.Documents) == 0 {
		t.Errorf("SearchAddress: %v, %+v", err, res)
	}
	if regions, err := c.Map.CoordToRegionCode(ctx, pangyo, nil); err != nil || len(regions) == 0 {
		t.Errorf("CoordToRegionCode: %v, %+v", err, regions)
	}
	if _, err := c.Map.CoordToAddress(ctx, pangyo, nil); err != nil {
		t.Errorf("CoordToAddress: %v", err)
	}
	if p, err := c.Map.TransformCoord(ctx, pangyo, WGS84, WTM); err != nil || p.X == 0 {
		t.Errorf("TransformCoord: %v, %+v", err, p)
	}
	if res, err := c.Map.SearchKeyword(ctx, "카카오", &PlaceSearchOptions{Center: &pangyo, Radius: 1000, Size: 1}); err != nil {
		t.Errorf("SearchKeyword: %v, %+v", err, res)
	}
	if _, err := c.Map.SearchCategory(ctx, CategoryCafe, &PlaceSearchOptions{Center: &pangyo, Radius: 500, Size: 1}); err != nil {
		t.Errorf("SearchCategory: %v", err)
	}
}
