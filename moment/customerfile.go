package moment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	kakao "github.com/donghquinn/gokakao"
)

// CustomerFileStatusTransformError means Kakao could not fetch the CSV file
// of a URL-registered customer file (URL 유형에서 CSV파일 등록 실패).
const CustomerFileStatusTransformError CustomerFileStatus = "TRANSFORM_ERROR"

// CustomerFile is a customer file (고객파일): a list of advertising IDs
// that display ad groups can target. Customer files are being retired;
// register new targets as [CustomerData].
//
// The registration responses also fill the counts, file URLs and FileType.
type CustomerFile struct {
	ID          kakao.Int64 `json:"id"`
	AdAccountID kakao.Int64 `json:"adAccountId"`
	Name        string      `json:"name"`
	// AdidListKey is the registration key (고객파일 등록 Key).
	AdidListKey string `json:"adidListKey"`
	// CustomerFileStatus is read from "customerFileStatus", or from
	// "status" as the URL update response sends it.
	CustomerFileStatus CustomerFileStatus `json:"customerFileStatus"`
	// PopulationScore is the number of matched Kakao users (타겟 모수),
	// available up to 6 hours after registration.
	PopulationScore kakao.Int64 `json:"populationScore"`
	// Ready reports whether the file can be used for targeting.
	Ready bool `json:"ready"`
	// Type is how the file was registered, e.g. "URL".
	Type CustomerDataType `json:"type"`
	// SourceURL is the registered URL, partly masked by Kakao.
	SourceURL string `json:"sourceUrl"`
	// Renewable reports whether Kakao refetches SourceURL daily.
	Renewable bool `json:"renewable"`
	// RenewalExpireDateTime ends the daily refetch, 90 days after the
	// target was created.
	RenewalExpireDateTime kakao.Time `json:"renewalExpireDateTime"`
	// TransformStatus is the state of fetching SourceURL, e.g. STANDBY or
	// COMPLETED (undocumented).
	TransformStatus   string     `json:"transformStatus"`
	TransformDateTime kakao.Time `json:"transformDateTime"`
	// CreatedDate is when this version was registered; after an update
	// OriginalCreatedDate is when the first version was.
	CreatedDate           kakao.Time `json:"createdDate"`
	OriginalCreatedDate   kakao.Time `json:"originalCreatedDate"`
	LastModifyRequestDate kakao.Time `json:"lastModifyRequestDate"`
	LastModifiedDate      kakao.Time `json:"lastModifiedDate"`
	PopulationUpdateDate  kakao.Time `json:"populationUpdateDate"`

	// FileType is always ADID.
	FileType       CustomerDataFileType `json:"fileType"`
	SuccessCount   int                  `json:"successCount"`
	FailedCount    int                  `json:"failedCount"`
	SuccessFileURL string               `json:"successFileUrl"`
	FailedFileURL  string               `json:"failedFileUrl"`
}

