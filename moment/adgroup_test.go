package moment

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	kakao "github.com/donghquinn/gokakao"
)

var adGroupTestDay = `["1","1","1","1","1","1","1","1","1","1","1","1","1","1","1","1","1","1","1","1","1","1","1","1"]`

// adGroupTestDisplayJSON is the display ad group example of the docs.
var adGroupTestDisplayJSON = `{
  "id": 54322, "name": "카카오 비즈보드_방문_202205201557", "config": "ON", "dynamicTarget": null,
  "pacing": "NONE", "pricingType": "CPC", "bidAmount": 0, "bidStrategy": "AUTOBID",
  "bidStrategyTarget": {"type": "TARGET_CPC", "value": 250},
  "statusDescription": "운영중", "status": ["LIVE"], "deviceTypes": ["ANDROID", "IOS"],
  "placements": ["DAUM", "NETWORK", "KAKAO_TALK", "KAKAO_SERVICE"],
  "targeting": {
    "type": "NORMAL", "adAccountId": null, "ageType": "ALL", "genderType": "ALL", "locationType": "AREA",
    "locations": [
      {"value": "E", "description": "광주광역시", "depth1Name": "광주광역시"},
      {"value": "O", "description": "충청남도", "depth1Name": "충청남도"}
    ],
    "depth2Locations": [
      {"value": "B7222", "description": "경기도 여주시", "depth1Name": "경기도", "depth2Name": "여주시"},
      {"value": "I1009", "description": "서울특별시 도봉구", "depth1Name": "광주광역시", "depth2Name": "도봉구"}
    ],
    "depth3Locations": [
      {"value": "A70052424", "description": "강원도 삼척시 원덕읍", "depth1Name": "강원도", "depth2Name": "삼척시", "depth3Name": "원덕읍"},
      {"value": "M55550301", "description": "전북특별자치도 전주시 덕진구 금암1동", "depth1Name": "전북특별자치도", "depth2Name": "전주시 덕진구", "depth3Name": "금암1동", "deprecated": true}
    ],
    "ufoInterests": ["I0202", "I020601", "I03", "I320104"],
    "ufoDeprecatedInterests": ["I320104"],
    "ufoBusinessTypes": ["P1021", "P2003"],
    "ufoDeprecatedBusinessTypes": ["P2003"]
  },
  "schedule": {
    "detailTime": false, "beginDate": "2022-05-20", "beginTime": "00:00:00",
    "mondayTime": ` + adGroupTestDay + `, "tuesdayTime": ` + adGroupTestDay + `, "wednesdayTime": ` + adGroupTestDay + `,
    "thursdayTime": ` + adGroupTestDay + `, "fridayTime": ` + adGroupTestDay + `, "saturdayTime": ` + adGroupTestDay + `,
    "sundayTime": ` + adGroupTestDay + `, "lateNight": false
  },
  "campaign": {
    "id": 33626, "name": "픽셀&SDK수정_카카오 비즈보드_방문_202204211136",
    "campaignTypeGoal": {"campaignType": "TALK_BIZ_BOARD", "goal": "VISITING"},
    "objective": null, "dailyBudgetAmount": null, "config": "ON", "statusDescription": "운영중",
    "trackId": null, "adAccountId": 27429, "status": ["LIVE"], "systemConfig": "ON",
    "isDailyBudgetAmountOver": false, "adminStop": false
  },
  "useWifiOnly": false, "creativeCount": 0, "systemConfig": "ON",
  "allAvailableDeviceType": false, "allAvailablePlacement": true, "adult": false,
  "dailyBudgetAmount": 100000, "isDailyBudgetAmountOver": false, "isValidPeriod": true,
  "createdDate": "2022-05-20T15:57:51", "lastModifiedDate": "2022-05-20T15:57:51", "adminStop": false
}`

// adGroupTestMessageJSON is the message ad group create example of the
// docs.
const adGroupTestMessageJSON = `{
  "id": 541620, "name": "메시지_광고그룹", "config": "ON", "dynamicTarget": false,
  "pricingType": "CPMS", "bidAmount": 15, "bidStrategy": "MANUAL",
  "totalBudget": 1500, "totalBudgetWithVAT": 1650, "status": ["NO_AVAILABLE_CREATIVE"],
  "deviceTypes": ["IOS", "ANDROID"], "placements": ["KAKAO_TALK"],
  "targeting": {"type": "NORMAL", "adAccountId": null, "ageType": "ALL", "genderType": "ALL", "locationType": "ALL"},
  "schedule": {
    "detailTime": false, "beginDate": "2023-11-15", "beginTime": "13:00:00",
    "endDate": "2023-12-15", "endTime": "23:59:59.999999999", "lateNight": false
  },
  "messageSendingInfo": {
    "price": 15, "contractCount": 100, "sendRate": 0, "pushAlarm": true,
    "startedAt": null, "finishedAt": null, "status": "SAVE", "syncStatus": "READY",
    "ageVerification": false, "longTerm": false
  },
  "profileId": "_ZQxd",
  "campaign": {
    "id": 34097, "name": "카카오톡 채널_도달_202311031251",
    "campaignTypeGoal": {"campaignType": "TALK_CHANNEL", "goal": "REACH"},
    "objective": {"type": "TALK_CHANNEL", "detailType": "SEND_MESSAGE", "value": "_ZQxd"},
    "dailyBudgetAmount": null, "config": "ON", "statusDescription": "운영중", "trackId": null,
    "adAccountId": 27429, "status": ["LIVE"], "systemConfig": "ON", "isDailyBudgetAmountOver": false
  },
  "useWifiOnly": false, "creativeCount": 0, "systemConfig": "ON", "bidStrategyTarget": null,
  "allAvailableDeviceType": true, "allAvailablePlacement": false, "adult": false,
  "isDailyBudgetAmountOver": false, "isValidPeriod": false,
  "createdDate": "2023-11-14T15:36:40.469166", "lastModifiedDate": "2023-11-14T15:36:40.522281"
}`

