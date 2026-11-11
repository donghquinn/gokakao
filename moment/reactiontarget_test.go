package moment

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"
)

const reactionTargetJSON = `{"adAccountId":1234,"id":1,"audienceType":"MESSAGE","collectDuration":90,"cohortStatus":"WAITING","name":"%s","baseAds":[{"campaign":{"id":56,"name":"첫번째_캠페인","campaignTypeGoal":{"campaignType":"TALK_CHANNEL","goal":"REACH"}},"adGroup":{"id":78,"name":"첫번째_광고그룹"},"operation":"ONLY","firstIndicator":"OPEN"}],"createdDate":"2020-01-01 00:00:00","lastModifiedDate":"%s"}`

// reactionTargetCheck checks the decoded create, rename and update examples.
func reactionTargetCheck(t *testing.T, r *ReactionTarget, name string, modified time.Time) {
	t.Helper()
	kst := time.FixedZone("KST", 9*60*60)
	if r.ID != 1 || r.AdAccountID != 1234 || r.AudienceType != AudienceMessage || r.CollectDuration != 90 ||
		r.CohortStatus != ReactionTargetWaiting || r.Name != name ||
		!r.CreatedDate.Equal(time.Date(2020, 1, 1, 0, 0, 0, 0, kst)) || !r.LastModifiedDate.Equal(modified) || len(r.BaseAds) != 1 {
		t.Fatalf("target = %+v", r)
	}
	a := r.BaseAds[0]
	if a.Campaign == nil || a.Campaign.ID != 56 || a.Campaign.CampaignTypeGoal.CampaignType != CampaignTypeTalkChannel ||
		a.AdGroup == nil || a.AdGroup.ID != 78 || a.AdGroup.Name != "첫번째_광고그룹" ||
		a.Operation != OperationOnly || a.FirstIndicator != IndicatorOpen {
		t.Errorf("base ad = %+v", a)
	}
}

// reactionTargetBody fills the name and last modified date of the example.
func reactionTargetBody(name, modified string) string {
	return fmt.Sprintf(reactionTargetJSON, name, modified)
}

func TestReactionTargetService_List(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/targetings/cohort/list", 200, `[{"adAccountId":1234,"id":1,"audienceType":"MESSAGE","name":"첫번째_광고반응타겟","baseAds":[{"campaign":{"id":5678,"name":"첫번째_캠페인","campaignTypeGoal":{"campaignType":"TALK_CHANNEL","goal":"REACH"}},"adGroup":{"id":20425,"name":"첫번째_광고그룹"},"operation":"ONLY","firstIndicator":"OPEN"}],"collectDuration":90,"cohortStatus":"AVAILABLE","popultaionScore":321,"createdDate":"2020-01-01 00:00:00","lastModifiedDate":"2020-01-01 00:00:00"}]`)

	rs, err := c.ReactionTargets.List(t.Context(), 1111)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodGet || req.Path != "/openapi/v4/targetings/cohort/list" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	if len(rs) != 1 || rs[0].ID != 1 || rs[0].Name != "첫번째_광고반응타겟" || rs[0].CohortStatus != ReactionTargetAvailable ||
		rs[0].PopulationScore != 321 || len(rs[0].BaseAds) != 1 || rs[0].BaseAds[0].AdGroup.ID != 20425 || rs[0].CreatedDate.IsZero() {
		t.Errorf("targets = %+v", rs)
	}
	if _, err := c.ReactionTargets.List(t.Context(), 0); err == nil {
		t.Error("missing ad account: want error")
	}
	var bad ReactionTarget
	if err := json.Unmarshal([]byte(`{"id":"x"}`), &bad); err == nil {
		t.Error("bad ID: want error")
	}
}

