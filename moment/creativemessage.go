package moment

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	kakao "github.com/donghquinn/gokakao"
)

// MessageImage is the image of a message creative. Kakao sends either the
// Image shape (size, width, height) or the Thumbnail shape (fileSize,
// imageWidth, imageHeight), so both are decoded. Personalized message
// templates with an image variable only set ValueWithVariable.
type MessageImage struct {
	URL         string      `json:"url"`
	FileName    string      `json:"fileName"`
	Size        kakao.Int64 `json:"size"`
	FileSize    kakao.Int64 `json:"fileSize"`
	Width       int         `json:"width"`
	Height      int         `json:"height"`
	ImageWidth  int         `json:"imageWidth"`
	ImageHeight int         `json:"imageHeight"`
	MimeType    string      `json:"mimeType"`
	ImageHash   string      `json:"imageHash"`
	// ValueWithVariable is the image variable of a personalized message,
	// e.g. "${image_url1}".
	ValueWithVariable string `json:"valueWithVariable"`
}

// MessageCreativeElement is the message of a message or personalized
// message creative as returned by the creative APIs. It extends
// [MessageElement] with the image and thumbnails these responses carry.
type MessageCreativeElement struct {
	MessageElement
	// Image replaces MessageElement.Image; see MessageImage.
	Image            *MessageImage `json:"image"`
	Thumbnail        *Thumbnail    `json:"thumbnail"`
	MessageThumbnail *Thumbnail    `json:"messageThumbnail"`
}

// ImageVariable is a variable image of a personalized message template,
// e.g. {ValueWithVariable: "${image_url1}"}. The variable's value must be a
// URL from [PersonalizedMessageService.UploadImages].
type ImageVariable struct {
	ValueWithVariable string `json:"valueWithVariable"`
}

// NewButtonAssetGroup is a button of a message creative to create or
// update. Landing URLs must be http:// or https://; carousel formats only
// allow LANDING_URL.
type NewButtonAssetGroup struct {
	// Ordering is 0 or 1 for BASIC_TEXT_MESSAGE and 0 otherwise; carousel
	// buttons are numbered per card (card 1: 0 and 1, card 2: 2 and 3, ...).
	Ordering int `json:"ordering"`
	// Title is up to 8 characters; the first carousel commerce button must
	// be "구매하기".
	Title       string           `json:"title"`
	LandingType AssetLandingType `json:"landingType"`
	// PCLandingURL, for LANDING_URL, sends PC KakaoTalk to a separate URL.
	PCLandingURL     string `json:"pcLandingUrl,omitempty"`
	MobileLandingURL string `json:"mobileLandingUrl,omitempty"`
	// ChannelCouponID is for CHANNEL_COUPON; see [MessageService.Coupons].
	ChannelCouponID int64 `json:"channelCouponId,omitempty"`
	// ChannelPostID is for CHANNEL_POST; see [MessageService.Posts].
	ChannelPostID int64 `json:"channelPostId,omitempty"`
	// AdViewID is for AD_VIEW; see [MessageService.AdViews].
	AdViewID int64 `json:"adViewId,omitempty"`
}

// NewItemAssetGroup is a list item or carousel card of a message creative
// to create or update.
type NewItemAssetGroup struct {
	LandingType AssetLandingType `json:"landingType"`
	// Title is the item copy: up to 25 (first) or 30 characters for wide
	// lists, 25 for carousel commerce and 20 for carousel feed.
	Title string `json:"title,omitempty"`
	// Description is the carousel feed copy, up to 180 characters.
	Description string `json:"description,omitempty"`
	// PriceAmount, PriceCurrencyCode and DiscountedPriceAmount are required
	// for carousel commerce cards. Amounts are integers 0~99999999 and the
	// discount must be at least 1% below the price.
	PriceAmount           string       `json:"priceAmount,omitempty"`
	PriceCurrencyCode     CurrencyCode `json:"priceCurrencyCode,omitempty"`
	DiscountedPriceAmount string       `json:"discountedPriceAmount,omitempty"`
	MobileLandingURL      string       `json:"mobileLandingUrl,omitempty"`
	PCLandingURL          string       `json:"pcLandingUrl,omitempty"`
	ChannelPostID         int64        `json:"channelPostId,omitempty"`
	ChannelCouponID       int64        `json:"channelCouponId,omitempty"`
	// ImageFile is the item image for wide image, wide list, carousel
	// commerce (not the intro card) and carousel feed messages.
	ImageFile *UploadFile `json:"-"`
	// Image is a variable image instead of ImageFile, for personalized
	// message templates.
	Image *ImageVariable `json:"image,omitempty"`
}

