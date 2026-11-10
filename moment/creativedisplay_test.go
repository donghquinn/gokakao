package moment

import (
	"net/http"
	"testing"
)

const creativeDisplayCreateJSON = `{
  "id": 1,
  "creativeId": 2,
  "adGroupId": 1234,
  "format": "IMAGE_BANNER",
  "name": "첫번째_이미지배너",
  "landingInfo": {"landingType": "AD_VIEW", "adViewItem": {"id": 1}},
  "pcLandingUrl": null,
  "mobileLandingUrl": "http://www.daum.net",
  "rspvLandingUrl": null,
  "frequencyCapType": "AUTO",
  "frequencyCap": null,
  "frequencyCapTime": null,
  "config": "ON",
  "systemConfig": "ON",
  "reviewStatus": "WAITING",
  "image": {"url": "http://xxx.kakao.co.kr/sample_image_banner_xxx.jpeg", "fileName": "sample.jpeg", "width": 640, "height": 100, "size": 50000},
  "altText": "소재설명",
  "isAiLabeled": false,
  "statusDescription": "심사중",
  "createdDate": "2020-01-01T00:00:00.00000",
  "lastModifiedDate": "2020-01-01T00:00:00.000000",
  "rejectedReason": []
}`

func creativeDisplayAssertCreated(t *testing.T, cr *Creative) {
	t.Helper()
	if cr.ID != 1 || cr.CreativeID != 2 || cr.AdGroupID != 1234 || cr.Format != CreativeFormatImageBanner ||
		cr.Name != "첫번째_이미지배너" || cr.LandingInfo == nil || cr.LandingInfo.AdViewItem == nil || cr.LandingInfo.AdViewItem.ID != 1 ||
		cr.MobileLandingURL != "http://www.daum.net" || cr.PCLandingURL != "" || cr.FrequencyCapType != FrequencyCapAuto ||
		cr.FrequencyCap != 0 || cr.SystemConfig != SystemConfigOn || cr.ReviewStatus != ReviewWaiting ||
		cr.AltText != "소재설명" || cr.AILabeled || cr.StatusDescription != "심사중" || cr.CreatedDate.Year() != 2020 ||
		cr.Image == nil || cr.Image.FileName != "sample.jpeg" || len(cr.RejectedReason) != 0 {
		t.Errorf("creative = %+v", cr)
	}
}

func TestCreativeService_CreateDisplay_JSON(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v4/creatives", 200, creativeDisplayCreateJSON)
	ai := false

	cr, err := c.Creatives.CreateDisplay(t.Context(), 5, NewDisplayCreative{
		AdGroupID:    1234,
		Format:       CreativeFormatImageBanner,
		ImageFileURL: "https://www.image.com/sample.png",
		AltText:      "소재설명",
		LandingInfo:  &LandingInfo{LandingType: LandingTypeAdView, AdViewItem: &AdViewItem{ID: 1}},
		Name:         "첫번째_이미지배너",
		AILabeled:    &ai,
	})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "5")
	creativeAssertJSON(t, req, `{
  "adGroupId": 1234,
  "format": "IMAGE_BANNER",
  "imageFileUrl": "https://www.image.com/sample.png",
  "altText": "소재설명",
  "landingInfo": {"landingType": "AD_VIEW", "adViewItem": {"id": 1}},
  "name": "첫번째_이미지배너",
  "isAiLabeled": false
}`)
	creativeDisplayAssertCreated(t, cr)
}

