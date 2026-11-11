package moment

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	kakao "github.com/donghquinn/gokakao"
)

// CustomerData is customer data (고객데이터): a list of user identifiers
// that display and message ad groups can target.
//
// The registration responses of [CustomerDataService.Create] and
// [CustomerDataService.Update] also fill AdGroupIDs and the counts.
type CustomerData struct {
	ID          kakao.Int64 `json:"id"`
	AdAccountID kakao.Int64 `json:"adAccountId"`
	Name        string      `json:"name"`
	// KuidListKey is the registration key (고객데이터 등록 Key).
	KuidListKey string             `json:"kuidListKey"`
	Status      CustomerDataStatus `json:"status"`
	// Ready reports whether the data can be used for targeting.
	Ready    bool                 `json:"ready"`
	FileType CustomerDataFileType `json:"fileType"`
	Type     CustomerDataType     `json:"type"`
	// SourceURL is the registered URL, partly masked by Kakao.
	SourceURL string `json:"sourceUrl"`
	// Renewable reports whether Kakao refetches SourceURL daily.
	Renewable bool `json:"renewable"`
	// RenewalExpireDateTime ends the daily refetch, 90 days after the
	// target was created.
	RenewalExpireDateTime kakao.Time `json:"renewalExpireDateTime"`
	// PopulationScore is the number of matched Kakao users (타겟 모수).
	PopulationScore      kakao.Int64 `json:"populationScore"`
	PopulationUpdateDate kakao.Time  `json:"populationUpdateDate"`
	// CreatedDate is when this version was registered; after an update
	// OriginalCreatedDate is when the first version was.
	CreatedDate         kakao.Time `json:"createdDate"`
	OriginalCreatedDate kakao.Time `json:"originalCreatedDate"`
	LastModifiedDate    kakao.Time `json:"lastModifiedDate"`

	// AdGroupIDs are the message ad groups of MESSAGE_RETARGET data.
	AdGroupIDs []kakao.Int64 `json:"adGroupIds"`
	// SuccessCount, FailedCount and DuplicationCount count the identifiers
	// of an uploaded file.
	SuccessCount     int `json:"successCount"`
	FailedCount      int `json:"failedCount"`
	DuplicationCount int `json:"duplicationCount"`
}

// AdGroupAndCampaign is an ad group with its campaign, or a message,
// that uses a target (사용 현황) or can be used to create an ad reaction
// target. Either Campaign and AdGroup or MessageAd is set.
type AdGroupAndCampaign struct {
	Campaign  *Campaign      `json:"campaign"`
	AdGroup   *TargetAdGroup `json:"adGroup"`
	MessageAd *MessageAd     `json:"messageAd"`
}

// TargetAdGroup is an ad group in [AdGroupAndCampaign].
type TargetAdGroup struct {
	AdGroup
	AdGroupType AdGroupType `json:"adGroupType"`
}

