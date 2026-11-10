package keywordad

import kakao "github.com/donghquinn/gokakao"

// Shared types from the Keyword Ad reference (타입 정보).
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/reference

// Config is the on/off setting of an ad account, campaign, ad group,
// keyword or creative link.
type Config string

// Configs.
const (
	ConfigOn  Config = "ON"
	ConfigOff Config = "OFF"
	ConfigDel Config = "DEL" // deleted; only in responses
)

// MemberType is the user's role in an ad account.
type MemberType string

// Member types.
const (
	MemberMaster MemberType = "MASTER"
	MemberMember MemberType = "MEMBER"
)

// Status is the delivery status of a campaign, ad group, keyword or
// creative. ON means it can run; the OFF_BY_* values say why it cannot.
type Status string

// Statuses. Not every value applies to every object.
const (
	StatusOn                         Status = "ON"
	StatusOff                        Status = "OFF"
	StatusWaiting                    Status = "WAITING" // under review
	StatusDeleted                    Status = "DELETED"
	StatusOffByAdAccountAdminStop    Status = "OFF_BY_AD_ACCOUNT_ADMIN_STOP"
	StatusOffByAdAccount             Status = "OFF_BY_AD_ACCOUNT"
	StatusOffByAdAccountUserBalance  Status = "OFF_BY_AD_ACCOUNT_USER_BALANCE"
	StatusOffByNotConnectedBizWallet Status = "OFF_BY_NOT_CONNECTED_BIZ_WALLET"
	StatusOffByBizChannelAdmin       Status = "OFF_BY_BIZ_CHANNEL_ADMIN"
	StatusOffByBizChannelWaiting     Status = "OFF_BY_BIZ_CHANNEL_WAITING"
	StatusOffByCampaignDailyBudget   Status = "OFF_BY_CAMPAIGN_DAILY_BUDGET"
	StatusOffByCampaign              Status = "OFF_BY_CAMPAIGN"
	StatusOffByAdGroup               Status = "OFF_BY_AD_GROUP"
	StatusOffByAdGroupDailyBudget    Status = "OFF_BY_AD_GROUP_DAILY_BUDGET"
	StatusOffByAdGroupDate           Status = "OFF_BY_AD_GROUP_DATE"
	StatusOffByAdGroupHour           Status = "OFF_BY_AD_GROUP_HOUR"
	StatusOffByAdGroupDay            Status = "OFF_BY_AD_GROUP_DAY"
)

// ReviewStatus is the review state of a keyword, creative or biz channel.
type ReviewStatus string

// Review statuses.
const (
	ReviewApproved ReviewStatus = "APPROVED"
	ReviewWaiting  ReviewStatus = "WAITING"
	ReviewRejected ReviewStatus = "REJECTED" // 심사보류
)

// DeviceType is a device class for targeting and bid weights.
type DeviceType string

// Device types.
const (
	DeviceMobile DeviceType = "MOBILE"
	DevicePC     DeviceType = "PC"
)

// Placement is where premium link ads run.
type Placement string

// Placements.
const (
	PlacementSearchKakao    Placement = "PREMIUM_LINK_SEARCH_KAKAO"
	PlacementSearchPartner  Placement = "PREMIUM_LINK_SEARCH_PARTNER" // unavailable from 2026-08-26
	PlacementContentKakao   Placement = "PREMIUM_LINK_CONTENT_KAKAO"
	PlacementContentPartner Placement = "PREMIUM_LINK_CONTENT_PARTNER"
)

// EventStatus is the collection state of a Pixel & SDK event.
type EventStatus string

// Event statuses.
const (
	EventCollecting    EventStatus = "COLLECTION"
	EventPreCollection EventStatus = "PRE_COLLECTION"
)

// AssetType is a kind of creative extension asset.
type AssetType string

