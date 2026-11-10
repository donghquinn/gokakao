package kakao

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/donghquinn/gokakao/template"
)

var textMsg = template.Text{Text: "hi", Link: template.Link{WebURL: "https://example.com"}}

func TestTalkService_SendToMe(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /v2/api/talk/memo/default/send", 200, `{"result_code":0}`)
	srv.HandleJSON("POST /v2/api/talk/memo/send", 200, `{"result_code":0}`)
	srv.HandleJSON("POST /v2/api/talk/memo/scrap/send", 200, `{"result_code":0}`)

	if err := c.Talk.SendToMe(t.Context(), "AT", textMsg); err != nil {
		t.Fatal(err)
	}
	req := srv.LastRequest()
	if req.Header.Get("Authorization") != "Bearer AT" || req.Header.Get("Content-Type") != contentTypeForm {
		t.Errorf("headers = %v", req.Header)
	}
	if got := req.Form().Get("template_object"); got != `{"object_type":"text","text":"hi","link":{"web_url":"https://example.com"}}` {
		t.Errorf("template_object = %s", got)
	}

	if err := c.Talk.SendCustomToMe(t.Context(), "AT", CustomTemplate{ID: 12345, Args: map[string]string{"TITLE": "타이틀입니다."}}); err != nil {
		t.Fatal(err)
	}
	if f := srv.LastRequest().Form(); f.Get("template_id") != "12345" || f.Get("template_args") != `{"TITLE":"타이틀입니다."}` {
		t.Errorf("form = %v", f)
	}

	if err := c.Talk.SendScrapToMe(t.Context(), "AT", "https://developers.kakao.com", &ScrapOptions{Template: &CustomTemplate{ID: 7}}); err != nil {
		t.Fatal(err)
	}
	if f := srv.LastRequest().Form(); f.Get("request_url") != "https://developers.kakao.com" || f.Get("template_id") != "7" || f.Has("template_args") {
		t.Errorf("form = %v", f)
	}
	if err := c.Talk.SendScrapToMe(t.Context(), "AT", "https://developers.kakao.com", nil); err != nil {
		t.Fatal(err)
	}
	if f := srv.LastRequest().Form(); len(f) != 1 {
		t.Errorf("form = %v", f)
	}
}

func TestTalkService_SendToMe_errors(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /v2/api/talk/memo/default/send", 200, `{"result_code":1}`)
	srv.HandleJSON("POST /v2/api/talk/memo/send", 403, `{"msg":"insufficient scopes.","code":-402,"required_scopes":["talk_message"]}`)

	if err := c.Talk.SendToMe(t.Context(), "AT", textMsg); err == nil || !strings.Contains(err.Error(), "result_code 1") {
		t.Errorf("non-zero result_code: err = %v", err)
	}
	var e *Error
	err := c.Talk.SendCustomToMe(t.Context(), "AT", CustomTemplate{ID: 1})
	if !errors.As(err, &e) || !slices.Equal(e.RequiredScopes, []string{"talk_message"}) {
		t.Errorf("consent: err = %v", err)
	}
	if err := c.Talk.SendToMe(t.Context(), "AT", nil); err == nil {
		t.Error("nil template: want error")
	}
	if err := c.Talk.SendCustomToMe(t.Context(), "AT", CustomTemplate{}); err == nil {
		t.Error("zero template ID: want error")
	}
	if err := c.Talk.SendScrapToMe(t.Context(), "AT", "", nil); err == nil {
		t.Error("empty URL: want error")
	}
	if err := c.Talk.SendScrapToMe(t.Context(), "AT", "https://x", &ScrapOptions{Template: &CustomTemplate{}}); err == nil {
		t.Error("scrap with zero template ID: want error")
	}
	if err := c.Talk.SendToMe(t.Context(), "", textMsg); !errors.Is(err, ErrMissingCredential) {
		t.Errorf("no token: err = %v", err)
	}
}

