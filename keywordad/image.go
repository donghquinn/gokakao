package keywordad

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	kakao "github.com/donghquinn/gokakao"
)

// Image is an image for creative extension assets.
type Image struct {
	ID kakao.Int64 `json:"id"`
	// URL downloads the image.
	URL      string `json:"url"`
	FileName string `json:"fileName"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	// Size is the file size in bytes.
	Size         int64      `json:"size"`
	LastUsedDate kakao.Time `json:"lastUsedDate"`
}

// ImagePage is a page of images.
type ImagePage struct {
	Content []Image  `json:"content"`
	Page    PageInfo `json:"page"`
	// RecentlyUsed reports whether the list is ordered by last use.
	RecentlyUsed bool `json:"recentlyUsed"`
}

// ImageListOptions are the optional parameters of [ImageService.List].
type ImageListOptions struct {
	// Page is the 0-based page number.
	Page int
	// Limit is the page size; Kakao's default is 10 and maximum 1000.
	Limit int
	// RecentlyUsed orders by last use instead of newest upload.
	RecentlyUsed bool
}

// List returns a page of an ad account's images.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/image-mgmt
func (s *ImageService) List(ctx context.Context, adAccountID int64, opts *ImageListOptions) (*ImagePage, error) {
	q := url.Values{}
	if opts != nil {
		if opts.RecentlyUsed {
			q.Set("recentlyUsed", strconv.FormatBool(true))
		}
		limitQuery(q, opts.Page, opts.Limit)
	}
	var page ImagePage
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v1/images", adAccountID: adAccountID, query: q}, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get returns an image.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/image-mgmt
func (s *ImageService) Get(ctx context.Context, adAccountID, imageID int64) (*Image, error) {
	var img Image
	if err := s.c.do(ctx, call{method: http.MethodGet, path: idPath("/v1/images/%d", imageID), adAccountID: adAccountID}, &img); err != nil {
		return nil, err
	}
	return &img, nil
}

// Upload uploads a JPG or PNG image of at most 450x450 pixels and 400 KB.
// Kakao allows one image upload per 10 seconds.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/image-mgmt
func (s *ImageService) Upload(ctx context.Context, adAccountID int64, filename string, data []byte) (*Image, error) {
	if filename == "" || len(data) == 0 {
		return nil, errors.New("keywordad: image file name and data are required")
	}
	m := &kakao.Multipart{Files: []kakao.MultipartFile{{
		Field:       "imageFile",
		Filename:    filename,
		ContentType: http.DetectContentType(data),
		Data:        data,
	}}}
	var img Image
	if err := s.c.do(ctx, call{method: http.MethodPost, path: "/v1/images", adAccountID: adAccountID, multipart: m}, &img); err != nil {
		return nil, err
	}
	return &img, nil
}

// UploadURL registers the image at an http or https URL. The image limits
// and rate limit of [ImageService.Upload] apply.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/image-mgmt
func (s *ImageService) UploadURL(ctx context.Context, adAccountID int64, imageURL string) (*Image, error) {
	if imageURL == "" {
		return nil, errors.New("keywordad: image URL is required")
	}
	body := struct {
		ImageFileURL string `json:"imageFileUrl"`
	}{imageURL}
	var img Image
	if err := s.c.do(ctx, call{method: http.MethodPost, path: "/v1/images", adAccountID: adAccountID, body: body}, &img); err != nil {
		return nil, err
	}
	return &img, nil
}

// Delete deletes an image.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/image-mgmt
func (s *ImageService) Delete(ctx context.Context, adAccountID, imageID int64) error {
	return s.c.do(ctx, call{method: http.MethodDelete, path: idPath("/v1/images/%d", imageID), adAccountID: adAccountID}, nil)
}