// Asset types.
const (
	AssetThumbnail       AssetType = "THUMBNAIL"
	AssetMultiThumbnail  AssetType = "MULTI_THUMBNAIL"
	AssetAdditionalTitle AssetType = "ADDITIONAL_TITLE"
	AssetAdditionalLink  AssetType = "ADDITIONAL_LINK"
	AssetSubject         AssetType = "SUBJECT"
	AssetPriceTable      AssetType = "PRICE_TABLE"
	AssetCalculate       AssetType = "CALCULATE"
	AssetPhoneNumber     AssetType = "PHONE_NUMBER"
	AssetTalkChannel     AssetType = "TALK_CHANNEL"
)

// MetricsGroup is a group of report metrics.
type MetricsGroup string

// Metrics groups.
const (
	MetricsBasic              MetricsGroup = "BASIC" // the only group for ad account reports
	MetricsAddition           MetricsGroup = "ADDITION"
	MetricsPixelSDKConversion MetricsGroup = "PIXEL_SDK_CONVERSION"
	MetricsKakaoFriend        MetricsGroup = "KAKAO_FRIEND" // talk channel search ads only
)

// Company is a business (사업자).
type Company struct {
	BusinessRegistrationNumber string `json:"businessRegistrationNumber"`
	Name                       string `json:"name"`
}

// AccessAccount is a login a reviewer can use to check a members-only
// site (검수계정).
type AccessAccount struct {
	ID       string `json:"id"`
	Password string `json:"password"`
}

// Address is a postal address.
type Address struct {
	BaseAddress   string `json:"baseAddress,omitempty"`
	DetailAddress string `json:"detailAddress,omitempty"`
	ZipCode       string `json:"zipCode,omitempty"`
}

// LandingInfo holds landing URLs. Use RspvLandingURL for a responsive page
// or the PC and mobile URLs.
type LandingInfo struct {
	RspvLandingURL   string `json:"rspvLandingUrl,omitempty"`
	PCLandingURL     string `json:"pcLandingUrl,omitempty"`
	MobileLandingURL string `json:"mobileLandingUrl,omitempty"`
}

// WebsiteURL holds website URLs. Use RspvURL for a responsive site or the
// PC and mobile URLs.
type WebsiteURL struct {
	RspvURL   string `json:"rspvUrl,omitempty"`
	PCURL     string `json:"pcUrl,omitempty"`
	MobileURL string `json:"mobileUrl,omitempty"`
}

// License is a medical advertising review number (의료광고 심의필).
type License struct {
	LicenseNumber    string `json:"licenseNumber,omitempty"`
	ReviewDocumentID string `json:"reviewDocumentId,omitempty"`
}

// BidStrategyType chooses where bids are set.
type BidStrategyType string

// Bid strategy types.
const (
	BidByAdGroup BidStrategyType = "AD_GROUP" // the ad group's default bid
	BidManual    BidStrategyType = "MANUAL"   // a fixed bid per keyword
)

// BidStrategy is a keyword bid strategy.
type BidStrategy struct {
	Type BidStrategyType `json:"type"`
	// BidAmount is the fixed bid in KRW, required for BidManual.
	BidAmount int `json:"bidAmount,omitempty"`
}

// Schedule is when an ad group runs.
type Schedule struct {
	// BeginDate and EndDate are "yyyy-MM-dd". BeginDate defaults to today
	// and cannot be in the past.
	BeginDate string `json:"beginDate,omitempty"`
	EndDate   string `json:"endDate,omitempty"`
	// DetailTime enables per-hour settings. When set, each day lists 24
	// "1" (show) or "0" (hide) values; an omitted day shows all hours.
	DetailTime    bool     `json:"detailTime,omitempty"`
	MondayTime    []string `json:"mondayTime,omitempty"`
	TuesdayTime   []string `json:"tuesdayTime,omitempty"`
	WednesdayTime []string `json:"wednesdayTime,omitempty"`
	ThursdayTime  []string `json:"thursdayTime,omitempty"`
	FridayTime    []string `json:"fridayTime,omitempty"`
	SaturdayTime  []string `json:"saturdayTime,omitempty"`
	SundayTime    []string `json:"sundayTime,omitempty"`
}

