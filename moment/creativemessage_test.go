package moment

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

const creativeMessageJSON = `{
  "id": 12345,
  "creativeId": 12345,
  "name": "카카오톡 채널_도달_20210625",
  "adGroupId": 11223,
  "format": "BASIC_TEXT_MESSAGE",
  "config": "ON",
  "systemConfig": "ON",
  "statusDescription": "발송 대기",
  "creativeStatus": "OPERATING",
  "createdDate": "2021-06-25T17:04:02.883575",
  "lastModifiedDate": "2021-06-25T17:04:06.291245",
  "messageElement": {
    "id": 12345,
    "adAccountId": 123,
    "profileId": "_xbHxd",
    "name": "카카오톡 채널_도달_20210625",
    "creativeFormat": "BASIC_TEXT_MESSAGE",
    "title": "홍보문구입니다.",
    "image": {
      "fileSize": 168816,
      "url": "//t1.kakaocdn.net/b2/creative/759/d7961bd0662a240f43f047d3116a25f3.jpg",
      "fileName": "풀뷰 1280x720.jpg",
      "imageWidth": 1280,
      "imageHeight": 720,
      "mimeType": "image/jpeg",
      "imageHash": "35156f0c1393434ced4be21423d08a6a"
    },
    "shareFlag": true,
    "adFlag": true,
    "thumbnail": {
      "fileSize": 168816,
      "url": "//t1.kakaocdn.net/b2/creative/759/d7961bd0662a240f43f047d3116a25f3.jpg",
      "fileName": "풀뷰 1280x720.jpg",
      "imageWidth": 1280,
      "imageHeight": 720,
      "mimeType": "image/jpeg",
      "imageHash": "35156f0c1393434ced4be21423d08a6a"
    },
    "buttonAssetGroups": [
      {"ordering": 0, "title": "버튼1", "pcLandingUrl": "http://www.daum.net", "mobileLandingUrl": "https://www.kakaocorp.com", "landingType": "LANDING_URL"}
    ],
    "thumbnailUrl": "//t1.kakaocdn.net/b2/creative/759/d7961bd0662a240f43f047d3116a25f3.jpg",
    "messageThumbnail": {
      "fileSize": 168816,
      "url": "//t1.kakaocdn.net/b2/creative/759/d7961bd0662a240f43f047d3116a25f3.jpg",
      "fileName": "풀뷰 1280x720.jpg",
      "imageWidth": 1280,
      "imageHeight": 720,
      "mimeType": "image/jpeg",
      "imageHash": "35156f0c1393434ced4be21423d08a6a"
    },
    "createdDate": "2021-06-25T17:04:02.883575",
    "lastModifiedDate": "2021-06-25T17:04:06.291245"
  }
}`

// creativeMessageNew is the docs' basic text example.
func creativeMessageNew() NewMessageCreative {
	return NewMessageCreative{
		AdGroupID: 39688,
		Format:    MessageTypeBasicText,
		MessageElement: NewMessageElement{
			ProfileID: "_Xxo",
			Title:     "홍보문구",
			ButtonAssetGroups: []NewButtonAssetGroup{
				{Ordering: 0, LandingType: AssetLandingURL, Title: "버튼1", PCLandingURL: "http://www.daum.net", MobileLandingURL: "http://www.kakaocorp.com"},
				{Ordering: 1, LandingType: AssetLandingChannelPost, ChannelPostID: 1, Title: "소식보기"},
			},
			ShareFlag: true,
			AdFlag:    true,
			CSInfo:    "02-1234-5678",
			ImageFile: &UploadFile{Filename: "banner.png", ContentType: "image/png", Data: []byte("png")},
		},
	}
}

// creativeMessageWantFields are the form fields of creativeMessageNew.
var creativeMessageWantFields = url.Values{
	"adGroupId":                     {"39688"},
	"format":                        {"BASIC_TEXT_MESSAGE"},
	"messageElement.creativeFormat": {"BASIC_TEXT_MESSAGE"},
	"messageElement.profileId":      {"_Xxo"},
	"messageElement.title":          {"홍보문구"},
	"messageElement.buttonAssetGroups[0].ordering":         {"0"},
	"messageElement.buttonAssetGroups[0].landingType":      {"LANDING_URL"},
	"messageElement.buttonAssetGroups[0].title":            {"버튼1"},
	"messageElement.buttonAssetGroups[0].pcLandingUrl":     {"http://www.daum.net"},
	"messageElement.buttonAssetGroups[0].mobileLandingUrl": {"http://www.kakaocorp.com"},
	"messageElement.buttonAssetGroups[1].ordering":         {"1"},
	"messageElement.buttonAssetGroups[1].landingType":      {"CHANNEL_POST"},
	"messageElement.buttonAssetGroups[1].channelPostId":    {"1"},
	"messageElement.buttonAssetGroups[1].title":            {"소식보기"},
	"messageElement.shareFlag":                             {"true"},
	"messageElement.adFlag":                                {"true"},
	"messageElement.csInfo":                                {"02-1234-5678"},
}

