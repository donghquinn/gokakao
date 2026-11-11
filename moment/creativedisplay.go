package moment

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	kakao "github.com/donghquinn/gokakao"
)

// Slide is a slide item of an IMAGE_NATIVE or VIDEO_NATIVE creative
// (슬라이드, formerly AssetGroup). Landing URLs must be http:// or
// https://.
type Slide struct {
	// ImageFileURL (JSON requests) or ImageFile (multipart requests) is the
	// slide image: 500x500 pixels or larger (1:1), up to 500 KB, JPG or PNG.
	ImageFileURL string      `json:"imageFileUrl,omitempty"`
	ImageFile    *UploadFile `json:"-"`
	Title        string      `json:"title,omitempty"` // up to 25 characters; required
	PCLandingURL string      `json:"pcLandingUrl,omitempty"`
	// MobileLandingURL and RspvLandingURL are the mobile and responsive
	// landing URLs.
	MobileLandingURL string `json:"mobileLandingUrl,omitempty"`
	RspvLandingURL   string `json:"rspvLandingUrl,omitempty"`
	// Description is not allowed for IMAGE_NATIVE and VIDEO_NATIVE slides.
	Description string `json:"description,omitempty"`
	// PriceAmount and DiscountedPriceAmount are 0~9999999999: integers for
	// KRW and JPY, up to two decimals for USD and EUR. They require
	// PriceCurrencyCode, and DiscountedPriceAmount must be at least 1%
	// below PriceAmount.
	PriceAmount           string       `json:"priceAmount,omitempty"`
	PriceCurrencyCode     CurrencyCode `json:"priceCurrencyCode,omitempty"`
	DiscountedPriceAmount string       `json:"discountedPriceAmount,omitempty"`
}