// List returns the ad account's customer data.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/customer-data
func (s *CustomerDataService) List(ctx context.Context, adAccountID int64) ([]CustomerData, error) {
	if adAccountID == 0 {
		return nil, errors.New("moment: ad account ID is required")
	}
	var res []CustomerData
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v4/customerData", adAccountID: adAccountID}, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// Get returns customer data. Kakao allows one call per customer data per 5
// seconds.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/customer-data
func (s *CustomerDataService) Get(ctx context.Context, adAccountID, customerDataID int64) (*CustomerData, error) {
	if adAccountID == 0 || customerDataID == 0 {
		return nil, errors.New("moment: ad account ID and customer data ID are required")
	}
	var d CustomerData
	if err := s.c.do(ctx, call{method: http.MethodGet, path: idPath("/v4/customerData/%d", customerDataID), adAccountID: adAccountID}, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// NewCustomerData describes customer data to register from CSV files.
type NewCustomerData struct {
	// Name is up to 120 characters.
	Name     string
	FileType CustomerDataFileType
	// Files are up to 10 CSV files of at most 200 MB in total, in the
	// format of FileType. ContentType defaults to text/csv and Filename
	// to "customer.csv".
	// Required unless FileType is MESSAGE_RETARGET.
	Files []UploadFile
	// AdGroupIDs are the message ad groups whose recipients to target;
	// required, instead of Files, for MESSAGE_RETARGET.
	AdGroupIDs []int64
}

// Create registers customer data from CSV files, sent as
// multipart/form-data. An account holds up to 50 customer data, and Kakao
// matches users within an hour. Kakao allows one call per user and ad
// account per 5 seconds.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/customer-data
func (s *CustomerDataService) Create(ctx context.Context, adAccountID int64, d NewCustomerData) (*CustomerData, error) {
	if adAccountID == 0 {
		return nil, errors.New("moment: ad account ID is required")
	}
	if d.Name == "" {
		return nil, errors.New("moment: customer data name is required")
	}
	m, err := customerDataMultipart(d.FileType, d.Files, d.AdGroupIDs)
	if err != nil {
		return nil, err
	}
	m.Fields.Set("name", d.Name)
	var res CustomerData
	if err := s.c.do(ctx, call{method: http.MethodPost, path: "/v4/customerData", adAccountID: adAccountID, multipart: m}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// customerDataMultipart builds the file type, files and ad group IDs of a
// customer data upload.
func customerDataMultipart(ft CustomerDataFileType, files []UploadFile, adGroupIDs []int64) (*kakao.Multipart, error) {
	if ft == "" {
		return nil, errors.New("moment: customer data file type is required")
	}
	m := &kakao.Multipart{Fields: url.Values{"fileType": {string(ft)}}}
	if ft == CustomerDataFileMessageRetarget {
		if len(adGroupIDs) == 0 || len(files) > 0 {
			return nil, errors.New("moment: MESSAGE_RETARGET customer data takes ad group IDs and no files")
		}
		for _, id := range adGroupIDs {
			m.Fields.Add("adGroupIds", strconv.FormatInt(id, 10))
		}
		return m, nil
	}
	if len(adGroupIDs) > 0 {
		return nil, errors.New("moment: ad group IDs are only for MESSAGE_RETARGET customer data")
	}
	csv, err := csvFiles(files, "customer.csv")
	if err != nil {
		return nil, err
	}
	m.Files = csv
	return m, nil
}

// csvFiles validates 1 to 10 CSV files and sends them as "files".
func csvFiles(files []UploadFile, defaultName string) ([]kakao.MultipartFile, error) {
	if len(files) == 0 || len(files) > 10 {
		return nil, errors.New("moment: 1 to 10 CSV files are required")
	}
	out := make([]kakao.MultipartFile, len(files))
	for i, f := range files {
		if f.Filename == "" {
			f.Filename = defaultName
		}
		if f.ContentType == "" {
			f.ContentType = "text/csv"
		}
		part, err := f.part("files")
		if err != nil {
			return nil, err
		}
		out[i] = part
	}
	return out, nil
}

// NewCustomerDataURL describes customer data to register from a CSV URL.
type NewCustomerDataURL struct {
	// Name is up to 120 characters.
	Name string `json:"name"`
	// FileType cannot be MESSAGE_RETARGET.
	FileType CustomerDataFileType `json:"fileType"`
	// SourceURL is a public http:// or https:// URL of a CSV file, with no
	// size limit.
	SourceURL string `json:"sourceUrl"`
	// Renewable makes Kakao refetch SourceURL daily for 90 days.
	Renewable bool `json:"renewable"`
}

// CreateFromURL registers customer data from a CSV file URL. Kakao matches
// users within an hour. Kakao allows one call per user and ad account per
// 5 seconds.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/customer-data
func (s *CustomerDataService) CreateFromURL(ctx context.Context, adAccountID int64, d NewCustomerDataURL) (*CustomerData, error) {
	if adAccountID == 0 {
		return nil, errors.New("moment: ad account ID is required")
	}
	if d.Name == "" || d.FileType == "" || d.SourceURL == "" {
		return nil, errors.New("moment: customer data name, file type and source URL are required")
	}
	if d.FileType == CustomerDataFileMessageRetarget {
		return nil, errors.New("moment: MESSAGE_RETARGET customer data cannot be registered from a URL")
	}
	var res CustomerData
	if err := s.c.do(ctx, call{method: http.MethodPost, path: "/v4/customerData/url", adAccountID: adAccountID, body: d}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// CustomerDataFiles is the new content of file-registered customer data
// for [CustomerDataService.Update]. Its fields follow [NewCustomerData].
type CustomerDataFiles struct {
	FileType   CustomerDataFileType
	Files      []UploadFile
	AdGroupIDs []int64
}

// Update replaces the files (or message ad groups) of customer data that
// was registered from files and is COMPLETE, sent as multipart/form-data.
// It fails while a message ad uses the data. Until Kakao finishes matching,
// within an hour, ad groups keep targeting the previous users. Kakao may
// return a new ID. Kakao allows one call per customer data per 5 seconds.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/customer-data
func (s *CustomerDataService) Update(ctx context.Context, adAccountID, customerDataID int64, d CustomerDataFiles) (*CustomerData, error) {
	if adAccountID == 0 || customerDataID == 0 {
		return nil, errors.New("moment: ad account ID and customer data ID are required")
	}
	m, err := customerDataMultipart(d.FileType, d.Files, d.AdGroupIDs)
	if err != nil {
		return nil, err
	}
	m.Fields.Set("id", strconv.FormatInt(customerDataID, 10))
	var res CustomerData
	if err := s.c.do(ctx, call{method: http.MethodPut, path: "/v4/customerData", adAccountID: adAccountID, multipart: m}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// CustomerDataURLChanges is the new source of URL-registered customer data
// for [CustomerDataService.UpdateFromURL].
type CustomerDataURLChanges struct {
	SourceURL string `json:"sourceUrl"`
	// Renewable makes Kakao refetch SourceURL daily.
	Renewable bool `json:"renewable"`
	// FileType is listed as required by Kakao's field table but omitted
	// from its example; it is sent when set and cannot be
	// MESSAGE_RETARGET.
	FileType CustomerDataFileType `json:"fileType,omitempty"`
}

// UpdateFromURL replaces the URL and renewal option of customer data that
// was registered from a URL. Until Kakao finishes matching, within an
// hour, ad groups keep targeting the previous users. Kakao may return a
// new ID. Kakao allows one call per customer data per 5 seconds.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/customer-data
func (s *CustomerDataService) UpdateFromURL(ctx context.Context, adAccountID, customerDataID int64, d CustomerDataURLChanges) (*CustomerData, error) {
	if adAccountID == 0 || customerDataID == 0 {
		return nil, errors.New("moment: ad account ID and customer data ID are required")
	}
	if d.SourceURL == "" {
		return nil, errors.New("moment: customer data source URL is required")
	}
	if d.FileType == CustomerDataFileMessageRetarget {
		return nil, errors.New("moment: MESSAGE_RETARGET customer data cannot be registered from a URL")
	}
	body := struct {
		ID int64 `json:"id"`
		CustomerDataURLChanges
	}{customerDataID, d}
	var res CustomerData
	if err := s.c.do(ctx, call{method: http.MethodPut, path: "/v4/customerData/url", adAccountID: adAccountID, body: body}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Rename changes the name of customer data, up to 120 characters.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/customer-data
func (s *CustomerDataService) Rename(ctx context.Context, adAccountID, customerDataID int64, name string) (*CustomerData, error) {
	if adAccountID == 0 || customerDataID == 0 || name == "" {
		return nil, errors.New("moment: ad account ID, customer data ID and name are required")
	}
	body := struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}{customerDataID, name}
	var res CustomerData
	if err := s.c.do(ctx, call{method: http.MethodPut, path: "/v4/customerData/name", adAccountID: adAccountID, body: body}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Delete deletes customer data and its uploaded files. It fails while an
// ad group uses the data.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/customer-data
func (s *CustomerDataService) Delete(ctx context.Context, adAccountID, customerDataID int64) error {
	if adAccountID == 0 || customerDataID == 0 {
		return errors.New("moment: ad account ID and customer data ID are required")
	}
	return s.c.do(ctx, call{method: http.MethodDelete, path: idPath("/v4/customerData/%d", customerDataID), adAccountID: adAccountID}, nil)
}

// DeleteMany deletes several customer data. Data used by ad groups or
// messages is not deleted and is reported in the result.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/customer-data
func (s *CustomerDataService) DeleteMany(ctx context.Context, adAccountID int64, customerDataIDs []int64) (*BulkDeleteResult, error) {
	return deleteMany(ctx, s.c, adAccountID, "/v4/customerData", "customerDataIds", customerDataIDs)
}

// Usages returns the ad groups, with their campaigns, and messages that use
// customer data (사용 현황).
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/customer-data
func (s *CustomerDataService) Usages(ctx context.Context, adAccountID, customerDataID int64) ([]AdGroupAndCampaign, error) {
	if adAccountID == 0 || customerDataID == 0 {
		return nil, errors.New("moment: ad account ID and customer data ID are required")
	}
	var res []AdGroupAndCampaign
	if err := s.c.do(ctx, call{method: http.MethodGet, path: idPath("/v4/customerData/usages/%d", customerDataID), adAccountID: adAccountID}, &res); err != nil {
		return nil, err
	}
	return res, nil
}