func creativeMessageAssert(t *testing.T, cr *Creative) {
	t.Helper()
	if cr.ID != 12345 || cr.CreativeID != 12345 || cr.AdGroupID != 11223 || MessageType(cr.Format) != MessageTypeBasicText ||
		cr.StatusDescription != "발송 대기" || cr.CreativeStatus != CreativeStatusOperating || cr.CreatedDate.Year() != 2021 {
		t.Errorf("creative = %+v", cr)
	}
	m := cr.MessageElement
	if m == nil {
		t.Fatal("no message element")
	}
	if m.ID != 12345 || m.AdAccountID != 123 || m.ProfileID != "_xbHxd" || m.CreativeFormat != MessageTypeBasicText ||
		m.Title != "홍보문구입니다." || !m.ShareFlag || !m.AdFlag || m.ThumbnailURL == "" || m.CreatedDate != "2021-06-25T17:04:02.883575" {
		t.Errorf("element = %+v", m)
	}
	if m.Image == nil || m.Image.FileSize != 168816 || m.Image.ImageWidth != 1280 || m.Image.MimeType != "image/jpeg" || m.Image.ImageHash == "" {
		t.Errorf("image = %+v", m.Image)
	}
	if m.Thumbnail == nil || m.Thumbnail.ImageHeight != 720 || m.MessageThumbnail == nil || m.MessageThumbnail.FileSize != 168816 {
		t.Errorf("thumbnails = %+v %+v", m.Thumbnail, m.MessageThumbnail)
	}
	if len(m.ButtonAssetGroups) != 1 || m.ButtonAssetGroups[0].Title != "버튼1" || m.ButtonAssetGroups[0].LandingType != AssetLandingURL {
		t.Errorf("buttons = %+v", m.ButtonAssetGroups)
	}
}

func TestCreativeService_CreateMessage(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v4/creatives", 200, creativeMessageJSON)

	cr, err := c.Creatives.CreateMessage(t.Context(), 5, creativeMessageNew())
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "5")
	fields, files := creativeParseMultipart(t, req)
	if !reflect.DeepEqual(fields, creativeMessageWantFields) {
		t.Errorf("fields = %v", fields)
	}
	if f := files["messageElement.imageFile"]; len(files) != 1 || f.Filename != "banner.png" || f.Data != "png" {
		t.Errorf("files = %v", files)
	}
	creativeMessageAssert(t, cr)
}

func TestCreativeService_CreateMessage_Items(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v4/creatives", 200, creativeMessageJSON)

	cr := NewMessageCreative{AdGroupID: 1, Format: MessageTypeCarouselCommerce, Name: "캐러셀", MessageElement: NewMessageElement{
		ProfileID: "_Xxo", CSInfo: "1", HasIntro: true, IntroLandingType: "LANDING_URL",
		ItemAssetGroups: []NewItemAssetGroup{{
			LandingType: AssetLandingURL, Title: "카드", PriceAmount: "1000", PriceCurrencyCode: CurrencyKRW, DiscountedPriceAmount: "900",
			MobileLandingURL: "http://m", ImageFile: &UploadFile{Filename: "c.jpg", Data: []byte("jpg")},
		}},
	}}
	if _, err := c.Creatives.CreateMessage(t.Context(), 5, cr); err != nil {
		t.Fatal(err)
	}
	fields, files := creativeParseMultipart(t, assertRequest(t, srv, "5"))
	if fields.Get("name") != "캐러셀" || fields.Get("messageElement.hasIntro") != "true" || fields.Get("messageElement.shareFlag") != "false" ||
		fields.Get("messageElement.itemAssetGroups[0].priceAmount") != "1000" || fields.Get("messageElement.itemAssetGroups[0].priceCurrencyCode") != "KRW" ||
		fields.Get("messageElement.creativeFormat") != "CAROUSEL_COMMERCE_MESSAGE" {
		t.Errorf("fields = %v", fields)
	}
	if f := files["messageElement.itemAssetGroups[0].imageFile"]; f.Filename != "c.jpg" || f.Data != "jpg" {
		t.Errorf("files = %v", files)
	}
}

