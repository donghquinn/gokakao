package moment

import (
	"bytes"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/donghquinn/gokakao/internal/kakaotest"
)

// customerDataFile is an uploaded file part.
type customerDataFile struct {
	Name, ContentType, Data string
}

// customerDataParseMultipart parses a multipart/form-data request body.
func customerDataParseMultipart(t *testing.T, req kakaotest.Request) (map[string][]string, map[string][]customerDataFile) {
	t.Helper()
	mt, params, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil || mt != "multipart/form-data" {
		t.Fatalf("Content-Type = %q", req.Header.Get("Content-Type"))
	}
	form, err := multipart.NewReader(bytes.NewReader(req.Body), params["boundary"]).ReadForm(1 << 20)
	if err != nil {
		t.Fatal(err)
	}
	files := map[string][]customerDataFile{}
	for field, fhs := range form.File {
		for _, fh := range fhs {
			f, err := fh.Open()
			if err != nil {
				t.Fatal(err)
			}
			data, _ := io.ReadAll(f)
			_ = f.Close()
			files[field] = append(files[field], customerDataFile{fh.Filename, fh.Header.Get("Content-Type"), string(data)})
		}
	}
	return form.Value, files
}

const customerDataJSON = `{"id":2656,"adAccountId":27429,"name":"고객데이터 등록","kuidListKey":"c8010decafde484eb33284b53a290c30","status":"COMPLETE","ready":true,"fileType":"ADID","type":"URL","sourceUrl":"https://sample-url.com/*****valid1.csv","renewable":true,"renewalExpireDateTime":"2026-09-01T00:03:58","populationScore":7170,"populationUpdateDate":"2026-06-03T00:13:59","createdDate":"2026-06-03T00:03:59","originalCreatedDate":"2026-06-01T00:03:59","lastModifiedDate":"2026-06-03T00:13:59"}`

func customerDataCheck(t *testing.T, d *CustomerData, name string) {
	t.Helper()
	kst := time.FixedZone("KST", 9*60*60)
	if d.ID != 2656 || d.AdAccountID != 27429 || d.Name != name || d.KuidListKey != "c8010decafde484eb33284b53a290c30" ||
		d.Status != CustomerDataStatusComplete || !d.Ready || d.FileType != CustomerDataFileADID || d.Type != CustomerDataURL ||
		d.SourceURL != "https://sample-url.com/*****valid1.csv" || !d.Renewable || d.PopulationScore != 7170 ||
		!d.RenewalExpireDateTime.Equal(time.Date(2026, 9, 1, 0, 3, 58, 0, kst)) ||
		!d.PopulationUpdateDate.Equal(time.Date(2026, 6, 3, 0, 13, 59, 0, kst)) ||
		!d.CreatedDate.Equal(time.Date(2026, 6, 3, 0, 3, 59, 0, kst)) ||
		!d.OriginalCreatedDate.Equal(time.Date(2026, 6, 1, 0, 3, 59, 0, kst)) ||
		!d.LastModifiedDate.Equal(time.Date(2026, 6, 3, 0, 13, 59, 0, kst)) {
		t.Errorf("customer data = %+v", d)
	}
}

