package moment

import (
	"encoding/json"

	kakao "github.com/donghquinn/gokakao"
)

// Shared types from the Kakao Moment type reference (타입 정보).
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/type-info

// ---------------------------------------------------------------------------
// Enums
// ---------------------------------------------------------------------------

// ActionButton is the call-to-action button of a creative (행동유도버튼).
type ActionButton string

// Action buttons.
const (
	ActionButtonLink        ActionButton = "LINK"        // 바로가기
	ActionButtonMore        ActionButton = "MORE"        // 알아보기
	ActionButtonReservation ActionButton = "RESERVATION" // 예약하기
	ActionButtonPurchase    ActionButton = "PURCHASE"    // 구매하기
	ActionButtonRequest     ActionButton = "REQUEST"     // 신청하기
	ActionButtonInquire     ActionButton = "INQUIRE"     // 문의하기
	ActionButtonInstall     ActionButton = "INSTALL"     // 설치하기
	ActionButtonRun         ActionButton = "RUN"         // 실행하기
	ActionButtonDownload    ActionButton = "DOWNLOAD"    // 다운로드
	ActionButtonLaunch      ActionButton = "LAUNCH"      // 사용하기
	ActionButtonAddFriend   ActionButton = "ADD_FRIEND"  // 채널 추가
	ActionButtonSubscribe   ActionButton = "SUBSCRIBE"   // 소식 받기
	ActionButtonMoreVideo   ActionButton = "MORE_VIDEO"  // 동영상 더보기
	ActionButtonJoin        ActionButton = "JOIN"        // 가입하기
	ActionButtonCoupon      ActionButton = "COUPON"      // 쿠폰 받기
	ActionButtonPlay        ActionButton = "PLAY"        // 게임하기
)

// AdGroupType is the kind of an ad group (광고그룹 유형).
type AdGroupType string

// Ad group types.
const (
	AdGroupTypeDisplay       AdGroupType = "DISPLAY"        // 디스플레이
	AdGroupTypeDirectMessage AdGroupType = "DIRECT_MESSAGE" // 메시지
)

// AdViewType is the layout of an ad view (애드뷰 타입).
type AdViewType string

// Ad view types.
const (
	AdViewCompact AdViewType = "COMPACT" // 콤팩트뷰
	AdViewFull    AdViewType = "FULL"    // 풀뷰
)

// AdViewTemplateType is the template of an ad view (애드뷰 템플릿 타입).
type AdViewTemplateType string

// Ad view template types.
const (
	AdViewTemplateImage    AdViewTemplateType = "IMAGE"    // 이미지
	AdViewTemplateVideo    AdViewTemplateType = "VIDEO"    // 동영상
	AdViewTemplateCarousel AdViewTemplateType = "CAROUSEL" // 캐러셀
	AdViewTemplateScroll   AdViewTemplateType = "SCROLL"   // 스크롤
)

// AppTrackingType is what an app tracker is used for. The docs list it as
// a table of String values.
type AppTrackingType string

// App tracking types.
const (
	AppTrackingRetargeting AppTrackingType = "RETARGETING" // 리타게팅
	AppTrackingConversion  AppTrackingType = "CONVERSION"  // 전환
	AppTrackingBoth        AppTrackingType = "BOTH"        // 리타게팅과 전환 모두
)

// AssetLandingType is the landing of a message asset group or group
// (랜딩 유형).
type AssetLandingType string

// Asset landing types. Item asset groups do not use AssetLandingAdView.
const (
	AssetLandingURL           AssetLandingType = "LANDING_URL"    // URL
	AssetLandingChannelCoupon AssetLandingType = "CHANNEL_COUPON" // 쿠폰
	AssetLandingChannelPost   AssetLandingType = "CHANNEL_POST"   // 소식
	AssetLandingAdView        AssetLandingType = "AD_VIEW"        // 애드뷰
)

// AudienceType says whether a targeting uses an audience (오디언스 사용 여부).
type AudienceType string

// Audience types.
const (
	AudienceNormal  AudienceType = "NORMAL"  // 일반 타겟팅 설정
	AudienceDisplay AudienceType = "DISPLAY" // 디스플레이 오디언스 관리기능 설정
	AudienceMessage AudienceType = "MESSAGE" // 메시지 오디언스 관리기능 설정
)

// BidStrategyTargetType is an automatic bidding option (자동 입찰 옵션).
type BidStrategyTargetType string

// Bid strategy target types.
const (
	BidStrategyTargetCPC               BidStrategyTargetType = "TARGET_CPC"                // CPC 비용 목표
	BidStrategyTargetCPA               BidStrategyTargetType = "TARGET_CPA"                // CPA 비용 목표
	BidStrategyTargetROAS              BidStrategyTargetType = "TARGET_ROAS"               // ROAS 목표
	BidStrategyMaximizeConversionValue BidStrategyTargetType = "MAXIMIZE_CONVERSION_VALUE" // 전환 가치 최대화
)

// BizFormFlowType is the kind of a business form (비즈니스폼 유형).
type BizFormFlowType string

// Business form flow types.
const (
	BizFormApply   BizFormFlowType = "APPLY"  // 바로응모형
	BizFormSurvey  BizFormFlowType = "SURVEY" // 설문조사형
	BizFormReserve BizFormFlowType = "RESERV" // 신청예약형
)

// CampaignType is the type of a campaign (캠페인 유형).
type CampaignType string

// Campaign types.
const (
	CampaignTypeTalkBizBoard         CampaignType = "TALK_BIZ_BOARD"          // 카카오 비즈보드
	CampaignTypeDisplay              CampaignType = "DISPLAY"                 // 디스플레이
	CampaignTypeVideo                CampaignType = "VIDEO"                   // 동영상
	CampaignTypeProductCatalog       CampaignType = "PRODUCT_CATALOG"         // 상품 카탈로그
	CampaignTypeTalkBizBoardReserved CampaignType = "TALK_BIZ_BOARD_RESERVED" // 카카오 비즈보드 CPT
	CampaignTypeFocusFullView        CampaignType = "FOCUS_FULL_VIEW"         // 포커스 풀뷰
	CampaignTypePCTalkRichPop        CampaignType = "PC_TALK_RICH_POP"        // 리치팝 올데이
	CampaignTypeProfileFullView      CampaignType = "PROFILE_FULL_VIEW"       // 프로필 풀뷰
	CampaignTypePCTalkBottom         CampaignType = "PC_TALK_BOTTOM"          // 포커스 보드
	CampaignTypeDaumShopping         CampaignType = "DAUM_SHOPPING"           // 다음쇼핑
	CampaignTypeTalkChannel          CampaignType = "TALK_CHANNEL"            // 카카오톡 채널
	CampaignTypePersonalMessage      CampaignType = "PERSONAL_MESSAGE"        // 개인화 메시지
)

// NetworkCategoryCode is a sub-category of the network placement
// (네트워크 게재지면 하위 목록 정보). The docs call it "Code".
type NetworkCategoryCode string

// Network category codes (IAB categories).
const (
	NetworkCategoryIAB1  NetworkCategoryCode = "IAB1"  // 예술 및 엔터테인먼트
	NetworkCategoryIAB10 NetworkCategoryCode = "IAB10" // 홈/정원
	NetworkCategoryIAB11 NetworkCategoryCode = "IAB11" // 법, 정부와 정치
	NetworkCategoryIAB12 NetworkCategoryCode = "IAB12" // 언론
	NetworkCategoryIAB13 NetworkCategoryCode = "IAB13" // 개인금융
	NetworkCategoryIAB14 NetworkCategoryCode = "IAB14" // 사회
	NetworkCategoryIAB15 NetworkCategoryCode = "IAB15" // 과학
	NetworkCategoryIAB16 NetworkCategoryCode = "IAB16" // 애완동물
	NetworkCategoryIAB17 NetworkCategoryCode = "IAB17" // 스포츠
	NetworkCategoryIAB18 NetworkCategoryCode = "IAB18" // 스타일/패션
	NetworkCategoryIAB19 NetworkCategoryCode = "IAB19" // 컴퓨터/기술
	NetworkCategoryIAB2  NetworkCategoryCode = "IAB2"  // 자동차
	NetworkCategoryIAB20 NetworkCategoryCode = "IAB20" // 여행
	NetworkCategoryIAB21 NetworkCategoryCode = "IAB21" // 부동산
	NetworkCategoryIAB22 NetworkCategoryCode = "IAB22" // 쇼핑
	NetworkCategoryIAB23 NetworkCategoryCode = "IAB23" // 종교
	NetworkCategoryIAB24 NetworkCategoryCode = "IAB24" // 미분류
	NetworkCategoryIAB3  NetworkCategoryCode = "IAB3"  // 산업
	NetworkCategoryIAB4  NetworkCategoryCode = "IAB4"  // 직업
	NetworkCategoryIAB5  NetworkCategoryCode = "IAB5"  // 교육
	NetworkCategoryIAB6  NetworkCategoryCode = "IAB6"  // 가족 및 육아
	NetworkCategoryIAB7  NetworkCategoryCode = "IAB7"  // 건강 및 운동
	NetworkCategoryIAB8  NetworkCategoryCode = "IAB8"  // 식음료
	NetworkCategoryIAB9  NetworkCategoryCode = "IAB9"  // 취미
)

// TalkSectionCode is a KakaoTalk sub-section of the placement
// (게재지면 카카오톡 하위 섹션 카테고리 목록 정보). The docs call it "Code".
type TalkSectionCode string

// KakaoTalk section codes.
const (
	TalkSectionChat          TalkSectionCode = "KKO99-1" // 채팅지면 노출
	TalkSectionOpenChat      TalkSectionCode = "KKO99-2" // 오픈채팅지면 노출
	TalkSectionFriends       TalkSectionCode = "KKO99-4" // 친구지면 노출
	TalkSectionFriendNews    TalkSectionCode = "KKO99-5" // 친구소식지면 노출
	TalkSectionShortForm     TalkSectionCode = "KKO99-6" // 숏폼지면 노출
	TalkSectionOther         TalkSectionCode = "KKO99-7" // 기타지면 노출
	TalkSectionKakaoShopping TalkSectionCode = "KKO99-8" // 카카오쇼핑지면 노출
)