// adGroupTestPersonalizedJSON is the personalized message ad group example
// of the docs.
const adGroupTestPersonalizedJSON = `{
  "id": 1, "name": "테스트 개인화 메시지 광고그룹명", "config": "ON", "pricingType": "CPMS",
  "status": ["NO_AVAILABLE_CREATIVE"], "deviceTypes": ["ANDROID", "IOS"], "placements": ["KAKAO_TALK"],
  "profileId": "_Xxju",
  "campaign": {
    "id": 1, "name": "개인화 메시지_도달_202306150913",
    "campaignTypeGoal": {"campaignType": "PERSONAL_MESSAGE", "goal": "REACH"},
    "objective": {"type": "TALK_CHANNEL", "detailType": "SEND_MESSAGE", "value": "_Xxju"},
    "dailyBudgetAmount": null, "config": "ON", "statusDescription": "운영중", "trackId": null,
    "adAccountId": 39543, "status": ["LIVE"], "systemConfig": "ON", "isDailyBudgetAmountOver": false
  },
  "creativeCount": 0, "systemConfig": "ON", "adult": false,
  "createdDate": "2023-06-16T14:14:15.124859", "lastModifiedDate": "2023-06-16T14:14:15.94692"
}`

const adGroupTestNotFound = `{"code": -813, "msg": "KakaoMomentException", "extras": {"detailCode": 32001, "detailMsg": "광고그룹이 존재하지 않습니다."}}`

func adGroupCheckDisplay(t *testing.T, g *AdGroupDetail) {
	t.Helper()
	if g.ID != 54322 || g.Name != "카카오 비즈보드_방문_202205201557" || g.Config != ConfigOn || g.Pacing != PacingNone ||
		g.PricingType != PricingCPC || g.BidAmount != 0 || g.BidStrategy != BidStrategyAutoBid ||
		g.StatusDescription != "운영중" || !slices.Equal(g.Status, []AdGroupStatus{AdGroupStatusLive}) ||
		!slices.Equal(g.DeviceTypes, []DeviceType{DeviceAndroid, DeviceIOS}) ||
		!slices.Equal(g.Placements, []Placement{PlacementDaum, PlacementNetwork, PlacementKakaoTalk, PlacementKakaoService}) ||
		g.UseWifiOnly || g.CreativeCount != 0 || g.SystemConfig != SystemConfigOn || g.AllAvailableDeviceType ||
		!g.AllAvailablePlacement || g.Adult || g.DailyBudgetAmount != 100000 || g.DailyBudgetAmountOver || !g.ValidPeriod {
		t.Errorf("ad group = %+v", g)
	}
	if b := g.BidStrategyTarget; b == nil || b.Type != BidStrategyTargetCPC || b.Value != 250 {
		t.Errorf("bid strategy target = %+v", b)
	}
	want := time.Date(2022, 5, 20, 15, 57, 51, 0, time.FixedZone("KST", 9*3600))
	if !g.CreatedDate.Equal(want) || !g.LastModifiedDate.Equal(want) {
		t.Errorf("dates = %v, %v", g.CreatedDate, g.LastModifiedDate)
	}
	tg := g.Targeting
	if tg == nil || tg.Type != AudienceNormal || tg.AgeType != SelectionAll || tg.GenderType != SelectionAll ||
		tg.LocationType != LocationTypeArea || len(tg.Locations) != 2 || tg.Locations[1].Value != "O" ||
		len(tg.Depth2Locations) != 2 || tg.Depth2Locations[0].Depth2Name != "여주시" ||
		len(tg.Depth3Locations) != 2 || !tg.Depth3Locations[1].Deprecated || tg.Depth3Locations[0].Depth3Name != "원덕읍" ||
		!slices.Equal(tg.UfoInterests, []string{"I0202", "I020601", "I03", "I320104"}) ||
		!slices.Equal(tg.UfoDeprecatedInterests, []string{"I320104"}) ||
		!slices.Equal(tg.UfoBusinessTypes, []string{"P1021", "P2003"}) ||
		!slices.Equal(tg.UfoDeprecatedBusinessTypes, []string{"P2003"}) {
		t.Errorf("targeting = %+v", tg)
	}
	if s := g.Schedule; s == nil || s.BeginDate != "2022-05-20" || s.BeginTime != "00:00:00" || s.DetailTime ||
		s.LateNight || len(s.MondayTime) != 24 || s.SundayTime[23] != "1" {
		t.Errorf("schedule = %+v", s)
	}
	if c := g.Campaign; c == nil || c.ID != 33626 || c.AdAccountID != 27429 ||
		c.CampaignTypeGoal.CampaignType != CampaignTypeTalkBizBoard || c.CampaignTypeGoal.Goal != GoalVisiting {
		t.Errorf("campaign = %+v", c)
	}
}

func adGroupCheckMessage(t *testing.T, g *AdGroupDetail) {
	t.Helper()
	if g.ID != 541620 || g.PricingType != PricingCPMS || g.BidAmount != 15 || g.BidStrategy != BidStrategyManual ||
		g.TotalBudget != 1500 || g.TotalBudgetWithVAT != 1650 ||
		!slices.Equal(g.Status, []AdGroupStatus{AdGroupStatusNoAvailableCreative}) || g.ProfileID != "_ZQxd" ||
		!g.AllAvailableDeviceType || g.BidStrategyTarget != nil || g.ValidPeriod {
		t.Errorf("ad group = %+v", g)
	}
	if s := g.Schedule; s == nil || s.BeginTime != "13:00:00" || s.EndDate != "2023-12-15" || s.EndTime != "23:59:59.999999999" {
		t.Errorf("schedule = %+v", s)
	}
	if m := g.MessageSendingInfo; m == nil || m.Price != 15 || m.ContractCount != 100 || !m.PushAlarm ||
		m.Status != MessageSendSave || m.SyncStatus != MessageSyncReady || m.LongTerm {
		t.Errorf("message sending info = %+v", m)
	}
	if c := g.Campaign; c == nil || c.Objective == nil || c.Objective.Value != "_ZQxd" || c.Objective.Type != ObjectiveTalkChannel {
		t.Errorf("campaign = %+v", c)
	}
	if g.Targeting == nil || g.Targeting.LocationType != LocationTypeAll {
		t.Errorf("targeting = %+v", g.Targeting)
	}
	if g.CreatedDate.Nanosecond() != 469166000 {
		t.Errorf("created = %v", g.CreatedDate)
	}
}

