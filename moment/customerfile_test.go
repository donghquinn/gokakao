package moment

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
	"time"
)

var customerFileKST = time.FixedZone("KST", 9*60*60)

func TestCustomerFileService_List(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/customerFiles", 200, `[{"id":1,"adAccountId":1234,"name":"첫번째_고객파일","adidListKey":"a1234b567890123cde45f6g7890hij23","customerFileStatus":"COMPLETE","ready":true,"createdDate":"2020-01-01 00:00","lastModifyRequestDate":"2020-01-01 00:00","originalCreatedDate":"2020-01-01 00:00","populationUpdateDate":"2020-01-01 14:00"}]`)

	fs, err := c.CustomerFiles.List(t.Context(), 1111)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodGet || req.Path != "/openapi/v4/customerFiles" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	if len(fs) != 1 {
		t.Fatalf("files = %+v", fs)
	}
	f := fs[0]
	if f.ID != 1 || f.AdAccountID != 1234 || f.Name != "첫번째_고객파일" || f.AdidListKey != "a1234b567890123cde45f6g7890hij23" ||
		f.CustomerFileStatus != CustomerFileStatusComplete || !f.Ready ||
		!f.CreatedDate.Equal(time.Date(2020, 1, 1, 0, 0, 0, 0, customerFileKST)) ||
		!f.LastModifyRequestDate.Equal(f.CreatedDate.Time) || !f.OriginalCreatedDate.Equal(f.CreatedDate.Time) ||
		!f.PopulationUpdateDate.Equal(time.Date(2020, 1, 1, 14, 0, 0, 0, customerFileKST)) {
		t.Errorf("file = %+v", f)
	}
	if _, err := c.CustomerFiles.List(t.Context(), 0); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestCustomerFile_UnmarshalJSON(t *testing.T) {
	t.Parallel()
	for _, in := range []string{`{"createdDate":"yesterday"}`, `{"id":"x"}`} {
		var f CustomerFile
		if err := json.Unmarshal([]byte(in), &f); err == nil {
			t.Errorf("Unmarshal(%s): want error", in)
		}
	}
	var f CustomerFile
	if err := json.Unmarshal([]byte(`{"status":"WAITING","customerFileStatus":"COMPLETE"}`), &f); err != nil || f.CustomerFileStatus != CustomerFileStatusComplete {
		t.Errorf("status = %q, %v", f.CustomerFileStatus, err)
	}
}

func TestCustomerFileService_Get(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/customerFiles/{id}", 200, `{"id":2656,"adAccountId":27429,"name":"고객파일 등록","adidListKey":"c8010decafde484eb33284b53a290c30","customerFileStatus":"MODIFYING","populationScore":7170,"ready":true,"createdDate":"2024-02-06T00:03:59","lastModifyRequestDate":"2024-02-06T15:35:44","originalCreatedDate":"2024-02-05T16:58:20","populationUpdateDate":"2024-02-06T01:34:00","type":"URL","sourceUrl":"https://sample-url.com/*****valid1.csv","renewable":true,"renewalExpireDateTime":"2024-05-06T00:03:58","transformStatus":"COMPLETED","transformDateTime":"2024-02-06T00:16:19"}`)

	f, err := c.CustomerFiles.Get(t.Context(), 1111, 2656)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodGet || req.Path != "/openapi/v4/customerFiles/2656" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	if f.ID != 2656 || f.CustomerFileStatus != CustomerFileStatusModifying || f.PopulationScore != 7170 || f.Type != CustomerDataURL ||
		f.SourceURL != "https://sample-url.com/*****valid1.csv" || !f.Renewable || f.TransformStatus != "COMPLETED" ||
		!f.TransformDateTime.Equal(time.Date(2024, 2, 6, 0, 16, 19, 0, customerFileKST)) ||
		!f.RenewalExpireDateTime.Equal(time.Date(2024, 5, 6, 0, 3, 58, 0, customerFileKST)) ||
		!f.LastModifyRequestDate.Equal(time.Date(2024, 2, 6, 15, 35, 44, 0, customerFileKST)) {
		t.Errorf("file = %+v", f)
	}
	if _, err := c.CustomerFiles.Get(t.Context(), 1111, 0); err == nil {
		t.Error("missing customer file: want error")
	}
}

const customerFileUploadResponse = `{"id":1,"adAccountId":1234,"name":"첫번째_고객파일","successCount":100,"failedCount":0,"successFileUrl":"https://stwg.kakaocdn.net/success.csv","fileType":"ADID","customerFileStatus":"WAITING","createdDate":"2020-01-01 00:00:00"}`

// customerFileCheckUpload checks the decoded upload example.
func customerFileCheckUpload(t *testing.T, f *CustomerFile) {
	t.Helper()
	if f.ID != 1 || f.AdAccountID != 1234 || f.SuccessCount != 100 || f.SuccessFileURL != "https://stwg.kakaocdn.net/success.csv" ||
		f.FileType != CustomerDataFileADID || f.CustomerFileStatus != CustomerFileStatusWaiting ||
		!f.CreatedDate.Equal(time.Date(2020, 1, 1, 0, 0, 0, 0, customerFileKST)) {
		t.Errorf("file = %+v", f)
	}
}

func TestCustomerFileService_Create(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v4/customerFiles", 200, customerFileUploadResponse)

	f, err := c.CustomerFiles.Create(t.Context(), 1111, NewCustomerFile{Name: "첫번째_고객파일", Files: []UploadFile{
		{Filename: "sample1.csv", Data: []byte("1")}, {Filename: "sample2.csv", Data: []byte("2")}, {Data: []byte("3")},
	}})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodPost || req.Path != "/openapi/v4/customerFiles" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	fields, files := customerDataParseMultipart(t, req)
	wantFiles := map[string][]customerDataFile{"files": {{"sample1.csv", "text/csv", "1"}, {"sample2.csv", "text/csv", "2"}, {"customer.csv", "text/csv", "3"}}}
	if !reflect.DeepEqual(fields, map[string][]string{"name": {"첫번째_고객파일"}}) || !reflect.DeepEqual(files, wantFiles) {
		t.Errorf("multipart = %v %v", fields, files)
	}
	customerFileCheckUpload(t, f)

	n := len(srv.Requests())
	for _, nf := range []NewCustomerFile{{Files: []UploadFile{{Data: []byte("1")}}}, {Name: "n"}} {
		if _, err := c.CustomerFiles.Create(t.Context(), 1111, nf); err == nil {
			t.Errorf("Create(%+v): want error", nf)
		}
	}
	if _, err := c.CustomerFiles.Create(t.Context(), 0, NewCustomerFile{Name: "n", Files: []UploadFile{{Data: []byte("1")}}}); err == nil {
		t.Error("missing ad account: want error")
	}
	if len(srv.Requests()) != n {
		t.Error("invalid input was sent")
	}
}

func TestCustomerFileService_CreateFromURL(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v4/customerFiles/url", 200, `{"id":1234,"adAccountId":12345,"name":"test9","adidListKey":"1be324a5e8314cec9ddc526e2e9858a7","customerFileStatus":"WAITING","ready":false,"createdDate":"2024-01-25T19:31:41.571411","lastModifiedDate":"2024-01-25T19:31:41.571411","originalCreatedDate":"2024-01-25T19:31:41.489213","populationUpdateDate":null,"type":"URL","sourceUrl":"https://sample-url.com/*****e-file.csv","renewable":false,"renewalExpireDateTime":null,"transformStatus":"STANDBY","transformDateTime":null}`)

	f, err := c.CustomerFiles.CreateFromURL(t.Context(), 1111, NewCustomerFileURL{Name: "test9", SourceURL: "https://sample-url.com/this-is-sample-file.csv", Renewable: true})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodPost || req.Path != "/openapi/v4/customerFiles/url" ||
		string(req.Body) != `{"name":"test9","fileType":"ADID","sourceUrl":"https://sample-url.com/this-is-sample-file.csv","renewable":true}` {
		t.Errorf("request = %s %s %s", req.Method, req.Path, req.Body)
	}
	if f.ID != 1234 || f.CustomerFileStatus != CustomerFileStatusWaiting || f.TransformStatus != "STANDBY" || !f.PopulationUpdateDate.IsZero() ||
		!f.CreatedDate.Equal(time.Date(2024, 1, 25, 19, 31, 41, 571411000, customerFileKST)) || f.LastModifiedDate.IsZero() {
		t.Errorf("file = %+v", f)
	}

	n := len(srv.Requests())
	for _, nf := range []NewCustomerFileURL{{SourceURL: "https://a"}, {Name: "n"}} {
		if _, err := c.CustomerFiles.CreateFromURL(t.Context(), 1111, nf); err == nil {
			t.Errorf("CreateFromURL(%+v): want error", nf)
		}
	}
	if _, err := c.CustomerFiles.CreateFromURL(t.Context(), 0, NewCustomerFileURL{Name: "n", SourceURL: "https://a"}); err == nil {
		t.Error("missing ad account: want error")
	}
	if len(srv.Requests()) != n {
		t.Error("invalid input was sent")
	}
}

func TestCustomerFileService_Update(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PUT /openapi/v4/customerFiles", 200, customerFileUploadResponse)

	f, err := c.CustomerFiles.Update(t.Context(), 1111, 1234, []UploadFile{{Filename: "sample1.csv", Data: []byte("1")}})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodPut || req.Path != "/openapi/v4/customerFiles" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	fields, files := customerDataParseMultipart(t, req)
	if !reflect.DeepEqual(fields, map[string][]string{"id": {"1234"}}) ||
		!reflect.DeepEqual(files, map[string][]customerDataFile{"files": {{"sample1.csv", "text/csv", "1"}}}) {
		t.Errorf("multipart = %v %v", fields, files)
	}
	customerFileCheckUpload(t, f)

	n := len(srv.Requests())
	if _, err := c.CustomerFiles.Update(t.Context(), 1111, 1234, nil); err == nil {
		t.Error("missing files: want error")
	}
	if _, err := c.CustomerFiles.Update(t.Context(), 1111, 0, []UploadFile{{Data: []byte("1")}}); err == nil {
		t.Error("missing customer file: want error")
	}
	if len(srv.Requests()) != n {
		t.Error("invalid input was sent")
	}
}

func TestCustomerFileService_UpdateFromURL(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PUT /openapi/v4/customerFiles/url", 200, `{"id":2611,"version":0,"createdDate":"2024-01-31T18:24:32","lastModifiedDate":"2024-01-31T18:24:32","adAccountId":27429,"name":"test1","fileUrl":null,"payloadFileUrl":null,"status":"WAITING","adidListKey":"df3078a4a3f44cb6bc9041d1ca77c73f","fileType":"ADID","type":"URL","sourceUrl":"ttps://sample-url.com/*****e-file.csv","renewable":false,"renewalExpireDateTime":null,"payloadFileUploadDate":null,"dataRegStatusTargeting":"STANDBY","dataRegStatusPopulation":"STANDBY","transformStatus":"STANDBY","transformDateTime":null,"originalCreatedDate":"2024-01-30T14:34:22","populationUpdateDate":null,"modifyStatus":"WAITING","maskedSourceUrl":"https://sample-url.com/*****e-file.csv"}`)

	f, err := c.CustomerFiles.UpdateFromURL(t.Context(), 1111, 9999, CustomerFileURLChanges{SourceURL: "https://sample-url.com/customer_file_invalid1.csv"})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodPut || req.Path != "/openapi/v4/customerFiles/url" ||
		string(req.Body) != `{"id":9999,"fileType":"ADID","sourceUrl":"https://sample-url.com/customer_file_invalid1.csv","renewable":false}` {
		t.Errorf("request = %s %s %s", req.Method, req.Path, req.Body)
	}
	if f.ID != 2611 || f.CustomerFileStatus != CustomerFileStatusWaiting || f.Name != "test1" ||
		!f.OriginalCreatedDate.Equal(time.Date(2024, 1, 30, 14, 34, 22, 0, customerFileKST)) {
		t.Errorf("file = %+v", f)
	}

	n := len(srv.Requests())
	if _, err := c.CustomerFiles.UpdateFromURL(t.Context(), 1111, 9999, CustomerFileURLChanges{}); err == nil {
		t.Error("missing URL: want error")
	}
	if _, err := c.CustomerFiles.UpdateFromURL(t.Context(), 1111, 0, CustomerFileURLChanges{SourceURL: "https://a"}); err == nil {
		t.Error("missing customer file: want error")
	}
	if len(srv.Requests()) != n {
		t.Error("invalid input was sent")
	}
}

func TestCustomerFileService_Rename(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("PUT /openapi/v4/customerFiles/name", 200, `{"id":2687,"adAccountId":27429,"name":"변경한_고객파일이름","adidListKey":"da4a57899e43428bb7e47b09ac9c4311","customerFileStatus":"COMPLETE","ready":true,"createdDate":"2024-02-07T00:03:38","lastModifiedDate":"2024-02-07T17:58:11.591512","originalCreatedDate":"2024-02-06T17:12:15","populationUpdateDate":"2024-02-07T01:34:00","type":"URL","sourceUrl":"https://sample-url.com/*****valid1.csv","renewable":true,"renewalExpireDateTime":"2024-05-07T00:03:38","transformStatus":"COMPLETED","transformDateTime":"2024-02-07T00:16:23"}`)

	f, err := c.CustomerFiles.Rename(t.Context(), 1111, 123, "변경한_고객파일이름")
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodPut || req.Path != "/openapi/v4/customerFiles/name" || string(req.Body) != `{"id":123,"name":"변경한_고객파일이름"}` {
		t.Errorf("request = %s %s %s", req.Method, req.Path, req.Body)
	}
	if f.ID != 2687 || f.Name != "변경한_고객파일이름" || f.CustomerFileStatus != CustomerFileStatusComplete ||
		!f.LastModifiedDate.Equal(time.Date(2024, 2, 7, 17, 58, 11, 591512000, customerFileKST)) {
		t.Errorf("file = %+v", f)
	}
	if _, err := c.CustomerFiles.Rename(t.Context(), 1111, 123, ""); err == nil {
		t.Error("missing name: want error")
	}
}

func TestCustomerFileService_Delete(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("DELETE /openapi/v4/customerFiles/{id}", 200, ``)

	if err := c.CustomerFiles.Delete(t.Context(), 1111, 5); err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodDelete || req.Path != "/openapi/v4/customerFiles/5" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	if err := c.CustomerFiles.Delete(t.Context(), 0, 5); err == nil {
		t.Error("missing ad account: want error")
	}
}

func TestCustomerFileService_DeleteMany(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("DELETE /openapi/v4/customerFiles", 200, `{"successCount":1,"failCount":1,"errorMessages":["타겟을 사용 중인 오디언스가 있습니다."]}`)

	res, err := c.CustomerFiles.DeleteMany(t.Context(), 1111, []int64{5, 6})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodDelete || req.Path != "/openapi/v4/customerFiles" || req.Query.Get("customerFileIds") != "5,6" {
		t.Errorf("request = %s %s?%v", req.Method, req.Path, req.Query)
	}
	if res.SuccessCount != 1 || res.FailCount != 1 || res.ErrorMessages[0] != "타겟을 사용 중인 오디언스가 있습니다." {
		t.Errorf("result = %+v", res)
	}
	if _, err := c.CustomerFiles.DeleteMany(t.Context(), 1111, nil); err == nil {
		t.Error("missing IDs: want error")
	}
}

func TestCustomerFileService_Usages(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/customerFiles/usages/{id}", 200, `[{"adGroup":{"id":56,"name":"첫번째_광고그룹","adGroupStatus":["LIVE"],"adGroupType":"DISPLAY"},"campaign":{"id":78,"name":"첫번째_캠페인","campaignTypeGoal":{"campaignType":"DISPLAY","goal":"VISITING"}}}]`)

	us, err := c.CustomerFiles.Usages(t.Context(), 1111, 5)
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111")
	if req.Method != http.MethodGet || req.Path != "/openapi/v4/customerFiles/usages/5" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	if len(us) != 1 || us[0].AdGroup == nil || us[0].Campaign == nil || us[0].MessageAd != nil ||
		us[0].AdGroup.ID != 56 || us[0].AdGroup.Name != "첫번째_광고그룹" || us[0].AdGroup.AdGroupType != AdGroupTypeDisplay ||
		us[0].AdGroup.AdGroupStatus[0] != AdGroupStatusLive || us[0].Campaign.ID != 78 || us[0].Campaign.CampaignTypeGoal.CampaignType != CampaignTypeDisplay {
		t.Errorf("usages = %+v", us)
	}
	if _, err := c.CustomerFiles.Usages(t.Context(), 1111, 0); err == nil {
		t.Error("missing customer file: want error")
	}
}