// Config is the on/off setting of an ad account, campaign, ad group or
// creative.
type Config string

// Configs.
const (
	ConfigOn  Config = "ON"  // 활성화
	ConfigOff Config = "OFF" // 비활성화
	ConfigDel Config = "DEL" // 삭제
)

// CouponBookTitleType is the kind of a message coupon (쿠폰 유형).
type CouponBookTitleType string

// Coupon book title types.
const (
	CouponBookDiscountPrice CouponBookTitleType = "DISCOUNT_PRICE" // 할인금액
	CouponBookDiscountRate  CouponBookTitleType = "DISCOUNT_RATE"  // 할인율
	CouponBookFreeShipping  CouponBookTitleType = "FREE_SHIPPING"  // 배송비 할인
	CouponBookFreeGift      CouponBookTitleType = "FREE_GIFT"      // 무료 증정
	CouponBookUpgrade       CouponBookTitleType = "UPGRADE"        // 업그레이드
)

// CreativeReviewStatus is the review state of an original creative
// (원 소재 심사 상태).
type CreativeReviewStatus string

// Creative review statuses. The modification states appear only when the
// original creative is approved or stopped by an administrator.
const (
	ReviewApproved             CreativeReviewStatus = "APPROVED"              // 승인
	ReviewWaiting              CreativeReviewStatus = "WAITING"               // 심사중
	ReviewRejected             CreativeReviewStatus = "REJECTED"              // 보류 (심사보류 또는 관리자정지)
	ReviewModificationWaiting  CreativeReviewStatus = "MODIFICATION_WAITING"  // 수정사항심사중
	ReviewModificationRejected CreativeReviewStatus = "MODIFICATION_REJECTED" // 수정사항심사보류
)

// CurrencyCode is a currency (통화).
type CurrencyCode string

// Currency codes.
const (
	CurrencyKRW CurrencyCode = "KRW" // ₩
	CurrencyUSD CurrencyCode = "USD" // $
	CurrencyJPY CurrencyCode = "JPY" // ¥
	CurrencyEUR CurrencyCode = "EUR" // €
)

// CustomerDataStatus is the processing state of customer data (고객데이터).
type CustomerDataStatus string

// Customer data statuses.
const (
	CustomerDataStatusWaiting        CustomerDataStatus = "WAITING"         // 모수추출 대기중
	CustomerDataStatusComplete       CustomerDataStatus = "COMPLETE"        // 모수추출 완료
	CustomerDataStatusDelete         CustomerDataStatus = "DELETE"          // 삭제 또는 삭제중
	CustomerDataStatusError          CustomerDataStatus = "ERROR"           // 그 외 비정상
	CustomerDataStatusModifying      CustomerDataStatus = "MODIFYING"       // 수정된 모수 준비중
	CustomerDataStatusTransformError CustomerDataStatus = "TRANSFORM_ERROR" // URL 유형에서 CSV 파일 등록 실패
)

// CustomerDataFileType is the kind of identifiers in customer data or a
// friend group file.
type CustomerDataFileType string

// Customer data file types.
const (
	// CustomerDataFileADID is an advertising ID, or its SHA-256 hash taken
	// over the lowercase ID with dashes (광고 식별자).
	CustomerDataFileADID CustomerDataFileType = "ADID"
	// CustomerDataFilePhoneNumber is a phone number, digits or with
	// hyphens, optionally with the country code (전화번호).
	CustomerDataFilePhoneNumber CustomerDataFileType = "PHONE_NUMBER"
	// CustomerDataFileHashedPhoneNumber is the hex SHA-256 of the phone
	// number with country code and no leading zero, e.g. 821011111111
	// (해시 전화번호).
	CustomerDataFileHashedPhoneNumber CustomerDataFileType = "HASHED_PHONE_NUMBER"
	// CustomerDataFileHashedEmail is a hex SHA-256 email (해시 이메일).
	CustomerDataFileHashedEmail CustomerDataFileType = "HASHED_EMAIL"
	// CustomerDataFileAppUserID is a positive integer app user ID
	// (앱 유저 아이디).
	CustomerDataFileAppUserID CustomerDataFileType = "APP_USER_ID"
	// CustomerDataFileMessageRetarget is message recipients
	// (메시지 발송 대상자).
	CustomerDataFileMessageRetarget CustomerDataFileType = "MESSAGE_RETARGET"
)

// CustomerDataType is how customer data is registered.
type CustomerDataType string

// Customer data types.
const (
	CustomerDataCSVFile CustomerDataType = "CSV_FILE" // 파일로 등록
	CustomerDataURL     CustomerDataType = "URL"      // URL로 등록
)

// CustomerFileStatus is the processing state of a customer file (고객파일).
// Customer files are being retired in favor of customer data.
type CustomerFileStatus string

// Customer file statuses.
const (
	CustomerFileStatusWaiting   CustomerFileStatus = "WAITING"   // 모수추출 대기중
	CustomerFileStatusComplete  CustomerFileStatus = "COMPLETE"  // 모수추출 완료
	CustomerFileStatusDelete    CustomerFileStatus = "DELETE"    // 삭제 또는 삭제중
	CustomerFileStatusError     CustomerFileStatus = "ERROR"     // 그 외 비정상
	CustomerFileStatusModifying CustomerFileStatus = "MODIFYING" // 수정된 모수 준비중
)

// CustomTargetingType is the kind of a custom target (맞춤타겟 유형).
type CustomTargetingType string

// Custom targeting types.
const (
	CustomTargetingCustomerData    CustomTargetingType = "CUSTOMERDATA"    // 고객데이터
	CustomTargetingPlusFriendGroup CustomTargetingType = "PLUSFRIENDGROUP" // 친구그룹
	CustomTargetingPlusFriend      CustomTargetingType = "PLUSFRIEND"      // 카카오톡 채널친구
	CustomTargetingCohort          CustomTargetingType = "COHORT"          // 광고반응타겟
	CustomTargetingSyncApp         CustomTargetingType = "SYNCAPP"         // 카카오 로그인 사용자
)

// DatePreset is a predefined report period.
type DatePreset string

// Date presets.
const (
	DatePresetToday     DatePreset = "TODAY"      // 오늘 00시부터 현재 전 시간까지
	DatePresetYesterday DatePreset = "YESTERDAY"  // 어제 00시부터 24시까지
	DatePresetLast7Day  DatePreset = "LAST_7DAY"  // 어제까지 7일, 오늘 미포함
	DatePresetLast14Day DatePreset = "LAST_14DAY" // 어제까지 14일, 오늘 미포함
	DatePresetLast30Day DatePreset = "LAST_30DAY" // 어제까지 30일, 오늘 미포함
	DatePresetThisMonth DatePreset = "THIS_MONTH" // 이번달 1일부터 오늘까지
	DatePresetLastMonth DatePreset = "LAST_MONTH" // 지난달 1일부터 마지막일까지
)

// DeviceType is a device class.
type DeviceType string

// Device types.
const (
	DeviceAndroid DeviceType = "ANDROID" // 안드로이드
	DeviceIOS     DeviceType = "IOS"     // iOS
	DevicePC      DeviceType = "PC"      // PC
)

// ReportLevel is the level a report is aggregated at
// (보고서 조회 레벨 기준). The docs call it "Dimension".
type ReportLevel string

// Report levels. The comment is the key the level adds to [Dimensions].
const (
	ReportLevelAdAccount ReportLevel = "AD_ACCOUNT" // ad_account_id: 광고계정
	ReportLevelCampaign  ReportLevel = "CAMPAIGN"   // campaign_id: 캠페인
	ReportLevelAdGroup   ReportLevel = "AD_GROUP"   // ad_group_id: 광고그룹
	ReportLevelCreative  ReportLevel = "CREATIVE"   // creative_id: 소재
)

// ReportDimension is what a report is broken down by (보고서 조회 기준).
// The docs call it "Dimension". Every dimension also reports "total".
type ReportDimension string

// Report dimensions.
const (
	// DimensionCreativeFormat is the creative format (소재 형식): IMAGE_BANNER,
	// IMAGE_NATIVE, VIDEO_NATIVE, IMAGE_BOX, CATALOG_DYNAMIC,
	// SERVICE_CONTENT for display and BASIC_TEXT_MESSAGE, WIDE_LIST_MESSAGE,
	// WIDE_MESSAGE, CAROUSEL_COMMERCE_MESSAGE, CAROUSEL_FEED_MESSAGE,
	// PREMIUM_VIDEO_MESSAGE, CATALOG_MESSAGE, SPOTLIGHT_MESSAGE for messages.
	DimensionCreativeFormat ReportDimension = "CREATIVE_FORMAT"
	// DimensionPlacement is the placement (게재지면): KAKAO_TALK, DAUM,
	// KAKAO_STORY, KAKAO_SERVICE, NETWORK.
	DimensionPlacement ReportDimension = "PLACEMENT"
	// DimensionAge is the legacy age band ((구) 연령): 15, 20 (20~25),
	// 26, 31, 41, 51 (51~70), -1 (알 수 없음).
	DimensionAge ReportDimension = "AGE"
	// DimensionAgeBand is the age band (연령): 15, 20, 25, ... 65 in
	// five-year steps, 70 (70이상), 0 (알 수 없음).
	DimensionAgeBand ReportDimension = "AGE_BAND"
	// DimensionGender is the gender (성별): M, F, N (알 수 없음).
	DimensionGender ReportDimension = "GENDER"
	// DimensionAgeGender is legacy age and gender ((구) 연령+성별).
	DimensionAgeGender ReportDimension = "AGE_GENDER"
	// DimensionAgeBandGender is age band and gender (연령+성별).
	DimensionAgeBandGender ReportDimension = "AGE_BAND_GENDER"
	// DimensionLocation is the region (지역): A 강원, B 경기, C 경남, D 경북,
	// E 광주, F 대구, G 대전, H 부산, I 서울, Q 세종, J 울산, K 인천,
	// R 전남광주통합특별시, L 전남, M 전북, N 제주, O 충남, P 충북, Z 해외,
	// U 알 수 없음.
	DimensionLocation ReportDimension = "LOCATION"
	// DimensionDeviceType is the device (디바이스): PC, Android, iOS,
	// N/A (기타).
	DimensionDeviceType ReportDimension = "DEVICE_TYPE"
	// DimensionHour is the hour of day (시간대): "00" through "23".
	DimensionHour ReportDimension = "HOUR"
)