// NewDisplayCreative describes a display creative to create.
//
// Media can be sent as http(s) URLs (the ...FileURL fields, sent as JSON)
// or as local files (the ...File fields, sent as multipart/form-data);
// setting any file switches the whole request to multipart. Which media are
// required depends on Format:
//
//   - IMAGE_BANNER: image (960x300, up to 300 KB, PNG-24) and AltText.
//   - IMAGE_NATIVE: profile image and image, Title, ProfileName,
//     Description, ActionButton and AltText.
//   - VIDEO_NATIVE: profile image and video, Title, ProfileName,
//     Description, ActionButton and one landing URL.
//
// At least one landing URL is required. RspvLandingURL cannot be combined
// with the others, except with MobileLandingURL in bizboard visit and
// conversion campaigns that optimize a Pixel & SDK event.
type NewDisplayCreative struct {
	AdGroupID int64          `json:"adGroupId"`
	Format    CreativeFormat `json:"format"`
	// Name is up to 50 characters; Kakao generates one when empty.
	Name string `json:"name,omitempty"`
	// AltText is the accessible description (소재 설명), up to 30
	// characters without special characters.
	AltText string `json:"altText,omitempty"`
	// LandingInfo is for creatives under bizboard (카카오 비즈보드)
	// campaigns.
	LandingInfo *LandingInfo `json:"landingInfo,omitempty"`
	// PCLandingURL is not allowed in bizboard campaigns.
	PCLandingURL     string `json:"pcLandingUrl,omitempty"`
	MobileLandingURL string `json:"mobileLandingUrl,omitempty"`
	RspvLandingURL   string `json:"rspvLandingUrl,omitempty"`

	ImageFileURL string      `json:"imageFileUrl,omitempty"`
	ImageFile    *UploadFile `json:"-"`
	// ProfileImageFileURL or ProfileImageFile is the profile image of
	// native creatives: 300x300 pixels or larger (1:1), up to 500 KB.
	ProfileImageFileURL string      `json:"profileImageFileUrl,omitempty"`
	ProfileImageFile    *UploadFile `json:"-"`
	// VideoFileURL or VideoFile is the video of VIDEO_NATIVE: 9:16, 16:9 or
	// 1:1, up to 1 GB, AVI, FLV or MP4.
	VideoFileURL string      `json:"videoFileUrl,omitempty"`
	VideoFile    *UploadFile `json:"-"`
	// UploadThumbnailImageFileURL or UploadThumbnailImageFile is the custom
	// thumbnail of VIDEO_NATIVE (맞춤 썸네일), up to 500 KB.
	UploadThumbnailImageFileURL string      `json:"uploadThumbnailImageFileUrl,omitempty"`
	UploadThumbnailImageFile    *UploadFile `json:"-"`
	// Slides are 1~10 slide items of native creatives.
	Slides []Slide `json:"slides,omitempty"`

	Title        string       `json:"title,omitempty"`       // up to 25 characters
	ProfileName  string       `json:"profileName,omitempty"` // up to 20 characters
	Description  string       `json:"description,omitempty"` // up to 45 characters
	ActionButton ActionButton `json:"actionButton,omitempty"`
	// ActionButtonFontColor is "#000000" or "#FFFFFF" and
	// ActionButtonBackgroundColor any color but #FFFFFF, #F0F0F0, #000000
	// and #FEE500; set both or neither.
	ActionButtonFontColor       string `json:"actionButtonFontColor,omitempty"`
	ActionButtonBackgroundColor string `json:"actionButtonBackgroundColor,omitempty"`

	// Opinion is a note for the review (의견), up to 1,000 characters.
	Opinion string `json:"opinion,omitempty"`
	// OpinionProofFileURLs (JSON requests) or OpinionProofFiles (multipart
	// requests) are up to 10 supporting files (증빙), 10 MB in total, JPG,
	// PNG or PDF.
	OpinionProofFileURLs []string     `json:"opinionProofFileUrlList,omitempty"`
	OpinionProofFiles    []UploadFile `json:"-"`

	FrequencyCapType FrequencyCapType `json:"frequencyCapType,omitempty"`
	// FrequencyCapTime (1~24 hours) and FrequencyCap (1~5 impressions) are
	// required for FrequencyCapDayImp.
	FrequencyCapTime int `json:"frequencyCapTime,omitempty"`
	FrequencyCap     int `json:"frequencyCap,omitempty"`
	// AILabeled shows the generative AI mark (AI 활용 표기). nil sends
	// nothing: false on create, the current value on update.
	AILabeled *bool `json:"isAiLabeled,omitempty"`
}

// multipart reports whether any file is set.
func (cr *NewDisplayCreative) multipart() bool {
	if cr.ImageFile != nil || cr.ProfileImageFile != nil || cr.VideoFile != nil ||
		cr.UploadThumbnailImageFile != nil || len(cr.OpinionProofFiles) > 0 {
		return true
	}
	for _, s := range cr.Slides {
		if s.ImageFile != nil {
			return true
		}
	}
	return false
}

// files returns the file parts of cr.
func (cr *NewDisplayCreative) files() ([]kakao.MultipartFile, error) {
	var files []kakao.MultipartFile
	add := func(field string, f *UploadFile) error {
		if f == nil {
			return nil
		}
		p, err := f.part(field)
		if err != nil {
			return err
		}
		files = append(files, p)
		return nil
	}
	for _, f := range []struct {
		field string
		file  *UploadFile
	}{
		{"imageFile", cr.ImageFile},
		{"profileImageFile", cr.ProfileImageFile},
		{"videoFile", cr.VideoFile},
		{"uploadThumbnailImageFile", cr.UploadThumbnailImageFile},
	} {
		if err := add(f.field, f.file); err != nil {
			return nil, err
		}
	}
	for i, s := range cr.Slides {
		if err := add("slides["+strconv.Itoa(i)+"].imageFile", s.ImageFile); err != nil {
			return nil, err
		}
	}
	for i := range cr.OpinionProofFiles {
		if err := add("opinionProofFileList["+strconv.Itoa(i)+"]", &cr.OpinionProofFiles[i]); err != nil {
			return nil, err
		}
	}
	return files, nil
}