func TestReactionTargetService_Get(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/targetings/cohort/{id}", 200, `{"adAccountId":65721,"id":100943,"collectDuration":90,"cohortStatus":"WAITING","name":"first_engagement_target","audienceType":"MESSAGE","baseAds":[{"campaign":{"id":82141,"name":"first_campaign","campaignTypeGoal":{"campaignType":"TALK_CHANNEL","goal":"REACH"}},"adGroup":{"id":101326,"name":"first_ad_group"},"operation":"ONLY","firstIndicator":"OPEN"},{"messageAd":{"id":"msg-ad-1336211109228568577"},"operation":"ONLY","firstIndicator":"OPEN"}],"populationScore":0,"createdDate":"2025-04-15T13:27:57.603841","lastModifiedDate":"2025-04-15T13:27:57.60384","populationUpdateDate":null}`)

	r, err := c.ReactionTargets.Get(t.Context(), 1111, 100943)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodGet || req.Path != "/openapi/v4/targetings/cohort/100943" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	kst := time.FixedZone("KST", 9*60*60)
	if r.ID != 100943 || r.AdAccountID != 65721 || len(r.BaseAds) != 2 || r.BaseAds[1].MessageAd == nil ||
		r.BaseAds[1].MessageAd.ID != "msg-ad-1336211109228568577" || r.BaseAds[1].Campaign != nil ||
		!r.PopulationUpdateDate.IsZero() || !r.CreatedDate.Equal(time.Date(2025, 4, 15, 13, 27, 57, 603841000, kst)) {
		t.Errorf("target = %+v", r)
	}

	ch := r.Changes()
	want := []NewReactionBaseAd{
		{CampaignID: 82141, CampaignType: CampaignTypeTalkChannel, AdGroupID: 101326, Operation: OperationOnly, FirstIndicator: IndicatorOpen},
		{MessageAdID: "msg-ad-1336211109228568577", Operation: OperationOnly, FirstIndicator: IndicatorOpen},
	}
	if !reflect.DeepEqual(ch.BaseAds, want) {
		t.Errorf("changes = %+v", ch)
	}

	if _, err := c.ReactionTargets.Get(t.Context(), 1111, 0); err == nil {
		t.Error("missing target: want error")
	}
}

func TestReactionTargetService_Creatables(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/targetings/cohort/creatables", 200, `[{"campaign":{"id":1234,"name":"첫번째_캠페인","campaignTypeGoal":{"campaignType":"DISPLAY","goal":"VISITING"}},"adGroup":[{"id":56,"name":"첫번째_광고그룹"},{"id":78,"name":"두번째_광고그룹"}]},{"campaign":{"id":9},"adGroup":{"id":10,"name":"단일"}},{"campaign":{"id":11},"adGroup":null}]`)

	cs, err := c.ReactionTargets.Creatables(t.Context(), 1111, ReactionTargetCreatablesQuery{CampaignType: CampaignTypeDisplay, Goal: GoalVisiting, SearchKeyword: "HG"})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodGet || req.Path != "/openapi/v4/targetings/cohort/creatables" ||
		req.Query.Get("campaignType") != "DISPLAY" || req.Query.Get("goal") != "VISITING" || req.Query.Get("searchKeyword") != "HG" {
		t.Errorf("request = %s %s?%v", req.Method, req.Path, req.Query)
	}
	if len(cs) != 3 || cs[0].Campaign.ID != 1234 || cs[0].Campaign.CampaignTypeGoal.Goal != GoalVisiting ||
		!reflect.DeepEqual(cs[0].AdGroups, []AdGroup{{ID: 56, Name: "첫번째_광고그룹"}, {ID: 78, Name: "두번째_광고그룹"}}) ||
		!reflect.DeepEqual(cs[1].AdGroups, []AdGroup{{ID: 10, Name: "단일"}}) || cs[2].AdGroups != nil {
		t.Errorf("creatables = %+v", cs)
	}

	n := len(srv.Requests())
	for _, q := range []ReactionTargetCreatablesQuery{
		{Goal: GoalVisiting, SearchKeyword: "HG"},
		{CampaignType: CampaignTypeDisplay, SearchKeyword: "HG"},
		{CampaignType: CampaignTypeDisplay, Goal: GoalVisiting},
	} {
		if _, err := c.ReactionTargets.Creatables(t.Context(), 1111, q); err == nil {
			t.Errorf("Creatables(%+v): want error", q)
		}
	}
	if _, err := c.ReactionTargets.Creatables(t.Context(), 0, ReactionTargetCreatablesQuery{CampaignType: CampaignTypeDisplay, Goal: GoalVisiting, SearchKeyword: "HG"}); err == nil {
		t.Error("missing ad account: want error")
	}
	if len(srv.Requests()) != n {
		t.Error("invalid input was sent")
	}
	for _, in := range []string{`{"campaign":1}`, `{"adGroup":{"id":"x"}}`, `{"adGroup":[1]}`} {
		var cc CreatableCampaign
		if err := json.Unmarshal([]byte(in), &cc); err == nil {
			t.Errorf("Unmarshal(%s): want error", in)
		}
	}
}