// Gender is a targeting gender (성별).
type Gender string

// Genders. Reports also use "N" (알 수 없음).
const (
	GenderMale   Gender = "M" // 남자
	GenderFemale Gender = "F" // 여자
)

// Goal is a campaign goal (캠페인 목표).
type Goal string

// Goals.
const (
	GoalVisiting   Goal = "VISITING"   // 방문
	GoalConversion Goal = "CONVERSION" // 전환
	GoalReach      Goal = "REACH"      // 도달
	GoalView       Goal = "VIEW"       // 조회
)

// InclusionType says whether a target is included or excluded.
type InclusionType string

// Inclusion types.
const (
	InclusionInclude InclusionType = "INCLUDE" // 포함
	InclusionExclude InclusionType = "EXCLUDE" // 제외
)

// Indicator is an ad reaction kind (반응 종류).
type Indicator string

// Indicators.
const (
	IndicatorPlay       Indicator = "PLAY"       // 영상 플레이
	IndicatorClick      Indicator = "CLICK"      // 클릭
	IndicatorConversion Indicator = "CONVERSION" // 전환
	IndicatorOpen       Indicator = "OPEN"       // OPEN
)

// LandingType is the landing of a creative (랜딩 유형).
type LandingType string

// Landing types.
const (
	LandingTypePFBot          LandingType = "PF_BOT"          // 챗봇
	LandingTypeAdView         LandingType = "AD_VIEW"         // 애드뷰
	LandingTypeMessage        LandingType = "MESSAGE"         // 메시지
	LandingTypeChannelWebview LandingType = "CHANNEL_WEBVIEW" // 채널웹뷰
	LandingTypeBizForm        LandingType = "BIZ_FORM"        // 비즈니스폼 (지원 종료 예정)
	LandingTypeBizFormPlus    LandingType = "BIZ_FORM_PLUS"   // 비즈니스폼+
	LandingTypePayBuy         LandingType = "PAY_BUY"         // 카카오페이 구매
)

// LocationType chooses the regions of a demographic targeting
// (지역 선택 유형).
type LocationType string

// Location types.
const (
	LocationTypeAll LocationType = "ALL" // 전체 (국내, 해외)
	// LocationTypeDomestic is domestic only (국내만).
	// TODO(review): the docs spell this "Domestic"; confirm it is not
	// "DOMESTIC".
	LocationTypeDomestic LocationType = "Domestic"
	LocationTypeArea     LocationType = "AREA" // 지역 선택
)

// MemberType is the user's role in an ad account.
type MemberType string

// Member types.
const (
	MemberMaster MemberType = "MASTER" // 마스터 권한
	MemberMember MemberType = "MEMBER" // 멤버 권한
)

// MessageType is the format of a message (메시지 유형).
type MessageType string

// Message types.
const (
	MessageTypeBasicText        MessageType = "BASIC_TEXT_MESSAGE"        // 기본 텍스트
	MessageTypeWide             MessageType = "WIDE_MESSAGE"              // 와이드 이미지
	MessageTypeWideList         MessageType = "WIDE_LIST_MESSAGE"         // 와이드 리스트
	MessageTypeCarouselCommerce MessageType = "CAROUSEL_COMMERCE_MESSAGE" // 캐러셀 커머스형
	MessageTypeCarouselFeed     MessageType = "CAROUSEL_FEED_MESSAGE"     // 캐러셀 피드형
	MessageTypePremiumVideo     MessageType = "PREMIUM_VIDEO_MESSAGE"     // 프리미엄 동영상
	MessageTypeCatalog          MessageType = "CATALOG_MESSAGE"           // 카탈로그
	MessageTypeSpotlight        MessageType = "SPOTLIGHT_MESSAGE"         // 스포트라이트
)

// MessageSendStatus is the sending state of a message (메시지 발송 상태).
type MessageSendStatus string

// Message send statuses.
const (
	MessageSendReady    MessageSendStatus = "READY"    // 발송 대기
	MessageSending      MessageSendStatus = "SENDING"  // 발송중
	MessageSendPause    MessageSendStatus = "PAUSE"    // 발송중지
	MessageSendFinished MessageSendStatus = "FINISHED" // 발송완료
	MessageSendStopped  MessageSendStatus = "STOPPED"  // 즉시종료
	MessageSendExpired  MessageSendStatus = "EXPIRED"  // 기간만료
	MessageSendSave     MessageSendStatus = "SAVE"     // 저장
)

// MessageSyncStatus is the sync state with the sending system
// (발송시스템과의 연동 상태).
type MessageSyncStatus string

// Message sync statuses.
const (
	MessageSyncReady   MessageSyncStatus = "READY"   // 연동 대기
	MessageSyncSuccess MessageSyncStatus = "SUCCESS" // 연동 성공
	MessageSyncFailed  MessageSyncStatus = "FAILED"  // 연동 실패
)

// MessageTargetingType is the audience of a message ad (타게팅 타입).
type MessageTargetingType string

// Message targeting types. The optimization types are for optimized
// message ads (최적화 메시지광고).
const (
	MessageTargetingWhole              MessageTargetingType = "WHOLE"               // 전체 친구
	MessageTargetingSmart              MessageTargetingType = "SMART"               // 추천 친구
	MessageTargetingDirect             MessageTargetingType = "DIRECT"              // 직접 설정
	MessageTargetingOptimizationWhole  MessageTargetingType = "OPTIMIZATION_WHOLE"  // 전체 친구 최적화
	MessageTargetingOptimizationDirect MessageTargetingType = "OPTIMIZATION_DIRECT" // 직접 설정 최적화
)

// MetricsGroup is a group of report metrics. The comments list the metric
// keys each group adds to [Metrics].
type MetricsGroup string

// Metrics groups.
const (
	// MetricsBasic (기본지표): imp, click, ctr, cost, video_play_3s, vtr.
	MetricsBasic MetricsGroup = "BASIC"
	// MetricsAddition (추가지표): reach, cost_per_imp, cost_per_click,
	// cost_per_reach, cost_per_video_play_3s.
	MetricsAddition MetricsGroup = "ADDITION"
	// MetricsMessage (메시지 기본지표): msg_send, msg_open, msg_click, cost,
	// msg_send_fail.
	MetricsMessage MetricsGroup = "MESSAGE"
	// MetricsMessageAddition (메시지 추가지표): cost_per_msg_send,
	// cost_per_msg_open, cost_per_msg_click, msg_open_rate, msg_click_rate.
	MetricsMessageAddition MetricsGroup = "MESSAGE_ADDITION"
	// MetricsPixelSDKConversion (픽셀 & SDK 전환 지표): for each of the
	// events cmpt_reg (가입 완료), view_cart (장바구니 보기), purchase (구매),
	// participation (잠재 고객), signup (서비스 신청), app_install (앱 설치),
	// add_to_wishlist (관심상품추가) and add_to_cart (장바구니추가), with
	// window W of 1d or 7d: conv_<event>_W, cost_per_conv_<event>_W and
	// conv_<event>_W_rate. Also conv_purchase_p_W (구매 금액),
	// conv_purchase_p_per_cost_W (ROAS) and
	// conv_estimated_purchase_p_per_cost_W (추정 ROAS).
	MetricsPixelSDKConversion MetricsGroup = "PIXEL_SDK_CONVERSION"
	// MetricsPlusFriend (카카오친구 지표): conv_kf_pf_add_W,
	// cost_per_conv_kf_pf_add_W and conv_kf_pf_add_W_rate for W of 1d
	// or 7d.
	MetricsPlusFriend MetricsGroup = "PLUS_FRIEND"
	// MetricsVideo (동영상 지표): video_play, cost_per_video_play_3s,
	// video_play_auto, video_play_touch, video_unmute, video_play_5s,
	// video_play_10s, video_play_15s, video_play_30s, video_play_60s,
	// video_play_25p, video_play_50p, video_play_75p, video_play_100p.
	MetricsVideo MetricsGroup = "VIDEO"
	// MetricsSlideClick (슬라이드 지표): slide1_click through slide10_click
	// and slide_others_click.
	MetricsSlideClick MetricsGroup = "SLIDE_CLICK"
	// MetricsMessageClick (메시지 클릭지표): msg_click_share,
	// msg_click_media, msg_click_title, msg_click_text, msg_play_start,
	// msg_play_Ns; msg_click_list1..4 and msg_play_listK_start /
	// msg_play_listK_Ns for K 1..4 (N in 3, 5, 10, 15, 30, 60);
	// msg_click_button1..2; msg_click_carousel_button1..6,
	// msg_click_carousel_share1..6, msg_click_carousel1..6_coupon;
	// msg_click_catalog_item1..7; msg_click_carousel_more,
	// msg_click_coupon, msg_click_others.
	MetricsMessageClick MetricsGroup = "MESSAGE_CLICK"
	// MetricsAdView (애드뷰 지표): ADV_CLICK_BUTTON1, ADV_CLICK_BUTTON2,
	// ADV_CLICK_BUTTON_SHARE and, for N 1..10, ADV_TH<N>_VIEW,
	// ADV_TH<N>_VIDEO_PLAY_AUTO, ADV_TH<N>_VIDEO_PLAY_TOUCH,
	// ADV_TH<N>_VIDEO_PLAY_3S.
	MetricsAdView MetricsGroup = "ADVIEW"
	// MetricsBizBoard (비즈보드 지표): EXPANDABLE_IMP,
	// MOTION_VIDEO_PLAY_AUTO, EXPANDABLE_VIDEO_PLAY_AUTO,
	// MOTION_VIDEO_PLAY_3S, EXPANDABLE_VIDEO_PLAY_3S, EXPANDABLE_CLICK_CTA,
	// EXPANDABLE_CLICK_IMAGE1, EXPANDABLE_VIEWABLE_GOODS_1..5,
	// EXPANDABLE_CLICK_GOODS_1..5.
	MetricsBizBoard MetricsGroup = "BIZ_BOARD"
	// MetricsSPB (보드 지표): channel_home_click, channel_add_click,
	// article1_click through article10_click.
	MetricsSPB MetricsGroup = "SPB"
	// MetricsTrigger (전면형 유도 지표): trigger_viewable.
	MetricsTrigger MetricsGroup = "TRIGGER"
)