// adGroupBody decodes the JSON body of the last request.
func adGroupBody(t *testing.T, b []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("body %s: %v", b, err)
	}
	return m
}

// adGroupJSON re-encodes v for comparison.
func adGroupJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestAdGroupService_List(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/adGroups", 200, `{"content": [
		{"id": 1111, "name": "광고그룹 1", "config": "ON", "userConfig": "ON", "systemConfig": "ON"},
		{"id": 1112, "name": "광고그룹 2", "config": "OFF", "userConfig": "ON", "systemConfig": "ADMIN_STOP"}]}`)

	gs, err := c.AdGroups.List(t.Context(), 27429, 1234, &AdGroupListOptions{Configs: []Config{ConfigOn}})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "27429")
	if req.Method != http.MethodGet || req.Query.Get("campaignId") != "1234" || !slices.Equal(req.Query["config"], []string{"ON"}) {
		t.Errorf("request = %s %v", req.Method, req.Query)
	}
	if len(gs) != 2 || gs[0].ID != 1111 || gs[0].Name != "광고그룹 1" || gs[0].Config != ConfigOn || gs[0].UserConfig != ConfigOn ||
		gs[1].Config != ConfigOff || gs[1].SystemConfig != SystemConfigAdminStop {
		t.Errorf("ad groups = %+v", gs)
	}

	if _, err := c.AdGroups.List(t.Context(), 27429, 1234, nil); err != nil {
		t.Fatal(err)
	}
	if q := srv.LastRequest().Query; len(q) != 1 {
		t.Errorf("nil options query = %v", q)
	}
	if _, err := c.AdGroups.List(t.Context(), 27429, 0, nil); err == nil {
		t.Error("missing campaign ID: want error")
	}
	if _, err := c.AdGroups.List(t.Context(), 0, 1234, nil); err == nil {
		t.Error("missing ad account ID: want error")
	}
}

func TestAdGroupService_Get(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/adGroups/54322", 200, adGroupTestDisplayJSON)
	srv.HandleJSON("GET /openapi/v4/adGroups/541620", 200, adGroupTestMessageJSON)
	srv.HandleJSON("GET /openapi/v4/adGroups/9", 400, adGroupTestNotFound)

	g, err := c.AdGroups.Get(t.Context(), 27429, 54322)
	if err != nil {
		t.Fatal(err)
	}
	if req := assertRequest(t, srv, "27429"); req.Method != http.MethodGet {
		t.Errorf("method = %s", req.Method)
	}
	adGroupCheckDisplay(t, g)

	m, err := c.AdGroups.Get(t.Context(), 27429, 541620)
	if err != nil {
		t.Fatal(err)
	}
	adGroupCheckMessage(t, m)

	_, err = c.AdGroups.Get(t.Context(), 27429, 9)
	var kerr *kakao.Error
	if !errors.As(err, &kerr) {
		t.Fatalf("err = %v, want *kakao.Error", err)
	}
	if _, err := c.AdGroups.Get(t.Context(), 27429, 0); err == nil {
		t.Error("missing ad group ID: want error")
	}
}

func adGroupTestNewDisplay() NewDisplayAdGroup {
	day := []string{"1", "1", "1", "0", "0", "1", "0", "1", "1", "1", "1", "1", "1", "1", "1", "1", "1", "1", "1", "1", "1", "1", "1", "1"}
	f := false
	return NewDisplayAdGroup{
		CampaignID:             5678,
		Name:                   "카카오톡_비즈보드_등록",
		Placements:             []Placement{PlacementKakaoTalk},
		SectionCategories:      []TalkSectionCode{TalkSectionChat},
		AllAvailableDeviceType: &f,
		AllAvailablePlacement:  &f,
		DeviceTypes:            []DeviceType{DeviceIOS, DeviceAndroid},
		Targeting: AdGroupTargetingSpec{
			AgeType:                SelectionNotAll,
			Ages:                   []string{"15", "20"},
			GenderType:             SelectionNotAll,
			Genders:                []Gender{GenderMale, GenderFemale},
			CustomerDataTargetings: []CustomerDataTarget{{CustomerDataID: 1, InclusionType: InclusionInclude}},
			CohortTargetings:       []CohortTarget{{CohortID: 1, InclusionType: InclusionExclude}},
			CustomerFileTargetings: []CustomerFileTarget{{CustomerFileID: 1, InclusionType: InclusionInclude}},
			TalkChannelTargetings:  []TalkChannelTarget{{ProfileID: "_xbHxd", InclusionType: InclusionExclude}},
			SyncAppTargetings:      []SyncAppTargeting{{ProfileID: "_XLd", InclusionType: InclusionInclude}},
			TrackerTargetings: []TrackerTarget{{
				TrackID: "123456789", InclusionType: InclusionInclude, Term: 180,
				EventCode: "PageView", TrackRuleID: "*", TrackRuleName: "방문",
			}},
			LocationType:    LocationTypeArea,
			Locations:       []Location{{Value: "E", Description: "광주광역시", Depth1Name: "광주광역시"}},
			Depth2Locations: []Depth2Location{{Value: "B7222", Description: "경기도 여주시", Depth1Name: "경기도", Depth2Name: "여주시"}},
		},
		DailyBudgetAmount: 100000,
		BidStrategy:       BidStrategyAutoBid,
		BidStrategyTarget: &BidStrategyTarget{Type: BidStrategyTargetCPC, Value: 250},
		PricingType:       PricingCPC,
		Pacing:            PacingNone,
		Schedule: Schedule{
			BeginDate: "2020-01-01", EndDate: "2020-01-31", LateNight: true,
			MondayTime: day, SundayTime: day,
		},
	}
}

func TestAdGroupService_CreateDisplay(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v4/adGroups", 200, adGroupTestDisplayJSON)

	g, err := c.AdGroups.CreateDisplay(t.Context(), 27429, adGroupTestNewDisplay())
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "27429")
	if req.Method != http.MethodPost || req.Header.Get("Content-Type") != "application/json" {
		t.Errorf("request = %s %v", req.Method, req.Header)
	}
	want := `{"campaign":{"id":5678},"name":"카카오톡_비즈보드_등록","placements":["KAKAO_TALK"],` +
		`"sectionCategories":["KKO99-1"],"allAvailableDeviceType":false,"allAvailablePlacement":false,` +
		`"deviceTypes":["IOS","ANDROID"],"targeting":{"ageType":"NOT_ALL","ages":["15","20"],"genderType":"NOT_ALL",` +
		`"genders":["M","F"],"locationType":"AREA","locations":[{"value":"E","description":"광주광역시","depth1Name":"광주광역시"}],` +
		`"depth2Locations":[{"value":"B7222","description":"경기도 여주시","depth1Name":"경기도","depth2Name":"여주시"}],` +
		`"customerDataTargetings":[{"customerDataId":1,"inclusionType":"INCLUDE"}],` +
		`"customerFileTargetings":[{"customerFileId":1,"inclusionType":"INCLUDE"}],` +
		`"trackerTargetings":[{"trackId":"123456789","inclusionType":"INCLUDE","eventCode":"PageView","trackRuleId":"*","trackRuleName":"방문","term":180}],` +
		`"cohortTargetings":[{"cohortId":1,"inclusionType":"EXCLUDE"}],` +
		`"talkChannelTargetings":[{"profileId":"_xbHxd","inclusionType":"EXCLUDE"}],` +
		`"syncAppTargetings":[{"profileId":"_XLd","inclusionType":"INCLUDE"}]},` +
		`"dailyBudgetAmount":100000,"bidStrategy":"AUTOBID","pricingType":"CPC","bidAmount":0,` +
		`"bidStrategyTarget":{"type":"TARGET_CPC","value":250},"pacing":"NONE",` +
		`"schedule":{"beginDate":"2020-01-01","endDate":"2020-01-31","lateNight":true,"detailTime":false,` +
		`"mondayTime":["1","1","1","0","0","1","0","1","1","1","1","1","1","1","1","1","1","1","1","1","1","1","1","1"],` +
		`"sundayTime":["1","1","1","0","0","1","0","1","1","1","1","1","1","1","1","1","1","1","1","1","1","1","1","1"]},"type":"DISPLAY"}`
	if string(req.Body) != want {
		t.Errorf("body =\n%s\nwant\n%s", req.Body, want)
	}
	adGroupCheckDisplay(t, g)

	bad := []func(*NewDisplayAdGroup){
		func(g *NewDisplayAdGroup) { g.CampaignID = 0 },
		func(g *NewDisplayAdGroup) { g.Placements = nil },
		func(g *NewDisplayAdGroup) { g.DeviceTypes = nil },
		func(g *NewDisplayAdGroup) { g.DailyBudgetAmount = 0 },
		func(g *NewDisplayAdGroup) { g.Pacing = "" },
		func(g *NewDisplayAdGroup) { g.Schedule.BeginDate = "" },
	}
	for i, f := range bad {
		ng := adGroupTestNewDisplay()
		f(&ng)
		if _, err := c.AdGroups.CreateDisplay(t.Context(), 27429, ng); err == nil {
			t.Errorf("case %d: want error", i)
		}
	}
	if _, err := c.AdGroups.CreateDisplay(t.Context(), 0, adGroupTestNewDisplay()); err == nil {
		t.Error("missing ad account ID: want error")
	}
}

func TestAdGroupService_UpdateDisplay(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/adGroups/54322", 200, adGroupTestDisplayJSON)
	srv.HandleJSON("PUT /openapi/v4/adGroups", 200, adGroupTestDisplayJSON)

	cur, err := c.AdGroups.Get(t.Context(), 27429, 54322)
	if err != nil {
		t.Fatal(err)
	}
	ch := cur.DisplayChanges()
	ch.Name = "새 이름"
	g, err := c.AdGroups.UpdateDisplay(t.Context(), 27429, 54322, ch)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "27429")
	if req.Method != http.MethodPut {
		t.Errorf("method = %s", req.Method)
	}
	body := adGroupBody(t, req.Body)
	if body["id"] != 54322.0 || adGroupJSON(t, body["campaign"]) != `{"id":33626}` || body["name"] != "새 이름" ||
		body["type"] != "DISPLAY" || body["allAvailablePlacement"] != true || body["allAvailableDeviceType"] != false ||
		body["adult"] != false || body["dailyBudgetAmount"] != 100000.0 || body["bidStrategy"] != "AUTOBID" ||
		body["pricingType"] != "CPC" || body["bidAmount"] != 0.0 || body["pacing"] != "NONE" ||
		adGroupJSON(t, body["bidStrategyTarget"]) != `{"type":"TARGET_CPC","value":250}` ||
		adGroupJSON(t, body["placements"]) != `["DAUM","NETWORK","KAKAO_TALK","KAKAO_SERVICE"]` {
		t.Errorf("body = %s", req.Body)
	}
	if got := adGroupJSON(t, body["targeting"]); !strings.Contains(got, `"ufoInterests":["I0202","I020601","I03","I320104"]`) ||
		strings.Contains(got, "Deprecated") || !strings.Contains(got, `"locationType":"AREA"`) ||
		!strings.Contains(got, `"deprecated":true`) {
		t.Errorf("targeting = %s", got)
	}
	if got := adGroupJSON(t, body["schedule"]); strings.Contains(got, "beginTime") || !strings.Contains(got, `"beginDate":"2022-05-20"`) {
		t.Errorf("schedule = %s", got)
	}
	adGroupCheckDisplay(t, g)

	if _, err := c.AdGroups.UpdateDisplay(t.Context(), 27429, 54322, DisplayAdGroupChanges{}); err == nil {
		t.Error("empty changes: want error")
	}
	if _, err := c.AdGroups.UpdateDisplay(t.Context(), 27429, 0, ch); err == nil {
		t.Error("missing ad group ID: want error")
	}
}

func TestAdGroupTargeting_Spec(t *testing.T) {
	t.Parallel()
	var nilTargeting *AdGroupTargeting
	if s := nilTargeting.Spec(); s.Type != "" || s.Locations != nil {
		t.Errorf("nil spec = %+v", s)
	}
	tg := AdGroupTargeting{
		ID: 7, Type: AudienceDisplay,
		CustomerDataTargetings:     []CustomerDataTargeting{{ID: 1, InclusionType: InclusionInclude, Name: "d"}},
		CustomerFileTargetings:     []CustomerFileTargeting{{ID: 2, InclusionType: InclusionExclude}},
		TrackerTargetings:          []TrackerTargeting{{ID: 9, TrackID: "t", InclusionType: InclusionInclude, EventCode: "*", TrackRuleID: "*", TrackRuleName: "n", Term: 30}},
		CohortTargetings:           []CohortTargeting{{ID: 3, CohortID: 4, InclusionType: InclusionInclude}, {ID: 5, InclusionType: InclusionExclude}},
		TalkChannelTargetings:      []TalkChannelTargeting{{ID: 6, ProfileID: "123", InclusionType: InclusionInclude}},
		SyncAppTargetings:          []SyncAppTargeting{{ProfileID: "_a", InclusionType: InclusionInclude}},
		TalkChannelGroupTargetings: []TalkChannelGroupTargeting{{TalkChannelGroupFileID: 8, Name: "g", InclusionType: InclusionInclude, FileType: CustomerDataFilePhoneNumber, GroupKey: "k"}},
	}
	got := adGroupJSON(t, tg.Spec())
	want := `{"id":7,"type":"DISPLAY","customerDataTargetings":[{"customerDataId":1,"inclusionType":"INCLUDE"}],` +
		`"customerFileTargetings":[{"customerFileId":2,"inclusionType":"EXCLUDE"}],` +
		`"trackerTargetings":[{"trackId":"t","inclusionType":"INCLUDE","eventCode":"*","trackRuleId":"*","trackRuleName":"n","term":30}],` +
		`"cohortTargetings":[{"cohortId":4,"inclusionType":"INCLUDE"},{"cohortId":5,"inclusionType":"EXCLUDE"}],` +
		`"talkChannelTargetings":[{"profileId":"123","inclusionType":"INCLUDE"}],` +
		`"syncAppTargetings":[{"profileId":"_a","inclusionType":"INCLUDE"}],` +
		`"talkChannelGroupTargetings":[{"talkChannelGroupFileId":8,"name":"g","inclusionType":"INCLUDE","fileType":"PHONE_NUMBER","groupKey":"k"}]}`
	if got != want {
		t.Errorf("spec =\n%s\nwant\n%s", got, want)
	}
}

func TestOptimizationStatuses_UnmarshalJSON(t *testing.T) {
	t.Parallel()
	for in, want := range map[string]OptimizationStatuses{
		`"LEVEL_1"`:             {OptimizationStatusLevel1},
		`["LEVEL_2","INVALID"]`: {OptimizationStatusLevel2, OptimizationStatusInvalid},
		`""`:                    nil,
		`null`:                  nil,
	} {
		var got OptimizationStatuses
		if err := json.Unmarshal([]byte(in), &got); err != nil {
			t.Fatalf("%s: %v", in, err)
		}
		if !slices.Equal(got, want) {
			t.Errorf("%s = %v, want %v", in, got, want)
		}
	}
	var s OptimizationStatuses
	if err := json.Unmarshal([]byte(`1`), &s); err == nil {
		t.Error("number: want error")
	}
	if err := json.Unmarshal([]byte(`"x`), &s); err == nil {
		t.Error("bad string: want error")
	}
}