// validate checks the inputs common to create and update.
func (cr *NewDisplayCreative) validate() error {
	if cr.AdGroupID == 0 {
		return errors.New("moment: ad group ID is required")
	}
	switch cr.Format {
	case CreativeFormatImageBanner, CreativeFormatImageNative, CreativeFormatVideoNative:
	default:
		return errors.New("moment: format must be IMAGE_BANNER, IMAGE_NATIVE or VIDEO_NATIVE")
	}
	if len(cr.Slides) > 10 {
		return errors.New("moment: at most 10 slides are allowed")
	}
	if len(cr.OpinionProofFileURLs)+len(cr.OpinionProofFiles) > 10 {
		return errors.New("moment: at most 10 opinion proof files are allowed")
	}
	if cr.multipart() && len(cr.OpinionProofFileURLs) > 0 {
		return errors.New("moment: opinion proof file URLs cannot be sent with file uploads; use OpinionProofFiles")
	}
	if cr.multipart() && cr.UploadThumbnailImageFileURL != "" {
		return errors.New("moment: a thumbnail URL cannot be sent with file uploads; use UploadThumbnailImageFile")
	}
	return nil
}

// displayCall builds the JSON or multipart request for v, whose JSON
// encoding holds the fields of cr.
func displayCall(method string, adAccountID int64, cr *NewDisplayCreative, v any) (call, error) {
	c := call{method: method, path: "/v4/creatives", adAccountID: adAccountID}
	if !cr.multipart() {
		c.body = v
		return c, nil
	}
	fields, err := creativeForm(v)
	if err != nil {
		return call{}, err
	}
	files, err := cr.files()
	if err != nil {
		return call{}, err
	}
	c.multipart = &kakao.Multipart{Fields: fields, Files: files}
	return c, nil
}