// ObjectiveType is the kind of an ad objective (광고 목표 설정의 유형).
type ObjectiveType string

// Objective types.
const (
	ObjectiveTalkChannel ObjectiveType = "TALK_CHANNEL"  // 카카오톡채널
	ObjectivePixelAndSDK ObjectiveType = "PIXEL_AND_SDK" // 픽셀&SDK
)

// ObjectiveDetailType is the conversion to optimize for (최적화할 전환 종류).
type ObjectiveDetailType string

// Objective detail types.
const (
	ObjectiveDetailPurchase             ObjectiveDetailType = "PURCHASE"              // 구매
	ObjectiveDetailAppInstall           ObjectiveDetailType = "APP_INSTALL"           // 앱설치
	ObjectiveDetailCompleteRegistration ObjectiveDetailType = "COMPLETE_REGISTRATION" // 회원가입
	ObjectiveDetailSignUp               ObjectiveDetailType = "SIGN_UP"               // 서비스신청
	ObjectiveDetailCart                 ObjectiveDetailType = "CART"                  // 장바구니
	ObjectiveDetailParticipation        ObjectiveDetailType = "PARTICIPATION"         // 잠재고객
	ObjectiveDetailAddFriend            ObjectiveDetailType = "ADD_FRIEND"            // 채널 친구 추가; response only, for TALK_CHANNEL
)

// Operation combines ad reactions in an ad reaction target (연산 종류).
type Operation string

// Operations.
const (
	OperationOnly  Operation = "ONLY"  // ONLY
	OperationAnd   Operation = "AND"   // AND
	OperationMinus Operation = "MINUS" // MINUS
	OperationOr    Operation = "OR"    // OR
)

// OptimizationMethod is how an optimized message ad splits sends
// (메시지 최적화 방식).
type OptimizationMethod string

// Optimization methods.
const (
	OptimizationAICustom OptimizationMethod = "AI_CUSTOM" // AI 맞춤 최적화
	OptimizationRatio    OptimizationMethod = "RATIO"     // 비중 직접 설정
)

// OptimizationStatus is the machine-learning stage of a conversion
// campaign. It is null (the empty string) for other goals.
type OptimizationStatus string

// Optimization statuses.
const (
	OptimizationStatusLevel1  OptimizationStatus = "LEVEL_1" // 1단계 제한된 머신러닝
	OptimizationStatusLevel2  OptimizationStatus = "LEVEL_2" // 2단계 기본 머신러닝
	OptimizationStatusLevel3  OptimizationStatus = "LEVEL_3" // 3단계 심화된 머신러닝
	OptimizationStatusLevel4  OptimizationStatus = "LEVEL_4" // 4단계 최적의 머신러닝
	OptimizationStatusInvalid OptimizationStatus = "INVALID" // 조회 불가
	OptimizationStatusReady   OptimizationStatus = "READY"   // 학습 대기
)

// Pacing controls how fast a daily budget is spent (게재 속도).
type Pacing string

// Pacings.
const (
	PacingNone   Pacing = "NONE"   // 설정 해제
	PacingQuick  Pacing = "QUICK"  // 빠른 게재: 예산을 최대한 사용
	PacingNormal Pacing = "NORMAL" // 일반 게재: 일예산을 시간대별로 분할
)

// PeriodUnit is a week or month schedule unit (주/월 구분).
type PeriodUnit string

// Period units.
const (
	PeriodWeek  PeriodUnit = "WEEK"  // 주 단위
	PeriodMonth PeriodUnit = "MONTH" // 월 단위
)

// Placement is where ads run (게재지면).
type Placement string

// Placements.
const (
	PlacementKakaoTalk    Placement = "KAKAO_TALK"    // 카카오톡
	PlacementDaum         Placement = "DAUM"          // 다음
	PlacementKakaoService Placement = "KAKAO_SERVICE" // 카카오 서비스
	PlacementNetwork      Placement = "NETWORK"       // 네트워크
)

// PricingType is a billing model (과금 방식).
type PricingType string

// Pricing types.
const (
	PricingCPC  PricingType = "CPC"  // 클릭당 과금
	PricingCPM  PricingType = "CPM"  // 1,000회 노출당 과금
	PricingCPA  PricingType = "CPA"  // 클릭 후 24시간 내 친구추가 전환당 과금
	PricingCPV  PricingType = "CPV"  // 동영상 조회당 과금
	PricingCPMS PricingType = "CPMS" // 메시지 발송당 과금; 메시지 광고그룹
)

// SelectionType says whether all ages or genders are selected
// (전체 선택 유형).
type SelectionType string

// Selection types.
const (
	SelectionAll    SelectionType = "ALL"     // 전체 선택
	SelectionNotAll SelectionType = "NOT_ALL" // 부분 선택
)

// CampaignStatus is the delivery status of a campaign (Status: 캠페인).
type CampaignStatus string

// Campaign statuses.
const (
	CampaignStatusReady                       CampaignStatus = "READY"                               // 집행 예정
	CampaignStatusLive                        CampaignStatus = "LIVE"                                // 집행 중
	CampaignStatusFinished                    CampaignStatus = "FINISHED"                            // 집행 완료
	CampaignStatusOff                         CampaignStatus = "OFF"                                 // 사용자OFF
	CampaignStatusDeleted                     CampaignStatus = "DELETED"                             // 삭제
	CampaignStatusExceedDailyBudget           CampaignStatus = "EXCEED_DAILY_BUDGET"                 // 캠페인 일 예산 초과
	CampaignStatusPaused                      CampaignStatus = "PAUSED"                              // 일시 중지
	CampaignStatusNoAvailableCreative         CampaignStatus = "NO_AVAILABLE_CREATIVE"               // 집행 가능한 소재가 없음
	CampaignStatusCanceled                    CampaignStatus = "CANCELED"                            // 계약해지
	CampaignStatusSystemConfigExternalService CampaignStatus = "SYSTEM_CONFIG_EXTERNAL_SERVICE_STOP" // 연결 서비스 제한으로 운영불가
	CampaignStatusAdAccountUnavailable        CampaignStatus = "ADACCOUNT_UNAVAILABLE"               // 광고계정 운영불가
)

// AdGroupStatus is the delivery status of an ad group (Status: 광고그룹).
type AdGroupStatus string

// Ad group statuses.
const (
	AdGroupStatusReady                 AdGroupStatus = "READY"                    // 집행 예정
	AdGroupStatusLive                  AdGroupStatus = "LIVE"                     // 집행 중
	AdGroupStatusFinished              AdGroupStatus = "FINISHED"                 // 집행 완료
	AdGroupStatusOff                   AdGroupStatus = "OFF"                      // 사용자OFF
	AdGroupStatusDeleted               AdGroupStatus = "DELETED"                  // 삭제
	AdGroupStatusExceedDailyBudget     AdGroupStatus = "EXCEED_DAILY_BUDGET"      // 광고 그룹 일 예산 초과
	AdGroupStatusPaused                AdGroupStatus = "PAUSED"                   // 일시 중지
	AdGroupStatusNoAvailableCreative   AdGroupStatus = "NO_AVAILABLE_CREATIVE"    // 집행 가능한 소재가 없음
	AdGroupStatusCanceled              AdGroupStatus = "CANCELED"                 // 계약해지
	AdGroupStatusCampaignUnavailable   AdGroupStatus = "CAMPAIGN_UNAVAILABLE"     // 캠페인 운영불가
	AdGroupStatusSystemConfigAdminStop AdGroupStatus = "SYSTEM_CONFIG_ADMIN_STOP" // 관리자 정지
)

// CreativeStatus is the operating status of a creative (Status: 소재).
// Message creatives can only be OPERATING, INVALID_DATE, OFF, DELETED or
// ADGROUP_UNAVAILABLE.
type CreativeStatus string

// Creative statuses.
const (
	CreativeStatusOperating             CreativeStatus = "OPERATING"                // 운영가능
	CreativeStatusUnapproved            CreativeStatus = "UNAPPROVED"               // 심사미승인 (심사중, 심사보류)
	CreativeStatusInvalidDate           CreativeStatus = "INVALID_DATE"             // 기간오류
	CreativeStatusMonitoringRejected    CreativeStatus = "MONITORING_REJECTED"      // 모니터링 보류
	CreativeStatusSystemConfigAdminStop CreativeStatus = "SYSTEM_CONFIG_ADMIN_STOP" // 관리자 정지
	CreativeStatusOff                   CreativeStatus = "OFF"                      // 사용자OFF
	CreativeStatusDeleted               CreativeStatus = "DELETED"                  // 삭제
	CreativeStatusAdGroupUnavailable    CreativeStatus = "ADGROUP_UNAVAILABLE"      // 광고그룹 운영불가
	CreativeStatusSystemConfigVoid      CreativeStatus = "SYSTEM_CONFIG_VOID"       // 운영불가 콘텐츠오류
)

// SystemConfig is a system-imposed state (시스템 설정 상태).
type SystemConfig string