func TestAdGroupService_UpdateDailyBudget(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PUT /openapi/v4/adGroups/dailyBudgetAmount", 200, ``)

	if err := c.AdGroups.UpdateDailyBudget(t.Context(), 27429, 1234, 5000000); err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "27429")
	if req.Method != http.MethodPut || string(req.Body) != `{"id":1234,"dailyBudgetAmount":5000000}` {
		t.Errorf("request = %s %s", req.Method, req.Body)
	}
	if err := c.AdGroups.UpdateDailyBudget(t.Context(), 27429, 1234, 9990); err == nil {
		t.Error("too small: want error")
	}
	if err := c.AdGroups.UpdateDailyBudget(t.Context(), 27429, 0, 50000); err == nil {
		t.Error("missing ad group ID: want error")
	}
}

func TestAdGroupService_UpdateBidAmount(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PUT /openapi/v4/adGroups/bidAmount", 200, ``)

	if err := c.AdGroups.UpdateBidAmount(t.Context(), 27429, 1234, 5000); err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "27429")
	if req.Method != http.MethodPut || string(req.Body) != `{"id":1234,"bidAmount":5000}` {
		t.Errorf("request = %s %s", req.Method, req.Body)
	}
	if err := c.AdGroups.UpdateBidAmount(t.Context(), 27429, 1234, 0); err == nil {
		t.Error("zero bid: want error")
	}
	if err := c.AdGroups.UpdateBidAmount(t.Context(), 0, 1234, 10); err == nil {
		t.Error("missing ad account ID: want error")
	}
}

