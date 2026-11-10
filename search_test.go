package kakao

import (
	"errors"
	"slices"
	"testing"
	"time"
)

func TestSearchService(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithRESTKey("rest"))
	srv.HandleFile("GET /v2/search/web", "testdata/search/web.json")
	srv.HandleFile("GET /v2/search/vclip", "testdata/search/vclip.json")
	srv.HandleFile("GET /v2/search/image", "testdata/search/image.json")
	srv.HandleFile("GET /v2/search/blog", "testdata/search/blog.json")
	srv.HandleFile("GET /v2/search/cafe", "testdata/search/cafe.json")
	srv.HandleFile("GET /v3/search/book", "testdata/search/book.json")
	ctx := t.Context()

	t.Run("web", func(t *testing.T) {
		res, err := c.Search.Web(ctx, "이효리", &SearchOptions{Sort: SortRecency, Page: 2, Size: 50})
		if err != nil {
			t.Fatal(err)
		}
		req := srv.LastRequest()
		if req.Header.Get("Authorization") != "KakaoAK rest" {
			t.Errorf("Authorization = %q", req.Header.Get("Authorization"))
		}
		if q := req.Query; q.Get("query") != "이효리" || q.Get("sort") != "recency" || q.Get("page") != "2" || q.Get("size") != "50" {
			t.Errorf("query = %v", q)
		}
		if res.Meta != (SearchMeta{TotalCount: 897323, PageableCount: 775, IsEnd: false}) {
			t.Errorf("meta = %+v", res.Meta)
		}
		d := res.Documents[0]
		if d.Title != "<b>이효리</b> - 나무위키" || d.URL == "" || d.Contents == "" {
			t.Errorf("doc = %+v", d)
		}
		if want := time.Date(2017, 6, 14, 0, 0, 0, 0, time.FixedZone("", 9*3600)); !d.Datetime.Equal(want) {
			t.Errorf("datetime = %v", d.Datetime)
		}
	})

	t.Run("nil options send only query", func(t *testing.T) {
		if _, err := c.Search.Web(ctx, "q", nil); err != nil {
			t.Fatal(err)
		}
		if q := srv.LastRequest().Query; len(q) != 1 || q.Get("query") != "q" {
			t.Errorf("query = %v", q)
		}
	})

	t.Run("video", func(t *testing.T) {
		res, err := c.Search.Video(ctx, "AOA", nil)
		if err != nil {
			t.Fatal(err)
		}
		d := res.Documents[0]
		if d.PlayTime != 185 || d.Author != "_SBS" || d.Thumbnail == "" || d.Datetime.IsZero() {
			t.Errorf("doc = %+v", d)
		}
	})

	t.Run("image", func(t *testing.T) {
		res, err := c.Search.Image(ctx, "설현", nil)
		if err != nil {
			t.Fatal(err)
		}
		d := res.Documents[0]
		if d.Width != 540 || d.Height != 457 || d.Collection != "news" || d.DisplaySitename != "한국경제TV" || d.ImageURL == "" || d.DocURL == "" || d.ThumbnailURL == "" {
			t.Errorf("doc = %+v", d)
		}
	})

	t.Run("blog", func(t *testing.T) {
		res, err := c.Search.Blog(ctx, "https://brunch.co.kr/@tourism 집짓기", nil)
		if err != nil {
			t.Fatal(err)
		}
		if !res.Meta.IsEnd || res.Documents[0].BlogName != "정란수의 브런치" {
			t.Errorf("res = %+v", res)
		}
		if got := srv.LastRequest().Query.Get("query"); got != "https://brunch.co.kr/@tourism 집짓기" {
			t.Errorf("query = %q", got)
		}
	})

	t.Run("cafe", func(t *testing.T) {
		res, err := c.Search.Cafe(ctx, "아이유", nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.Documents[0].CafeName != "쭉빵카페" {
			t.Errorf("doc = %+v", res.Documents[0])
		}
	})

	t.Run("book", func(t *testing.T) {
		res, err := c.Search.Book(ctx, "미움받을 용기", &BookSearchOptions{Target: BookTargetTitle, Sort: SortLatest, Size: 5})
		if err != nil {
			t.Fatal(err)
		}
		if q := srv.LastRequest().Query; q.Get("target") != "title" || q.Get("sort") != "latest" || q.Get("size") != "5" {
			t.Errorf("query = %v", q)
		}
		d := res.Documents[0]
		if d.Title != "미움받을 용기" || d.ISBN != "8996991341 9788996991342" || d.Price != 14900 || d.SalePrice != 13410 ||
			!slices.Equal(d.Authors, []string{"기시미 이치로", "고가 후미타케"}) || !slices.Equal(d.Translators, []string{"전경아"}) ||
			d.Publisher != "인플루엔셜" || d.Status != "정상판매" {
			t.Errorf("doc = %+v", d)
		}
	})

	t.Run("book nil options", func(t *testing.T) {
		if _, err := c.Search.Book(ctx, "x", nil); err != nil {
			t.Fatal(err)
		}
		if q := srv.LastRequest().Query; len(q) != 1 {
			t.Errorf("query = %v", q)
		}
	})
}

func TestSearchService_errors(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithRESTKey("rest"))
	srv.HandleJSON("GET /v2/search/web", 400, `{"errorType":"InvalidArgument","message":"page is more than max"}`)

	if _, err := c.Search.Web(t.Context(), "", nil); !errors.Is(err, errEmptyQuery) {
		t.Errorf("empty query: err = %v", err)
	}
	if _, err := c.Search.Book(t.Context(), "", nil); !errors.Is(err, errEmptyQuery) {
		t.Errorf("empty book query: err = %v", err)
	}
	_, err := c.Search.Web(t.Context(), "q", &SearchOptions{Page: 51})
	var e *Error
	if !errors.As(err, &e) || e.Type != "InvalidArgument" || e.Msg != "page is more than max" {
		t.Errorf("err = %v", err)
	}

	noKey, _ := newTestClient(t)
	if _, err := noKey.Search.Web(t.Context(), "q", nil); !errors.Is(err, ErrMissingCredential) {
		t.Errorf("no key: err = %v", err)
	}
}