func TestCustomerDataService_List(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/customerData", 200, `[`+customerDataJSON+`]`)

	ds, err := c.CustomerData.List(t.Context(), 1111)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodGet || req.Path != "/openapi/v4/customerData" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	if len(ds) != 1 {
		t.Fatalf("customer data = %+v", ds)
	}
	customerDataCheck(t, &ds[0], "고객데이터 등록")
	if _, err := c.CustomerData.List(t.Context(), 0); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestCustomerDataService_Get(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/customerData/{id}", 200, customerDataJSON)

	d, err := c.CustomerData.Get(t.Context(), 1111, 2656)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodGet || req.Path != "/openapi/v4/customerData/2656" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	customerDataCheck(t, d, "고객데이터 등록")
	if _, err := c.CustomerData.Get(t.Context(), 1111, 0); err == nil {
		t.Error("missing customer data: want error")
	}
}

func TestCustomerDataService_Create(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v4/customerData", 200, `{"id":2656,"adAccountId":27429,"name":"고객데이터 등록","fileType":"ADID","status":"WAITING","successCount":980,"failedCount":15,"duplicationCount":5,"createdDate":"2026-06-03T00:03:59"}`)

	d, err := c.CustomerData.Create(t.Context(), 1111, NewCustomerData{
		Name: "첫번째_고객데이터", FileType: CustomerDataFileADID,
		Files: []UploadFile{
			{Filename: "sample1.csv", ContentType: "text/plain", Data: []byte("a\n")},
			{Data: []byte("b\n")},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodPost || req.Path != "/openapi/v4/customerData" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	fields, files := customerDataParseMultipart(t, req)
	if !reflect.DeepEqual(fields, map[string][]string{"name": {"첫번째_고객데이터"}, "fileType": {"ADID"}}) {
		t.Errorf("fields = %v", fields)
	}
	wantFiles := map[string][]customerDataFile{"files": {{"sample1.csv", "text/plain", "a\n"}, {"customer.csv", "text/csv", "b\n"}}}
	if !reflect.DeepEqual(files, wantFiles) {
		t.Errorf("files = %v", files)
	}
	if d.ID != 2656 || d.Status != CustomerDataStatusWaiting || d.SuccessCount != 980 || d.FailedCount != 15 || d.DuplicationCount != 5 || d.CreatedDate.IsZero() {
		t.Errorf("customer data = %+v", d)
	}

	if _, err := c.CustomerData.Create(t.Context(), 1111, NewCustomerData{Name: "m", FileType: CustomerDataFileMessageRetarget, AdGroupIDs: []int64{7, 8}}); err != nil {
		t.Fatal(err)
	}
	fields, files = customerDataParseMultipart(t, srv.LastRequest())
	if !reflect.DeepEqual(fields["adGroupIds"], []string{"7", "8"}) || fields["fileType"][0] != "MESSAGE_RETARGET" || len(files) != 0 {
		t.Errorf("message retarget = %v %v", fields, files)
	}

	n := len(srv.Requests())
	csv := []UploadFile{{Data: []byte("x")}}
	for _, d := range []NewCustomerData{
		{FileType: CustomerDataFileADID, Files: csv},
		{Name: "n", Files: csv},
		{Name: "n", FileType: CustomerDataFileADID},
		{Name: "n", FileType: CustomerDataFileADID, Files: make([]UploadFile, 11)},
		{Name: "n", FileType: CustomerDataFileADID, Files: []UploadFile{{Filename: "empty.csv"}}},
		{Name: "n", FileType: CustomerDataFileADID, Files: csv, AdGroupIDs: []int64{1}},
		{Name: "n", FileType: CustomerDataFileMessageRetarget},
		{Name: "n", FileType: CustomerDataFileMessageRetarget, AdGroupIDs: []int64{1}, Files: csv},
	} {
		if _, err := c.CustomerData.Create(t.Context(), 1111, d); err == nil {
			t.Errorf("Create(%+v): want error", d)
		}
	}
	if _, err := c.CustomerData.Create(t.Context(), 0, NewCustomerData{Name: "n", FileType: CustomerDataFileADID, Files: csv}); err == nil {
		t.Error("missing ad account: want error")
	}
	if len(srv.Requests()) != n {
		t.Error("invalid input was sent")
	}
}

func TestCustomerDataService_CreateFromURL(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v4/customerData/url", 200, `{"id":2657,"adAccountId":27429,"name":"고객데이터 URL 등록","fileType":"ADID","status":"WAITING","createdDate":"2026-06-03T00:03:59","sourceUrl":"https://sample-url.com/*****ience.csv","renewable":true,"renewalExpireDateTime":"2026-09-01T00:03:58"}`)

	d, err := c.CustomerData.CreateFromURL(t.Context(), 1111, NewCustomerDataURL{Name: "고객데이터 URL 등록", FileType: CustomerDataFileADID, SourceURL: "https://sample-url.com/audience.csv", Renewable: true})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodPost || req.Path != "/openapi/v4/customerData/url" ||
		string(req.Body) != `{"name":"고객데이터 URL 등록","fileType":"ADID","sourceUrl":"https://sample-url.com/audience.csv","renewable":true}` {
		t.Errorf("request = %s %s %s", req.Method, req.Path, req.Body)
	}
	if d.ID != 2657 || d.SourceURL != "https://sample-url.com/*****ience.csv" || !d.Renewable || d.RenewalExpireDateTime.IsZero() {
		t.Errorf("customer data = %+v", d)
	}

	n := len(srv.Requests())
	for _, d := range []NewCustomerDataURL{
		{FileType: CustomerDataFileADID, SourceURL: "https://a"},
		{Name: "n", SourceURL: "https://a"},
		{Name: "n", FileType: CustomerDataFileADID},
		{Name: "n", FileType: CustomerDataFileMessageRetarget, SourceURL: "https://a"},
	} {
		if _, err := c.CustomerData.CreateFromURL(t.Context(), 1111, d); err == nil {
			t.Errorf("CreateFromURL(%+v): want error", d)
		}
	}
	if _, err := c.CustomerData.CreateFromURL(t.Context(), 0, NewCustomerDataURL{Name: "n", FileType: CustomerDataFileADID, SourceURL: "https://a"}); err == nil {
		t.Error("missing ad account: want error")
	}
	if len(srv.Requests()) != n {
		t.Error("invalid input was sent")
	}
}

func TestCustomerDataService_Update(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PUT /openapi/v4/customerData", 200, `{"id":2658,"adAccountId":27429,"name":"고객데이터 등록","fileType":"ADID","status":"WAITING","successCount":1200,"failedCount":0,"duplicationCount":3,"createdDate":"2026-06-05T09:00:00"}`)

	d, err := c.CustomerData.Update(t.Context(), 1111, 2656, CustomerDataFiles{FileType: CustomerDataFileADID, Files: []UploadFile{{Filename: "audience_new.csv", Data: []byte("x")}}})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodPut || req.Path != "/openapi/v4/customerData" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	fields, files := customerDataParseMultipart(t, req)
	if !reflect.DeepEqual(fields, map[string][]string{"id": {"2656"}, "fileType": {"ADID"}}) ||
		!reflect.DeepEqual(files, map[string][]customerDataFile{"files": {{"audience_new.csv", "text/csv", "x"}}}) {
		t.Errorf("multipart = %v %v", fields, files)
	}
	if d.ID != 2658 || d.SuccessCount != 1200 || d.DuplicationCount != 3 {
		t.Errorf("customer data = %+v", d)
	}

	n := len(srv.Requests())
	if _, err := c.CustomerData.Update(t.Context(), 1111, 0, CustomerDataFiles{FileType: CustomerDataFileADID, Files: []UploadFile{{Data: []byte("x")}}}); err == nil {
		t.Error("missing customer data: want error")
	}
	if _, err := c.CustomerData.Update(t.Context(), 1111, 2656, CustomerDataFiles{FileType: CustomerDataFileADID}); err == nil {
		t.Error("missing files: want error")
	}
	if len(srv.Requests()) != n {
		t.Error("invalid input was sent")
	}
}

func TestCustomerDataService_UpdateFromURL(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PUT /openapi/v4/customerData/url", 200, `{"id":2659,"adAccountId":27429,"name":"고객데이터 URL 등록","fileType":"ADID","status":"WAITING","createdDate":"2026-06-05T09:00:00","sourceUrl":"https://sample-url.com/*****e_v2.csv","renewable":false,"renewalExpireDateTime":"2026-09-01T00:03:58"}`)

	d, err := c.CustomerData.UpdateFromURL(t.Context(), 1111, 2657, CustomerDataURLChanges{SourceURL: "https://sample-url.com/audience_v2.csv"})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodPut || req.Path != "/openapi/v4/customerData/url" ||
		string(req.Body) != `{"id":2657,"sourceUrl":"https://sample-url.com/audience_v2.csv","renewable":false}` {
		t.Errorf("request = %s %s %s", req.Method, req.Path, req.Body)
	}
	if d.ID != 2659 || d.Renewable || d.SourceURL != "https://sample-url.com/*****e_v2.csv" {
		t.Errorf("customer data = %+v", d)
	}

	if _, err := c.CustomerData.UpdateFromURL(t.Context(), 1111, 2657, CustomerDataURLChanges{SourceURL: "https://a", Renewable: true, FileType: CustomerDataFileHashedEmail}); err != nil {
		t.Fatal(err)
	}
	if got := string(srv.LastRequest().Body); got != `{"id":2657,"sourceUrl":"https://a","renewable":true,"fileType":"HASHED_EMAIL"}` {
		t.Errorf("body = %s", got)
	}

	n := len(srv.Requests())
	for _, ch := range []CustomerDataURLChanges{{}, {SourceURL: "https://a", FileType: CustomerDataFileMessageRetarget}} {
		if _, err := c.CustomerData.UpdateFromURL(t.Context(), 1111, 2657, ch); err == nil {
			t.Errorf("UpdateFromURL(%+v): want error", ch)
		}
	}
	if _, err := c.CustomerData.UpdateFromURL(t.Context(), 0, 2657, CustomerDataURLChanges{SourceURL: "https://a"}); err == nil {
		t.Error("missing ad account: want error")
	}
	if len(srv.Requests()) != n {
		t.Error("invalid input was sent")
	}
}

func TestCustomerDataService_Rename(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PUT /openapi/v4/customerData/name", 200, `{"id":2656,"adAccountId":27429,"name":"고객데이터 이름 변경","kuidListKey":"c8010decafde484eb33284b53a290c30","ready":true,"status":"COMPLETE","fileType":"ADID","type":"URL","sourceUrl":"https://sample-url.com/*****valid1.csv","renewable":true,"renewalExpireDateTime":"2026-09-01T00:03:58","populationScore":7170,"createdDate":"2026-06-03T00:03:59","lastModifiedDate":"2026-06-03T00:13:59","originalCreatedDate":"2026-06-01T00:03:59","populationUpdateDate":"2026-06-03T00:13:59"}`)

	d, err := c.CustomerData.Rename(t.Context(), 1111, 2656, "고객데이터 이름 변경")
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodPut || req.Path != "/openapi/v4/customerData/name" || string(req.Body) != `{"id":2656,"name":"고객데이터 이름 변경"}` {
		t.Errorf("request = %s %s %s", req.Method, req.Path, req.Body)
	}
	customerDataCheck(t, d, "고객데이터 이름 변경")
	if _, err := c.CustomerData.Rename(t.Context(), 1111, 2656, ""); err == nil {
		t.Error("missing name: want error")
	}
}

func TestCustomerDataService_Delete(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("DELETE /openapi/v4/customerData/{id}", 200, ``)

	if err := c.CustomerData.Delete(t.Context(), 1111, 2656); err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodDelete || req.Path != "/openapi/v4/customerData/2656" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	if err := c.CustomerData.Delete(t.Context(), 1111, 0); err == nil {
		t.Error("missing customer data: want error")
	}
}

func TestCustomerDataService_DeleteMany(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("DELETE /openapi/v4/customerData", 200, `{"successCount":1,"failCount":1,"errorMessages":["타겟을 사용 중인 오디언스가 있습니다."]}`)

	res, err := c.CustomerData.DeleteMany(t.Context(), 1111, []int64{2656, 2657})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodDelete || req.Path != "/openapi/v4/customerData" || req.Query.Get("customerDataIds") != "2656,2657" {
		t.Errorf("request = %s %s?%v", req.Method, req.Path, req.Query)
	}
	if res.SuccessCount != 1 || res.FailCount != 1 || len(res.ErrorMessages) != 1 {
		t.Errorf("result = %+v", res)
	}
	if _, err := c.CustomerData.DeleteMany(t.Context(), 0, []int64{1}); err == nil {
		t.Error("missing ad account: want error")
	}
}

const customerDataUsagesJSON = `[{"adGroup":{"id":105488,"name":"first_ad_group","adGroupStatus":["LIVE"],"adGroupType":"DISPLAY"},"campaign":{"id":62286,"name":"first_campaign","campaignTypeGoal":{"campaignType":"DISPLAY","goal":"VISITING"}}},{"messageAd":{"id":"msg-ad-1362737813386051585","name":"first_message","opStatus":["READY"]}}]`

// customerDataCheckUsages checks the decoded usage example.
func customerDataCheckUsages(t *testing.T, us []AdGroupAndCampaign) {
	t.Helper()
	if len(us) != 2 || us[0].AdGroup == nil || us[0].Campaign == nil || us[0].MessageAd != nil {
		t.Fatalf("usages = %+v", us)
	}
	g := TargetAdGroup{AdGroup: AdGroup{ID: 105488, Name: "first_ad_group", AdGroupStatus: []AdGroupStatus{AdGroupStatusLive}}, AdGroupType: AdGroupTypeDisplay}
	if !reflect.DeepEqual(*us[0].AdGroup, g) || us[0].Campaign.ID != 62286 || us[0].Campaign.CampaignTypeGoal.Goal != GoalVisiting {
		t.Errorf("usage = %+v %+v", us[0].AdGroup, us[0].Campaign)
	}
	m := us[1].MessageAd
	if m == nil || m.ID != "msg-ad-1362737813386051585" || m.Name != "first_message" || len(m.OpStatus) != 1 || m.OpStatus[0] != "READY" {
		t.Errorf("message usage = %+v", us[1])
	}
}

func TestCustomerDataService_Usages(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/customerData/usages/{id}", 200, customerDataUsagesJSON)

	us, err := c.CustomerData.Usages(t.Context(), 1111, 2656)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodGet || req.Path != "/openapi/v4/customerData/usages/2656" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	customerDataCheckUsages(t, us)
	if _, err := c.CustomerData.Usages(t.Context(), 1111, 0); err == nil {
		t.Error("missing customer data: want error")
	}
}