func TestTalkService_SendToFriends(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	partial := `{"successful_receiver_uuids":["abcdefg0001"],"failure_info":[{"code":-532,"msg":"daily message limit per sender has been exceeded.","receiver_uuids":["abcdefg0002","abcdefg0003"]}]}`
	srv.HandleJSON("POST /v1/api/talk/friends/message/default/send", 200, partial)
	srv.HandleJSON("POST /v1/api/talk/friends/message/send", 200, `{"successful_receiver_uuids":["a"]}`)
	srv.HandleJSON("POST /v1/api/talk/friends/message/scrap/send", 200, `{"successful_receiver_uuids":["a"]}`)
	uuids := []string{"abcdefg0001", "abcdefg0002", "abcdefg0003"}

	res, err := c.Talk.SendToFriends(t.Context(), "AT", uuids, textMsg)
	if err != nil {
		t.Fatal(err)
	}
	f := srv.LastRequest().Form()
	if f.Get("receiver_uuids") != `["abcdefg0001","abcdefg0002","abcdefg0003"]` || !f.Has("template_object") {
		t.Errorf("form = %v", f)
	}
	if !slices.Equal(res.SuccessfulReceiverUUIDs, []string{"abcdefg0001"}) || len(res.FailureInfo) != 1 ||
		res.FailureInfo[0].Code != -532 || !slices.Equal(res.FailureInfo[0].ReceiverUUIDs, uuids[1:]) {
		t.Errorf("res = %+v", res)
	}

	if _, err := c.Talk.SendCustomToFriends(t.Context(), "AT", []string{"a"}, CustomTemplate{ID: 3, Args: map[string]string{"K": "V"}}); err != nil {
		t.Fatal(err)
	}
	if f := srv.LastRequest().Form(); f.Get("template_id") != "3" || f.Get("template_args") != `{"K":"V"}` || f.Get("receiver_uuids") != `["a"]` {
		t.Errorf("form = %v", f)
	}

	if _, err := c.Talk.SendScrapToFriends(t.Context(), "AT", []string{"a"}, "https://x", nil); err != nil {
		t.Fatal(err)
	}
	if f := srv.LastRequest().Form(); f.Get("request_url") != "https://x" {
		t.Errorf("form = %v", f)
	}

	for _, n := range []int{0, 6} {
		if _, err := c.Talk.SendToFriends(t.Context(), "AT", make([]string, n), textMsg); err == nil {
			t.Errorf("%d receivers: want error", n)
		}
	}
	if _, err := c.Talk.SendToFriends(t.Context(), "AT", []string{"a"}, nil); err == nil {
		t.Error("nil template: want error")
	}
	if _, err := c.Talk.SendCustomToFriends(t.Context(), "AT", []string{"a"}, CustomTemplate{}); err == nil {
		t.Error("zero template ID: want error")
	}
	if _, err := c.Talk.SendScrapToFriends(t.Context(), "AT", []string{"a"}, "", nil); err == nil {
		t.Error("empty URL: want error")
	}
}

func TestTalkService_SendToFriends_singleFailure(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /v1/api/talk/friends/message/default/send", 400, `{"msg":"not a friend","code":-502}`)
	_, err := c.Talk.SendToFriends(t.Context(), "AT", []string{"a"}, textMsg)
	var e *Error
	if !errors.As(err, &e) || e.Code != -502 {
		t.Errorf("err = %v", err)
	}
}

func TestTalkService_Profile(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /v1/api/talk/profile", 200, `{"id":1234567,"nickName":"홍길동","profileImageURL":"https://xxx.kakao.co.kr/.../aaa.jpg","thumbnailURL":"https://xxx.kakao.co.kr/.../bbb.jpg"}`)
	p, err := c.Talk.Profile(t.Context(), "AT")
	if err != nil {
		t.Fatal(err)
	}
	if *p != (TalkProfile{ID: 1234567, Nickname: "홍길동", ProfileImageURL: "https://xxx.kakao.co.kr/.../aaa.jpg", ThumbnailURL: "https://xxx.kakao.co.kr/.../bbb.jpg"}) {
		t.Errorf("profile = %+v", p)
	}
}

