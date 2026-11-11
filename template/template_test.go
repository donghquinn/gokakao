package template

import (
	"encoding/json"
	"reflect"
	"testing"
)

// assertJSON checks that v marshals to JSON equivalent to want.
func assertJSON(t *testing.T, v any, want string) {
	t.Helper()
	got, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var g, w any
	if err := json.Unmarshal(got, &g); err != nil {
		t.Fatalf("invalid JSON %s: %v", got, err)
	}
	if err := json.Unmarshal([]byte(want), &w); err != nil {
		t.Fatalf("bad want: %v", err)
	}
	if !reflect.DeepEqual(g, w) {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}

// The feed example from the KakaoTalk Message REST API docs.
const docFeed = `{
  "object_type": "feed",
  "content": {
    "title": "오늘의 디저트",
    "description": "아메리카노, 빵, 케익",
    "image_url": "https://mud-kage.kakao.com/dn/NTmhS/btqfEUdFAUf/FjKzkZsnoeE4o19klTOVI1/openlink_640x640s.jpg",
    "image_width": 640,
    "image_height": 640,
    "link": {
      "web_url": "http://www.daum.net",
      "mobile_web_url": "http://m.daum.net",
      "android_execution_params": "contentId=100",
      "ios_execution_params": "contentId=100"
    }
  },
  "item_content": {
    "profile_text": "Kakao",
    "profile_image_url": "https://mud-kage.kakao.com/dn/Q2iNx/btqgeRgV54P/VLdBs9cvyn8BJXB3o7N8UK/kakaolink40_original.png",
    "title_image_url": "https://mud-kage.kakao.com/dn/Q2iNx/btqgeRgV54P/VLdBs9cvyn8BJXB3o7N8UK/kakaolink40_original.png",
    "title_image_text": "Cheese cake",
    "title_image_category": "Cake",
    "items": [
      {"item": "Cake1", "item_op": "1000원"},
      {"item": "Cake2", "item_op": "2000원"}
    ],
    "sum": "Total",
    "sum_op": "15000원"
  },
  "social": {"like_count": 100, "comment_count": 200, "shared_count": 300, "view_count": 400, "subscriber_count": 500},
  "buttons": [
    {"title": "웹으로 이동", "link": {"web_url": "http://www.daum.net", "mobile_web_url": "http://m.daum.net"}},
    {"title": "앱으로 이동", "link": {"android_execution_params": "contentId=100", "ios_execution_params": "contentId=100"}}
  ]
}`

func TestFeed(t *testing.T) {
	t.Parallel()
	img := "https://mud-kage.kakao.com/dn/Q2iNx/btqgeRgV54P/VLdBs9cvyn8BJXB3o7N8UK/kakaolink40_original.png"
	f := Feed{
		Content: Content{
			Title:       "오늘의 디저트",
			Description: "아메리카노, 빵, 케익",
			ImageURL:    "https://mud-kage.kakao.com/dn/NTmhS/btqfEUdFAUf/FjKzkZsnoeE4o19klTOVI1/openlink_640x640s.jpg",
			ImageWidth:  640,
			ImageHeight: 640,
			Link: Link{
				WebURL: "http://www.daum.net", MobileWebURL: "http://m.daum.net",
				AndroidExecutionParams: "contentId=100", IOSExecutionParams: "contentId=100",
			},
		},
		ItemContent: &ItemContent{
			ProfileText: "Kakao", ProfileImageURL: img, TitleImageURL: img,
			TitleImageText: "Cheese cake", TitleImageCategory: "Cake",
			Items: []Item{{Item: "Cake1", ItemOp: "1000원"}, {Item: "Cake2", ItemOp: "2000원"}},
			Sum:   "Total", SumOp: "15000원",
		},
		Social: &Social{LikeCount: 100, CommentCount: 200, SharedCount: 300, ViewCount: 400, SubscriberCount: 500},
		Buttons: []Button{
			{Title: "웹으로 이동", Link: Link{WebURL: "http://www.daum.net", MobileWebURL: "http://m.daum.net"}},
			{Title: "앱으로 이동", Link: Link{AndroidExecutionParams: "contentId=100", IOSExecutionParams: "contentId=100"}},
		},
	}
	assertJSON(t, f, docFeed)
	assertJSON(t, &f, docFeed) // pointer receivers marshal the same
}

func TestTemplates(t *testing.T) {
	t.Parallel()
	link := Link{WebURL: "https://developers.kakao.com"}
	tests := []struct {
		name string
		tmpl Template
		want string
	}{
		{
			"text",
			Text{Text: "안녕하세요", Link: link, ButtonTitle: "바로 확인"},
			`{"object_type":"text","text":"안녕하세요","link":{"web_url":"https://developers.kakao.com"},"button_title":"바로 확인"}`,
		},
		{
			"list",
			List{
				HeaderTitle: "WEEKLY MAGAZINE", HeaderLink: link,
				Contents: []Content{{Title: "a", Link: link}, {Title: "b", ImageURL: "https://x/b.png", Link: link}},
			},
			`{"object_type":"list","header_title":"WEEKLY MAGAZINE","header_link":{"web_url":"https://developers.kakao.com"},
			  "contents":[{"title":"a","link":{"web_url":"https://developers.kakao.com"}},{"title":"b","image_url":"https://x/b.png","link":{"web_url":"https://developers.kakao.com"}}]}`,
		},
		{
			"commerce",
			Commerce{
				Content:  Content{ImageURL: "https://x/p.png", Description: "desc", Link: link},
				Commerce: CommerceInfo{ProductName: "Ivory long dress", RegularPrice: 208800, DiscountPrice: 146160, DiscountRate: 30, CurrencyUnit: "$", CurrencyUnitPosition: CurrencyBefore},
			},
			`{"object_type":"commerce","content":{"image_url":"https://x/p.png","description":"desc","link":{"web_url":"https://developers.kakao.com"}},
			  "commerce":{"product_name":"Ivory long dress","regular_price":208800,"discount_price":146160,"discount_rate":30,"currency_unit":"$","currency_unit_position":1}}`,
		},
		{
			"commerce default position and zero price",
			Commerce{Content: Content{Title: "t", Link: link}, Commerce: CommerceInfo{}},
			`{"object_type":"commerce","content":{"title":"t","link":{"web_url":"https://developers.kakao.com"}},"commerce":{"regular_price":0}}`,
		},
		{
			"location",
			Location{Address: "경기 성남시 분당구 판교역로 235", AddressTitle: "카카오판교오피스", Content: Content{Title: "t", Link: link}, Social: &Social{LikeCount: 1}},
			`{"object_type":"location","address":"경기 성남시 분당구 판교역로 235","address_title":"카카오판교오피스",
			  "content":{"title":"t","link":{"web_url":"https://developers.kakao.com"}},"social":{"like_count":1}}`,
		},
		{
			"calendar",
			Calendar{IDType: CalendarEvent, ID: "abc", Content: Content{Title: "일정 제목", Link: link}, Buttons: []Button{{Title: "일정 정보 보기", Link: link}}},
			`{"object_type":"calendar","id_type":"event","id":"abc","content":{"title":"일정 제목","link":{"web_url":"https://developers.kakao.com"}},
			  "buttons":[{"title":"일정 정보 보기","link":{"web_url":"https://developers.kakao.com"}}]}`,
		},
		{
			"empty feed",
			Feed{},
			`{"object_type":"feed","content":{"link":{}}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertJSON(t, tt.tmpl, tt.want)
		})
	}
}

func TestObjectType(t *testing.T) {
	t.Parallel()
	types := []Template{Feed{}, List{}, Commerce{}, Location{}, Text{}, Calendar{}}
	want := []string{"feed", "list", "commerce", "location", "text", "calendar"}
	for i, tmpl := range types {
		if got := tmpl.ObjectType(); got != want[i] {
			t.Errorf("%T.ObjectType() = %q, want %q", tmpl, got, want[i])
		}
	}
}

func TestMarshal_emptyObject(t *testing.T) {
	t.Parallel()
	b, err := marshal(Text{}, struct{}{})
	if err != nil || string(b) != `{"object_type":"text"}` {
		t.Errorf("got %s, %v", b, err)
	}
	if _, err := marshal(Text{}, func() {}); err == nil {
		t.Error("unencodable fields: want error")
	}
}