// CreateDisplay creates a display creative (IMAGE_BANNER, IMAGE_NATIVE or
// VIDEO_NATIVE). It sends JSON, or multipart/form-data when cr has files.
// Kakao allows one call per second per user and ad account.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/creatives-display
func (s *CreativeService) CreateDisplay(ctx context.Context, adAccountID int64, cr NewDisplayCreative) (*Creative, error) {
	if err := cr.validate(); err != nil {
		return nil, err
	}
	switch cr.Format {
	case CreativeFormatImageBanner:
		if cr.ImageFileURL == "" && cr.ImageFile == nil {
			return nil, errors.New("moment: IMAGE_BANNER requires an image")
		}
	case CreativeFormatImageNative:
		if (cr.ImageFileURL == "" && cr.ImageFile == nil) || (cr.ProfileImageFileURL == "" && cr.ProfileImageFile == nil) {
			return nil, errors.New("moment: IMAGE_NATIVE requires an image and a profile image")
		}
	case CreativeFormatVideoNative:
		if (cr.VideoFileURL == "" && cr.VideoFile == nil) || (cr.ProfileImageFileURL == "" && cr.ProfileImageFile == nil) {
			return nil, errors.New("moment: VIDEO_NATIVE requires a video and a profile image")
		}
	}
	c, err := displayCall(http.MethodPost, adAccountID, &cr, cr)
	if err != nil {
		return nil, err
	}
	var res Creative
	if err := s.c.do(ctx, c, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// DisplayCreativeChanges describes a display creative update. Kakao
// replaces the creative with the request, so required fields (such as
// AltText or Title) must be sent again; start from
// [Creative.DisplayChanges]. Omitted media keep the registered files, an
// empty Name keeps the name and a nil AILabeled keeps the AI mark.
type DisplayCreativeChanges struct {
	// ID is the original creative ID (원본 소재 번호).
	ID int64 `json:"id"`
	NewDisplayCreative
}

// DisplayChanges returns the update that keeps cr as it is, for
// [CreativeService.UpdateDisplay]. The main, profile, video and thumbnail
// media and the opinion proof files are left out, so Kakao keeps the
// registered ones; slides carry their current image URLs.
func (cr *Creative) DisplayChanges() DisplayCreativeChanges {
	ai := cr.AILabeled
	ch := DisplayCreativeChanges{
		ID: int64(cr.ID),
		NewDisplayCreative: NewDisplayCreative{
			AdGroupID:        int64(cr.AdGroupID),
			Format:           cr.Format,
			Name:             cr.Name,
			AltText:          cr.AltText,
			LandingInfo:      cr.LandingInfo,
			PCLandingURL:     cr.PCLandingURL,
			MobileLandingURL: cr.MobileLandingURL,
			RspvLandingURL:   cr.RspvLandingURL,
			Title:            cr.Title,
			ProfileName:      cr.ProfileName,
			Description:      cr.Description,
			ActionButton:     cr.ActionButton,
			Opinion:          cr.Opinion,
			FrequencyCapType: cr.FrequencyCapType,
			FrequencyCapTime: int(cr.FrequencyCapTime),
			FrequencyCap:     int(cr.FrequencyCap),
			AILabeled:        &ai,
		},
	}
	for _, g := range cr.AssetGroups {
		var img string
		if g.Image != nil {
			img = g.Image.URL
		}
		ch.Slides = append(ch.Slides, Slide{
			ImageFileURL:          img,
			Title:                 g.Title,
			PCLandingURL:          g.PCLandingURL,
			MobileLandingURL:      g.MobileLandingURL,
			RspvLandingURL:        g.RspvLandingURL,
			PriceAmount:           g.PriceAmount,
			PriceCurrencyCode:     g.PriceCurrencyCode,
			DiscountedPriceAmount: g.DiscountedPriceAmount,
		})
	}
	return ch
}

// UpdateDisplay updates a display creative; see [DisplayCreativeChanges].
// It sends JSON, or multipart/form-data when ch has files. Kakao allows one
// call per second per user and ad account.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/creatives-display
func (s *CreativeService) UpdateDisplay(ctx context.Context, adAccountID int64, ch DisplayCreativeChanges) (*Creative, error) {
	if ch.ID == 0 {
		return nil, errors.New("moment: creative ID is required")
	}
	if err := ch.validate(); err != nil {
		return nil, err
	}
	c, err := displayCall(http.MethodPut, adAccountID, &ch.NewDisplayCreative, ch)
	if err != nil {
		return nil, err
	}
	var res Creative
	if err := s.c.do(ctx, c, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// CopyDisplay copies up to 20 display creatives into an ad group and
// returns the copies. Only IMAGE_BANNER, IMAGE_NATIVE and VIDEO_NATIVE
// creatives of bizboard, display and video campaigns can be copied. Kakao
// allows one call per 5 seconds per user and ad account.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/creatives-display
func (s *CreativeService) CopyDisplay(ctx context.Context, adAccountID, adGroupID int64, creativeIDs []int64) ([]Creative, error) {
	if len(creativeIDs) > 20 {
		return nil, errors.New("moment: at most 20 creatives can be copied at once")
	}
	return s.copyCreatives(ctx, adAccountID, adGroupID, creativeIDs)
}

// OpinionProof is the review opinion and supporting files of a display
// creative (의견증빙).
type OpinionProof struct {
	ID      kakao.Int64 `json:"id"`
	Opinion string      `json:"opinion"`
	// Files are the supporting files (opinionProofFileList).
	Files []OpinionFile `json:"opinionProofFileList"`
}

// OpinionProof returns the review opinion and supporting files of a display
// creative.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/creatives-display
func (s *CreativeService) OpinionProof(ctx context.Context, adAccountID, creativeID int64) (*OpinionProof, error) {
	var p OpinionProof
	if err := s.c.do(ctx, call{method: http.MethodGet, path: idPath("/v4/creatives/%d/opinionProof", creativeID), adAccountID: adAccountID}, &p); err != nil {
		return nil, err
	}
	return &p, nil
}