func TestAdGroupService_UpdatePacing(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PUT /openapi/v4/adGroups/pacing", 200, ``)

	if err := c.AdGroups.UpdatePacing(t.Context(), 27429, 1234, PacingQuick); err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "27429")
	if req.Method != http.MethodPut || string(req.Body) != `{"id":1234,"pacing":"QUICK"}` {
		t.Errorf("request = %s %s", req.Method, req.Body)
	}
	if err := c.AdGroups.UpdatePacing(t.Context(), 27429, 1234, ""); err == nil {
		t.Error("empty pacing: want error")
	}
	if err := c.AdGroups.UpdatePacing(t.Context(), 27429, 0, PacingQuick); err == nil {
		t.Error("missing ad group ID: want error")
	}
}

func adGroupTestNewMessage() NewMessageAdGroup {
	return NewMessageAdGroup{
		CampaignID:             1,
		Name:                   "메시지_광고그룹",
		AllAvailableDeviceType: true,
		DeviceTypes:            []DeviceType{DeviceIOS, DeviceAndroid},
		MessageSendingInfo:     MessageSendingInfo{ContractCount: 100, Price: 15, PushAlarm: true},
		Targeting:              AdGroupTargetingSpec{AgeType: SelectionAll, GenderType: SelectionAll, LocationType: LocationTypeAll},
		Schedule:               MessageSchedule{BeginDate: "2023-11-15", BeginTime: "13:00:00"},
	}
}

