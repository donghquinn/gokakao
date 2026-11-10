package kakao

import (
	"testing"
)

func TestPushService_Register(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithAdminKey("admin"))
	srv.HandleJSON("POST /v2/push/register", 200, `-1`)

	days, err := c.Push.Register(t.Context(), Device{UUID: 9876543211234, DeviceID: "0f365b39", PushType: PushFCM, PushToken: "APA91b"})
	if err != nil || days != -1 {
		t.Fatalf("Register = %d, %v", days, err)
	}
	req := srv.LastRequest()
	if req.Header.Get("Authorization") != "KakaoAK admin" {
		t.Errorf("Authorization = %q", req.Header.Get("Authorization"))
	}
	if !equalValues(req.Form(), map[string][]string{"uuid": {"9876543211234"}, "device_id": {"0f365b39"}, "push_type": {"fcm"}, "push_token": {"APA91b"}}) {
		t.Errorf("form = %v", req.Form())
	}
	if _, err := c.Push.Register(t.Context(), Device{UUID: 1}); err == nil {
		t.Error("incomplete device: want error")
	}
}

func TestPushService_Tokens(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithAdminKey("admin"))
	srv.HandleJSON("GET /v2/push/tokens", 200, `[{"user_id":"9876543211234","uuid":"9876543211234","device_id":"0f365b39","push_type":"fcm","push_token":"APA91b","created_at":"2014-07-29T06:24:12Z","updated_at":"2014-07-29T06:24:12Z"}]`)

	tokens, err := c.Push.Tokens(t.Context(), 9876543211234)
	if err != nil {
		t.Fatal(err)
	}
	if q := srv.LastRequest().Query; q.Get("uuid") != "9876543211234" || q.Has("uuids") {
		t.Errorf("query = %v", q)
	}
	if len(tokens) != 1 || tokens[0].UUID != "9876543211234" || tokens[0].PushType != PushFCM || tokens[0].CreatedAt.IsZero() {
		t.Errorf("tokens = %+v", tokens)
	}

	if _, err := c.Push.Tokens(t.Context(), 1, 2); err != nil {
		t.Fatal(err)
	}
	if q := srv.LastRequest().Query; q.Get("uuids") != `["1","2"]` || q.Has("uuid") {
		t.Errorf("query = %v", q)
	}
	if _, err := c.Push.Tokens(t.Context()); err == nil {
		t.Error("no UUIDs: want error")
	}
	if _, err := c.Push.Tokens(t.Context(), make([]int64, 101)...); err == nil {
		t.Error("101 UUIDs: want error")
	}
}

func TestPushService_Deregister(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithAdminKey("admin"))
	srv.HandleJSON("POST /v2/push/deregister", 200, ``)

	if err := c.Push.Deregister(t.Context(), 1, "dev", PushAPNs); err != nil {
		t.Fatal(err)
	}
	if !equalValues(srv.LastRequest().Form(), map[string][]string{"uuid": {"1"}, "device_id": {"dev"}, "push_type": {"apns"}}) {
		t.Errorf("form = %v", srv.LastRequest().Form())
	}
	if err := c.Push.Deregister(t.Context(), 1, "dev", ""); err != nil {
		t.Fatal(err)
	}
	if f := srv.LastRequest().Form(); len(f) != 1 {
		t.Errorf("device_id without push_type must deregister all: %v", f)
	}
	if err := c.Push.Deregister(t.Context(), 0, "", ""); err == nil {
		t.Error("zero UUID: want error")
	}
}

func TestPushService_Send(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithAdminKey("admin"))
	srv.HandleJSON("POST /v2/push/send", 200, ``)
	badge := 3

	err := c.Push.Send(t.Context(), []int64{1, 2}, PushMessage{
		APNs: &APNsMessage{
			Badge: &badge, Sound: "sound_file", PushAlert: new(bool), ContentAvailable: 1,
			Message:     map[string]any{"loc-key": "COMMENT_ALERT_FORMAT", "loc-args": []string{"홍길동", "2"}},
			CustomField: map[string]any{"article_id": "111"}, APNsEnv: "sandbox",
		},
		FCM: &FCMMessage{Collapse: "articleId123", TimeToLive: 17200, Priority: "high", CustomField: map[string]any{"article_id": 111}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	f := srv.LastRequest().Form()
	if f.Get("uuids") != `["1","2"]` || f.Has("bypass") {
		t.Errorf("form = %v", f)
	}
	assertBodyJSON(t, []byte(f.Get("push_message")), `{
		"for_apns":{"badge":3,"sound":"sound_file","push_alert":false,"content-available":1,
			"message":{"loc-key":"COMMENT_ALERT_FORMAT","loc-args":["홍길동","2"]},"custom_field":{"article_id":"111"},"apns_env":"sandbox"},
		"for_fcm":{"collapse":"articleId123","time_to_live":17200,"priority":"high","custom_field":{"article_id":111}}}`)

	if err := c.Push.Send(t.Context(), []int64{1}, PushMessage{}, &SendOptions{Bypass: true}); err != nil {
		t.Fatal(err)
	}
	if f := srv.LastRequest().Form(); f.Get("bypass") != "true" || f.Get("push_message") != `{}` {
		t.Errorf("form = %v", f)
	}

	if err := c.Push.Send(t.Context(), nil, PushMessage{}, nil); err == nil {
		t.Error("no UUIDs: want error")
	}
	if err := c.Push.Send(t.Context(), []int64{1, 2}, PushMessage{}, &SendOptions{Bypass: true}); err == nil {
		t.Error("bypass with 2 UUIDs: want error")
	}
	if err := c.Push.Send(t.Context(), []int64{1}, PushMessage{APNs: &APNsMessage{Message: func() {}}}, nil); err == nil {
		t.Error("unencodable message: want error")
	}
}
