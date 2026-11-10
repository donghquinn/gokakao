package moment

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestTypes_decode(t *testing.T) {
	t.Parallel()

	var c Campaign
	err := json.Unmarshal([]byte(`{"id":"1234","adAccountId":5678,"name":"캠페인","campaignTypeGoal":{"campaignType":"DISPLAY","goal":"CONVERSION"},"objective":{"type":"PIXEL_AND_SDK","value":"999","detailType":"PURCHASE"},"dailyBudgetAmount":"50000","config":"ON","isDailyBudgetAmountOver":true,"statusDescription":"일 예산 초과","trackId":"t1"}`), &c)
	if err != nil {
		t.Fatal(err)
	}
	want := Campaign{
		ID: 1234, AdAccountID: 5678, Name: "캠페인",
		CampaignTypeGoal:  CampaignTypeGoal{CampaignType: CampaignTypeDisplay, Goal: GoalConversion},
		Objective:         &Objective{Type: ObjectivePixelAndSDK, Value: "999", DetailType: ObjectiveDetailPurchase},
		DailyBudgetAmount: 50000, Config: ConfigOn, DailyBudgetAmountOver: true,
		StatusDescription: "일 예산 초과", TrackID: "t1",
	}
	if !reflect.DeepEqual(c, want) {
		t.Errorf("campaign = %+v", c)
	}

	var m Metrics
	if err := json.Unmarshal([]byte(`{"imp":4,"click":0,"ctr":"0.5","cost":null}`), &m); err != nil {
		t.Fatal(err)
	}
	if m["imp"] != 4 || m["ctr"] != 0.5 || m["cost"] != 0 {
		t.Errorf("metrics = %v", m)
	}
}

func TestTypes_roundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   any
		out  any
		json string
	}{
		{
			name: "optimization targeting",
			in: OptimizationTargeting{
				Type:               MessageTargetingOptimizationDirect,
				OptimizationTarget: OptimizationTarget{Type: OptimizationRatio, SendRatios: []int{60, 40}},
				Devices:            []DeviceType{DeviceAndroid},
			},
			out:  &OptimizationTargeting{},
			json: `{"type":"OPTIMIZATION_DIRECT","optimizationTarget":{"type":"RATIO","sendRatios":[60,40]},"devices":["ANDROID"]}`,
		},
		{
			name: "schedule",
			in:   Schedule{BeginDate: "2026-01-01", LateNight: true},
			out:  &Schedule{},
			json: `{"beginDate":"2026-01-01","lateNight":true,"detailTime":false}`,
		},
		{
			name: "landing info",
			in:   LandingInfo{LandingType: LandingTypeAdView, AdViewItem: &AdViewItem{ID: 42}},
			out:  &LandingInfo{},
			json: `{"landingType":"AD_VIEW","adViewItem":{"id":42}}`,
		},
		{
			name: "demographics",
			in: Demographics{
				GenderType: SelectionNotAll, Genders: []Gender{GenderFemale},
				LocationType: LocationTypeArea, Depth1Locations: []Location{{Value: "I"}},
			},
			out:  &Demographics{},
			json: `{"genderType":"NOT_ALL","genders":["F"],"locationType":"AREA","depth1Locations":[{"value":"I"}]}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			b, err := json.Marshal(tt.in)
			if err != nil {
				t.Fatal(err)
			}
			if string(b) != tt.json {
				t.Errorf("marshal = %s, want %s", b, tt.json)
			}
			if err := json.Unmarshal(b, tt.out); err != nil {
				t.Fatal(err)
			}
			if got := reflect.ValueOf(tt.out).Elem().Interface(); !reflect.DeepEqual(got, tt.in) {
				t.Errorf("round trip = %+v, want %+v", got, tt.in)
			}
		})
	}
}
