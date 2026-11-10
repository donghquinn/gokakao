package moment

import (
	"net/http"
	"testing"
)

// targetingAssertGet checks that a request was a GET to path.
func targetingAssertGet(t *testing.T, method, gotPath, path string) {
	t.Helper()
	if method != http.MethodGet || gotPath != path {
		t.Errorf("request = %s %s, want GET %s", method, gotPath, path)
	}
}

func TestTargetingService_Depth1Locations(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/targetings/location/depth1", 200, `[
	  {"value": "A", "description": "강원도", "depth1Name": "강원도"},
	  {"value": "B", "description": "경기도", "depth1Name": "경기도"},
	  {"value": "C", "description": "경상남도", "depth1Name": "경상남도"}]`)

	ls, err := c.Targeting.Depth1Locations(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "")
	targetingAssertGet(t, req.Method, req.Path, "/openapi/v4/targetings/location/depth1")
	if len(ls) != 3 || ls[1].Value != "B" || ls[1].Description != "경기도" || ls[2].Depth1Name != "경상남도" {
		t.Errorf("locations = %+v", ls)
	}
}

func TestTargetingService_Depth2Locations(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/targetings/location/depth2", 200, `[
	  {"value": "I1000", "description": "서울특별시 강남구", "depth1Name": "서울특별시", "depth2Name": "강남구"},
	  {"value": "I1001", "description": "서울특별시 강동구", "depth1Name": "서울특별시", "depth2Name": "강동구"}]`)

	ls, err := c.Targeting.Depth2Locations(t.Context(), "서울")
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "")
	targetingAssertGet(t, req.Method, req.Path, "/openapi/v4/targetings/location/depth2")
	if req.Query.Get("keyword") != "서울" {
		t.Errorf("query = %v", req.Query)
	}
	if len(ls) != 2 || ls[0].Value != "I1000" || ls[1].Depth2Name != "강동구" || ls[1].Depth1Name != "서울특별시" {
		t.Errorf("locations = %+v", ls)
	}
	if _, err := c.Targeting.Depth2Locations(t.Context(), ""); err == nil {
		t.Error("empty keyword: want error")
	}
}

func TestTargetingService_Depth3Locations(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/targetings/location/depth3", 200, `[
	  {"value": "I10000101", "description": "서울특별시 강남구", "depth1Name": "서울특별시", "depth2Name": "강남구", "depth3Name": "논현1동"},
	  {"value": "I10000102", "description": "서울특별시 강동구", "depth1Name": "서울특별시", "depth2Name": "강남구", "depth3Name": "논현2동"}]`)

	ls, err := c.Targeting.Depth3Locations(t.Context(), "논현")
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "")
	targetingAssertGet(t, req.Method, req.Path, "/openapi/v4/targetings/location/depth3")
	if req.Query.Get("keyword") != "논현" {
		t.Errorf("query = %v", req.Query)
	}
	if len(ls) != 2 || ls[0].Value != "I10000101" || ls[1].Depth3Name != "논현2동" || ls[1].Depth2Name != "강남구" {
		t.Errorf("locations = %+v", ls)
	}
	if _, err := c.Targeting.Depth3Locations(t.Context(), ""); err == nil {
		t.Error("empty keyword: want error")
	}
}

func TestTargetingService_AdServingCategories(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/targetings/placement/adServingCategories", 200, `[
	  {"code": "IAB16", "name": "애완동물"}, {"code": "IAB8", "name": "식음료"}, {"code": "IAB15", "name": "과학"}]`)

	cs, err := c.Targeting.AdServingCategories(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "")
	targetingAssertGet(t, req.Method, req.Path, "/openapi/v4/targetings/placement/adServingCategories")
	if len(cs) != 3 || cs[0].Code != NetworkCategoryIAB16 || cs[0].Name != "애완동물" || cs[2].Code != NetworkCategoryIAB15 {
		t.Errorf("categories = %+v", cs)
	}
}

func TestTargetingService_SectionCategories(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/targetings/placement/sectionCategories", 200, `[
	  {"code": "KKO99-1", "name": "채팅지면 노출"}, {"code": "KKO99-2", "name": "오픈채팅지면 노출"},
	  {"code": "KKO99-4", "name": "친구지면 노출"}, {"code": "KKO99-5", "name": "친구소식지면 노출"},
	  {"code": "KKO99-6", "name": "숏폼지면 노출"}, {"code": "KKO99-7", "name": "기타지면 노출"},
	  {"code": "KKO99-8", "name": "카카오쇼핑지면 노출"}]`)

	cs, err := c.Targeting.SectionCategories(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "")
	targetingAssertGet(t, req.Method, req.Path, "/openapi/v4/targetings/placement/sectionCategories")
	if len(cs) != 7 || cs[0].Code != TalkSectionChat || cs[6].Code != TalkSectionKakaoShopping || cs[6].Name != "카카오쇼핑지면 노출" {
		t.Errorf("categories = %+v", cs)
	}
}

func TestTargetingService_Cohorts(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/adGroups/targetings/cohort/availables", 200, `[
	  {"adAccountId": 1234, "id": 1, "audienceType": "DISPLAY", "name": "첫번째_광고반응타겟",
	   "baseAds": [{"campaign": {"id": 5678, "name": "첫번째_캠페인", "campaignTypeGoal": {"campaignType": "DISPLAY", "goal": "VISITING"}},
	     "adGroup": {"id": 9012, "name": "첫번째_광고그룹"}, "operation": "MINUS", "firstIndicator": "OPEN", "secondIndicator": "CLICK"}],
	   "collectDuration": 90, "status": "AVAILABLE", "createdDate": "2018-02-07 10:46:12", "lastModifiedDate": "2020-07-07 19:15:38"},
	  {"adAccountId": 1234, "id": 2, "audienceType": "DISPLAY", "name": "두번째_광고반응타겟",
	   "baseAds": [{"campaign": {"id": 5678, "name": "두번째_캠페인", "campaignTypeGoal": {"campaignType": "DISPLAY", "goal": "CONVERSION"}},
	     "adGroup": {"id": 9012, "name": "두번째_광고그룹"}, "operation": "ONLY", "firstIndicator": "CONVERSION"}],
	   "collectDuration": 90, "status": "AVAILABLE", "createdDate": "2020-01-01 00:00:00", "lastModifiedDate": "2020-01-01 00:00:00"}]`)

	cs, err := c.Targeting.Cohorts(t.Context(), 1234)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1234")
	targetingAssertGet(t, req.Method, req.Path, "/openapi/v4/adGroups/targetings/cohort/availables")
	if len(cs) != 2 || cs[0].ID != 1 || cs[0].AdAccountID != 1234 || cs[0].AudienceType != AudienceDisplay ||
		cs[0].CollectDuration != 90 || cs[0].Status != TargetableCohortAvailable || cs[0].CreatedDate.Year() != 2018 ||
		cs[0].LastModifiedDate.Month() != 7 || cs[1].Name != "두번째_광고반응타겟" {
		t.Fatalf("cohorts = %+v", cs)
	}
	b := cs[0].BaseAds
	if len(b) != 1 || b[0].Operation != OperationMinus || b[0].FirstIndicator != IndicatorOpen || b[0].SecondIndicator != IndicatorClick ||
		b[0].Campaign == nil || b[0].Campaign.ID != 5678 || b[0].Campaign.CampaignTypeGoal.Goal != GoalVisiting ||
		b[0].AdGroup == nil || b[0].AdGroup.ID != 9012 {
		t.Errorf("base ads = %+v", b)
	}
	if _, err := c.Targeting.Cohorts(t.Context(), 0); err == nil {
		t.Error("missing ad account ID: want error")
	}
}

func TestTargetingService_TrackerEvents(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/adGroups/targetings/trackers/eventCreatables/3830610710806837838", 200, `[
	  {"lastEventDate": "2019.10.16 19:36", "eventCode": "*", "eventName": "모든 이벤트", "trackRuleId": "*", "population": 15300000},
	  {"trackRule": {"trackRuleId": "*", "trackId": "3830610710806837838", "eventCode": "PageView", "name": "방문", "ruleSetStr": "[]"},
	   "lastEventDate": "2019.10.16 19:36", "eventCode": "PageView", "eventName": "방문", "trackRuleId": "*", "population": 15290918}]`)

	es, err := c.Targeting.TrackerEvents(t.Context(), 1234, "3830610710806837838", 30)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1234")
	targetingAssertGet(t, req.Method, req.Path, "/openapi/v4/adGroups/targetings/trackers/eventCreatables/3830610710806837838")
	if req.Query.Get("term") != "30" {
		t.Errorf("query = %v", req.Query)
	}
	if len(es) != 2 || es[0].EventCode != "*" || es[0].TrackRule != nil || es[0].Population != 15300000 ||
		es[0].LastEventDate != "2019.10.16 19:36" || es[1].EventName != "방문" || es[1].TrackRuleID != "*" {
		t.Fatalf("events = %+v", es)
	}
	if r := es[1].TrackRule; r == nil || r.TrackID != "3830610710806837838" || r.EventCode != "PageView" || r.Name != "방문" ||
		string(r.RuleSetStr) != `"[]"` {
		t.Errorf("track rule = %+v", r)
	}
	for _, tc := range []struct {
		trackID string
		term    int
	}{{"", 30}, {"x", 0}, {"x", 181}} {
		if _, err := c.Targeting.TrackerEvents(t.Context(), 1234, tc.trackID, tc.term); err == nil {
			t.Errorf("%+v: want error", tc)
		}
	}
	if _, err := c.Targeting.TrackerEvents(t.Context(), 0, "x", 1); err == nil {
		t.Error("missing ad account ID: want error")
	}
}

func TestTargetingService_CustomerData(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/adGroups/targetings/customerData/availables", 200, `[
	  {"id": 1, "adAccountId": 1234, "name": "첫번째_고객데이터", "kuidListKey": "123456789abcdefghijklmn",
	   "fileType": "ADID", "status": "COMPLETE", "populationScore": 100, "ready": true,
	   "createdDate": "2020-01-01 00:00", "lastModifyRequestDate": "2020-01-01 00:00"}]`)

	ds, err := c.Targeting.CustomerData(t.Context(), 1234)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1234")
	targetingAssertGet(t, req.Method, req.Path, "/openapi/v4/adGroups/targetings/customerData/availables")
	if len(ds) != 1 || ds[0].ID != 1 || ds[0].AdAccountID != 1234 || ds[0].KuidListKey != "123456789abcdefghijklmn" ||
		ds[0].FileType != CustomerDataFileADID || ds[0].Status != CustomerDataStatusComplete || ds[0].PopulationScore != 100 ||
		!ds[0].Ready || ds[0].CreatedDate != "2020-01-01 00:00" || ds[0].LastModifyRequestDate != "2020-01-01 00:00" {
		t.Errorf("customer data = %+v", ds)
	}
	if _, err := c.Targeting.CustomerData(t.Context(), 0); err == nil {
		t.Error("missing ad account ID: want error")
	}
}

func TestTargetingService_CustomerFiles(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/adGroups/targetings/customerFiles/availables", 200, `[
	  {"id": 1, "adAccountId": 1234, "name": "첫번째_고객파일", "adidListKey": "123456789abcdefghijklmn",
	   "status": "COMPLETE", "populationScore": 100, "ready": true,
	   "createdDate": "2020-01-01 00:00", "lastModifyRequestDate": "2020-01-01 00:00"}]`)

	fs, err := c.Targeting.CustomerFiles(t.Context(), 1234)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1234")
	targetingAssertGet(t, req.Method, req.Path, "/openapi/v4/adGroups/targetings/customerFiles/availables")
	if len(fs) != 1 || fs[0].ID != 1 || fs[0].Name != "첫번째_고객파일" || fs[0].AdidListKey != "123456789abcdefghijklmn" ||
		fs[0].Status != CustomerFileStatusComplete || fs[0].PopulationScore != 100 || !fs[0].Ready || fs[0].CreatedDate != "2020-01-01 00:00" {
		t.Errorf("customer files = %+v", fs)
	}
	if _, err := c.Targeting.CustomerFiles(t.Context(), 0); err == nil {
		t.Error("missing ad account ID: want error")
	}
}

func TestTargetingService_FriendGroups(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/adGroups/targetings/talkChannelGroupFiles/availables", 200, `[
	  {"id": 1468, "profileId": "_Xxju", "name": "예제1", "talkChannelGroupFileStatus": "COMPLETE",
	   "fileType": "MESSAGE_RETARGET", "groupKey": "GK1", "createdDate": "2022-07-25T16:12:43", "populationUpdateDate": "2022-07-25T16:50:30"},
	  {"id": 1392, "profileId": "_Xxju", "name": "예제2", "talkChannelGroupFileStatus": "COMPLETE",
	   "fileType": "PHONE_NUMBER", "groupKey": "GK2", "createdDate": "2022-06-10T13:57:49", "populationUpdateDate": "2022-06-10T14:33:30"},
	  {"id": 528, "profileId": "_Xxju", "name": "예제3", "talkChannelGroupFileStatus": "COMPLETE",
	   "fileType": "MESSAGE_RETARGET", "groupKey": "GK3", "createdDate": "2020-08-10T15:28:05"}]`)

	gs, err := c.Targeting.FriendGroups(t.Context(), 1234, "_Xxju")
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1234")
	targetingAssertGet(t, req.Method, req.Path, "/openapi/v4/adGroups/targetings/talkChannelGroupFiles/availables")
	if req.Query.Get("profileId") != "_Xxju" {
		t.Errorf("query = %v", req.Query)
	}
	if len(gs) != 3 || gs[0].ID != 1468 || gs[0].ProfileID != "_Xxju" || gs[0].Status != TargetableFriendGroupComplete ||
		gs[0].FileType != CustomerDataFileMessageRetarget || gs[0].GroupKey != "GK1" || gs[0].CreatedDate.Day() != 25 ||
		gs[0].PopulationUpdateDate.Hour() != 16 || gs[1].FileType != CustomerDataFilePhoneNumber || !gs[2].PopulationUpdateDate.IsZero() {
		t.Errorf("friend groups = %+v", gs)
	}
	if _, err := c.Targeting.FriendGroups(t.Context(), 1234, ""); err == nil {
		t.Error("missing profile ID: want error")
	}
	if _, err := c.Targeting.FriendGroups(t.Context(), 0, "_Xxju"); err == nil {
		t.Error("missing ad account ID: want error")
	}
}

func TestTargetingService_TalkChannels(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/adGroups/targetings/talkChannels/availables", 200, `[
	  {"id": "_xbHxd", "name": "첫번째_친구", "searchId": "첫번째_친구"},
	  {"id": "_ZQxd", "name": "두번째_친구", "searchId": "두번째_친구"}]`)

	cs, err := c.Targeting.TalkChannels(t.Context(), 1234)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1234")
	targetingAssertGet(t, req.Method, req.Path, "/openapi/v4/adGroups/targetings/talkChannels/availables")
	if len(cs) != 2 || cs[0].ID != "_xbHxd" || cs[1].Name != "두번째_친구" || cs[1].SearchID != "두번째_친구" {
		t.Errorf("channels = %+v", cs)
	}
	if _, err := c.Targeting.TalkChannels(t.Context(), 0); err == nil {
		t.Error("missing ad account ID: want error")
	}
}

func TestTargetingService_LoginUserChannels(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/adGroups/targetings/syncProfiles/availables", 200, `[
	  {"id": "_XLd", "name": "프로필있는_첫번째_친구", "searchId": "now_profile_exist"},
	  {"id": "_cWn", "name": "프로필있는_두번째_친구", "searchId": "after_profile_exist"}]`)

	cs, err := c.Targeting.LoginUserChannels(t.Context(), 1234)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1234")
	targetingAssertGet(t, req.Method, req.Path, "/openapi/v4/adGroups/targetings/syncProfiles/availables")
	if len(cs) != 2 || cs[0].ID != "_XLd" || cs[0].SearchID != "now_profile_exist" || cs[1].ID != "_cWn" {
		t.Errorf("channels = %+v", cs)
	}
	if _, err := c.Targeting.LoginUserChannels(t.Context(), 0); err == nil {
		t.Error("missing ad account ID: want error")
	}
}

func TestTargetingService_Categories(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/targetings/data/categories", 200, `{
	  "type": "INTEREST",
	  "data": [{"id": "I01", "name": "여가/일상", "children": [
	    {"id": "I0101", "name": "소셜활동", "children": []},
	    {"id": "I0102", "name": "TV/라디오", "children": [{"id": "I010201", "name": "드라마", "children": []}, {}]},
	    {}]}]}`)

	cs, err := c.Targeting.Categories(t.Context(), CategoryInterest)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "")
	targetingAssertGet(t, req.Method, req.Path, "/openapi/v4/targetings/data/categories")
	if req.Query.Get("type") != "INTEREST" {
		t.Errorf("query = %v", req.Query)
	}
	if cs.Type != CategoryInterest || len(cs.Data) != 1 || cs.Data[0].ID != "I01" || len(cs.Data[0].Children) != 3 ||
		cs.Data[0].Children[1].Children[0].Name != "드라마" || cs.Data[0].Children[1].Children[0].ID != "I010201" {
		t.Errorf("categories = %+v", cs)
	}
	if _, err := c.Targeting.Categories(t.Context(), ""); err == nil {
		t.Error("empty type: want error")
	}
}

func TestTargetingService_MapLocations(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/targetings/location/map", 200, `{
	  "query": "판교역로 235",
	  "items": [{"poi": "N33885142", "labelName": "경기 성남시 분당구 판교역로 235", "address": "경기 성남시 분당구 삼평동 681",
	    "longitude": 127.108638414972, "latitude": 37.402095568027, "geohash": "wydku2",
	    "centerLongitude": 127.1063232421875, "centerLatitude": 37.40020751953125}]}`)

	m, err := c.Targeting.MapLocations(t.Context(), "판교역로 235")
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "")
	targetingAssertGet(t, req.Method, req.Path, "/openapi/v4/targetings/location/map")
	if req.Query.Get("keyword") != "판교역로 235" {
		t.Errorf("query = %v", req.Query)
	}
	if m.Query != "판교역로 235" || len(m.Items) != 1 {
		t.Fatalf("result = %+v", m)
	}
	it := m.Items[0]
	if it.POI != "N33885142" || it.LabelName != "경기 성남시 분당구 판교역로 235" || it.Address != "경기 성남시 분당구 삼평동 681" ||
		it.Geohash != "wydku2" || it.Longitude != 127.108638414972 || it.Latitude != 37.402095568027 ||
		it.CenterLongitude != 127.1063232421875 || it.CenterLatitude != 37.40020751953125 {
		t.Errorf("item = %+v", it)
	}
	if _, err := c.Targeting.MapLocations(t.Context(), ""); err == nil {
		t.Error("empty keyword: want error")
	}
}