func TestAdGroupService_CreateMessage(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v4/adGroups", 200, adGroupTestMessageJSON)

	g, err := c.AdGroups.CreateMessage(t.Context(), 27429, adGroupTestNewMessage())
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "27429")
	want := `{"campaign":{"id":1},"name":"메시지_광고그룹","placements":["KAKAO_TALK"],"allAvailableDeviceType":true,` +
		`"allAvailablePlacement":false,"deviceTypes":["IOS","ANDROID"],"messageSendingInfo":{"contractCount":100,` +
		`"longTerm":false,"price":15,"pushAlarm":true,"sendRate":0,"status":"SAVE","syncStatus":"READY","ageVerification":false},` +
		`"targeting":{"ageType":"ALL","genderType":"ALL","locationType":"ALL"},"adult":false,"bidStrategy":"MANUAL",` +
		`"pricingType":"CPMS","bidAmount":15,"pacing":"NONE","schedule":{"detailTime":false,"beginDate":"2023-11-15",` +
		`"beginTime":"13:00:00","lateNight":false},"type":"DIRECT_MESSAGE"}`
	if req.Method != http.MethodPost || string(req.Body) != want {
		t.Errorf("body =\n%s\nwant\n%s", req.Body, want)
	}
	adGroupCheckMessage(t, g)

	for i, f := range []func(*NewMessageAdGroup){
		func(g *NewMessageAdGroup) { g.CampaignID = 0 },
		func(g *NewMessageAdGroup) { g.DeviceTypes = nil },
		func(g *NewMessageAdGroup) { g.Schedule.BeginDate = "" },
	} {
		ng := adGroupTestNewMessage()
		f(&ng)
		if _, err := c.AdGroups.CreateMessage(t.Context(), 27429, ng); err == nil {
			t.Errorf("case %d: want error", i)
		}
	}
	if _, err := c.AdGroups.CreateMessage(t.Context(), 0, adGroupTestNewMessage()); err == nil {
		t.Error("missing ad account ID: want error")
	}
}

func TestAdGroupService_UpdateMessage(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/adGroups/541620", 200, adGroupTestMessageJSON)
	srv.HandleJSON("PUT /openapi/v4/adGroups", 200, `{
	  "id": 541620, "name": "메시지발송_광고그룹", "config": "ON", "pricingType": "CPMS", "bidAmount": 15,
	  "bidStrategy": "MANUAL", "totalBudget": 270, "totalBudgetWithVAT": 297, "status": ["FINISHED"],
	  "placements": ["KAKAO_TALK"],
	  "messageSendingInfo": {"price": 15, "contractCount": 18, "sendRate": 0, "pushAlarm": true,
	    "startedAt": "2022-05-18T16:18:14", "finishedAt": "2022-05-18T16:18:26", "status": "FINISHED",
	    "syncStatus": "SUCCESS", "ageVerification": false, "longTerm": false},
	  "profileId": "_ZQxd", "creativeCount": 1, "createdDate": "2022-05-17T11:39:10"
	}`)

	cur, err := c.AdGroups.Get(t.Context(), 27429, 541620)
	if err != nil {
		t.Fatal(err)
	}
	ch := cur.MessageChanges()
	ch.MessageSendingInfo.ContractCount = 33
	g, err := c.AdGroups.UpdateMessage(t.Context(), 27429, 541620, ch)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "27429")
	want := `{"id":541620,"campaign":{"id":34097},"name":"메시지_광고그룹","placements":["KAKAO_TALK"],` +
		`"allAvailableDeviceType":true,"allAvailablePlacement":false,"deviceTypes":["IOS","ANDROID"],` +
		`"messageSendingInfo":{"contractCount":33,"longTerm":false,"price":15,"pushAlarm":true,"sendRate":0,` +
		`"status":"SAVE","syncStatus":"READY","ageVerification":false},` +
		`"targeting":{"type":"NORMAL","ageType":"ALL","genderType":"ALL","locationType":"ALL"},"adult":false,` +
		`"bidStrategy":"MANUAL","pricingType":"CPMS","bidAmount":15,"pacing":"NONE",` +
		`"schedule":{"detailTime":false,"beginDate":"2023-11-15","beginTime":"13:00:00","endDate":"2023-12-15",` +
		`"endTime":"23:59:59.999999999","lateNight":false},"type":"DIRECT_MESSAGE"}`
	if req.Method != http.MethodPut || string(req.Body) != want {
		t.Errorf("body =\n%s\nwant\n%s", req.Body, want)
	}
	if g.TotalBudget != 270 || g.MessageSendingInfo == nil || g.MessageSendingInfo.Status != MessageSendFinished ||
		g.MessageSendingInfo.StartedAt != "2022-05-18T16:18:14" || g.CreativeCount != 1 {
		t.Errorf("ad group = %+v", g)
	}

	ch.DailyBudgetAmount = 100000
	if _, err := c.AdGroups.UpdateMessage(t.Context(), 27429, 541620, ch); err != nil {
		t.Fatal(err)
	}
	if body := adGroupBody(t, srv.LastRequest().Body); body["dailyBudgetAmount"] != 100000.0 {
		t.Errorf("daily budget = %v", body["dailyBudgetAmount"])
	}
	if _, err := c.AdGroups.UpdateMessage(t.Context(), 27429, 541620, MessageAdGroupChanges{}); err == nil {
		t.Error("empty changes: want error")
	}
	if _, err := c.AdGroups.UpdateMessage(t.Context(), 27429, 0, ch); err == nil {
		t.Error("missing ad group ID: want error")
	}
	var empty AdGroupDetail
	if ch := empty.MessageChanges(); ch.CampaignID != 0 || ch.Schedule.BeginDate != "" {
		t.Errorf("empty changes = %+v", ch)
	}
	if ch := empty.DisplayChanges(); ch.CampaignID != 0 || ch.BidStrategyTarget != nil {
		t.Errorf("empty display changes = %+v", ch)
	}
}