func TestReactionTargetService_Create(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v4/targetings/cohort", 200, reactionTargetBody("첫번째_광고반응", "2020-01-01 00:00:00"))

	r, err := c.ReactionTargets.Create(t.Context(), 1111, NewReactionTarget{
		Name: "첫번째_광고반응", AudienceType: AudienceMessage,
		BaseAds: []NewReactionBaseAd{{CampaignID: 56, CampaignType: CampaignTypeTalkChannel, AdGroupID: 78, Operation: OperationOnly, FirstIndicator: IndicatorOpen}},
	})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	want := `{"name":"첫번째_광고반응","audienceType":"MESSAGE","baseAds":[{"campaign":{"id":56,"campaignTypeGoal":{"campaignType":"TALK_CHANNEL"}},"adGroup":{"id":78},"operation":"ONLY","firstIndicator":"OPEN"}]}`
	if req.Method != http.MethodPost || req.Path != "/openapi/v4/targetings/cohort" || string(req.Body) != want {
		t.Errorf("request = %s %s %s", req.Method, req.Path, req.Body)
	}
	kst := time.FixedZone("KST", 9*60*60)
	reactionTargetCheck(t, r, "첫번째_광고반응", time.Date(2020, 1, 1, 0, 0, 0, 0, kst))

	if _, err := c.ReactionTargets.Create(t.Context(), 1111, NewReactionTarget{
		Name: "n",
		BaseAds: []NewReactionBaseAd{
			{CampaignID: 1, AdGroupID: 2, Operation: OperationMinus, FirstIndicator: IndicatorOpen, SecondIndicator: IndicatorClick, ThirdIndicator: IndicatorPlay},
			{MessageAdID: "msg-ad-1", Operation: OperationOnly, FirstIndicator: IndicatorClick},
		},
	}); err != nil {
		t.Fatal(err)
	}
	want = `{"name":"n","baseAds":[{"campaign":{"id":1},"adGroup":{"id":2},"operation":"MINUS","firstIndicator":"OPEN","secondIndicator":"CLICK","thirdIndicator":"PLAY"},{"messageAd":{"id":"msg-ad-1"},"operation":"ONLY","firstIndicator":"CLICK"}]}`
	if got := string(srv.LastRequest().Body); got != want {
		t.Errorf("body = %s", got)
	}

	n := len(srv.Requests())
	ok := NewReactionBaseAd{CampaignID: 1, AdGroupID: 2, Operation: OperationOnly, FirstIndicator: IndicatorClick}
	for _, nr := range []NewReactionTarget{
		{BaseAds: []NewReactionBaseAd{ok}},
		{Name: "n"},
		{Name: "n", BaseAds: []NewReactionBaseAd{{CampaignID: 1, Operation: OperationOnly, FirstIndicator: IndicatorClick}}},
		{Name: "n", BaseAds: []NewReactionBaseAd{{CampaignID: 1, AdGroupID: 2, FirstIndicator: IndicatorClick}}},
		{Name: "n", BaseAds: []NewReactionBaseAd{{CampaignID: 1, AdGroupID: 2, Operation: OperationOnly}}},
	} {
		if _, err := c.ReactionTargets.Create(t.Context(), 1111, nr); err == nil {
			t.Errorf("Create(%+v): want error", nr)
		}
	}
	if _, err := c.ReactionTargets.Create(t.Context(), 0, NewReactionTarget{Name: "n", BaseAds: []NewReactionBaseAd{ok}}); err == nil {
		t.Error("missing ad account: want error")
	}
	if len(srv.Requests()) != n {
		t.Error("invalid input was sent")
	}
}