func TestCreativeService_CreateMessage_Validation(t *testing.T) {
	t.Parallel()
	c, _ := newTestClient(t)
	el := NewMessageElement{ProfileID: "p", CSInfo: "1"}
	img := &UploadFile{Filename: "a", Data: []byte("x")}
	for name, cr := range map[string]NewMessageCreative{
		"no ad group":      {Format: MessageTypeBasicText, MessageElement: el},
		"premium video":    {AdGroupID: 1, Format: MessageTypePremiumVideo, MessageElement: el},
		"format mismatch":  {AdGroupID: 1, Format: MessageTypeBasicText, MessageElement: NewMessageElement{CreativeFormat: MessageTypeWide, ProfileID: "p", CSInfo: "1"}},
		"no profile":       {AdGroupID: 1, Format: MessageTypeBasicText, MessageElement: NewMessageElement{CSInfo: "1"}},
		"image and var":    {AdGroupID: 1, Format: MessageTypeBasicText, MessageElement: NewMessageElement{ProfileID: "p", CSInfo: "1", ImageFile: img, Image: &ImageVariable{"${image_url1}"}}},
		"item image twice": {AdGroupID: 1, Format: MessageTypeWide, MessageElement: NewMessageElement{ProfileID: "p", CSInfo: "1", ItemAssetGroups: []NewItemAssetGroup{{ImageFile: img, Image: &ImageVariable{"x"}}}}},
		"empty file":       {AdGroupID: 1, Format: MessageTypeBasicText, MessageElement: NewMessageElement{ProfileID: "p", CSInfo: "1", ImageFile: &UploadFile{}}},
		"empty item file":  {AdGroupID: 1, Format: MessageTypeWide, MessageElement: NewMessageElement{ProfileID: "p", CSInfo: "1", ItemAssetGroups: []NewItemAssetGroup{{ImageFile: &UploadFile{}}}}},
	} {
		if _, err := c.Creatives.CreateMessage(t.Context(), 5, cr); err == nil {
			t.Errorf("%s: want error", name)
		}
	}
}

func TestCreativeService_UpdateMessage(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PUT /openapi/v4/creatives", 200, creativeMessageJSON)

	cr, err := c.Creatives.UpdateMessage(t.Context(), 5, MessageCreativeChanges{ID: 12345, NewMessageCreative: creativeMessageNew()})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "5")
	if req.Method != http.MethodPut {
		t.Errorf("method = %s", req.Method)
	}
	fields, files := creativeParseMultipart(t, req)
	want := url.Values{"id": {"12345"}}
	for k, v := range creativeMessageWantFields {
		want[k] = v
	}
	if !reflect.DeepEqual(fields, want) {
		t.Errorf("fields = %v", fields)
	}
	if len(files) != 1 {
		t.Errorf("files = %v", files)
	}
	creativeMessageAssert(t, cr)

	if _, err := c.Creatives.UpdateMessage(t.Context(), 5, MessageCreativeChanges{NewMessageCreative: creativeMessageNew()}); err == nil {
		t.Error("missing ID: want error")
	}
	if _, err := c.Creatives.UpdateMessage(t.Context(), 5, MessageCreativeChanges{ID: 1}); err == nil {
		t.Error("missing ad group: want error")
	}
}