// System configs.
const (
	SystemConfigOn                  SystemConfig = "ON"                    // 활성화
	SystemConfigVoid                SystemConfig = "VOID"                  // 운영불가 (시스템 무효처리)
	SystemConfigAdminStop           SystemConfig = "ADMIN_STOP"            // 관리자정지
	SystemConfigExternalServiceStop SystemConfig = "EXTERNAL_SERVICE_STOP" // 연결 서비스 제한
)

// TimeUnit is the aggregation unit of a report (보고서 집계단위). Both
// units allow a period of up to one month.
type TimeUnit string

// Time units.
const (
	TimeUnitDay TimeUnit = "DAY" // 일
	TimeUnitAll TimeUnit = "ALL" // 조회 기간 전체
)

// ---------------------------------------------------------------------------
// Structs
// ---------------------------------------------------------------------------

// AdGroup is an ad group referenced by an ad reaction target.
type AdGroup struct {
	ID   kakao.Int64 `json:"id"`
	Name string      `json:"name"`
	// TODO(review): the docs type this "Enum: Status[]" without saying
	// which Status; ad group statuses are assumed.
	AdGroupStatus []AdGroupStatus `json:"adGroupStatus"`
}

// AdView is an ad view (애드뷰).
type AdView struct {
	// ID goes in LandingInfo.AdViewItem.ID when creating a creative.
	ID           kakao.Int64        `json:"id"`
	Name         string             `json:"name"`
	Type         AdViewType         `json:"type"`
	TemplateType AdViewTemplateType `json:"templateType"`
}

// AdViewItem references an ad view from a landing.
type AdViewItem struct {
	ID kakao.Int64 `json:"id"`
}

// AssetGroup is a slide of a creative (슬라이드).
type AssetGroup struct {
	AssetGroupID kakao.Int64 `json:"assetGroupId"`
	Ordering     int         `json:"ordering"`
	Image        *Image      `json:"image"`
	Title        string      `json:"title"`
	Description  string      `json:"description"` // 홍보 문구
	// PriceAmount and DiscountedPriceAmount are strings in this type.
	PriceAmount           string       `json:"priceAmount"`
	PriceCurrencyCode     CurrencyCode `json:"priceCurrencyCode"`
	DiscountedPriceAmount string       `json:"discountedPriceAmount"`
	RspvLandingURL        string       `json:"rspvLandingUrl"`
	PCLandingURL          string       `json:"pcLandingUrl"`
	MobileLandingURL      string       `json:"mobileLandingUrl"`
}

// BaseAd is the ad an ad reaction target is built from (광고 반응 기준).
type BaseAd struct {
	AdGroup   *AdGroup  `json:"adGroup"`
	Campaign  *Campaign `json:"campaign"`
	Operation Operation `json:"operation"` // ONLY, MINUS or AND
	// FirstIndicator and SecondIndicator are the reactions combined by
	// Operation.
	FirstIndicator  Indicator  `json:"firstIndicator"`
	SecondIndicator Indicator  `json:"secondIndicator"`
	MessageAd       *MessageAd `json:"messageAd"`
}

// BidStrategyTarget is an automatic bidding goal.
type BidStrategyTarget struct {
	Type BidStrategyTargetType `json:"type"`
	// Value is the cost or ROAS goal; 0 for
	// BidStrategyMaximizeConversionValue.
	Value kakao.Int64 `json:"value"`
}

// BusinessForm is a business form linked to Kakao Moment (비즈니스폼, being
// retired).
type BusinessForm struct {
	ID          kakao.Int64 `json:"id"`
	AdAccountID kakao.Int64 `json:"adAccountId"`
	// BizFormID goes in LandingInfo.BizFormID when creating a creative.
	BizFormID kakao.Int64     `json:"bizFormId"`
	Title     string          `json:"title"`
	FlowType  BizFormFlowType `json:"flowType"`
	// LinkedDate, BeginDate and EndDate are "yyyy-MM-dd'T'HH:mm:ss".
	LinkedDate string `json:"linkedDate"`
	BeginDate  string `json:"beginDate"`
	EndDate    string `json:"endDate"`
}

// Button is a message button. Landing URLs must be http:// or https://.
type Button struct {
	Title string `json:"title,omitempty"`
	// PCLandingURL, when set, sends PC KakaoTalk to a separate URL.
	PCLandingURL     string `json:"pcLandingUrl,omitempty"`
	MobileLandingURL string `json:"mobileLandingUrl,omitempty"`
	// ChannelPostID is for a CHANNEL_POST landing (채널 소식).
	ChannelPostID kakao.Int64 `json:"channelPostId,omitempty"`
	// ChannelCouponID is for a CHANNEL_COUPON landing (채널 쿠폰).
	ChannelCouponID kakao.Int64 `json:"channelCouponId,omitempty"`
	// AdViewID is for an AD_VIEW landing (애드뷰).
	AdViewID           kakao.Int64 `json:"adViewId,omitempty"`
	WebembedChannelURL string      `json:"webembedChannelUrl,omitempty"`
}

// ButtonAssetGroup is a button of a message creative.
type ButtonAssetGroup struct {
	Ordering         kakao.Int64      `json:"ordering"`
	Title            string           `json:"title"`
	RspvLandingURL   string           `json:"rspvLandingUrl"`
	MobileLandingURL string           `json:"mobileLandingUrl"`
	PCLandingURL     string           `json:"pcLandingUrl"`
	AdViewID         kakao.Int64      `json:"adViewId"`
	ChannelPostID    kakao.Int64      `json:"channelPostId"`
	ChannelCouponID  kakao.Int64      `json:"channelCouponId"`
	Thumbnail        string           `json:"thumbnail"`
	Highlighted      bool             `json:"highlighted"`
	LandingType      AssetLandingType `json:"landingType"`
}

// Carousel is a carousel card of a message. Landing URLs must be http://
// or https://.
type Carousel struct {
	Title            string `json:"title,omitempty"`
	Description      string `json:"description,omitempty"`
	ImageURL         string `json:"imageUrl,omitempty"`
	PCLandingURL     string `json:"pcLandingUrl,omitempty"`
	MobileLandingURL string `json:"mobileLandingUrl,omitempty"`
	// ShareFlag enables sharing; not allowed for age-verified messages.
	ShareFlag         bool         `json:"shareFlag"`
	Buttons           []Button     `json:"buttons,omitempty"`
	PriceAmount       int          `json:"priceAmount,omitempty"`
	PriceCurrencyCode CurrencyCode `json:"priceCurrencyCode,omitempty"`
	// DiscountedPriceAmount must be at least 1% below PriceAmount.
	DiscountedPriceAmount int `json:"discountedPriceAmount,omitempty"`
}

// Campaign is a campaign (캠페인).
type Campaign struct {
	ID               kakao.Int64      `json:"id"`
	AdAccountID      kakao.Int64      `json:"adAccountId"`
	Name             string           `json:"name"`
	CampaignTypeGoal CampaignTypeGoal `json:"campaignTypeGoal"`
	Objective        *Objective       `json:"objective"`
	// DailyBudgetAmount is in KRW; 0 means no limit (한도 없음).
	DailyBudgetAmount     kakao.Int64 `json:"dailyBudgetAmount"`
	Config                Config      `json:"config"`
	DailyBudgetAmountOver bool        `json:"isDailyBudgetAmountOver"`
	// StatusDescription combines Config and DailyBudgetAmountOver: 운영중,
	// 일 예산 초과, 사용자 OFF, "사용자 OFF, 일 예산 초과" or 삭제됨.
	StatusDescription string `json:"statusDescription"`
	TrackID           string `json:"trackId"` // 전환추적설정 번호
}

// CampaignTypeGoal is a campaign type and goal pair (캠페인 유형 X 목표).
type CampaignTypeGoal struct {
	CampaignType CampaignType `json:"campaignType"`
	Goal         Goal         `json:"goal"`
}

// CategoryData is a custom-target category (맞춤타겟 카테고리).
type CategoryData struct {
	// ID is used in Targeting.UfoInterests or Targeting.UfoBusinessTypes.
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Children []CategoryData `json:"children"`
}

// ChannelCoupon is a KakaoTalk Channel coupon (채널 쿠폰).
type ChannelCoupon struct {
	ID         kakao.Int64 `json:"id"`
	Title      string      `json:"title"`
	Content    string      `json:"content"`
	StatusCode string      `json:"statusCode"`
	Permalink  string      `json:"permalink"`
	EntryImage *EntryImage `json:"entryImage"`
	CreatedAt  string      `json:"createdAt"`
}

// ChannelCoupons references a channel coupon from a coupon book asset
// group.
type ChannelCoupons struct {
	ID string `json:"id"`
}

// ChannelPost is a KakaoTalk Channel post (채널 소식).
type ChannelPost struct {
	ID     kakao.Int64 `json:"id"`
	Title  string      `json:"title"`
	Status string      `json:"status"`
	// Adult reports an age-verified post (연령인증 소식).
	Adult         bool   `json:"adult"`
	Permalink     string `json:"permalink"`
	CreatedDate   string `json:"createdDate"`
	PublishedDate string `json:"publishedDate"`
}

// Clip is the video of a clip link.
type Clip struct {
	ID           kakao.Int64 `json:"id"`
	Duration     float64     `json:"duration"` // 영상 길이
	ThumbnailURL string      `json:"thumbnailUrl"`
	Open         bool        `json:"isOpen"` // 공개 클립
}

// ClipLink is a clip link (클립링크).
type ClipLink struct {
	ID           kakao.Int64 `json:"id"`
	DisplayTitle string      `json:"displayTitle"`
	PlayCount    int         `json:"playCount"`
	LikeCount    int         `json:"likeCount"`
	CommentCount int         `json:"commentCount"`
	// TODO(review): the docs type CreateTime and UpdateTime as "Datetime"
	// without a format; kept as strings until the format is confirmed.
	CreateTime string `json:"createTime"`
	UpdateTime string `json:"updateTime"`
	Clip       *Clip  `json:"clip"`
	LinkURL    string `json:"linkUrl"`
}