// TalkChannel is a KakaoTalk Channel.
type TalkChannel struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

// ReviewDocumentType says what a review document is for.
type ReviewDocumentType string

// Review document types.
const (
	ReviewDocumentCreative   ReviewDocumentType = "CREATIVE"
	ReviewDocumentBizChannel ReviewDocumentType = "BIZ_CHANNEL"
)

// ReviewDocument is an uploaded review document (심사서류).
type ReviewDocument struct {
	ID       kakao.Int64 `json:"id"`
	FileName string      `json:"fileName"`
	// RequiredDocumentID identifies the kind of document required for the
	// business category, e.g. 1 (사업자등록증).
	RequiredDocumentID   kakao.Int64        `json:"requiredDocumentId"`
	RequiredDocumentName string             `json:"requiredDocumentName"`
	Type                 ReviewDocumentType `json:"type"`
	// URL downloads the document.
	URL string `json:"url"`
}

// Assets are a creative's extension assets.
type Assets struct {
	Thumbnail       *ThumbnailAsset       `json:"thumbnail,omitempty"`
	MultiThumbnails []MultiThumbnailAsset `json:"multiThumbnails,omitempty"`
	AdditionalTitle string                `json:"additionalTitle,omitempty"`
	AdditionalLinks []LinkAsset           `json:"additionalLinks,omitempty"`
	Subject         *SubjectAsset         `json:"subject,omitempty"`
	PriceTables     []PriceTableAsset     `json:"priceTables,omitempty"`
	Calculate       *CalculateAsset       `json:"calculate,omitempty"`
	PhoneNumber     string                `json:"phoneNumber,omitempty"`
	TalkChannelID   string                `json:"talkChannelId,omitempty"`
}

// ThumbnailAsset is a thumbnail image.
type ThumbnailAsset struct {
	ImageID string `json:"imageId"`
	Text    string `json:"text,omitempty"`
}

// MultiThumbnailAsset is one image of a multi-thumbnail.
type MultiThumbnailAsset struct {
	ImageID     string       `json:"imageId"`
	Text        string       `json:"text,omitempty"`
	LandingInfo *LandingInfo `json:"landingInfo,omitempty"`
}

// LinkAsset is an additional link.
type LinkAsset struct {
	Text        string       `json:"text"`
	LandingInfo *LandingInfo `json:"landingInfo,omitempty"`
}

// SubjectKind is the label of a subject asset (말머리).
type SubjectKind string

// Subject kinds.
const (
	SubjectDiscount SubjectKind = "DISCOUNT"
	SubjectEvent    SubjectKind = "EVENT"
	SubjectAuth     SubjectKind = "AUTH"
	SubjectRelation SubjectKind = "RELATION"
)

// SubjectAsset is a subject label with text.
type SubjectAsset struct {
	Subject     SubjectKind  `json:"subject"`
	Text        string       `json:"text,omitempty"`
	LandingInfo *LandingInfo `json:"landingInfo,omitempty"`
}

// PriceRange says how a price table price applies.
type PriceRange string

// Price ranges.
const (
	PriceStatic PriceRange = "STATIC" // exactly the price
	PriceTo     PriceRange = "TO"     // up to the price
	PriceFrom   PriceRange = "FROM"   // from the price
)

// PriceTableAsset is one row of a price table.
type PriceTableAsset struct {
	Text        string       `json:"text"`
	Price       int          `json:"price"` // KRW
	Range       PriceRange   `json:"range"`
	LandingInfo *LandingInfo `json:"landingInfo,omitempty"`
}

// CalculateAsset is a "calculate" link.
type CalculateAsset struct {
	LandingInfo *LandingInfo `json:"landingInfo,omitempty"`
}
