package moment

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestAudienceService_List(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/targetings/audiences", 200, `{"content":[{"id":1,"name":"첫번째_오디언스","audienceStatus":"AVAILABLE","type":"DISPLAY","createdDate":"2021-05-03T18:50:54","lastModifiedDate":"2021-05-03T18:50:54"},{"id":2,"name":"두번째_오디언스","audienceStatus":"AVAILABLE","type":"MESSAGE","createdDate":"2021-04-21T14:36:07","lastModifiedDate":"2021-04-21T14:43:23"},{"id":3,"name":"세번째_오디언스","audienceStatus":"SEED_NOT_ENOUGH","type":"DISPLAY","createdDate":"2021-04-21T14:35:32","lastModifiedDate":"2021-04-21T14:35:42"}]}`)

	as, err := c.Audiences.List(t.Context(), 1111)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodGet || req.Path != "/openapi/v4/targetings/audiences" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	kst := time.FixedZone("KST", 9*60*60)
	if len(as) != 3 || as[2].ID != 3 || as[2].AudienceStatus != AudienceStatusSeedNotEnough || as[1].Type != AudienceMessage ||
		!as[0].CreatedDate.Equal(time.Date(2021, 5, 3, 18, 50, 54, 0, kst)) {
		t.Errorf("audiences = %+v", as)
	}

	if _, err := c.Audiences.List(t.Context(), 0); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestAudienceService_Get(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/targetings/{id}", 200, `{"id":1,"type":"DISPLAY","audienceStatus":"AVAILABLE","name":"첫번째_오디언스","ageType":"ALL","genders":["F","M"],"genderType":"NOT_ALL","locations":[{"value":"E","description":"광주광역시","depth1Name":"광주광역시"},{"value":"M55550301","description":"전북특별자치도 전주시 덕진구 금암1동","depth1Name":"전북특별자치도","depth2Name":"전주시 덕진구","depth3Name":"금암1동","deprecated":true}],"ufoInterests":["I0202","I020601","I03","I320104"],"ufoDeprecatedInterests":["I320104"],"ufoBusinessTypes":["P1021","P2003"],"ufoDeprecatedBusinessTypes":[{"value":"P2003"}],"populationScore":100,"createdDate":"2021-04-14T17:49:35","lastModifiedDate":"2021-04-14T17:49:35"}`)

	a, err := c.Audiences.Get(t.Context(), 1111, 1)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodGet || req.Path != "/openapi/v4/targetings/1" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	if a.ID != 1 || a.Type != AudienceDisplay || a.AgeType != SelectionAll || a.GenderType != SelectionNotAll ||
		!reflect.DeepEqual(a.Genders, []Gender{GenderFemale, GenderMale}) || a.PopulationScore != 100 ||
		len(a.Locations) != 2 || !a.Locations[1].Deprecated || a.Locations[0].Value != "E" ||
		!reflect.DeepEqual(a.UfoInterests, []string{"I0202", "I020601", "I03", "I320104"}) ||
		!reflect.DeepEqual(a.UfoDeprecatedInterests, []string{"I320104"}) ||
		!reflect.DeepEqual(a.UfoBusinessTypes, []string{"P1021", "P2003"}) ||
		!reflect.DeepEqual(a.UfoDeprecatedBusinessTypes, []string{"P2003"}) {
		t.Errorf("audience = %+v", a)
	}

	if _, err := c.Audiences.Get(t.Context(), 1111, 0); err == nil {
		t.Error("missing audience: want error")
	}
	var bad Audience
	if err := json.Unmarshal([]byte(`{"ufoInterests":[1]}`), &bad); err == nil {
		t.Error("numeric key: want error")
	}
}

const audienceCreateResponse = `{"id":1,"type":"DISPLAY","adAccountId":1,"name":"오픈API생성","ageType":"NOT_ALL","ages":["15","20"],"genders":["M"],"genderType":"NOT_ALL","locations":[{"value":"E","description":"광주광역시","depth1Name":"광주광역시"},{"value":"O","description":"충청남도","depth1Name":"충청남도"}],"depth2Locations":[{"value":"B7222","description":"경기도 여주시","depth1Name":"경기도","depth2Name":"여주시"},{"value":"I1009","description":"서울특별시 도봉구","depth1Name":"광주광역시","depth2Name":"도봉구"}],"depth3Locations":[{"value":"A70052424","description":"강원도 삼척시 원덕읍","depth1Name":"강원도","depth2Name":"삼척시","depth3Name":"원덕읍"},{"value":"E13010702","description":"광주광역시 남구 백운2동","depth1Name":"광주광역시","depth2Name":"남구","depth3Name":"백운2동"}]}`

func TestAudienceService_Create(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v4/targetings", 200, audienceCreateResponse)

	a, err := c.Audiences.Create(t.Context(), 1, NewAudience{
		Name: "오픈API생성", Type: AudienceDisplay,
		AgeType: SelectionNotAll, Ages: []string{"20", "15"},
		GenderType: SelectionNotAll, Genders: []Gender{GenderMale},
		LocationType:     LocationTypeArea,
		Locations:        []Location{{Value: "E"}},
		Depth3Locations:  []Depth3Location{{Value: "A70052424"}},
		UfoInterests:     []string{"I0202"},
		CohortTargetings: []CohortTargeting{{ID: 7, Name: "반응", CohortID: 7, InclusionType: InclusionInclude}},
	})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1")
	want := `{"adAccountId":1,"name":"오픈API생성","type":"DISPLAY","ageType":"NOT_ALL","ages":["20","15"],"genderType":"NOT_ALL","genders":["M"],"locationType":"AREA","locations":[{"value":"E"}],"depth3Locations":[{"value":"A70052424"}],"ufoInterests":["I0202"],"cohortTargetings":[{"id":7,"name":"반응","cohortId":7,"inclusionType":"INCLUDE"}]}`
	if req.Method != http.MethodPost || req.Path != "/openapi/v4/targetings" || req.Header.Get("Content-Type") != "application/json" || string(req.Body) != want {
		t.Errorf("request = %s %s %s", req.Method, req.Path, req.Body)
	}
	if a.ID != 1 || a.AdAccountID != 1 || !reflect.DeepEqual(a.Ages, []string{"15", "20"}) || len(a.Depth2Locations) != 2 ||
		a.Depth2Locations[0].Depth2Name != "여주시" || a.Depth3Locations[1].Depth3Name != "백운2동" {
		t.Errorf("audience = %+v", a)
	}

	n := len(srv.Requests())
	valid := NewAudience{Name: "n", Type: AudienceDisplay, AgeType: SelectionAll, GenderType: SelectionAll, LocationType: LocationTypeAll}
	for _, mod := range []func(*NewAudience){
		func(a *NewAudience) { a.Name = "" },
		func(a *NewAudience) { a.Type = "" },
		func(a *NewAudience) { a.LocationType = "" },
		func(a *NewAudience) { a.Ages = []string{"20"} },
		func(a *NewAudience) { a.AgeType = SelectionNotAll },
		func(a *NewAudience) { a.GenderType = SelectionNotAll },
		func(a *NewAudience) { a.Genders = []Gender{GenderMale} },
		func(a *NewAudience) { a.LocationType = LocationTypeArea },
		func(a *NewAudience) { a.Depth2Locations = []Depth2Location{{Value: "B7222"}} },
	} {
		a := valid
		mod(&a)
		if _, err := c.Audiences.Create(t.Context(), 1, a); err == nil {
			t.Errorf("Create(%+v): want error", a)
		}
	}
	if _, err := c.Audiences.Create(t.Context(), 0, valid); err == nil {
		t.Error("missing ad account: want error")
	}
	if len(srv.Requests()) != n {
		t.Error("invalid input was sent")
	}
}

func TestAudienceService_Update(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PUT /openapi/v4/targetings", 200, audienceCreateResponse)

	cur := Audience{
		Name: "오픈API생성", Type: AudienceDisplay, AgeType: SelectionAll, GenderType: SelectionNotAll, Genders: []Gender{GenderMale},
		UfoInterests: []string{"I03"}, UfoDeprecatedInterests: []string{"I32"},
	}
	ch := cur.Changes()
	if ch.LocationType != LocationTypeAll || ch.Name != cur.Name || !reflect.DeepEqual(ch.UfoInterests, []string{"I03"}) {
		t.Errorf("changes = %+v", ch)
	}
	if (&Audience{Locations: []Location{{Value: "A"}}}).Changes().LocationType != LocationTypeArea {
		t.Error("changes with locations: want AREA")
	}
	if (&Audience{LocationType: LocationTypeDomestic}).Changes().LocationType != LocationTypeDomestic {
		t.Error("changes: want location type kept")
	}

	a, err := c.Audiences.Update(t.Context(), 1, 9, ch)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1")
	want := `{"id":9,"adAccountId":1,"name":"오픈API생성","type":"DISPLAY","ageType":"ALL","genderType":"NOT_ALL","genders":["M"],"locationType":"ALL","ufoInterests":["I03"]}`
	if req.Method != http.MethodPut || req.Path != "/openapi/v4/targetings" || string(req.Body) != want {
		t.Errorf("request = %s %s %s", req.Method, req.Path, req.Body)
	}
	if a.ID != 1 || len(a.Locations) != 2 {
		t.Errorf("audience = %+v", a)
	}

	n := len(srv.Requests())
	if _, err := c.Audiences.Update(t.Context(), 1, 0, ch); err == nil {
		t.Error("missing audience: want error")
	}
	if _, err := c.Audiences.Update(t.Context(), 1, 9, AudienceChanges{}); err == nil {
		t.Error("empty changes: want error")
	}
	if len(srv.Requests()) != n {
		t.Error("invalid input was sent")
	}
}

func TestAudienceService_Delete(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("DELETE /openapi/v4/targetings/{id}", 200, ``)

	if err := c.Audiences.Delete(t.Context(), 1111, 5); err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodDelete || req.Path != "/openapi/v4/targetings/5" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	if err := c.Audiences.Delete(t.Context(), 0, 5); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestAudienceService_DeleteMany(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("DELETE /openapi/v4/targetings", 200, `{"successCount":1,"failCount":1,"errorMessages":["타겟을 사용 중인 오디언스가 있습니다."]}`)

	res, err := c.Audiences.DeleteMany(t.Context(), 1111, []int64{1, 2})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodDelete || req.Path != "/openapi/v4/targetings" || req.Query.Get("audienceIds") != "1,2" {
		t.Errorf("request = %s %s?%v", req.Method, req.Path, req.Query)
	}
	want := BulkDeleteResult{SuccessCount: 1, FailCount: 1, ErrorMessages: []string{"타겟을 사용 중인 오디언스가 있습니다."}}
	if !reflect.DeepEqual(*res, want) {
		t.Errorf("result = %+v", res)
	}

	n := len(srv.Requests())
	for _, ids := range [][]int64{nil, {1, 0}} {
		if _, err := c.Audiences.DeleteMany(t.Context(), 1111, ids); err == nil {
			t.Errorf("DeleteMany(%v): want error", ids)
		}
	}
	if len(srv.Requests()) != n {
		t.Error("invalid input was sent")
	}
}

func TestAudienceService_AdGroups(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/targetings/{id}/adGroups", 200, `{"content":[{"campaignId":1,"campaignName":"첫번째_캠페인","campaignTypeGoal":{"campaignType":"TALK_BIZ_BOARD","goal":"VISITING"},"adGroupId":1,"adGroupName":"첫번째_광고그룹","adGroupStatus":["OFF"],"adGroupType":"DISPLAY"},{"campaignId":2,"campaignName":"두번째_캠페인","campaignTypeGoal":{"campaignType":"TALK_BIZ_BOARD","goal":"VISITING"},"adGroupId":3,"adGroupName":"세번째_광고그룹","adGroupStatus":["OFF"],"adGroupType":"DISPLAY"}]}`)

	gs, err := c.Audiences.AdGroups(t.Context(), 1111, 4)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodGet || req.Path != "/openapi/v4/targetings/4/adGroups" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	want := AudienceAdGroup{
		CampaignID: 2, CampaignName: "두번째_캠페인",
		CampaignTypeGoal: CampaignTypeGoal{CampaignType: CampaignTypeTalkBizBoard, Goal: GoalVisiting},
		AdGroupID:        3, AdGroupName: "세번째_광고그룹", AdGroupStatus: []AdGroupStatus{AdGroupStatusOff}, AdGroupType: AdGroupTypeDisplay,
	}
	if len(gs) != 2 || !reflect.DeepEqual(gs[1], want) {
		t.Errorf("ad groups = %+v", gs)
	}
	if _, err := c.Audiences.AdGroups(t.Context(), 1111, 0); err == nil {
		t.Error("missing audience: want error")
	}
}

func TestAudienceService_Unlink(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PUT /openapi/v4/targetings/{id}/unlink", 200, ``)

	if err := c.Audiences.Unlink(t.Context(), 1111, 4, 77); err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodPut || req.Path != "/openapi/v4/targetings/4/unlink" || req.Query.Get("adGroupId") != "77" || string(req.Body) != `{"adGroupId":77}` {
		t.Errorf("request = %s %s?%v %s", req.Method, req.Path, req.Query, req.Body)
	}
	if err := c.Audiences.Unlink(t.Context(), 1111, 4, 0); err == nil {
		t.Error("missing ad group: want error")
	}
}