func TestCreativeService_CreateDisplay_Multipart(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v4/creatives", 200, creativeDisplayCreateJSON)
	png := "\x89PNG\r\n\x1a\n banner"
	ai := false

	cr, err := c.Creatives.CreateDisplay(t.Context(), 5, NewDisplayCreative{
		AdGroupID:        1234,
		Format:           CreativeFormatImageBanner,
		ImageFile:        &UploadFile{Filename: "banner.png", Data: []byte(png)},
		AltText:          "소재설명",
		Name:             "첫번째_이미지배너",
		MobileLandingURL: "http://www.daum.net",
		LandingInfo:      &LandingInfo{LandingType: LandingTypeAdView, AdViewItem: &AdViewItem{ID: 1}},
		Slides:           []Slide{{Title: "s1", ImageFile: &UploadFile{Filename: "s.jpg", ContentType: "image/jpeg", Data: []byte("jpg")}}},
		OpinionProofFiles: []UploadFile{
			{Filename: "p.pdf", ContentType: "application/pdf", Data: []byte("pdf1")},
			{Filename: "q.pdf", ContentType: "application/pdf", Data: []byte("pdf2")},
		},
		AILabeled: &ai,
	})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "5")
	fields, files := creativeParseMultipart(t, req)
	for k, want := range map[string]string{
		"adGroupId":                 "1234",
		"format":                    "IMAGE_BANNER",
		"altText":                   "소재설명",
		"name":                      "첫번째_이미지배너",
		"mobileLandingUrl":          "http://www.daum.net",
		"landingInfo.landingType":   "AD_VIEW",
		"landingInfo.adViewItem.id": "1",
		"slides[0].title":           "s1",
		"isAiLabeled":               "false",
	} {
		if got := fields.Get(k); got != want {
			t.Errorf("field %s = %q, want %q", k, got, want)
		}
	}
	if len(fields) != 9 {
		t.Errorf("fields = %v", fields)
	}
	if f := files["imageFile"]; f.Filename != "banner.png" || f.ContentType != "image/png" || f.Data != png {
		t.Errorf("imageFile = %+v", f)
	}
	if f := files["slides[0].imageFile"]; f.Filename != "s.jpg" || f.Data != "jpg" {
		t.Errorf("slide file = %+v", f)
	}
	if f := files["opinionProofFileList[1]"]; f.Filename != "q.pdf" || f.ContentType != "application/pdf" || f.Data != "pdf2" {
		t.Errorf("proof file = %+v", f)
	}
	if len(files) != 4 {
		t.Errorf("files = %v", files)
	}
	creativeDisplayAssertCreated(t, cr)
}

func TestCreativeService_CreateDisplay_Validation(t *testing.T) {
	t.Parallel()
	c, _ := newTestClient(t)
	img := &UploadFile{Filename: "a.png", Data: []byte("x")}
	for name, cr := range map[string]NewDisplayCreative{
		"no ad group":       {Format: CreativeFormatImageBanner, ImageFileURL: "u"},
		"bad format":        {AdGroupID: 1, Format: "TEXT"},
		"banner no image":   {AdGroupID: 1, Format: CreativeFormatImageBanner},
		"native no profile": {AdGroupID: 1, Format: CreativeFormatImageNative, ImageFileURL: "u"},
		"video no video":    {AdGroupID: 1, Format: CreativeFormatVideoNative, ProfileImageFileURL: "u"},
		"proof URL in multipart": {
			AdGroupID: 1, Format: CreativeFormatImageBanner, ImageFile: img,
			OpinionProofFileURLs: []string{"https://a/b.pdf"},
		},
		"thumbnail URL in multipart": {
			AdGroupID: 1, Format: CreativeFormatVideoNative, VideoFile: img,
			ProfileImageFileURL: "u", UploadThumbnailImageFileURL: "u",
		},
		"too many slides": {AdGroupID: 1, Format: CreativeFormatImageNative, Slides: make([]Slide, 11)},
		"too many proofs": {AdGroupID: 1, Format: CreativeFormatImageBanner, ImageFileURL: "u", OpinionProofFileURLs: make([]string, 11)},
		"empty file":      {AdGroupID: 1, Format: CreativeFormatImageBanner, ImageFile: &UploadFile{Filename: "a"}},
	} {
		if _, err := c.Creatives.CreateDisplay(t.Context(), 5, cr); err == nil {
			t.Errorf("%s: want error", name)
		}
	}
}

const creativeDisplayUpdateJSON = `{
  "id": 1234,
  "creativeId": 1235,
  "adGroupId": 5678,
  "format": "IMAGE_BANNER",
  "name": "이미지_수정",
  "landingInfo": {"landingType": "AD_VIEW", "communicatorAdViewId": 1},
  "pcLandingUrl": null,
  "mobileLandingUrl": null,
  "rspvLandingUrl": null,
  "frequencyCapType": "AUTO",
  "frequencyCap": null,
  "frequencyCapTime": null,
  "config": "ON",
  "systemConfig": "ON",
  "reviewStatus": "WAITING",
  "image": {"url": "http://xxx.kakao.co.kr/sample_image_banner_xxx.jpg", "fileName": "sample_image_banner.jpg", "width": 640, "height": 100, "size": 50000},
  "altText": "소재설명_수정",
  "isAiLabeled": false,
  "statusDescription": "심사중",
  "createdDate": "2020-01-01T00:00:00.00000",
  "lastModifiedDate": "2020-01-01T00:00:00.000000",
  "rejectedReason": []
}`