func TestTalkService_Friends(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleFile("GET /v1/api/talk/friends", "testdata/talk/friends.json")

	page, err := c.Talk.Friends(t.Context(), "AT", &FriendsOptions{Offset: 3, Limit: 3, Descending: true, FriendOrder: FriendOrderNickname})
	if err != nil {
		t.Fatal(err)
	}
	q := srv.LastRequest().Query
	if q.Get("offset") != "3" || q.Get("limit") != "3" || q.Get("order") != "desc" || q.Get("friend_order") != "nickname" {
		t.Errorf("query = %v", q)
	}
	if page.TotalCount != 11 || page.FavoriteCount != 1 || len(page.Elements) != 3 || page.AfterURL == "" {
		t.Errorf("page = %+v", page)
	}
	if f := page.Elements[0]; f != (Friend{ID: 1234567, UUID: "abcdefg0001", Favorite: true, ProfileNickname: "이수민", ProfileThumbnailImage: "https://xxx.kakao.co.kr/.../aaa.jpg"}) {
		t.Errorf("friend = %+v", f)
	}
}

func TestTalkService_AllFriends(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.Handle("GET /v1/api/talk/friends", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("offset") {
		case "":
			_, _ = w.Write([]byte(`{"total_count":3,"elements":[{"uuid":"a"},{"uuid":"b"}],"after_url":"https://kapi.kakao.com/v1/api/talk/friends?offset=2&limit=2"}`))
		case "2":
			_, _ = w.Write([]byte(`{"total_count":3,"elements":[{"uuid":"c"}],"after_url":null}`))
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	})
	var got []string
	for f, err := range c.Talk.AllFriends(t.Context(), "AT", &FriendsOptions{Limit: 2}) {
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, f.UUID)
	}
	if !slices.Equal(got, []string{"a", "b", "c"}) {
		t.Errorf("uuids = %v", got)
	}
	if q := srv.Requests()[0].Query; q.Get("limit") != "2" {
		t.Errorf("first query = %v", q)
	}

	for f, err := range c.Talk.AllFriends(t.Context(), "AT", nil) {
		if err != nil || f.UUID != "a" {
			t.Errorf("first friend = %+v, %v", f, err)
		}
		break
	}
	if q := srv.LastRequest().Query; q.Get("limit") != "100" {
		t.Errorf("default page size: query = %v", q)
	}

	var gotErr error
	for _, err := range c.Talk.AllFriends(t.Context(), "", nil) {
		gotErr = err
	}
	if !errors.Is(gotErr, ErrMissingCredential) {
		t.Errorf("err = %v", gotErr)
	}
}

func TestTalkService_Profiles(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithAdminKey("admin"))
	srv.HandleJSON("GET /v2/api/talk/profiles", 200, `{"viewer":{"user_id":12345,"nickname":"xxx"},"targets":[{"user_id":9876,"nickname":"yyy","profile_image":"p","thumbnail_image":"t"},{"user_id":5432,"nickname":"zzz"}]}`)

	res, err := c.Talk.Profiles(t.Context(), []int64{9876, 5432}, &ProfilesOptions{ViewerID: 12345})
	if err != nil {
		t.Fatal(err)
	}
	req := srv.LastRequest()
	if req.Header.Get("Authorization") != "KakaoAK admin" {
		t.Errorf("Authorization = %q", req.Header.Get("Authorization"))
	}
	q := req.Query
	if q.Get("target_ids") != "[9876,5432]" || q.Get("target_id_type") != "user_id" || q.Get("viewer_id") != "12345" || q.Get("viewer_id_type") != "user_id" {
		t.Errorf("query = %v", q)
	}
	if res.Viewer == nil || res.Viewer.UserID != 12345 || len(res.Targets) != 2 ||
		res.Targets[0] != (TalkProfileInfo{UserID: 9876, Nickname: "yyy", ProfileImage: "p", ThumbnailImage: "t"}) {
		t.Errorf("res = %+v", res)
	}

	if _, err := c.Talk.Profiles(t.Context(), []int64{1}, nil); err != nil {
		t.Fatal(err)
	}
	if q := srv.LastRequest().Query; q.Has("viewer_id") || q.Has("viewer_id_type") {
		t.Errorf("viewer sent: %v", q)
	}
	if _, err := c.Talk.Profiles(t.Context(), nil, nil); err == nil {
		t.Error("no IDs: want error")
	}
}

func TestSendResult_JSON(t *testing.T) {
	t.Parallel()
	var r SendResult
	if err := json.Unmarshal([]byte(`{"successful_receiver_uuids":[]}`), &r); err != nil || r.SuccessfulReceiverUUIDs == nil || r.FailureInfo != nil {
		t.Errorf("r = %+v, %v", r, err)
	}
}