// NewMessageElement is the message of a message creative to create or
// update (messageElement).
type NewMessageElement struct {
	// CreativeFormat defaults to the creative's Format and must match it.
	CreativeFormat MessageType `json:"creativeFormat"`
	// ProfileID is the KakaoTalk Channel profile ID; required.
	ProfileID string `json:"profileId"`
	// Title is the copy or title, depending on the format: basic text copy
	// (400 characters, 300 with an image), wide image copy (76), wide list
	// title (20) or carousel commerce intro title (20). Personalized
	// templates may use up to 1,000 characters with variables.
	Title string `json:"title,omitempty"`
	// Description is the carousel commerce intro copy, up to 50 characters.
	Description       string                `json:"description,omitempty"`
	ButtonAssetGroups []NewButtonAssetGroup `json:"buttonAssetGroups,omitempty"`
	ItemAssetGroups   []NewItemAssetGroup   `json:"itemAssetGroups,omitempty"`
	// ShareFlag shows the share button. It must be false for age-verified
	// ad groups and for personalized messages.
	ShareFlag bool `json:"shareFlag"`
	// AdFlag marks an advertising message (광고성 메시지).
	AdFlag bool `json:"adFlag"`
	// ImageFile is the image of basic text messages or of the carousel
	// commerce intro card; other formats use item images.
	ImageFile *UploadFile `json:"-"`
	// Image is a variable image instead of ImageFile, for personalized
	// basic text templates.
	Image *ImageVariable `json:"image,omitempty"`
	// CSInfo is the customer service phone number (고객센터 전화번호);
	// required.
	CSInfo string `json:"csInfo"`
	// HasIntro and the Intro fields describe the carousel commerce intro
	// card.
	HasIntro              bool   `json:"hasIntro,omitempty"`
	IntroLandingType      string `json:"introLandingType,omitempty"`
	IntroMobileLandingURL string `json:"introMobileLandingUrl,omitempty"`
	IntroPCLandingURL     string `json:"introPcLandingUrl,omitempty"`
	// MobileLandingURL and PCLandingURL are the "more" landing of carousel
	// commerce and carousel feed messages.
	MobileLandingURL string `json:"mobileLandingUrl,omitempty"`
	PCLandingURL     string `json:"pcLandingUrl,omitempty"`
}

// NewMessageCreative describes a message or personalized message creative
// to create. Premium video messages and messages with coupon book asset
// groups cannot be created through the API.
type NewMessageCreative struct {
	AdGroupID int64       `json:"adGroupId"`
	Format    MessageType `json:"format"`
	// Name is up to 150 characters; Kakao generates one when empty.
	Name           string            `json:"name,omitempty"`
	MessageElement NewMessageElement `json:"messageElement"`
}