func TestCreativeService_UpdateDisplay(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PUT /openapi/v4/creatives", 200, creativeDisplayUpdateJSON)
	ai := false

	ch := DisplayCreativeChanges{ID: 1234, NewDisplayCreative: NewDisplayCreative{
		AdGroupID:        5678,
		Format:           CreativeFormatImageBanner,
		ImageFileURL:     "http://www.company.com/images/sample_image_banner.jpg",
		AltText:          "소재설명_수정",
		MobileLandingURL: "http://www.daum.net",
		Name:             "이미지_수정",
		AILabeled:        &ai,
	}}
	cr, err := c.Creatives.UpdateDisplay(t.Context(), 5, ch)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "5")
	if req.Method != http.MethodPut {
		t.Errorf("method = %s", req.Method)
	}
	creativeAssertJSON(t, req, `{
  "id": 1234,
  "adGroupId": 5678,
  "format": "IMAGE_BANNER",
  "imageFileUrl": "http://www.company.com/images/sample_image_banner.jpg",
  "altText": "소재설명_수정",
  "mobileLandingUrl": "http://www.daum.net",
  "name": "이미지_수정",
  "isAiLabeled": false
}`)
	if cr.ID != 1234 || cr.CreativeID != 1235 || cr.AdGroupID != 5678 || cr.Name != "이미지_수정" || cr.AltText != "소재설명_수정" {
		t.Errorf("creative = %+v", cr)
	}

	// Multipart: an image file without an image URL.
	ch.ImageFileURL = ""
	ch.ImageFile = &UploadFile{Filename: "b.png", ContentType: "image/png", Data: []byte("png")}
	if _, err := c.Creatives.UpdateDisplay(t.Context(), 5, ch); err != nil {
		t.Fatal(err)
	}
	fields, files := creativeParseMultipart(t, assertRequest(t, srv, "5"))
	if fields.Get("id") != "1234" || fields.Get("adGroupId") != "5678" || fields.Get("format") != "IMAGE_BANNER" ||
		fields.Get("altText") != "소재설명_수정" || fields.Get("isAiLabeled") != "false" || fields.Has("imageFileUrl") {
		t.Errorf("fields = %v", fields)
	}
	if f := files["imageFile"]; f.Filename != "b.png" || f.Data != "png" {
		t.Errorf("imageFile = %+v", f)
	}

	if _, err := c.Creatives.UpdateDisplay(t.Context(), 5, DisplayCreativeChanges{NewDisplayCreative: ch.NewDisplayCreative}); err == nil {
		t.Error("missing ID: want error")
	}
	if _, err := c.Creatives.UpdateDisplay(t.Context(), 5, DisplayCreativeChanges{ID: 1}); err == nil {
		t.Error("missing ad group: want error")
	}
}

func TestCreative_DisplayChanges(t *testing.T) {
	t.Parallel()
	cr := Creative{
		ID: 1, AdGroupID: 2, Format: CreativeFormatImageNative, Name: "n", AltText: "alt", Title: "t",
		ProfileName: "p", Description: "d", ActionButton: ActionButtonMore, AILabeled: true,
		FrequencyCapType: FrequencyCapDayImp, FrequencyCap: 3, FrequencyCapTime: 24,
		Image:       &Image{URL: "http://img"},
		AssetGroups: []AssetGroup{{Title: "s", Image: &Image{URL: "http://slide"}, PriceAmount: "100", PriceCurrencyCode: CurrencyKRW}, {Title: "s2"}},
	}
	ch := cr.DisplayChanges()
	if ch.ID != 1 || ch.AdGroupID != 2 || ch.Format != CreativeFormatImageNative || ch.Name != "n" || ch.AltText != "alt" ||
		ch.Title != "t" || ch.ProfileName != "p" || ch.Description != "d" || ch.ActionButton != ActionButtonMore ||
		ch.AILabeled == nil || !*ch.AILabeled || ch.FrequencyCap != 3 || ch.FrequencyCapTime != 24 || ch.ImageFileURL != "" {
		t.Errorf("changes = %+v", ch)
	}
	if len(ch.Slides) != 2 || ch.Slides[0].ImageFileURL != "http://slide" || ch.Slides[0].PriceAmount != "100" || ch.Slides[1].ImageFileURL != "" {
		t.Errorf("slides = %+v", ch.Slides)
	}
}