// CohortTargeting is an ad reaction target in a targeting (광고반응 타게팅).
type CohortTargeting struct {
	ID            kakao.Int64   `json:"id"`
	Name          string        `json:"name"`
	CohortID      kakao.Int64   `json:"cohortId"`
	InclusionType InclusionType `json:"inclusionType"`
}

// Company is a business (사업자).
type Company struct {
	BusinessRegistrationNumber string `json:"businessRegistrationNumber"`
	Name                       string `json:"name"`
}

// CouponBook is the coupon of a message. Landing URLs must be http:// or
// https://.
type CouponBook struct {
	CouponBookTitleType CouponBookTitleType `json:"couponBookTitleType,omitempty"`
	// CouponBookTitle depends on the type: up to 8 digits for
	// DISCOUNT_PRICE, up to 2 digits for DISCOUNT_RATE, none for
	// FREE_SHIPPING, text for FREE_GIFT and UPGRADE (7 or 12 characters
	// depending on the message type).
	CouponBookTitle string `json:"couponBookTitle,omitempty"`
	// Title is the description, 12 or 18 characters at most depending on
	// the message type.
	Title            string `json:"title,omitempty"`
	PCLandingURL     string `json:"pcLandingUrl,omitempty"`
	MobileLandingURL string `json:"mobileLandingUrl,omitempty"`
}

// CouponBookAssetGroup is the coupon of a message creative.
type CouponBookAssetGroup struct {
	CouponBookTitle     string `json:"couponBookTitle"`
	CouponBookTitleType string `json:"couponBookTitleType"`
	Title               string `json:"title"`
	// TODO(review): typed as a single ChannelCoupons object although the
	// name is plural; confirm whether it is an array.
	ChannelCoupons   *ChannelCoupons `json:"channelCoupons"`
	MobileLandingURL string          `json:"mobileLandingUrl"`
	PCLandingURL     string          `json:"pcLandingUrl"`
	Ordering         int             `json:"ordering"`
}

// CustomerDataTargeting is customer data in a targeting (고객데이터 타게팅).
type CustomerDataTargeting struct {
	ID            kakao.Int64        `json:"id"`
	InclusionType InclusionType      `json:"inclusionType"`
	Name          string             `json:"name"`
	Status        CustomerDataStatus `json:"status"`
}

// CustomerFileTargeting is a customer file in a targeting (고객파일 타게팅).
// Customer files are being retired; use customer data for new targets.
type CustomerFileTargeting struct {
	ID            kakao.Int64        `json:"id"`
	InclusionType InclusionType      `json:"inclusionType"`
	Name          string             `json:"name"`
	Status        CustomerFileStatus `json:"status"`
}

// CustomTargeting is a custom target of a message ad (맞춤타겟).
type CustomTargeting struct {
	Type          CustomTargetingType `json:"type"`
	ReferenceID   string              `json:"referenceId"`
	InclusionType InclusionType       `json:"inclusionType"`
	Name          string              `json:"name,omitempty"`
}

// Location is a province or metropolitan city (시/도, Depth1Location). The
// full list is in the administrative district CSV from the docs.
type Location struct {
	Value       string `json:"value"` // 시/도 키
	Description string `json:"description,omitempty"`
	Depth1Name  string `json:"depth1Name,omitempty"`
	// Deprecated reports a removed district; omitted otherwise.
	Deprecated bool `json:"deprecated,omitempty"`
}

// Depth2Location is a city, county or district (시/군/구).
type Depth2Location struct {
	Value       string `json:"value"` // 시/군/구 키
	Description string `json:"description,omitempty"`
	Depth1Name  string `json:"depth1Name,omitempty"`
	Depth2Name  string `json:"depth2Name,omitempty"`
	// Deprecated reports a removed district; omitted otherwise.
	Deprecated bool `json:"deprecated,omitempty"`
}

// Depth3Location is a town or neighborhood (동/읍/면).
type Depth3Location struct {
	Value       string `json:"value"` // 동/읍/면 키
	Description string `json:"description,omitempty"`
	Depth1Name  string `json:"depth1Name,omitempty"`
	Depth2Name  string `json:"depth2Name,omitempty"`
	Depth3Name  string `json:"depth3Name,omitempty"`
	// Deprecated reports a removed district; omitted otherwise.
	Deprecated bool `json:"deprecated,omitempty"`
}

// Demographics is the gender, age and region of a message targeting.
type Demographics struct {
	GenderType SelectionType `json:"genderType,omitempty"` // default ALL
	// Genders is only for GenderType NOT_ALL.
	Genders []Gender      `json:"genders,omitempty"`
	AgeType SelectionType `json:"ageType,omitempty"`
	// Ages is only for AgeType NOT_ALL: "15" (15~19), "20", "25", ... "65"
	// in five-year bands, "70" (70 이상).
	Ages         []string     `json:"ages,omitempty"`
	LocationType LocationType `json:"locationType,omitempty"`
	// The depth locations are only for LocationType AREA; any of them may
	// be set.
	Depth1Locations []Location       `json:"depth1Locations,omitempty"`
	Depth2Locations []Depth2Location `json:"depth2Locations,omitempty"`
	Depth3Locations []Depth3Location `json:"depth3Locations,omitempty"`
}

// Dimensions are the level and dimension values of a report row. Keys
// depend on the [ReportLevel] and [ReportDimension], e.g.
// {"creative_format": "IMAGE BANNER", "ad_account_id": "1234"}.
type Dimensions map[string]string

// EntryImage is the detail image of a channel coupon.
type EntryImage struct {
	Type         string `json:"type"` // image, video, ...
	URL          string `json:"url"`
	Thumbnail    string `json:"thumbnail"`
	PreviewImage string `json:"previewImage"` // video only
	PlayURL      string `json:"playUrl"`      // video only
	Name         string `json:"name"`
}

// FriendPeriod limits a message to users who became channel friends in the
// period (친구기간). Dates are "yyyy-MM-dd".
type FriendPeriod struct {
	StartDate string `json:"startDate,omitempty"`
	EndDate   string `json:"endDate,omitempty"`
}

// GroupInfo is a group (button or item) of a message creative.
type GroupInfo struct {
	Title              string           `json:"title"`
	Hash               string           `json:"hash"`
	Ordering           int              `json:"ordering"`
	Highlighted        bool             `json:"isHighlighted"`
	LandingType        AssetLandingType `json:"landingType"`
	Image              *ImageInfo       `json:"image"`
	PCLandingURL       string           `json:"pcLandingUrl"`
	MobileLandingURL   string           `json:"mobileLandingUrl"`
	ChannelPostID      kakao.Int64      `json:"channelPostId"`
	ChannelPostTitle   string           `json:"channelPostTitle"`
	ChannelCouponID    kakao.Int64      `json:"channelCouponId"`
	ChannelCouponTitle string           `json:"channelCouponTitle"`
	AdViewID           kakao.Int64      `json:"adViewId"`
}

// Image is an image of a creative.
type Image struct {
	Size     kakao.Int64 `json:"size"` // bytes
	URL      string      `json:"url"`
	FileName string      `json:"fileName"`
	Width    int         `json:"width"`
	Height   int         `json:"height"`
	MimeType string      `json:"mimeType"`
}

// ImageInfo is the image of a group.
type ImageInfo struct {
	URL         string      `json:"url"`
	ImageWidth  int         `json:"imageWidth"`
	ImageHeight int         `json:"imageHeight"`
	FileSize    kakao.Int64 `json:"fileSize"`
	FileName    string      `json:"fileName"`
}

// IntroCarousel is the intro card of a carousel commerce message. Landing
// URLs must be http:// or https://.
type IntroCarousel struct {
	Title            string `json:"title,omitempty"`       // up to 25 characters
	Description      string `json:"description,omitempty"` // up to 50 characters
	ImageURL         string `json:"imageUrl,omitempty"`
	PCLandingURL     string `json:"pcLandingUrl,omitempty"`
	MobileLandingURL string `json:"mobileLandingUrl,omitempty"`
}

// Item is an item of a message.
type Item struct {
	// ImageURL is the promotional image. For a KakaoTalk Channel video in
	// basic text, wide image or wide list items, the docs say to send
	// video.mediaId instead.
	// TODO(review): Item has no video field on the type page; add it once
	// the message docs define it.
	ImageURL string    `json:"imageUrl,omitempty"`
	Title    string    `json:"title,omitempty"`
	Landing  []Landing `json:"landing,omitempty"`
}

// ItemAssetGroup is a list item of a message creative.
type ItemAssetGroup struct {
	LandingType AssetLandingType `json:"landingType"`
	Ordering    int              `json:"ordering"`
	Title       string           `json:"title"`
	Description string           `json:"description"` // carousel feed only
	// PriceAmount and DiscountedPriceAmount are strings, for carousel
	// commerce cards.
	PriceAmount           string       `json:"priceAmount"`
	PriceCurrencyCode     CurrencyCode `json:"priceCurrencyCode"`
	DiscountedPriceAmount string       `json:"discountedPriceAmount"`
	// MobileLandingURL and PCLandingURL are set for LANDING_URL.
	MobileLandingURL string     `json:"mobileLandingUrl"`
	PCLandingURL     string     `json:"pcLandingUrl"`
	Image            *Image     `json:"image"`
	Thumbnail        *Thumbnail `json:"thumbnail"`
}

// Landing is the landing of a message item or "more" button.
type Landing struct {
	PCLandingURL       string      `json:"pcLandingUrl,omitempty"`
	MobileLandingURL   string      `json:"mobileLandingUrl,omitempty"`
	WebembedChannelURL string      `json:"webembedChannelUrl,omitempty"`
	ChannelPostID      kakao.Int64 `json:"channelPostId,omitempty"`
	ChannelCouponID    kakao.Int64 `json:"channelCouponId,omitempty"`
}

