//go:build integration

package kakao

import (
	"errors"
	"os"
	"testing"
	"time"
)

// Live tests against the real Kakao API. They only call read-only APIs.
// Run with the credentials you have; tests needing others are skipped:
//
//	KAKAO_REST_KEY=... [KAKAO_ADMIN_KEY=...] [KAKAO_ACCESS_TOKEN=...] \
//	  go test -tags=integration -run Integration -count=1 .
func integrationClient(t *testing.T, needAdmin bool) *Client {
	t.Helper()
	key := os.Getenv("KAKAO_REST_KEY")
	if key == "" {
		t.Skip("KAKAO_REST_KEY not set")
	}
	admin := os.Getenv("KAKAO_ADMIN_KEY")
	if needAdmin && admin == "" {
		t.Skip("KAKAO_ADMIN_KEY not set")
	}
	c, err := NewClient(WithRESTKey(key), WithAdminKey(admin), WithRetry(RetryPolicy{}))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func accessToken(t *testing.T) string {
	t.Helper()
	tok := os.Getenv("KAKAO_ACCESS_TOKEN")
	if tok == "" {
		t.Skip("KAKAO_ACCESS_TOKEN not set")
	}
	return tok
}

// check fails the test on err, but skips when the test app lacks the
// consent item or permission the API needs.
func check(t *testing.T, name string, err error) {
	t.Helper()
	switch {
	case err == nil:
	case errors.Is(err, ErrConsentRequired), errors.Is(err, ErrFeatureDisabled), errors.Is(err, ErrPermissionDenied):
		t.Logf("%s: skipped, app not configured for it: %v", name, err)
	default:
		t.Errorf("%s: %v", name, err)
	}
}

func TestIntegration_Search(t *testing.T) {
	c := integrationClient(t, false)
	ctx := t.Context()
	for name, call := range map[string]func() (int, error){
		"Web": func() (int, error) {
			r, err := c.Search.Web(ctx, "카카오", &SearchOptions{Size: 1})
			return count(r, err)
		},
		"Video": func() (int, error) {
			r, err := c.Search.Video(ctx, "카카오", &SearchOptions{Size: 1})
			return count(r, err)
		},
		"Image": func() (int, error) {
			r, err := c.Search.Image(ctx, "카카오", &SearchOptions{Size: 1})
			return count(r, err)
		},
		"Blog": func() (int, error) {
			r, err := c.Search.Blog(ctx, "카카오", &SearchOptions{Size: 1})
			return count(r, err)
		},
		"Cafe": func() (int, error) {
			r, err := c.Search.Cafe(ctx, "카카오", &SearchOptions{Size: 1})
			return count(r, err)
		},
		"Book": func() (int, error) {
			r, err := c.Search.Book(ctx, "미움받을 용기", &BookSearchOptions{Target: BookTargetTitle, Size: 1})
			return count(r, err)
		},
	} {
		if n, err := call(); err != nil || n == 0 {
			t.Errorf("%s: %d documents, %v", name, n, err)
		}
	}
}

func count[T any](r *SearchResult[T], err error) (int, error) {
	if err != nil {
		return 0, err
	}
	return len(r.Documents), nil
}

func TestIntegration_Map(t *testing.T) {
	c := integrationClient(t, false)
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
	if p, err := c.Map.TransformCoord(ctx, pangyo, &CoordOptions{InputCoord: WGS84, OutputCoord: WTM}); err != nil || p.X == 0 {
		t.Errorf("TransformCoord: %v, %+v", err, p)
	}
	if _, err := c.Map.SearchKeyword(ctx, "카카오", &PlaceSearchOptions{Center: &pangyo, Radius: 1000, Size: 1}); err != nil {
		t.Errorf("SearchKeyword: %v", err)
	}
	if _, err := c.Map.SearchCategory(ctx, CategoryCafe, &PlaceSearchOptions{Center: &pangyo, Radius: 500, Size: 1}); err != nil {
		t.Errorf("SearchCategory: %v", err)
	}
	dest := Point{X: 127.12629039752096, Y: 37.4199323570413}
	if r, err := c.Map.WalkingRoute(ctx, pangyo, dest, nil); err != nil || r.Properties.TotalDistance == 0 {
		t.Errorf("WalkingRoute: %v", err)
	}
	if img, err := c.Map.StaticMap(ctx, StaticMapParams{Width: 100, Height: 100, Center: &pangyo}); err != nil || len(img.Data) == 0 {
		t.Errorf("StaticMap: %v", err)
	}
}

func TestIntegration_OIDCKeys(t *testing.T) {
	c := integrationClient(t, false)
	if _, err := fetchJWKS(t.Context(), c); err != nil {
		t.Errorf("JWKS: %v", err)
	}
}

func TestIntegration_Admin(t *testing.T) {
	c := integrationClient(t, true)
	ctx := t.Context()
	_, err := c.User.UserIDs(ctx, &UserIDsOptions{Limit: 1})
	check(t, "UserIDs", err)
	now := time.Now()
	_, err = c.Calendar.Holidays(ctx, now, now.Add(30*24*time.Hour))
	check(t, "Holidays", err)
}

func TestIntegration_User(t *testing.T) {
	c := integrationClient(t, false)
	tok := accessToken(t)
	ctx := t.Context()

	info, err := c.User.TokenInfo(ctx, tok)
	if err != nil || info.ID == 0 {
		t.Fatalf("TokenInfo: %v, %+v", err, info)
	}
	u, err := c.User.Me(ctx, AccessToken(tok), nil)
	if err != nil || u.ID != info.ID {
		t.Errorf("Me: %v, %+v", err, u)
	}
	_, err = c.User.Scopes(ctx, AccessToken(tok))
	check(t, "Scopes", err)
	_, err = c.User.OIDCUserInfo(ctx, tok)
	check(t, "OIDCUserInfo", err)
	_, err = c.Talk.Profile(ctx, tok)
	check(t, "Talk.Profile", err)
	_, err = c.Talk.Friends(ctx, tok, &FriendsOptions{Limit: 1})
	check(t, "Talk.Friends", err)
	_, err = c.Channel.Relations(ctx, AccessToken(tok))
	check(t, "Channel.Relations", err)
	_, err = c.Calendar.Calendars(ctx, tok, "")
	check(t, "Calendar.Calendars", err)
}