// validate checks cr and defaults the element's format.
func (cr *NewMessageCreative) validate() error {
	if cr.AdGroupID == 0 {
		return errors.New("moment: ad group ID is required")
	}
	switch cr.Format {
	case MessageTypeBasicText, MessageTypeWide, MessageTypeWideList, MessageTypeCarouselCommerce, MessageTypeCarouselFeed:
	default:
		return errors.New("moment: format must be BASIC_TEXT_MESSAGE, WIDE_MESSAGE, WIDE_LIST_MESSAGE, CAROUSEL_COMMERCE_MESSAGE or CAROUSEL_FEED_MESSAGE")
	}
	e := &cr.MessageElement
	if e.CreativeFormat == "" {
		e.CreativeFormat = cr.Format
	}
	if e.CreativeFormat != cr.Format {
		return errors.New("moment: message element format must match the creative format")
	}
	if e.ProfileID == "" || e.CSInfo == "" {
		return errors.New("moment: profile ID and customer service number are required")
	}
	if e.ImageFile != nil && e.Image != nil {
		return errors.New("moment: set either an image file or an image variable")
	}
	for _, it := range e.ItemAssetGroups {
		if it.ImageFile != nil && it.Image != nil {
			return errors.New("moment: set either an item image file or an item image variable")
		}
	}
	return nil
}

// multipart returns the multipart/form-data body of v, whose JSON encoding
// holds the fields of cr.
func (cr *NewMessageCreative) multipart(v any) (*kakao.Multipart, error) {
	fields, err := creativeForm(v)
	if err != nil {
		return nil, err
	}
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
	if err := add("messageElement.imageFile", cr.MessageElement.ImageFile); err != nil {
		return nil, err
	}
	for i, it := range cr.MessageElement.ItemAssetGroups {
		if err := add("messageElement.itemAssetGroups["+strconv.Itoa(i)+"].imageFile", it.ImageFile); err != nil {
			return nil, err
		}
	}
	return &kakao.Multipart{Fields: fields, Files: files}, nil
}

// MessageCreativeChanges describes a message or personalized message
// creative update. Kakao replaces the message with the request, so start
// from [Creative.MessageChanges]. The format cannot change, and from 5
// minutes before sending only the name is applied.
type MessageCreativeChanges struct {
	// ID is the creative ID.
	// TODO(review): the update docs list no id field, but the shared PUT
	// endpoint needs one to identify the creative.
	ID int64 `json:"id"`
	NewMessageCreative
}

// MessageChanges returns the update that keeps the message creative cr as
// it is, for [CreativeService.UpdateMessage] and
// [CreativeService.UpdatePersonalizedMessage]. Image files are not carried
// over; set them again to replace the images. Image variables are kept.
func (cr *Creative) MessageChanges() MessageCreativeChanges {
	ch := MessageCreativeChanges{
		ID: int64(cr.ID),
		NewMessageCreative: NewMessageCreative{
			AdGroupID: int64(cr.AdGroupID),
			Format:    MessageType(cr.Format),
			Name:      cr.Name,
		},
	}
	m := cr.MessageElement
	if m == nil {
		return ch
	}
	e := NewMessageElement{
		CreativeFormat:        m.CreativeFormat,
		ProfileID:             m.ProfileID,
		Title:                 m.Title,
		Description:           m.Description,
		ShareFlag:             m.ShareFlag,
		AdFlag:                m.AdFlag,
		CSInfo:                m.CSInfo,
		HasIntro:              m.HasIntro,
		IntroLandingType:      m.IntroLandingType,
		IntroMobileLandingURL: m.IntroMobileLandingURL,
		IntroPCLandingURL:     m.IntroPCLandingURL,
	}
	if m.Image != nil && m.Image.ValueWithVariable != "" {
		e.Image = &ImageVariable{ValueWithVariable: m.Image.ValueWithVariable}
	}
	for _, b := range m.ButtonAssetGroups {
		e.ButtonAssetGroups = append(e.ButtonAssetGroups, NewButtonAssetGroup{
			Ordering:         int(b.Ordering),
			Title:            b.Title,
			LandingType:      b.LandingType,
			PCLandingURL:     b.PCLandingURL,
			MobileLandingURL: b.MobileLandingURL,
			ChannelCouponID:  int64(b.ChannelCouponID),
			ChannelPostID:    int64(b.ChannelPostID),
			AdViewID:         int64(b.AdViewID),
		})
	}
	for _, it := range m.ItemAssetGroups {
		e.ItemAssetGroups = append(e.ItemAssetGroups, NewItemAssetGroup{
			LandingType:           it.LandingType,
			Title:                 it.Title,
			Description:           it.Description,
			PriceAmount:           it.PriceAmount,
			PriceCurrencyCode:     it.PriceCurrencyCode,
			DiscountedPriceAmount: it.DiscountedPriceAmount,
			MobileLandingURL:      it.MobileLandingURL,
			PCLandingURL:          it.PCLandingURL,
		})
	}
	ch.MessageElement = e
	return ch
}

