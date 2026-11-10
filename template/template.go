// Package template builds KakaoTalk message templates (기본 템플릿) for the
// KakaoTalk Message API.
//
// Each template type marshals to the JSON template_object Kakao expects,
// including its object_type.
//
// Docs: https://developers.kakao.com/docs/ko/message-template/default
package template

import "encoding/json"

// Template is a default message template: [Feed], [List], [Commerce],
// [Location], [Text] or [Calendar].
type Template interface {
	// ObjectType returns the template's object_type, e.g. "feed".
	ObjectType() string
}

var (
	_ Template = Feed{}
	_ Template = List{}
	_ Template = Commerce{}
	_ Template = Location{}
	_ Template = Text{}
	_ Template = Calendar{}
)

// Link is where a content or button leads. At least one field is
// required. Web URL domains must be registered in the app settings.
type Link struct {
	// WebURL is used by KakaoTalk for PC.
	WebURL string `json:"web_url,omitempty"`
	// MobileWebURL is used by mobile KakaoTalk.
	MobileWebURL string `json:"mobile_web_url,omitempty"`
	// AndroidExecutionParams are appended to the Android app link. If
	// empty, MobileWebURL is used.
	AndroidExecutionParams string `json:"android_execution_params,omitempty"`
	// IOSExecutionParams are appended to the iOS app link. If empty,
	// MobileWebURL is used.
	IOSExecutionParams string `json:"ios_execution_params,omitempty"`
}

// Content is the main content of a message. At least one of Title,
// ImageURL or Description is required.
type Content struct {
	Title       string `json:"title,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`
	ImageWidth  int    `json:"image_width,omitempty"`
	ImageHeight int    `json:"image_height,omitempty"`
	// Description is shown below Title; together they take up to 4 lines.
	Description string `json:"description,omitempty"`
	Link        Link   `json:"link"`
}

// Button is a message button.
type Button struct {
	Title string `json:"title"`
	Link  Link   `json:"link"`
}

// Social shows engagement counts below the content.
type Social struct {
	LikeCount       int `json:"like_count,omitempty"`
	CommentCount    int `json:"comment_count,omitempty"`
	SharedCount     int `json:"shared_count,omitempty"`
	ViewCount       int `json:"view_count,omitempty"`
	SubscriberCount int `json:"subscriber_count,omitempty"`
}

// ItemContent is the item area of a feed message (feed type B).
type ItemContent struct {
	// ProfileText is shown in the header or profile area (up to 16 chars).
	ProfileText     string `json:"profile_text,omitempty"`
	ProfileImageURL string `json:"profile_image_url,omitempty"`
	// TitleImageText is the image item title (up to 2 lines, 24 chars).
	TitleImageText string `json:"title_image_text,omitempty"`
	TitleImageURL  string `json:"title_image_url,omitempty"`
	// TitleImageCategory is shown below the title (1 line, 14 chars).
	TitleImageCategory string `json:"title_image_category,omitempty"`
	// Items lists up to 5 text items.
	Items []Item `json:"items,omitempty"`
	// Sum and SumOp summarize the items, e.g. "Total" and "15000원". Both
	// are needed for the summary to show.
	Sum   string `json:"sum,omitempty"`
	SumOp string `json:"sum_op,omitempty"`
}

// Item is a text item of an [ItemContent].
type Item struct {
	Item   string `json:"item"`    // name, up to 6 chars
	ItemOp string `json:"item_op"` // value, e.g. a price
}

// CurrencyPosition places the currency unit of a [CommerceInfo].
type CurrencyPosition int

// Currency positions.
const (
	CurrencyAfter  CurrencyPosition = 0 // after the price (default)
	CurrencyBefore CurrencyPosition = 1 // before the price
)

// CommerceInfo is the product and price of a [Commerce] message.
type CommerceInfo struct {
	ProductName   string `json:"product_name,omitempty"`
	RegularPrice  int    `json:"regular_price"`
	DiscountPrice int    `json:"discount_price,omitempty"`
	DiscountRate  int    `json:"discount_rate,omitempty"`
	// FixedDiscountPrice cannot be combined with DiscountRate.
	FixedDiscountPrice int `json:"fixed_discount_price,omitempty"`
	// CurrencyUnit defaults to "원".
	CurrencyUnit         string           `json:"currency_unit,omitempty"`
	CurrencyUnitPosition CurrencyPosition `json:"currency_unit_position,omitempty"`
}