func TestCreativeService_CopyDisplay(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v4/creatives/copy", 200, `[
  {
    "id": 35678, "creativeId": 35678, "adGroupId": 23456, "format": "IMAGE_BANNER", "name": "이미지 소재 1",
    "landingInfo": {"landingType": "URL", "url": "http://m.daum.net"},
    "pcLandingUrl": null, "mobileLandingUrl": "http://m.daum.net", "rspvLandingUrl": null,
    "frequencyCap": null, "frequencyCapType": "AUTO", "config": "ON", "systemConfig": "ON",
    "reviewStatus": "APPROVED", "creativeStatus": "OPERATING",
    "image": {"url": "http://xxx.kakao.co.kr/sample_image_111.jpg", "fileName": "sample_image.jpg", "width": 640, "height": 480, "size": 100000},
    "altText": "이미지설명문구 2", "isAiLabeled": false, "title": "타이틀", "description": "홍보문구", "actionButton": "JOIN", "profileName": "프로필이름",
    "profileImage": {"url": "http://xxx.kakao.co.kr/sample_profile_111.jpg", "fileName": "sample_profile.jpg", "width": 300, "height": 300, "size": 50000},
    "video": {"url": "http://xxx.kakao.co.kr/sample_image_111.avi", "fileName": "sample_video.avi", "size": 500000, "duration": 3.0},
    "statusDescription": "운영중", "rejectedReason": [],
    "createdDate": "2020-01-01T00:00:00.000000", "lastModifiedDate": "2020-01-01T00:00:00.000000",
    "opinionProof": null,
    "thumbnailImage": {"url": "http://xxx.kakao.co.kr/sample_image_111.jpg", "fileName": "sample_image.jpg", "width": 640, "height": 480, "size": 100000}
  },
  {
    "id": 35679, "creativeId": 35679, "adGroupId": 23456, "format": "IMAGE_BANNER", "name": "이미지 소재 2",
    "landingInfo": {"landingType": "URL", "url": "http://m.daum.net"},
    "mobileLandingUrl": "http://m.daum.net", "frequencyCapType": "AUTO", "config": "ON", "systemConfig": "ON",
    "reviewStatus": "APPROVED", "creativeStatus": "OPERATING",
    "video": {"url": "http://xxx.kakao.co.kr/sample_image_222.avi", "fileName": "sample_video.avi", "size": 500000, "duration": 3.0},
    "statusDescription": "운영중", "rejectedReason": [], "opinionProof": null
  }
]`)

	list, err := c.Creatives.CopyDisplay(t.Context(), 5, 23456, []int64{34567, 34568})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "5")
	creativeAssertJSON(t, req, `{"adGroupId": 23456, "creativeIds": [34567, 34568]}`)
	if len(list) != 2 {
		t.Fatalf("list = %+v", list)
	}
	cr := list[0]
	if cr.ID != 35678 || cr.AdGroupID != 23456 || cr.ReviewStatus != ReviewApproved || cr.CreativeStatus != CreativeStatusOperating ||
		cr.ActionButton != ActionButtonJoin || cr.ProfileName != "프로필이름" || cr.ProfileImage == nil || cr.ProfileImage.Width != 300 ||
		cr.Video == nil || cr.Video.Duration != 3 || cr.Video.Size != 500000 || cr.ThumbnailImage == nil || cr.ThumbnailImage.Height != 480 ||
		cr.LandingInfo.URL != "http://m.daum.net" || cr.OpinionProof != nil {
		t.Errorf("copy = %+v", cr)
	}
	if list[1].ID != 35679 || list[1].Name != "이미지 소재 2" {
		t.Errorf("copy 2 = %+v", list[1])
	}

	if _, err := c.Creatives.CopyDisplay(t.Context(), 5, 1, make([]int64, 21)); err == nil {
		t.Error("21 creatives: want error")
	}
	if _, err := c.Creatives.CopyDisplay(t.Context(), 5, 1, nil); err == nil {
		t.Error("no creatives: want error")
	}
	if _, err := c.Creatives.CopyDisplay(t.Context(), 5, 0, []int64{1}); err == nil {
		t.Error("no ad group: want error")
	}
}

func TestCreativeService_OpinionProof(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/creatives/77/opinionProof", 200, `{
  "id": 1,
  "opinion": "의견증빙",
  "opinionProofFileList": [
    {"id": 11, "originalFileName": "sample_1.jpg", "downloadUrl": "https://xxx.kakao.com/sample_1.jpg"}
  ]
}`)

	p, err := c.Creatives.OpinionProof(t.Context(), 5, 77)
	if err != nil {
		t.Fatal(err)
	}
	assertRequest(t, srv, "5")
	if p.ID != 1 || p.Opinion != "의견증빙" || len(p.Files) != 1 || p.Files[0].ID != 11 ||
		p.Files[0].OriginalFileName != "sample_1.jpg" || p.Files[0].DownloadURL != "https://xxx.kakao.com/sample_1.jpg" {
		t.Errorf("proof = %+v", p)
	}
}