func TestCreative_MessageChanges(t *testing.T) {
	t.Parallel()
	cr := Creative{ID: 1, AdGroupID: 2, Format: CreativeFormat(MessageTypeWideList), Name: "n", MessageElement: &MessageCreativeElement{
		MessageElement: MessageElement{
			CreativeFormat: MessageTypeWideList, ProfileID: "p", Title: "t", ShareFlag: true, AdFlag: true, CSInfo: "cs",
			ButtonAssetGroups: []ButtonAssetGroup{{Ordering: 1, Title: "b", LandingType: AssetLandingAdView, AdViewID: 9}},
			ItemAssetGroups:   []ItemAssetGroup{{LandingType: AssetLandingChannelPost, Title: "i", PriceAmount: "5"}},
		},
		Image: &MessageImage{ValueWithVariable: "${image_url1}"},
	}}
	ch := cr.MessageChanges()
	e := ch.MessageElement
	if ch.ID != 1 || ch.AdGroupID != 2 || ch.Format != MessageTypeWideList || ch.Name != "n" ||
		e.CreativeFormat != MessageTypeWideList || e.ProfileID != "p" || e.Title != "t" || !e.ShareFlag || !e.AdFlag || e.CSInfo != "cs" ||
		e.Image == nil || e.Image.ValueWithVariable != "${image_url1}" {
		t.Errorf("changes = %+v", ch)
	}
	if len(e.ButtonAssetGroups) != 1 || e.ButtonAssetGroups[0].Ordering != 1 || e.ButtonAssetGroups[0].AdViewID != 9 ||
		len(e.ItemAssetGroups) != 1 || e.ItemAssetGroups[0].Title != "i" || e.ItemAssetGroups[0].PriceAmount != "5" {
		t.Errorf("groups = %+v", e)
	}

	empty := (&Creative{ID: 3}).MessageChanges()
	if empty.ID != 3 || empty.MessageElement.ProfileID != "" {
		t.Errorf("no element = %+v", empty)
	}
}

