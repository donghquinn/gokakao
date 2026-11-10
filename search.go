package kakao

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
)

// SearchService calls the Daum Search API.
// It authenticates with the REST API key.
//
// Docs: https://developers.kakao.com/docs/ko/daum-search/dev-guide
type SearchService struct {
	c *Client
}

// SearchSort orders search results.
type SearchSort string

// Search sort orders. Book search uses [SortLatest] instead of
// [SortRecency].
const (
	SortAccuracy SearchSort = "accuracy" // by relevance (default)
	SortRecency  SearchSort = "recency"  // newest first
	SortLatest   SearchSort = "latest"   // newest first, book search only
)

// SearchOptions are the optional parameters shared by the search APIs.
// Zero values are omitted, so Kakao's defaults apply.
type SearchOptions struct {
	Sort SearchSort
	// Page is the 1-based result page. The maximum depends on the API
	// (50 for most, 15 for video).
	Page int
	// Size is the number of documents per page. The maximum depends on
	// the API (50 for most, 30 for video, 80 for image).
	Size int
}

// BookTarget restricts which field a book search matches.
type BookTarget string

// Book search targets.
const (
	BookTargetTitle     BookTarget = "title"
	BookTargetISBN      BookTarget = "isbn"
	BookTargetPublisher BookTarget = "publisher"
	BookTargetPerson    BookTarget = "person"
)

// BookSearchOptions are the optional parameters of [SearchService.Book].
type BookSearchOptions struct {
	// Sort is SortAccuracy (default) or SortLatest.
	Sort   SearchSort
	Page   int
	Size   int
	Target BookTarget
}

// SearchMeta describes a page of search results.
type SearchMeta struct {
	// TotalCount is the number of matching documents.
	TotalCount int `json:"total_count"`
	// PageableCount is the number of documents that can be paged through.
	PageableCount int `json:"pageable_count"`
	// IsEnd reports whether this is the last page.
	IsEnd bool `json:"is_end"`
}

// SearchResult is a page of search results.
type SearchResult[T any] struct {
	Meta      SearchMeta `json:"meta"`
	Documents []T        `json:"documents"`
}

// WebDocument is a web search result.
type WebDocument struct {
	Title    string `json:"title"`
	Contents string `json:"contents"`
	URL      string `json:"url"`
	Datetime Time   `json:"datetime"`
}

// VideoDocument is a video search result.
type VideoDocument struct {
	Title     string `json:"title"`
	URL       string `json:"url"`
	Datetime  Time   `json:"datetime"`
	PlayTime  int    `json:"play_time"` // seconds
	Thumbnail string `json:"thumbnail"`
	Author    string `json:"author"`
}

// ImageDocument is an image search result.
type ImageDocument struct {
	Collection      string `json:"collection"`
	ThumbnailURL    string `json:"thumbnail_url"`
	ImageURL        string `json:"image_url"`
	Width           int    `json:"width"`
	Height          int    `json:"height"`
	DisplaySitename string `json:"display_sitename"`
	DocURL          string `json:"doc_url"`
	Datetime        Time   `json:"datetime"`
}

// BlogDocument is a blog search result.
type BlogDocument struct {
	Title     string `json:"title"`
	Contents  string `json:"contents"`
	URL       string `json:"url"`
	BlogName  string `json:"blogname"`
	Thumbnail string `json:"thumbnail"`
	Datetime  Time   `json:"datetime"`
}

// BookDocument is a book search result.
type BookDocument struct {
	Title    string `json:"title"`
	Contents string `json:"contents"`
	URL      string `json:"url"`
	// ISBN holds an ISBN-10, an ISBN-13, or both separated by a space.
	ISBN        string   `json:"isbn"`
	Datetime    Time     `json:"datetime"`
	Authors     []string `json:"authors"`
	Publisher   string   `json:"publisher"`
	Translators []string `json:"translators"`
	Price       int      `json:"price"`
	SalePrice   int      `json:"sale_price"`
	Thumbnail   string   `json:"thumbnail"`
	// Status is the sales status for display, e.g. "정상판매".
	Status string `json:"status"`
}