// UnmarshalJSON implements [json.Unmarshaler]. Besides the formats of
// [kakao.Time], it reads "yyyy-MM-dd HH:mm" dates, which Kakao's list
// example uses, as KST.
func (f *CustomerFile) UnmarshalJSON(b []byte) error {
	type plain CustomerFile
	var v struct {
		plain
		Status                CustomerFileStatus `json:"status"`
		RenewalExpireDateTime json.RawMessage    `json:"renewalExpireDateTime"`
		TransformDateTime     json.RawMessage    `json:"transformDateTime"`
		CreatedDate           json.RawMessage    `json:"createdDate"`
		OriginalCreatedDate   json.RawMessage    `json:"originalCreatedDate"`
		LastModifyRequestDate json.RawMessage    `json:"lastModifyRequestDate"`
		LastModifiedDate      json.RawMessage    `json:"lastModifiedDate"`
		PopulationUpdateDate  json.RawMessage    `json:"populationUpdateDate"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*f = CustomerFile(v.plain)
	if f.CustomerFileStatus == "" {
		f.CustomerFileStatus = v.Status
	}
	for _, d := range []struct {
		raw json.RawMessage
		dst *kakao.Time
	}{
		{v.RenewalExpireDateTime, &f.RenewalExpireDateTime},
		{v.TransformDateTime, &f.TransformDateTime},
		{v.CreatedDate, &f.CreatedDate},
		{v.OriginalCreatedDate, &f.OriginalCreatedDate},
		{v.LastModifyRequestDate, &f.LastModifyRequestDate},
		{v.LastModifiedDate, &f.LastModifiedDate},
		{v.PopulationUpdateDate, &f.PopulationUpdateDate},
	} {
		if err := customerFileTime(d.raw, d.dst); err != nil {
			return err
		}
	}
	return nil
}

// customerFileTime decodes raw into dst, completing "yyyy-MM-dd HH:mm" with
// seconds first.
func customerFileTime(raw json.RawMessage, dst *kakao.Time) error {
	if len(raw) == 0 {
		return nil
	}
	var s string
	if json.Unmarshal(raw, &s) == nil && len(s) == len("2006-01-02 15:04") && s[10] == ' ' {
		raw = []byte(strconv.Quote(s + ":00"))
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		return fmt.Errorf("moment: customer file: %w", err)
	}
	return nil
}

// List returns the ad account's customer files.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/customer-files
func (s *CustomerFileService) List(ctx context.Context, adAccountID int64) ([]CustomerFile, error) {
	if adAccountID == 0 {
		return nil, errors.New("moment: ad account ID is required")
	}
	var res []CustomerFile
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v4/customerFiles", adAccountID: adAccountID}, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// Get returns a customer file. Kakao allows one call per customer file per
// 5 seconds.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/customer-files
func (s *CustomerFileService) Get(ctx context.Context, adAccountID, customerFileID int64) (*CustomerFile, error) {
	if adAccountID == 0 || customerFileID == 0 {
		return nil, errors.New("moment: ad account ID and customer file ID are required")
	}
	var f CustomerFile
	if err := s.c.do(ctx, call{method: http.MethodGet, path: idPath("/v4/customerFiles/%d", customerFileID), adAccountID: adAccountID}, &f); err != nil {
		return nil, err
	}
	return &f, nil
}

// NewCustomerFile describes a customer file to register from CSV files.
type NewCustomerFile struct {
	// Name is up to 120 characters.
	Name string
	// Files are 1 to 10 CSV files of advertising IDs, at most 200 MB in
	// total. ContentType defaults to text/csv and Filename to
	// "customer.csv".
	Files []UploadFile
}

// Create registers a customer file from CSV files, sent as
// multipart/form-data. An account holds up to 50 customer files, and Kakao
// matches users within 6 hours. Kakao allows one call per user and ad
// account per 5 seconds.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/customer-files
func (s *CustomerFileService) Create(ctx context.Context, adAccountID int64, f NewCustomerFile) (*CustomerFile, error) {
	if adAccountID == 0 {
		return nil, errors.New("moment: ad account ID is required")
	}
	if f.Name == "" {
		return nil, errors.New("moment: customer file name is required")
	}
	files, err := csvFiles(f.Files, "customer.csv")
	if err != nil {
		return nil, err
	}
	m := &kakao.Multipart{Fields: map[string][]string{"name": {f.Name}}, Files: files}
	var res CustomerFile
	if err := s.c.do(ctx, call{method: http.MethodPost, path: "/v4/customerFiles", adAccountID: adAccountID, multipart: m}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// NewCustomerFileURL describes a customer file to register from a CSV
// URL. Its file type is always ADID.
type NewCustomerFileURL struct {
	// Name is up to 120 characters.
	Name string
	// SourceURL is a public http:// or https:// URL of a CSV file of
	// advertising IDs, with no size limit.
	SourceURL string
	// Renewable makes Kakao refetch SourceURL daily for 90 days.
	Renewable bool
}

// CreateFromURL registers a customer file from a CSV file URL. Kakao
// matches users within 6 hours. Kakao allows one call per user and ad
// account per 5 seconds.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/customer-files
func (s *CustomerFileService) CreateFromURL(ctx context.Context, adAccountID int64, f NewCustomerFileURL) (*CustomerFile, error) {
	if adAccountID == 0 {
		return nil, errors.New("moment: ad account ID is required")
	}
	if f.Name == "" || f.SourceURL == "" {
		return nil, errors.New("moment: customer file name and source URL are required")
	}
	body := struct {
		Name      string               `json:"name"`
		FileType  CustomerDataFileType `json:"fileType"`
		SourceURL string               `json:"sourceUrl"`
		Renewable bool                 `json:"renewable"`
	}{f.Name, CustomerDataFileADID, f.SourceURL, f.Renewable}
	var res CustomerFile
	if err := s.c.do(ctx, call{method: http.MethodPost, path: "/v4/customerFiles/url", adAccountID: adAccountID, body: body}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Update replaces the CSV files of a customer file that was registered from
// files and is COMPLETE, sent as multipart/form-data; files follow
// [NewCustomerFile].Files. Until Kakao finishes matching, within 6 hours,
// ad groups keep targeting the previous users. Kakao allows one call per
// customer file per 5 seconds.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/customer-files
func (s *CustomerFileService) Update(ctx context.Context, adAccountID, customerFileID int64, files []UploadFile) (*CustomerFile, error) {
	if adAccountID == 0 || customerFileID == 0 {
		return nil, errors.New("moment: ad account ID and customer file ID are required")
	}
	csv, err := csvFiles(files, "customer.csv")
	if err != nil {
		return nil, err
	}
	m := &kakao.Multipart{Fields: map[string][]string{"id": {strconv.FormatInt(customerFileID, 10)}}, Files: csv}
	var res CustomerFile
	if err := s.c.do(ctx, call{method: http.MethodPut, path: "/v4/customerFiles", adAccountID: adAccountID, multipart: m}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// CustomerFileURLChanges is the new source of a URL-registered customer
// file for [CustomerFileService.UpdateFromURL].
type CustomerFileURLChanges struct {
	SourceURL string
	// Renewable makes Kakao refetch SourceURL daily.
	Renewable bool
}

// UpdateFromURL replaces the URL and renewal option of a customer file that
// was registered from a URL. Until Kakao finishes matching, within 6 hours,
// ad groups keep targeting the previous users. Kakao allows one call per
// customer file per 5 seconds.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/customer-files
func (s *CustomerFileService) UpdateFromURL(ctx context.Context, adAccountID, customerFileID int64, f CustomerFileURLChanges) (*CustomerFile, error) {
	if adAccountID == 0 || customerFileID == 0 {
		return nil, errors.New("moment: ad account ID and customer file ID are required")
	}
	if f.SourceURL == "" {
		return nil, errors.New("moment: customer file source URL is required")
	}
	body := struct {
		ID        int64                `json:"id"`
		FileType  CustomerDataFileType `json:"fileType"`
		SourceURL string               `json:"sourceUrl"`
		Renewable bool                 `json:"renewable"`
	}{customerFileID, CustomerDataFileADID, f.SourceURL, f.Renewable}
	var res CustomerFile
	if err := s.c.do(ctx, call{method: http.MethodPut, path: "/v4/customerFiles/url", adAccountID: adAccountID, body: body}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Rename changes the name of a customer file, up to 120 characters.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/customer-files
func (s *CustomerFileService) Rename(ctx context.Context, adAccountID, customerFileID int64, name string) (*CustomerFile, error) {
	if adAccountID == 0 || customerFileID == 0 || name == "" {
		return nil, errors.New("moment: ad account ID, customer file ID and name are required")
	}
	body := struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}{customerFileID, name}
	var res CustomerFile
	if err := s.c.do(ctx, call{method: http.MethodPut, path: "/v4/customerFiles/name", adAccountID: adAccountID, body: body}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Delete deletes a customer file. It fails while an ad group uses the file.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/customer-files
func (s *CustomerFileService) Delete(ctx context.Context, adAccountID, customerFileID int64) error {
	if adAccountID == 0 || customerFileID == 0 {
		return errors.New("moment: ad account ID and customer file ID are required")
	}
	return s.c.do(ctx, call{method: http.MethodDelete, path: idPath("/v4/customerFiles/%d", customerFileID), adAccountID: adAccountID}, nil)
}

// DeleteMany deletes several customer files. Files used by ad groups are
// not deleted and are reported in the result.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/customer-files
func (s *CustomerFileService) DeleteMany(ctx context.Context, adAccountID int64, customerFileIDs []int64) (*BulkDeleteResult, error) {
	return deleteMany(ctx, s.c, adAccountID, "/v4/customerFiles", "customerFileIds", customerFileIDs)
}

// Usages returns the ad groups, with their campaigns, that use a customer
// file (사용 현황).
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/customer-files
func (s *CustomerFileService) Usages(ctx context.Context, adAccountID, customerFileID int64) ([]AdGroupAndCampaign, error) {
	if adAccountID == 0 || customerFileID == 0 {
		return nil, errors.New("moment: ad account ID and customer file ID are required")
	}
	var res []AdGroupAndCampaign
	if err := s.c.do(ctx, call{method: http.MethodGet, path: idPath("/v4/customerFiles/usages/%d", customerFileID), adAccountID: adAccountID}, &res); err != nil {
		return nil, err
	}
	return res, nil
}