func TestCreativeService_CopyMessage(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v4/creatives/copy", 200, `[
  {
    "id": 12346, "creativeId": 12346, "name": "카카오톡 채널_도달_20210625", "adGroupId": 11223, "format": "BASIC_TEXT_MESSAGE",
    "config": "ON", "systemConfig": "ON", "statusDescription": "발송 대기", "creativeStatus": "OPERATING",
    "createdDate": "2021-06-25T17:04:03", "lastModifiedDate": "2021-06-25T17:04:06",
    "messageElement": {
      "id": 78428, "adAccountId": 759, "profileId": "_xbHxd", "name": "카카오톡 채널_도달_20210625",
      "creativeFormat": "BASIC_TEXT_MESSAGE", "title": "홍보문구입니다.",
      "image": {"fileSize": 168816, "url": "//t1.kakaocdn.net/b2/creative/759/d7961bd0662a240f43f047d3116a25f3.jpg", "fileName": "풀뷰 1280x720.jpg", "imageWidth": 1280, "imageHeight": 720, "mimeType": "image/jpeg", "imageHash": "35156f0c1393434ced4be21423d08a6a"},
      "shareFlag": true, "adFlag": true,
      "buttonAssetGroups": [{"ordering": 0, "title": "버튼1", "pcLandingUrl": "http://www.daum.net", "mobileLandingUrl": "https://www.kakaocorp.com", "landingType": "LANDING_URL"}],
      "thumbnailUrl": "//t1.kakaocdn.net/b2/creative/759/d7961bd0662a240f43f047d3116a25f3.jpg",
      "createdDate": "2021-06-25T17:04:03", "lastModifiedDate": "2021-06-25T17:04:06"
    }
  }
]`)

	list, err := c.Creatives.CopyMessage(t.Context(), 5, 11223, []int64{12345})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "5")
	creativeAssertJSON(t, req, `{"adGroupId": 11223, "creativeIds": [12345]}`)
	if len(list) != 1 || list[0].ID != 12346 || list[0].MessageElement == nil || list[0].MessageElement.ID != 78428 ||
		list[0].MessageElement.AdAccountID != 759 || list[0].MessageElement.Image.ImageHeight != 720 || list[0].LastModifiedDate.Second() != 6 {
		t.Errorf("copy = %+v", list)
	}
	if _, err := c.Creatives.CopyMessage(t.Context(), 5, 11223, nil); err == nil {
		t.Error("no creatives: want error")
	}
}

const creativeMessagePersonalizedJSON = `{
  "id": 12345,
  "creativeId": 12345,
  "name": "카카오톡 채널_도달_20210625",
  "adGroupId": 11223,
  "format": "BASIC_TEXT_MESSAGE",
  "config": "ON",
  "statusDescription": "발송 대기",
  "creativeStatus": "OPERATING",
  "createdDate": "2021-06-25T17:04:02.883575",
  "lastModifiedDate": "2021-06-25T17:04:06.291245",
  "messageElement": {
    "id": 12345,
    "adAccountId": 123,
    "profileId": "_xbHxd",
    "name": "카카오톡 채널_도달_20210625",
    "creativeFormat": "BASIC_TEXT_MESSAGE",
    "title": "홍보문구입니다.",
    "image": {"valueWithVariable": "${image_url1}"},
    "shareFlag": true,
    "adFlag": true,
    "buttonAssetGroups": [
      {"ordering": 0, "title": "버튼1", "pcLandingUrl": "http://www.daum.net", "mobileLandingUrl": "http://www.google.com", "landingType": "LANDING_URL"}
    ],
    "thumbnailUrl": "//t1.kakaocdn.net/b2/creative/759/d7961bd0662a240f43f047d3116a25f3.jpg",
    "createdDate": "2021-06-25T17:04:02.883575",
    "lastModifiedDate": "2021-06-25T17:04:06.291245"
  }
}`

func TestCreativeService_CreatePersonalizedMessage(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v4/creatives", 200, creativeMessagePersonalizedJSON)

	cr, err := c.Creatives.CreatePersonalizedMessage(t.Context(), 5, NewMessageCreative{
		AdGroupID: 12345,
		Format:    MessageTypeBasicText,
		MessageElement: NewMessageElement{
			ProfileID: "_Xxo",
			Title:     "홍보문구",
			ButtonAssetGroups: []NewButtonAssetGroup{
				{Ordering: 0, LandingType: AssetLandingURL, Title: "버튼1", PCLandingURL: "${pc_url1}", MobileLandingURL: "${mobile_url1}"},
			},
			ShareFlag: true,
			AdFlag:    true,
			CSInfo:    "02-1234-5678",
			Image:     &ImageVariable{ValueWithVariable: "${image_url1}"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "5")
	fields, files := creativeParseMultipart(t, req)
	want := url.Values{
		"adGroupId":                     {"12345"},
		"format":                        {"BASIC_TEXT_MESSAGE"},
		"messageElement.creativeFormat": {"BASIC_TEXT_MESSAGE"},
		"messageElement.profileId":      {"_Xxo"},
		"messageElement.title":          {"홍보문구"},
		"messageElement.buttonAssetGroups[0].ordering":         {"0"},
		"messageElement.buttonAssetGroups[0].landingType":      {"LANDING_URL"},
		"messageElement.buttonAssetGroups[0].title":            {"버튼1"},
		"messageElement.buttonAssetGroups[0].pcLandingUrl":     {"${pc_url1}"},
		"messageElement.buttonAssetGroups[0].mobileLandingUrl": {"${mobile_url1}"},
		"messageElement.shareFlag":                             {"true"},
		"messageElement.adFlag":                                {"true"},
		"messageElement.csInfo":                                {"02-1234-5678"},
		"messageElement.image.valueWithVariable":               {"${image_url1}"},
	}
	if !reflect.DeepEqual(fields, want) || len(files) != 0 {
		t.Errorf("fields = %v, files = %v", fields, files)
	}
	if cr.ID != 12345 || cr.MessageElement == nil || cr.MessageElement.Image == nil || cr.MessageElement.Image.ValueWithVariable != "${image_url1}" ||
		len(cr.MessageElement.ButtonAssetGroups) != 1 || cr.MessageElement.ButtonAssetGroups[0].MobileLandingURL != "http://www.google.com" {
		t.Errorf("creative = %+v", cr)
	}

	if _, err := c.Creatives.CreatePersonalizedMessage(t.Context(), 5, NewMessageCreative{}); err == nil {
		t.Error("empty: want error")
	}
}

func TestCreativeService_UpdatePersonalizedMessage(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PUT /openapi/v4/creatives", 200, creativeMessageJSON)

	cr, err := c.Creatives.UpdatePersonalizedMessage(t.Context(), 5, MessageCreativeChanges{ID: 12345, NewMessageCreative: creativeMessageNew()})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "5")
	if req.Method != http.MethodPut {
		t.Errorf("method = %s", req.Method)
	}
	fields, files := creativeParseMultipart(t, req)
	if fields.Get("id") != "12345" || fields.Get("messageElement.buttonAssetGroups[1].channelPostId") != "1" || len(files) != 1 {
		t.Errorf("fields = %v, files = %v", fields, files)
	}
	creativeMessageAssert(t, cr)

	if _, err := c.Creatives.UpdatePersonalizedMessage(t.Context(), 5, MessageCreativeChanges{NewMessageCreative: creativeMessageNew()}); err == nil {
		t.Error("missing ID: want error")
	}
	if _, err := c.Creatives.UpdatePersonalizedMessage(t.Context(), 5, MessageCreativeChanges{ID: 1}); err == nil {
		t.Error("invalid: want error")
	}
}