// CafeDocument is a Daum Cafe search result.
type CafeDocument struct {
	Title     string `json:"title"`
	Contents  string `json:"contents"`
	URL       string `json:"url"`
	CafeName  string `json:"cafename"`
	Thumbnail string `json:"thumbnail"`
	Datetime  Time   `json:"datetime"`
}

// Web searches web documents.
//
// Docs: https://developers.kakao.com/docs/ko/daum-search/dev-guide
func (s *SearchService) Web(ctx context.Context, query string, opts *SearchOptions) (*SearchResult[WebDocument], error) {
	return search[WebDocument](ctx, s.c, "/v2/search/web", query, opts)
}

// Video searches videos.
//
// Docs: https://developers.kakao.com/docs/ko/daum-search/dev-guide
func (s *SearchService) Video(ctx context.Context, query string, opts *SearchOptions) (*SearchResult[VideoDocument], error) {
	return search[VideoDocument](ctx, s.c, "/v2/search/vclip", query, opts)
}

// Image searches images.
//
// Docs: https://developers.kakao.com/docs/ko/daum-search/dev-guide
func (s *SearchService) Image(ctx context.Context, query string, opts *SearchOptions) (*SearchResult[ImageDocument], error) {
	return search[ImageDocument](ctx, s.c, "/v2/search/image", query, opts)
}

// Blog searches blog posts. To search within one blog, prefix the query
// with the blog URL and a space, e.g. "https://brunch.co.kr/@tourism 집짓기".
//
// Docs: https://developers.kakao.com/docs/ko/daum-search/dev-guide
func (s *SearchService) Blog(ctx context.Context, query string, opts *SearchOptions) (*SearchResult[BlogDocument], error) {
	return search[BlogDocument](ctx, s.c, "/v2/search/blog", query, opts)
}

// Cafe searches Daum Cafe posts.
//
// Docs: https://developers.kakao.com/docs/ko/daum-search/dev-guide
func (s *SearchService) Cafe(ctx context.Context, query string, opts *SearchOptions) (*SearchResult[CafeDocument], error) {
	return search[CafeDocument](ctx, s.c, "/v2/search/cafe", query, opts)
}

// Book searches books.
//
// Docs: https://developers.kakao.com/docs/ko/daum-search/dev-guide
func (s *SearchService) Book(ctx context.Context, query string, opts *BookSearchOptions) (*SearchResult[BookDocument], error) {
	var so *SearchOptions
	var extra url.Values
	if opts != nil {
		so = &SearchOptions{Sort: opts.Sort, Page: opts.Page, Size: opts.Size}
		if opts.Target != "" {
			extra = url.Values{"target": {string(opts.Target)}}
		}
	}
	return searchWith[BookDocument](ctx, s.c, "/v3/search/book", query, so, extra)
}

var errEmptyQuery = errors.New("kakao: empty search query")

func search[T any](ctx context.Context, c *Client, path, query string, opts *SearchOptions) (*SearchResult[T], error) {
	return searchWith[T](ctx, c, path, query, opts, nil)
}

func searchWith[T any](ctx context.Context, c *Client, path, query string, opts *SearchOptions, extra url.Values) (*SearchResult[T], error) {
	if query == "" {
		return nil, errEmptyQuery
	}
	q := url.Values{"query": {query}}
	if opts != nil {
		setString(q, "sort", string(opts.Sort))
		setInt(q, "page", opts.Page)
		setInt(q, "size", opts.Size)
	}
	for k, v := range extra {
		q[k] = v
	}
	var res SearchResult[T]
	err := c.do(ctx, &request{method: http.MethodGet, host: HostDAPI, path: path, auth: authRESTKey, query: q}, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// setString sets key to v unless v is empty.
func setString(q url.Values, key, v string) {
	if v != "" {
		q.Set(key, v)
	}
}

// setInt sets key to v unless v is zero.
func setInt(q url.Values, key string, v int) {
	if v != 0 {
		q.Set(key, strconv.Itoa(v))
	}
}