func TestReactionTargetService_Rename(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PUT /openapi/v4/targetings/cohort/name", 200, reactionTargetBody("광고반응타겟_이름_수정", "2020-01-01 12:00:00"))

	r, err := c.ReactionTargets.Rename(t.Context(), 1111, 1, "광고반응타겟_이름_수정")
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodPut || req.Path != "/openapi/v4/targetings/cohort/name" || string(req.Body) != `{"id":1,"name":"광고반응타겟_이름_수정"}` {
		t.Errorf("request = %s %s %s", req.Method, req.Path, req.Body)
	}
	kst := time.FixedZone("KST", 9*60*60)
	reactionTargetCheck(t, r, "광고반응타겟_이름_수정", time.Date(2020, 1, 1, 12, 0, 0, 0, kst))
	if _, err := c.ReactionTargets.Rename(t.Context(), 1111, 0, "n"); err == nil {
		t.Error("missing target: want error")
	}
}

func TestReactionTargetService_Update(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PUT /openapi/v4/targetings/cohort", 200, reactionTargetBody("광고반응타겟_이름_수정", "2020-01-01 15:00:00"))

	r, err := c.ReactionTargets.Update(t.Context(), 1111, 1234, ReactionTargetChanges{BaseAds: []NewReactionBaseAd{
		{CampaignID: 56, CampaignType: CampaignTypeTalkChannel, AdGroupID: 78, Operation: OperationOnly, FirstIndicator: IndicatorOpen},
	}})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	want := `{"id":1234,"baseAds":[{"campaign":{"id":56,"campaignTypeGoal":{"campaignType":"TALK_CHANNEL"}},"adGroup":{"id":78},"operation":"ONLY","firstIndicator":"OPEN"}]}`
	if req.Method != http.MethodPut || req.Path != "/openapi/v4/targetings/cohort" || string(req.Body) != want {
		t.Errorf("request = %s %s %s", req.Method, req.Path, req.Body)
	}
	kst := time.FixedZone("KST", 9*60*60)
	reactionTargetCheck(t, r, "광고반응타겟_이름_수정", time.Date(2020, 1, 1, 15, 0, 0, 0, kst))

	n := len(srv.Requests())
	if _, err := c.ReactionTargets.Update(t.Context(), 1111, 1234, ReactionTargetChanges{}); err == nil {
		t.Error("missing base ads: want error")
	}
	if _, err := c.ReactionTargets.Update(t.Context(), 1111, 0, r.Changes()); err == nil {
		t.Error("missing target: want error")
	}
	if len(srv.Requests()) != n {
		t.Error("invalid input was sent")
	}
}

func TestReactionTargetService_Delete(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("DELETE /openapi/v4/targetings/cohort/{id}", 200, ``)

	if err := c.ReactionTargets.Delete(t.Context(), 1111, 7); err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodDelete || req.Path != "/openapi/v4/targetings/cohort/7" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	if err := c.ReactionTargets.Delete(t.Context(), 1111, 0); err == nil {
		t.Error("missing target: want error")
	}
}

func TestReactionTargetService_DeleteMany(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("DELETE /openapi/v4/targetings/cohort", 200, `{"successCount":1,"failCount":1,"errorMessages":["타겟을 사용 중인 오디언스가 있습니다."]}`)

	res, err := c.ReactionTargets.DeleteMany(t.Context(), 1111, []int64{7, 8})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodDelete || req.Path != "/openapi/v4/targetings/cohort" || req.Query.Get("cohortIds") != "7,8" {
		t.Errorf("request = %s %s?%v", req.Method, req.Path, req.Query)
	}
	if res.SuccessCount != 1 || res.FailCount != 1 || len(res.ErrorMessages) != 1 {
		t.Errorf("result = %+v", res)
	}
	if _, err := c.ReactionTargets.DeleteMany(t.Context(), 1111, nil); err == nil {
		t.Error("missing IDs: want error")
	}
}

func TestReactionTargetService_Usages(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/targetings/cohort/usages/{id}", 200, customerDataUsagesJSON)

	us, err := c.ReactionTargets.Usages(t.Context(), 1111, 7)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodGet || req.Path != "/openapi/v4/targetings/cohort/usages/7" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	customerDataCheckUsages(t, us)
	if _, err := c.ReactionTargets.Usages(t.Context(), 1111, 0); err == nil {
		t.Error("missing target: want error")
	}
}