func TestAdGroupService_EstimatedAudienceSize(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v4/targetings/populationScore", 200, `{"populationScore": 134}`)

	n, err := c.AdGroups.EstimatedAudienceSize(t.Context(), 27429, AudienceSizeQuery{
		CampaignTypeGoal:       CampaignTypeGoal{CampaignType: CampaignTypeTalkChannel, Goal: GoalReach},
		Objective:              &Objective{Type: ObjectiveTalkChannel, Value: "_ZQxd"},
		AllAvailableDeviceType: true,
		Targeting:              AdGroupTargetingSpec{Type: AudienceNormal, Ages: []string{"30"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "27429")
	want := `{"campaignTypeGoal":{"campaignType":"TALK_CHANNEL","goal":"REACH"},` +
		`"objective":{"type":"TALK_CHANNEL","value":"_ZQxd"},"placements":["KAKAO_TALK"],` +
		`"allAvailableDeviceType":true,"deviceTypes":[],"targeting":{"type":"NORMAL","ages":["30"]},` +
		`"allAvailablePlacement":false}`
	if req.Method != http.MethodPost || string(req.Body) != want {
		t.Errorf("body =\n%s\nwant\n%s", req.Body, want)
	}
	if n != 134 {
		t.Errorf("population = %d", n)
	}
	if _, err := c.AdGroups.EstimatedAudienceSize(t.Context(), 27429, AudienceSizeQuery{}); err == nil {
		t.Error("missing campaign type goal: want error")
	}
	if _, err := c.AdGroups.EstimatedAudienceSize(t.Context(), 0, AudienceSizeQuery{}); err == nil {
		t.Error("missing ad account ID: want error")
	}
}

func TestAdGroupService_CancelContract(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PUT /openapi/v4/adGroups/cancel/541620", 200, ``)

	if err := c.AdGroups.CancelContract(t.Context(), 27429, 541620); err != nil {
		t.Fatal(err)
	}
	if req := assertRequest(t, srv, "27429"); req.Method != http.MethodPut || len(req.Body) != 0 {
		t.Errorf("request = %s %s", req.Method, req.Body)
	}
	if err := c.AdGroups.CancelContract(t.Context(), 27429, 0); err == nil {
		t.Error("missing ad group ID: want error")
	}
}

func adGroupCheckPersonalized(t *testing.T, g *AdGroupDetail) {
	t.Helper()
	if g.ID != 1 || g.Name != "테스트 개인화 메시지 광고그룹명" || g.PricingType != PricingCPMS || g.ProfileID != "_Xxju" ||
		g.Campaign == nil || g.Campaign.CampaignTypeGoal.CampaignType != CampaignTypePersonalMessage ||
		g.Campaign.AdAccountID != 39543 || g.SystemConfig != SystemConfigOn || g.Targeting != nil {
		t.Errorf("ad group = %+v", g)
	}
}

func TestAdGroupService_CreatePersonalizedMessage(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v4/adGroups", 200, adGroupTestPersonalizedJSON)

	g, err := c.AdGroups.CreatePersonalizedMessage(t.Context(), 39543, NewPersonalizedMessageAdGroup{CampaignID: 1, Name: "테스트 개인화 메시지 광고그룹명"})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "39543")
	if req.Method != http.MethodPost || string(req.Body) != `{"campaign":{"id":1},"name":"테스트 개인화 메시지 광고그룹명"}` {
		t.Errorf("request = %s %s", req.Method, req.Body)
	}
	adGroupCheckPersonalized(t, g)

	if _, err := c.AdGroups.CreatePersonalizedMessage(t.Context(), 39543, NewPersonalizedMessageAdGroup{CampaignID: 1}); err != nil {
		t.Fatal(err)
	}
	if body := string(srv.LastRequest().Body); body != `{"campaign":{"id":1}}` {
		t.Errorf("minimal body = %s", body)
	}
	if _, err := c.AdGroups.CreatePersonalizedMessage(t.Context(), 39543, NewPersonalizedMessageAdGroup{}); err == nil {
		t.Error("missing campaign ID: want error")
	}
}

func TestAdGroupService_UpdatePersonalizedMessage(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PUT /openapi/v4/adGroups", 200, adGroupTestPersonalizedJSON)

	g, err := c.AdGroups.UpdatePersonalizedMessage(t.Context(), 39543, 11111, PersonalizedMessageAdGroupChanges{CampaignID: 11111, Name: "개인화 메시지_광고그룹 수정"})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "39543")
	if req.Method != http.MethodPut || string(req.Body) != `{"id":11111,"campaign":{"id":11111},"name":"개인화 메시지_광고그룹 수정"}` {
		t.Errorf("request = %s %s", req.Method, req.Body)
	}
	adGroupCheckPersonalized(t, g)

	if _, err := c.AdGroups.UpdatePersonalizedMessage(t.Context(), 39543, 11111, PersonalizedMessageAdGroupChanges{}); err == nil {
		t.Error("missing campaign ID: want error")
	}
	if _, err := c.AdGroups.UpdatePersonalizedMessage(t.Context(), 39543, 0, PersonalizedMessageAdGroupChanges{CampaignID: 1}); err == nil {
		t.Error("missing ad group ID: want error")
	}
}

func TestAdGroupService_SetConfig(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PUT /openapi/v4/adGroups/onOff", 200, ``)

	if err := c.AdGroups.SetConfig(t.Context(), 27429, 1234, ConfigOn); err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "27429")
	if req.Method != http.MethodPut || string(req.Body) != `{"id":1234,"config":"ON"}` {
		t.Errorf("request = %s %s", req.Method, req.Body)
	}
	if err := c.AdGroups.SetConfig(t.Context(), 27429, 1234, ConfigDel); err == nil {
		t.Error("DEL: want error")
	}
	if err := c.AdGroups.SetConfig(t.Context(), 0, 1234, ConfigOff); err == nil {
		t.Error("missing ad account ID: want error")
	}
}