// LandingInfo is the landing of a display creative. Which fields apply
// depends on LandingType.
type LandingInfo struct {
	LandingType LandingType `json:"landingType,omitempty"`
	// ProfileID is the KakaoTalk Channel profile for PF_BOT and
	// CHANNEL_WEBVIEW.
	ProfileID string `json:"profileId,omitempty"`
	// Action is the chatbot keyword for PF_BOT.
	Action     string      `json:"action,omitempty"`
	AdViewItem *AdViewItem `json:"adViewItem,omitempty"` // AD_VIEW
	URL        string      `json:"url,omitempty"`        // CHANNEL_WEBVIEW
	// BizFormID is for BIZ_FORM (being retired).
	BizFormID     kakao.Int64 `json:"bizFormId,omitempty"`
	BizFormPlusID kakao.Int64 `json:"bizFormplusId,omitempty"` // BIZ_FORM_PLUS
	// ChannelPostID is for a CHANNEL_POST landing.
	// TODO(review): CHANNEL_POST is not among the documented LandingType
	// values.
	ChannelPostID kakao.Int64 `json:"channelPostId,omitempty"`
	ShopKey       string      `json:"shopKey,omitempty"`   // PAY_BUY
	ProductID     string      `json:"productId,omitempty"` // PAY_BUY
}

// Message is the content of a message ad. Required fields depend on Type.
type Message struct {
	Type          MessageType    `json:"type"`
	MessageTitle  string         `json:"messageTitle,omitempty"`
	Description   string         `json:"description,omitempty"`
	Items         []Item         `json:"items,omitempty"`
	Buttons       []Button       `json:"buttons,omitempty"`
	CouponBook    *CouponBook    `json:"couponBook,omitempty"`
	IntroCarousel *IntroCarousel `json:"introCarousel,omitempty"`
	Carousels     []Carousel     `json:"carousels,omitempty"`
	// TODO(review): the Catalog type is not defined on the type page;
	// kept as raw JSON.
	Catalog    []json.RawMessage `json:"catalog,omitempty"`
	MoreButton *Landing          `json:"moreButton,omitempty"`
	// ButtonTitle and LandingURL are for spotlight messages.
	ButtonTitle string `json:"buttonTitle,omitempty"`
	LandingURL  string `json:"landingUrl,omitempty"`
	// AdFlag marks an advertising message; false is informational.
	AdFlag bool `json:"adFlag"`
	// ShareFlag enables sharing; not allowed for age-verified messages.
	ShareFlag bool `json:"shareFlag"`
}

// MessageAd is a message creative referenced by an ad reaction target.
type MessageAd struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// TODO(review): the docs type this "Enum: Status[]" without saying
	// which Status; creative statuses are assumed.
	OpStatus []CreativeStatus `json:"opStatus"`
}

// MessageElement is a message creative (메시지 소재).
type MessageElement struct {
	ID             kakao.Int64 `json:"id"`
	AdAccountID    kakao.Int64 `json:"adAccountId"`
	ProfileID      string      `json:"profileId"`
	ProfileName    string      `json:"profileName"`
	Name           string      `json:"name"`
	ShareFlag      bool        `json:"shareFlag"`
	AdFlag         bool        `json:"adFlag"`
	CreativeFormat MessageType `json:"creativeFormat"`
	// Title is the copy or title; where it shows depends on the format.
	Title       string `json:"title"`
	Description string `json:"description"` // carousel commerce intro card
	// Image is set for basic text and the carousel commerce intro card.
	Image *Image `json:"image"`
	// TODO(review): the docs type ButtonAssetGroups, ItemAssetGroups and
	// CouponBookAssetGroups as single objects; plural names suggest arrays.
	ButtonAssetGroups     []ButtonAssetGroup     `json:"buttonAssetGroups"`
	ItemAssetGroups       []ItemAssetGroup       `json:"itemAssetGroups"`
	ThumbnailURL          string                 `json:"thumbnailUrl"`
	CSInfo                string                 `json:"csInfo"` // 고객센터 전화번호
	CouponBookAssetGroups []CouponBookAssetGroup `json:"couponBookAssetGroups"`
	// CreatedDate and LastModifiedDate are "yyyy-MM-dd'T'HH:mm:ss".
	CreatedDate      string `json:"createdDate"`
	LastModifiedDate string `json:"lastModifiedDate"`
	// The Intro fields describe the carousel commerce intro card.
	HasIntro              bool   `json:"hasIntro"`
	IntroLandingType      string `json:"introLandingType"`
	IntroMobileLandingURL string `json:"introMobileLandingUrl"`
	IntroPCLandingURL     string `json:"introPcLandingUrl"`
}

// MessageSendingInfo is the sending setup of a message ad.
type MessageSendingInfo struct {
	// ContractCount is the purchased send count (구매발송수); it cannot be
	// below the estimated audience.
	ContractCount int `json:"contractCount"`
	// LongTerm also sends to new friends after the full send. It requires
	// the schedule's begin and end dates and times.
	LongTerm bool        `json:"longTerm"`
	Price    kakao.Int64 `json:"price"` // 발송단가
	// PushAlarm true sends without a push notification (푸시알림 보내지 않기).
	PushAlarm bool `json:"pushAlarm"`
	// SendRate spreads the send: 100, 500, 1000, 1500, 2000, or 0 for no
	// spreading (분산발송).
	SendRate        int               `json:"sendRate"`
	Status          MessageSendStatus `json:"status,omitempty"`
	SyncStatus      MessageSyncStatus `json:"syncStatus,omitempty"`
	AgeVerification bool              `json:"ageVerification"`
	StartedAt       string            `json:"startedAT,omitempty"`  // 발송 시작 시간
	FinishedAt      string            `json:"finishedAT,omitempty"` // 발송 완료 시간
}

// Metrics are the metric values of a report row, keyed by the metric names
// of the requested [MetricsGroup] values, e.g. {"imp": 4, "ctr": 0.0}.
type Metrics map[string]kakao.Float

// MoreButton is the "more" landing of a carousel (더보기).
type MoreButton struct {
	PCLandingURL       string      `json:"pcLandingUrl,omitempty"`
	MobileLandingURL   string      `json:"mobileLandingUrl,omitempty"`
	WebembedChannelURL string      `json:"webembedChannelUrl,omitempty"`
	ChannelPostID      kakao.Int64 `json:"channelPostId,omitempty"`
	ChannelCouponID    kakao.Int64 `json:"channelCouponId,omitempty"`
}

// Objective is the ad objective of a campaign (광고 목표 설정).
type Objective struct {
	Type ObjectiveType `json:"type"`
	// Value is the KakaoTalk Channel profile ID for TALK_CHANNEL or the
	// Pixel & SDK ID for PIXEL_AND_SDK.
	Value string `json:"value"`
	// DetailType is the conversion to optimize for PIXEL_AND_SDK.
	DetailType ObjectiveDetailType `json:"detailType,omitempty"`
}

// OpinionFile is a supporting file for a review opinion (의견증빙 파일).
type OpinionFile struct {
	ID               kakao.Int64 `json:"id"`
	OriginalFileName string      `json:"originalFileName"`
	DownloadURL      string      `json:"downloadUrl"`
}

// OptimizationTarget is the send split of an optimized message ad
// (메시지 최적화 설정).
type OptimizationTarget struct {
	Type OptimizationMethod `json:"type"`
	// SendRatios are percentages in the order of the ad's messages, each
	// 10~90 in steps of 10 and summing to 100. Required for RATIO; omit
	// for AI_CUSTOM.
	SendRatios []int `json:"sendRatios,omitempty"`
}

// OptimizationTargeting is the targeting of an optimized message ad
// (최적화 메시지광고). The fields after OptimizationTarget only apply to
// MessageTargetingOptimizationDirect, which requires at least one of them.
type OptimizationTargeting struct {
	Type               MessageTargetingType `json:"type"`
	OptimizationTarget OptimizationTarget   `json:"optimizationTarget"`
	// Devices limits devices; empty means no limit.
	Devices       []DeviceType      `json:"devices,omitempty"`
	Demographics  *Demographics     `json:"demographics,omitempty"`
	Customs       []CustomTargeting `json:"customs,omitempty"` // up to 10
	FriendPeriod  *FriendPeriod     `json:"friendPeriod,omitempty"`
	Interests     []UfoInterest     `json:"interests,omitempty"`
	BusinessTypes []UfoBusinessType `json:"businessTypes,omitempty"`
}

// RejectedReason is why a review was rejected (심사 보류 원인).
type RejectedReason struct {
	RejectedTitle   string `json:"rejectedTitle"`
	RejectedContent string `json:"rejectedContent"`
}

// ResponseMessage is the message of a regular message ad. Optimized
// message ads use [ResponseMultiMessage].
type ResponseMessage struct {
	MessageAdID string      `json:"messageAdId"`
	Name        string      `json:"name"` // up to 150 characters
	Type        MessageType `json:"type"`
	// MessageTitle and Description are set when the type uses them.
	MessageTitle  string         `json:"messageTitle"`
	Description   string         `json:"description"`
	Items         []Item         `json:"items"`
	Buttons       []Button       `json:"buttons"`
	CouponBook    *CouponBook    `json:"couponBook"`
	IntroCarousel *IntroCarousel `json:"introCarousel"`
	Carousels     []Carousel     `json:"carousels"`
	// TODO(review): the Catalog type is not defined on the type page;
	// kept as raw JSON.
	Catalog         []json.RawMessage `json:"catalog"`
	MoreButton      *Landing          `json:"moreButton"`
	ButtonTitle     string            `json:"buttonTitle"` // spotlight
	LandingURL      string            `json:"landingUrl"`  // spotlight
	AgeVerification bool              `json:"ageVerification"`
	AdFlag          bool              `json:"adFlag"`
	ShareFlag       bool              `json:"shareFlag"`
}

// ResponseMessageContent is one message of an optimized message ad. Only
// the fields of its type are set.
type ResponseMessageContent struct {
	Type          MessageType    `json:"type"`
	MessageTitle  string         `json:"messageTitle"`
	Description   string         `json:"description"`
	Items         []Item         `json:"items"`
	ButtonTitle   string         `json:"buttonTitle"` // spotlight
	LandingURL    string         `json:"landingUrl"`  // spotlight
	Buttons       []Button       `json:"buttons"`
	CouponBook    *CouponBook    `json:"couponBook"`
	IntroCarousel *IntroCarousel `json:"introCarousel"`
	Carousels     []Carousel     `json:"carousels"`
	MoreButton    *Landing       `json:"moreButton"`
	AdFlag        bool           `json:"adFlag"`
	ShareFlag     bool           `json:"shareFlag"`
}

