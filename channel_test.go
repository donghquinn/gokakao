package kakao

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestChannelService_Relations(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithAdminKey("admin"))
	srv.HandleJSON("GET /v2/api/talk/channels", 200, `{"user_id":1,"channels":[
		{"channel_uuid":"@ad","channel_public_id":"_frxjem","relation":"ADDED","created_at":"2023-01-01T00:00:00Z","updated_at":"2023-01-02T00:00:00Z"},
		{"channel_uuid":"@b","channel_public_id":"_xnrxjem","relation":"NONE"}]}`)

	res, err := c.Channel.Relations(t.Context(), AccessToken("AT"), "_frxjem", "_xnrxjem")
	if err != nil {
		t.Fatal(err)
	}
	q := srv.LastRequest().Query
	if q.Get("channel_ids") != "_frxjem,_xnrxjem" || q.Get("channel_id_type") != "channel_public_id" {
		t.Errorf("query = %v", q)
	}
	ch := res.Channels[0]
	if res.UserID != 1 || ch.Relation != RelationAdded || ch.ChannelUUID != "@ad" ||
		!ch.CreatedAt.Equal(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)) || res.Channels[1].Relation != RelationNone || !res.Channels[1].UpdatedAt.IsZero() {
		t.Errorf("res = %+v", res)
	}

	if _, err := c.Channel.Relations(t.Context(), AsAdmin(1)); err != nil {
		t.Fatal(err)
	}
	q = srv.LastRequest().Query
	if q.Has("channel_ids") || q.Get("target_id") != "1" || q.Get("target_id_type") != "user_id" {
		t.Errorf("admin query = %v", q)
	}
}

func TestChannelService_MultiRelations(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithAdminKey("admin"))
	srv.HandleJSON("GET /v2/api/talk/channels/multi", 200, `[{"user_id":1,"channels":[{"channel_public_id":"_a","relation":"BLOCKED"}]},{"user_id":2,"channels":[]}]`)

	res, err := c.Channel.MultiRelations(t.Context(), []int64{1, 2}, "_a")
	if err != nil {
		t.Fatal(err)
	}
	req := srv.LastRequest()
	if req.Header.Get("Authorization") != "KakaoAK admin" {
		t.Errorf("Authorization = %q", req.Header.Get("Authorization"))
	}
	if q := req.Query; q.Get("target_ids") != "1,2" || q.Get("target_id_type") != "user_id" || q.Get("channel_ids") != "_a" {
		t.Errorf("query = %v", q)
	}
	if len(res) != 2 || res[0].Channels[0].Relation != RelationBlocked {
		t.Errorf("res = %+v", res)
	}
	for _, n := range []int{0, 201} {
		if _, err := c.Channel.MultiRelations(t.Context(), make([]int64, n)); err == nil {
			t.Errorf("%d users: want error", n)
		}
	}
}

func TestChannelService_userFiles(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithRESTKey("rest"), WithAdminKey("admin"))
	srv.HandleJSON("POST /v1/talkchannel/create/target_user_file", 200, `{"file_id":437}`)
	srv.HandleJSON("GET /v1/talkchannel/target_user_file", 200, `{"empty_slot":27,"using_slot":3,"results":[{"file_id":437,"file_name":"vip고객리스트","status":"using","update_at":"2019-07-05T11:53:51","schema":{"생년월일":"string","연령":"number"}}]}`)
	srv.HandleJSON("POST /v1/talkchannel/update/target_users", 200, `{"file_id":437,"request_count":10,"success_count":9}`)
	srv.HandleJSON("POST /v1/talkchannel/delete/target_users", 200, ``)
	ctx := t.Context()

	id, err := c.Channel.CreateUserFile(ctx, "_ZeUTxl", "vip고객리스트", map[string]SchemaType{"생년월일": SchemaString, "연령": SchemaNumber})
	if err != nil || id != 437 {
		t.Fatalf("CreateUserFile = %d, %v", id, err)
	}
	req := srv.LastRequest()
	if req.Header.Get("Authorization") != "KakaoAK rest" || req.Header.Get("Content-Type") != contentTypeJSON {
		t.Errorf("headers = %v", req.Header)
	}
	assertBodyJSON(t, req.Body, `{"channel_public_id":"_ZeUTxl","file_name":"vip고객리스트","schema":{"생년월일":"string","연령":"number"}}`)

	files, err := c.Channel.UserFiles(ctx, "_ZeUTxl")
	if err != nil {
		t.Fatal(err)
	}
	if srv.LastRequest().Query.Get("channel_public_id") != "_ZeUTxl" {
		t.Errorf("query = %v", srv.LastRequest().Query)
	}
	if files.EmptySlot != 27 || files.UsingSlot != 3 || files.Results[0].Status != "using" || files.Results[0].Schema["연령"] != SchemaNumber {
		t.Errorf("files = %+v", files)
	}

	added, err := c.Channel.AddUsers(ctx, "_ZeUTxl", 437, UserIDApp, []FileUser{{ID: "12345", Field: map[string]any{"생년월일": "2000-01-01", "연령": 19}}})
	if err != nil {
		t.Fatal(err)
	}
	assertBodyJSON(t, srv.LastRequest().Body, `{"file_id":437,"channel_public_id":"_ZeUTxl","user_type":"app","users":[{"id":"12345","field":{"생년월일":"2000-01-01","연령":19}}]}`)
	if *added != (AddUsersResult{FileID: 437, RequestCount: 10, SuccessCount: 9}) {
		t.Errorf("added = %+v", added)
	}

	if err := c.Channel.RemoveUsers(ctx, "_ZeUTxl", 437, UserIDPhone, []string{"01012345678"}); err != nil {
		t.Fatal(err)
	}
	assertBodyJSON(t, srv.LastRequest().Body, `{"file_id":437,"channel_public_id":"_ZeUTxl","user_type":"phone","user_ids":["01012345678"]}`)
}

func TestChannelService_userFiles_adminKeyAndValidation(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithAdminKey("admin"))
	srv.HandleJSON("GET /v1/talkchannel/target_user_file", 200, `{"results":[]}`)
	if _, err := c.Channel.UserFiles(t.Context(), "_a"); err != nil {
		t.Fatal(err)
	}
	if got := srv.LastRequest().Header.Get("Authorization"); got != "KakaoAK admin" {
		t.Errorf("Authorization = %q, want admin key fallback", got)
	}

	ctx := t.Context()
	errs := []error{}
	_, err := c.Channel.CreateUserFile(ctx, "", "f", map[string]SchemaType{"a": SchemaNumber})
	errs = append(errs, err)
	_, err = c.Channel.CreateUserFile(ctx, "_a", "f", nil)
	errs = append(errs, err)
	_, err = c.Channel.UserFiles(ctx, "")
	errs = append(errs, err)
	_, err = c.Channel.AddUsers(ctx, "_a", 1, UserIDApp, nil)
	errs = append(errs, err)
	_, err = c.Channel.AddUsers(ctx, "_a", 1, UserIDApp, make([]FileUser, 2001))
	errs = append(errs, err)
	errs = append(errs, c.Channel.RemoveUsers(ctx, "_a", 1, UserIDApp, nil))
	for i, err := range errs {
		if err == nil {
			t.Errorf("case %d: want error", i)
		}
	}
}

func assertBodyJSON(t *testing.T, body []byte, want string) {
	t.Helper()
	var g, w any
	if err := json.Unmarshal(body, &g); err != nil {
		t.Fatalf("body %q: %v", body, err)
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(want)), &w); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(g, w) {
		t.Errorf("body = %s\nwant   %s", body, want)
	}
}