// sendMessageCreative creates (POST) or updates (PUT) a message creative as
// multipart/form-data, the only encoding the docs give.
func (s *CreativeService) sendMessageCreative(ctx context.Context, method string, adAccountID int64, cr *NewMessageCreative, v any) (*Creative, error) {
	m, err := cr.multipart(v)
	if err != nil {
		return nil, err
	}
	var res Creative
	if err := s.c.do(ctx, call{method: method, path: "/v4/creatives", adAccountID: adAccountID, multipart: m}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// CreateMessage creates a message creative under a KakaoTalk Channel reach
// campaign (카카오톡 채널 X 도달). Saving under an ad group whose send time
// has passed may send the message at once. The request is
// multipart/form-data with dotted field names. Kakao allows one call per
// second per user and ad account.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/creatives-message
func (s *CreativeService) CreateMessage(ctx context.Context, adAccountID int64, cr NewMessageCreative) (*Creative, error) {
	if err := cr.validate(); err != nil {
		return nil, err
	}
	return s.sendMessageCreative(ctx, http.MethodPost, adAccountID, &cr, cr)
}

// UpdateMessage updates a message creative; see [MessageCreativeChanges].
// Kakao allows one call per second per user and ad account.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/creatives-message
func (s *CreativeService) UpdateMessage(ctx context.Context, adAccountID int64, ch MessageCreativeChanges) (*Creative, error) {
	if ch.ID == 0 {
		return nil, errors.New("moment: creative ID is required")
	}
	if err := ch.validate(); err != nil {
		return nil, err
	}
	return s.sendMessageCreative(ctx, http.MethodPut, adAccountID, &ch.NewMessageCreative, ch)
}

// CopyMessage copies message creatives into an ad group and returns the
// copies. Premium video messages, coupon book messages and messages whose
// landing (coupon, post, ad view or business form) is no longer available
// cannot be copied. Kakao allows one call per 5 seconds per user and ad
// account.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/creatives-message
func (s *CreativeService) CopyMessage(ctx context.Context, adAccountID, adGroupID int64, creativeIDs []int64) ([]Creative, error) {
	return s.copyCreatives(ctx, adAccountID, adGroupID, creativeIDs)
}

// CreatePersonalizedMessage creates a personalized message template under a
// personalized message reach campaign (개인화 메시지 X 도달). Text, buttons
// and landing URLs may use variables such as ${user_name1}; images may be a
// fixed ImageFile or an Image variable. ShareFlag must be false. The
// request is multipart/form-data with dotted field names. Kakao allows one
// call per second per user and ad account.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/creatives-personalized-message
func (s *CreativeService) CreatePersonalizedMessage(ctx context.Context, adAccountID int64, cr NewMessageCreative) (*Creative, error) {
	if err := cr.validate(); err != nil {
		return nil, err
	}
	return s.sendMessageCreative(ctx, http.MethodPost, adAccountID, &cr, cr)
}

// UpdatePersonalizedMessage updates a personalized message template; see
// [MessageCreativeChanges]. Pending send requests of the template are
// stopped. Templates with coupon book asset groups cannot be updated.
// Kakao allows one call per second per user and ad account.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/creatives-personalized-message
func (s *CreativeService) UpdatePersonalizedMessage(ctx context.Context, adAccountID int64, ch MessageCreativeChanges) (*Creative, error) {
	if ch.ID == 0 {
		return nil, errors.New("moment: creative ID is required")
	}
	if err := ch.validate(); err != nil {
		return nil, err
	}
	return s.sendMessageCreative(ctx, http.MethodPut, adAccountID, &ch.NewMessageCreative, ch)
}