// ResponseMultiMessage is the message of an optimized message ad
// (최적화 메시지광고).
type ResponseMultiMessage struct {
	MessageAdID string `json:"messageAdId"`
	Name        string `json:"name"`
	// MessageAdType is always "MULTI_BROADCAST".
	MessageAdType   string `json:"messageAdType"`
	AgeVerification bool   `json:"ageVerification"`
	// Messages are in registration order.
	Messages []ResponseMessageContent `json:"messages"`
}

// Schedule is when an ad group runs. Each day's Time lists 24 "1" (show)
// or "0" (hide) values and is required when DetailTime is true.
type Schedule struct {
	// BeginDate ("yyyy-MM-dd") is today or later.
	BeginDate string `json:"beginDate"`
	// EndDate ("yyyy-MM-dd") is BeginDate or later.
	EndDate   string `json:"endDate,omitempty"`
	LateNight bool   `json:"lateNight"` // 심야 타게팅
	// DetailTime enables per-hour settings. With LateNight true and
	// DetailTime false, the default 22:00~06:59 applies.
	DetailTime    bool     `json:"detailTime"`
	MondayTime    []string `json:"mondayTime,omitempty"`
	TuesdayTime   []string `json:"tuesdayTime,omitempty"`
	WednesdayTime []string `json:"wednesdayTime,omitempty"`
	ThursdayTime  []string `json:"thursdayTime,omitempty"`
	FridayTime    []string `json:"fridayTime,omitempty"`
	SaturdayTime  []string `json:"saturdayTime,omitempty"`
	SundayTime    []string `json:"sundayTime,omitempty"`
}

// MessageSchedule is the second "Schedule" on the type page, with begin
// and end times and a period unit; message sending refers to it.
type MessageSchedule struct {
	DetailTime    bool       `json:"detailTime"`
	BeginDate     string     `json:"beginDate"`
	BeginTime     string     `json:"beginTime,omitempty"`
	EndDate       string     `json:"endDate,omitempty"`
	EndTime       string     `json:"endTime,omitempty"`
	MondayTime    []string   `json:"mondayTime,omitempty"`
	TuesdayTime   []string   `json:"tuesdayTime,omitempty"`
	WednesdayTime []string   `json:"wednesdayTime,omitempty"`
	ThursdayTime  []string   `json:"thursdayTime,omitempty"`
	FridayTime    []string   `json:"fridayTime,omitempty"`
	SaturdayTime  []string   `json:"saturdayTime,omitempty"`
	SundayTime    []string   `json:"sundayTime,omitempty"`
	LateNight     bool       `json:"lateNight"`
	PeriodUnit    PeriodUnit `json:"periodUnit,omitempty"`
}

// SyncAppTargeting targets Kakao Login users of a channel's app
// (카카오 로그인 이용자 타게팅).
type SyncAppTargeting struct {
	ProfileID     string        `json:"profileId"`
	InclusionType InclusionType `json:"inclusionType"`
}

// TalkChannelGroupTargeting is a friend group in a targeting (친구그룹
// 타게팅). Friend groups are being retired; use customer data instead.
type TalkChannelGroupTargeting struct {
	ID                     kakao.Int64   `json:"id"`
	TalkChannelGroupFileID kakao.Int64   `json:"talkChannelGroupFileId"`
	Name                   string        `json:"name"`
	InclusionType          InclusionType `json:"inclusionType"`
	// FileType is APP_USER_ID, PHONE_NUMBER or MESSAGE_RETARGET.
	FileType CustomerDataFileType `json:"fileType"`
	GroupKey string               `json:"groupKey"`
}

// TalkChannelTargeting is a KakaoTalk Channel in a targeting
// (카카오톡 채널 타게팅).
type TalkChannelTargeting struct {
	ID            kakao.Int64   `json:"id"`
	Name          string        `json:"name"`
	InclusionType InclusionType `json:"inclusionType"`
	// ProfileID is the encoded profile ID.
	// TODO(review): documented as Integer; kakao.Int64 accepts numbers and
	// numeric strings, but an encoded ID may be non-numeric.
	ProfileID   kakao.Int64 `json:"profileId"`
	ProfileName string      `json:"profileName"`
}

// Targeting is the targeting of a display ad group or audience.
// TODO(review): the docs mention locationType for this type but do not
// list it as a field.
type Targeting struct {
	// ID and Name identify a saved audience; unset for NORMAL.
	ID         kakao.Int64   `json:"id,omitempty"`
	Name       string        `json:"name,omitempty"`
	Type       AudienceType  `json:"type,omitempty"`
	AgeType    SelectionType `json:"ageType,omitempty"`
	Ages       []string      `json:"ages,omitempty"`
	GenderType SelectionType `json:"genderType,omitempty"`
	Genders    []Gender      `json:"genders,omitempty"`
	// Locations, Depth2Locations and Depth3Locations are 시/도, 시/군/구 and
	// 동/읍/면; at least one is required for a location type of AREA.
	Locations                  []Location                  `json:"locations,omitempty"`
	Depth2Locations            []Location                  `json:"depth2Locations,omitempty"`
	Depth3Locations            []Location                  `json:"depth3Locations,omitempty"`
	UfoInterests               []UfoInterest               `json:"ufoInterests,omitempty"`
	UfoDeprecatedInterests     []UfoInterest               `json:"ufoDeprecatedInterests,omitempty"`
	UfoBusinessTypes           []UfoBusinessType           `json:"ufoBusinessTypes,omitempty"`
	UfoDeprecatedBusinessTypes []UfoBusinessType           `json:"ufoDeprecatedBusinessTypes,omitempty"`
	TalkChannelTargetings      []TalkChannelTargeting      `json:"talkChannelTargetings,omitempty"`
	SyncAppTargetings          []SyncAppTargeting          `json:"syncAppTargetings,omitempty"`
	CustomerDataTargetings     []CustomerDataTargeting     `json:"customerDataTargetings,omitempty"`
	CustomerFileTargetings     []CustomerFileTargeting     `json:"customerFileTargetings,omitempty"`
	CohortTargetings           []CohortTargeting           `json:"cohortTargetings,omitempty"`
	TrackerTargetings          []TrackerTargeting          `json:"trackerTargetings,omitempty"`
	TalkChannelGroupTargeting  []TalkChannelGroupTargeting `json:"talkChannelGroupTargeting,omitempty"`
}

// MessageTargeting is the targeting of a message ad (Targeting: 메시지광고).
type MessageTargeting struct {
	Type MessageTargetingType `json:"type"`
	// OptimizationTarget is set when reading an optimized message ad.
	OptimizationTarget *TargetingOptimization `json:"optimizationTarget,omitempty"`
	Devices            []DeviceType           `json:"devices,omitempty"`
	Demographics       *Demographics          `json:"demographics,omitempty"`
	Customs            []CustomTargeting      `json:"customs,omitempty"`
	FriendPeriod       *FriendPeriod          `json:"friendPeriod,omitempty"`
	// Interests and BusinessTypes are read-only.
	Interests     []UfoInterest     `json:"interests,omitempty"`
	BusinessTypes []UfoBusinessType `json:"businessTypes,omitempty"`
}

// TargetingOptimization is the optimization setting returned in
// MessageTargeting.OptimizationTarget.
type TargetingOptimization struct {
	Type OptimizationMethod `json:"type"`
	// SendRatios are percentages in message order; empty for AI_CUSTOM.
	SendRatios []int `json:"sendRatios"`
}

// Thumbnail is a representative image (대표 이미지).
type Thumbnail struct {
	FileSize    kakao.Int64 `json:"fileSize"`
	URL         string      `json:"url"`
	FileName    string      `json:"fileName"`
	ImageWidth  int         `json:"imageWidth"`
	ImageHeight int         `json:"imageHeight"`
	MimeType    string      `json:"mimeType"`
}

// TrackerTargeting is a Pixel & SDK tracker in a targeting
// (픽셀&SDK 타게팅).
type TrackerTargeting struct {
	ID            kakao.Int64   `json:"id"`
	InclusionType InclusionType `json:"inclusionType"`
	TrackID       string        `json:"trackId"`
	EventCode     string        `json:"eventCode"`
	TrackRuleID   string        `json:"trackRuleId"`
	TrackRuleName string        `json:"trackRuleName"`
	Term          int           `json:"term"` // target days, 1~180
}

// TrackRule is a tracker rule (트래커 정보).
type TrackRule struct {
	TrackRuleID string   `json:"trackRuleId"`
	TrackID     string   `json:"trackId"`
	EventCode   string   `json:"eventCode"`
	Name        string   `json:"name"`
	RuleSetStr  []string `json:"ruleSetStr"` // 룰셋 정의
}

// UfoBusinessType is a business type for targeting (업종). The full list
// is in the business type CSV from the docs.
type UfoBusinessType struct {
	Value string `json:"value"`
}

// UfoInterest is an interest for targeting (관심사). The full list is in
// the interest CSV from the docs.
type UfoInterest struct {
	Value string `json:"value"`
}

// Video is a video file.
type Video struct {
	URL      string      `json:"url"`
	FileName string      `json:"fileName"`
	Size     kakao.Int64 `json:"size"`
	Duration string      `json:"duration"` // 재생시간
}

// DisplayVideo is the video of a display creative (Video: 디스플레이 소재).
type DisplayVideo struct {
	URL string `json:"url"`
}

// MessageVideo is the video of a message ad (Video: 메시지광고).
type MessageVideo struct {
	URL                string `json:"url"`
	AutoThumbnailURL   string `json:"autoThumbnailUrl"`
	UploadThumbnailURL string `json:"uploadThumbnailUrl"`
}