// Feed is an image-and-text message. Set ItemContent for feed type B.
type Feed struct {
	Content     Content      `json:"content"`
	ItemContent *ItemContent `json:"item_content,omitempty"`
	Social      *Social      `json:"social,omitempty"`
	// ButtonTitle replaces the default "자세히 보기" button title.
	ButtonTitle string `json:"button_title,omitempty"`
	// Buttons lists up to 2 buttons.
	Buttons []Button `json:"buttons,omitempty"`
}

// ObjectType implements [Template].
func (Feed) ObjectType() string { return "feed" }

// MarshalJSON implements [json.Marshaler].
func (t Feed) MarshalJSON() ([]byte, error) {
	type alias Feed
	return marshal(t, alias(t))
}

// List is a message with a header and 2-3 contents.
type List struct {
	HeaderTitle string    `json:"header_title"` // up to 200 chars
	HeaderLink  Link      `json:"header_link"`
	Contents    []Content `json:"contents"` // 2-3 contents
	ButtonTitle string    `json:"button_title,omitempty"`
	Buttons     []Button  `json:"buttons,omitempty"`
}

// ObjectType implements [Template].
func (List) ObjectType() string { return "list" }

// MarshalJSON implements [json.Marshaler].
func (t List) MarshalJSON() ([]byte, error) {
	type alias List
	return marshal(t, alias(t))
}

// Commerce is a product message with prices.
type Commerce struct {
	Content     Content      `json:"content"`
	Commerce    CommerceInfo `json:"commerce"`
	ButtonTitle string       `json:"button_title,omitempty"`
	Buttons     []Button     `json:"buttons,omitempty"`
}

// ObjectType implements [Template].
func (Commerce) ObjectType() string { return "commerce" }

// MarshalJSON implements [json.Marshaler].
func (t Commerce) MarshalJSON() ([]byte, error) {
	type alias Commerce
	return marshal(t, alias(t))
}

// Location is a message with an address that opens in Kakao Map.
type Location struct {
	// Address is e.g. "경기 성남시 분당구 판교역로 235".
	Address string `json:"address"`
	// AddressTitle is shown on the map view, e.g. "카카오판교오피스".
	AddressTitle string   `json:"address_title,omitempty"`
	Content      Content  `json:"content"`
	Social       *Social  `json:"social,omitempty"`
	ButtonTitle  string   `json:"button_title,omitempty"`
	Buttons      []Button `json:"buttons,omitempty"`
}

// ObjectType implements [Template].
func (Location) ObjectType() string { return "location" }

// MarshalJSON implements [json.Marshaler].
func (t Location) MarshalJSON() ([]byte, error) {
	type alias Location
	return marshal(t, alias(t))
}

// Text is a text-only message.
type Text struct {
	Text        string   `json:"text"` // up to 200 chars
	Link        Link     `json:"link"`
	ButtonTitle string   `json:"button_title,omitempty"`
	Buttons     []Button `json:"buttons,omitempty"`
}

// ObjectType implements [Template].
func (Text) ObjectType() string { return "text" }

// MarshalJSON implements [json.Marshaler].
func (t Text) MarshalJSON() ([]byte, error) {
	type alias Text
	return marshal(t, alias(t))
}

// CalendarIDType is the kind of ID in a [Calendar] message.
type CalendarIDType string

// Calendar ID types.
const (
	CalendarEvent        CalendarIDType = "event"    // a public event
	CalendarSubscription CalendarIDType = "calendar" // a subscription calendar
)

// Calendar is a message for a Talk Calendar public event or subscription
// calendar of the app's linked channel.
type Calendar struct {
	IDType  CalendarIDType `json:"id_type"`
	ID      string         `json:"id"`
	Content Content        `json:"content"`
	// Buttons adds at most 1 button to the default one.
	Buttons []Button `json:"buttons,omitempty"`
}

// ObjectType implements [Template].
func (Calendar) ObjectType() string { return "calendar" }

// MarshalJSON implements [json.Marshaler].
func (t Calendar) MarshalJSON() ([]byte, error) {
	type alias Calendar
	return marshal(t, alias(t))
}

// marshal encodes fields, a JSON object, with the object_type of t
// prepended.
func marshal(t Template, fields any) ([]byte, error) {
	body, err := json.Marshal(fields)
	if err != nil {
		return nil, err
	}
	head, err := json.Marshal(t.ObjectType())
	if err != nil {
		return nil, err
	}
	out := append([]byte(`{"object_type":`), head...)
	if len(body) > 2 { // not "{}"
		out = append(out, ',')
	}
	return append(out, body[1:]...), nil
}