func TestAdGroupService_Delete(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("DELETE /openapi/v4/adGroups/1234", 200, ``)
	srv.HandleJSON("DELETE /openapi/v4/adGroups/5678", 400, `{"code": -813, "msg": "KakaoMomentException",
		"extras": {"detailCode": 75007, "detailMsg": "카카오톡 채널_도달 그룹은 삭제할 수 없습니다.",
		"path": "/v2/moment/adGroups", "timestamp": "2018-10-01T10:16:14.294+0000"}}`)

	if err := c.AdGroups.Delete(t.Context(), 27429, 1234); err != nil {
		t.Fatal(err)
	}
	if req := assertRequest(t, srv, "27429"); req.Method != http.MethodDelete {
		t.Errorf("method = %s", req.Method)
	}
	var kerr *kakao.Error
	if err := c.AdGroups.Delete(t.Context(), 27429, 5678); !errors.As(err, &kerr) {
		t.Errorf("err = %v, want *kakao.Error", err)
	}
	if err := c.AdGroups.Delete(t.Context(), 27429, 0); err == nil {
		t.Error("missing ad group ID: want error")
	}
}

func TestAdGroupService_LatestSystemStopReason(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/adGroups/1234/latestSystemConfigHistory", 200, `{
	  "id": 1234, "systemConfig": "EXTERNAL_SERVICE_STOP", "reason": "채널 파트너센터에서 규제됨",
	  "detailReason": "메시지 집행 가이드, 운영정책을 위반한 내용이 포함되어 있어 발송 불가",
	  "createdDate": "2021-01-01T00:00:00", "lastModifiedDate": "2021-01-01T00:00:00"}`)

	r, err := c.AdGroups.LatestSystemStopReason(t.Context(), 27429, 1234)
	if err != nil {
		t.Fatal(err)
	}
	if req := assertRequest(t, srv, "27429"); req.Method != http.MethodGet {
		t.Errorf("method = %s", req.Method)
	}
	if r.ID != 1234 || r.SystemConfig != SystemConfigExternalServiceStop || r.Reason != "채널 파트너센터에서 규제됨" ||
		!strings.HasPrefix(r.DetailReason, "메시지 집행 가이드") || r.CreatedDate.Year() != 2021 || r.LastModifiedDate.IsZero() {
		t.Errorf("reason = %+v", r)
	}
	if _, err := c.AdGroups.LatestSystemStopReason(t.Context(), 27429, 0); err == nil {
		t.Error("missing ad group ID: want error")
	}
}

func TestAdGroupService_SystemStopReasons(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/adGroups/1234/systemConfigHistories", 200, `[
	  {"id": 1235, "systemConfig": "ADMIN_STOP", "reason": "해당 광고그룹은 관리자정지 조치 취해졌습니다.",
	   "createdDate": "2021-01-01T00:00:00", "lastModifiedDate": "2021-01-01T00:00:00"},
	  {"id": 1234, "systemConfig": "EXTERNAL_SERVICE_STOP", "reason": "채널 파트너센터에서 규제됨",
	   "detailReason": "메시지 집행 가이드, 운영정책을 위반한 내용이 포함되어 있어 발송 불가",
	   "createdDate": "2021-01-01T00:00:00", "lastModifiedDate": "2021-01-01T00:00:00"}]`)

	rs, err := c.AdGroups.SystemStopReasons(t.Context(), 27429, 1234)
	if err != nil {
		t.Fatal(err)
	}
	if req := assertRequest(t, srv, "27429"); req.Method != http.MethodGet {
		t.Errorf("method = %s", req.Method)
	}
	if len(rs) != 2 || rs[0].ID != 1235 || rs[0].SystemConfig != SystemConfigAdminStop || rs[0].DetailReason != "" ||
		rs[1].SystemConfig != SystemConfigExternalServiceStop || rs[1].DetailReason == "" {
		t.Errorf("reasons = %+v", rs)
	}
	if _, err := c.AdGroups.SystemStopReasons(t.Context(), 0, 1234); err == nil {
		t.Error("missing ad account ID: want error")
	}
}
