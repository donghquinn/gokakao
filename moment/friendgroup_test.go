package moment

import (
	"bytes"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"slices"
	"testing"
	"time"

	kakao "github.com/donghquinn/gokakao"
	"github.com/donghquinn/gokakao/internal/kakaotest"
)

func TestFriendGroupService_List(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/talkChannelGroupFiles", 200, `[
  {
    "id": 1,
    "profileId": 1,
    "name": "첫번째_친구",
    "talkChannelGroupFileStatus": "COMPLETE",
    "talkChannelProfileName": "첫번째_친구",
    "fileType": "APP_USER_ID",
    "groupKey": "${GROUP_KEY}",
    "createdDate": "2020-01-01 00:00:00"
  }
]`)

	gs, err := c.FriendGroups.List(t.Context(), 1111)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodGet || req.Path != "/openapi/v4/talkChannelGroupFiles" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	want := time.Date(2020, 1, 1, 0, 0, 0, 0, time.FixedZone("KST", 9*3600))
	if len(gs) != 1 || gs[0].ID != 1 || gs[0].ProfileID != "1" || gs[0].Name != "첫번째_친구" ||
		gs[0].Status != FriendGroupComplete || gs[0].TalkChannelProfileName != "첫번째_친구" ||
		gs[0].FileType != CustomerDataFileAppUserID || gs[0].GroupKey != "${GROUP_KEY}" || !gs[0].CreatedDate.Equal(want) {
		t.Errorf("groups = %+v", gs)
	}

	if _, err := c.FriendGroups.List(t.Context(), 0); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestFriendGroup_UnmarshalProfileID(t *testing.T) {
	t.Parallel()
	for in, want := range map[string]string{`{"profileId":"_ZQxd"}`: "_ZQxd", `{"profileId":12345}`: "12345", `{"profileId":null}`: "", `{}`: ""} {
		var g FriendGroup
		if err := g.UnmarshalJSON([]byte(in)); err != nil || g.ProfileID != want {
			t.Errorf("%s: ProfileID = %q, %v; want %q", in, g.ProfileID, err, want)
		}
	}
	var g FriendGroup
	if err := g.UnmarshalJSON([]byte(`{"profileId":true}`)); err == nil {
		t.Error("bool profileId: want error")
	}
}

// friendgroupForm parses the multipart body of req.
func friendgroupForm(t *testing.T, req kakaotest.Request) *multipart.Form {
	t.Helper()
	mt, params, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil || mt != "multipart/form-data" {
		t.Fatalf("Content-Type = %q, %v", req.Header.Get("Content-Type"), err)
	}
	form, err := multipart.NewReader(bytes.NewReader(req.Body), params["boundary"]).ReadForm(1 << 20)
	if err != nil {
		t.Fatal(err)
	}
	return form
}

func TestFriendGroupService_Create(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v4/talkChannelGroupFiles", 200, `{
  "adAccountId": 123456789,
  "id": 645645,
  "name": "친구그룹타겟이름",
  "fileType": "APP_USER_ID",
  "adGroupIds": 12345,
  "successCount": 20,
  "failedCount": 1,
  "failedFileUrl": "${FILE_URL}",
  "status": "WAITING",
  "createdDate": "2022-01-01 00:00:00"
}`)

	csv := []byte("id\n1234\n")
	u, err := c.FriendGroups.Create(t.Context(), 123456789, NewFriendGroup{
		Name: "친구그룹타겟이름", ProfileID: "12345", FileType: CustomerDataFileAppUserID,
		Files: []UploadFile{{Filename: "친구그룹파일.csv", Data: csv}, {Filename: "b.csv", ContentType: "text/plain", Data: csv}},
	})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "123456789")
	if req.Method != http.MethodPost || req.Path != "/openapi/v4/talkChannelGroupFiles" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	form := friendgroupForm(t, req)
	if !slices.Equal(form.Value["name"], []string{"친구그룹타겟이름"}) || !slices.Equal(form.Value["profileId"], []string{"12345"}) ||
		!slices.Equal(form.Value["fileType"], []string{"APP_USER_ID"}) || form.Value["adGroupIds"] != nil {
		t.Errorf("fields = %v", form.Value)
	}
	files := form.File["files"]
	if len(files) != 2 || files[0].Filename != "친구그룹파일.csv" || files[1].Filename != "b.csv" || files[1].Header.Get("Content-Type") != "text/plain" ||
		files[0].Header.Get("Content-Type") != "text/csv" {
		t.Fatalf("files = %+v", files)
	}
	f, _ := files[0].Open()
	data, _ := io.ReadAll(f)
	_ = f.Close()
	if !bytes.Equal(data, csv) {
		t.Errorf("file data = %q", data)
	}
	if u.AdAccountID != 123456789 || u.ID != 645645 || u.Name != "친구그룹타겟이름" || u.FileType != CustomerDataFileAppUserID ||
		!slices.Equal(u.AdGroupIDs, []kakao.Int64{12345}) || u.SuccessCount != 20 || u.FailedCount != 1 ||
		u.FailedFileURL != "${FILE_URL}" || u.Status != FriendGroupWaiting || u.CreatedDate.Year() != 2022 {
		t.Errorf("upload = %+v", u)
	}

	// MESSAGE_RETARGET sends ad group IDs and no files.
	if _, err := c.FriendGroups.Create(t.Context(), 1, NewFriendGroup{Name: "n", ProfileID: "p", FileType: CustomerDataFileMessageRetarget, AdGroupIDs: []int64{11, 22}}); err != nil {
		t.Fatal(err)
	}
	form = friendgroupForm(t, srv.LastRequest())
	if !slices.Equal(form.Value["adGroupIds"], []string{"11", "22"}) || len(form.File) != 0 {
		t.Errorf("retarget form = %v %v", form.Value, form.File)
	}

	for _, g := range []NewFriendGroup{
		{ProfileID: "p", FileType: CustomerDataFilePhoneNumber, Files: []UploadFile{{Filename: "a.csv", Data: csv}}},
		{Name: "n", FileType: CustomerDataFilePhoneNumber, Files: []UploadFile{{Filename: "a.csv", Data: csv}}},
		{Name: "n", ProfileID: "p", FileType: CustomerDataFilePhoneNumber},
		{Name: "n", ProfileID: "p", FileType: CustomerDataFileMessageRetarget},
		{Name: "n", ProfileID: "p", FileType: CustomerDataFileADID, Files: []UploadFile{{Filename: "a.csv", Data: csv}}},
		{Name: "n", ProfileID: "p"},
		{Name: "n", ProfileID: "p", FileType: CustomerDataFilePhoneNumber, Files: []UploadFile{{Data: csv}}},
	} {
		if _, err := c.FriendGroups.Create(t.Context(), 1, g); err == nil {
			t.Errorf("Create(%+v): want error", g)
		}
	}
	if _, err := c.FriendGroups.Create(t.Context(), 0, NewFriendGroup{Name: "n", ProfileID: "p", FileType: CustomerDataFileMessageRetarget, AdGroupIDs: []int64{1}}); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestFriendGroupUpload_UnmarshalAdGroupIDs(t *testing.T) {
	t.Parallel()
	for in, want := range map[string][]kakao.Int64{
		`{"adGroupIds":[1,"2"]}`: {1, 2},
		`{"adGroupIds":"3"}`:     {3},
		`{"adGroupIds":null}`:    nil,
		`{}`:                     nil,
	} {
		var u FriendGroupUpload
		if err := u.UnmarshalJSON([]byte(in)); err != nil || !slices.Equal(u.AdGroupIDs, want) {
			t.Errorf("%s: AdGroupIDs = %v, %v", in, u.AdGroupIDs, err)
		}
	}
	var u FriendGroupUpload
	if err := u.UnmarshalJSON([]byte(`{"adGroupIds":"x"}`)); err == nil {
		t.Error("bad adGroupIds: want error")
	}
}

func TestFriendGroupService_Rename(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PUT /openapi/v4/talkChannelGroupFiles/name", 200, `{
  "id": 1,
  "profileId": 1,
  "name": "첫번째_친구그룹_이름수정",
  "talkChannelGroupFileStatus": "COMPLETE",
  "talkChannelProfileName": "첫번째_친구그룹_이름수정",
  "fileType": "APP_USER_ID",
  "groupKey": "${GROUP_KEY}",
  "createdDate": "2020-01-01 00:00:00",
  "lastModifiedDate": "2020-01-01 15:00:00"
}`)

	g, err := c.FriendGroups.Rename(t.Context(), 1111, 1234, "첫번째_친구그룹_이름수정")
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodPut || string(req.Body) != `{"id":1234,"name":"첫번째_친구그룹_이름수정"}` {
		t.Errorf("request = %s %s", req.Method, req.Body)
	}
	if g.ID != 1 || g.Name != "첫번째_친구그룹_이름수정" || g.LastModifiedDate.Hour() != 15 {
		t.Errorf("group = %+v", g)
	}

	if _, err := c.FriendGroups.Rename(t.Context(), 1, 2, ""); err == nil {
		t.Error("missing name: want error")
	}
	if _, err := c.FriendGroups.Rename(t.Context(), 1, 0, "n"); err == nil {
		t.Error("missing group: want error")
	}
}

func TestFriendGroupService_Delete(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("DELETE /openapi/v4/talkChannelGroupFiles/1234", 200, ``)

	if err := c.FriendGroups.Delete(t.Context(), 1111, 1234); err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodDelete || req.Path != "/openapi/v4/talkChannelGroupFiles/1234" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	if err := c.FriendGroups.Delete(t.Context(), 0, 1); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestFriendGroupService_DeleteMany(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("DELETE /openapi/v4/talkChannelGroupFiles", 200, ``)

	if err := c.FriendGroups.DeleteMany(t.Context(), 1111, []int64{1, 2}); err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodDelete || req.Query.Get("ids") != "1,2" {
		t.Errorf("request = %s %v", req.Method, req.Query)
	}
	if err := c.FriendGroups.DeleteMany(t.Context(), 1111, nil); err == nil {
		t.Error("no IDs: want error")
	}
	if err := c.FriendGroups.DeleteMany(t.Context(), 0, []int64{1}); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestFriendGroupService_Usages(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/talkChannelGroupFiles/usages/1234", 200, `[
  {
    "adGroup": {
      "id": 105488,
      "name": "first_ad_group",
      "adGroupStatus": ["LIVE"],
      "adGroupType": "DISPLAY"
    },
    "campaign": {
      "id": 62286,
      "name": "first_campaign",
      "campaignTypeGoal": {
        "campaignType": "DISPLAY",
        "goal": "VISITING"
      }
    }
  },
  {
    "messageAd": {
      "id": "msg-ad-1362737813386051585",
      "name": "first_message",
      "opStatus": ["READY"]
    }
  }
]`)

	us, err := c.FriendGroups.Usages(t.Context(), 1111, 1234)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodGet || req.Path != "/openapi/v4/talkChannelGroupFiles/usages/1234" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	if len(us) != 2 || us[0].AdGroup == nil || us[0].AdGroup.ID != 105488 || !slices.Equal(us[0].AdGroup.AdGroupStatus, []AdGroupStatus{AdGroupStatusLive}) ||
		us[0].Campaign == nil || us[0].Campaign.ID != 62286 || us[0].Campaign.CampaignTypeGoal.Goal != GoalVisiting || us[0].MessageAd != nil ||
		us[1].MessageAd == nil || us[1].MessageAd.ID != "msg-ad-1362737813386051585" || us[1].MessageAd.Name != "first_message" {
		t.Errorf("usages = %+v", us)
	}
	if _, err := c.FriendGroups.Usages(t.Context(), 1111, 0); err == nil {
		t.Error("missing group: want error")
	}
}
