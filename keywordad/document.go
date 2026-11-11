package keywordad

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	kakao "github.com/donghquinn/gokakao"
)

// BizChannelDocuments are the review documents of a biz channel.
type BizChannelDocuments struct {
	BizChannelID   kakao.Int64      `json:"bizChannelId"`
	BizChannelName string           `json:"bizChannelName"`
	Documents      []ReviewDocument `json:"documents"`
}

// List returns a biz channel's review documents.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/document
func (s *DocumentService) List(ctx context.Context, adAccountID, bizChannelID int64) (*BizChannelDocuments, error) {
	if adAccountID == 0 || bizChannelID == 0 {
		return nil, errors.New("keywordad: ad account ID and biz channel ID are required")
	}
	q := url.Values{"bizChannelId": {strconv.FormatInt(bizChannelID, 10)}}
	var d BizChannelDocuments
	if err := s.c.do(ctx, call{method: http.MethodGet, path: "/v1/documents", adAccountID: adAccountID, query: q}, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// DocumentFile is a local file to upload. Supported formats are JPG, JPEG,
// GIF, PNG, TIF, HWP, DOC, DOCX and PDF, up to 10 MB.
type DocumentFile struct {
	// Filename defaults to "document".
	Filename string
	// ContentType defaults to application/octet-stream.
	ContentType string
	Data        []byte
}

// NewDocument describes a review document to register. Set exactly one of
// FileURL and File.
type NewDocument struct {
	BizChannelID int64
	Type         ReviewDocumentType
	// RequiredDocumentID is the kind of document required for the business
	// category, e.g. 1 (사업자등록증); see RequiredDocument in the
	// reference.
	RequiredDocumentID int64
	// FileURL is an http:// or https:// URL of the file, up to 255
	// characters, that Kakao downloads.
	FileURL string
	// File is uploaded as multipart/form-data.
	File *DocumentFile
}

// Create registers a review document. It is limited to 1 request per 10
// seconds.
//
// Docs: https://developers.kakao.com/docs/ko/keyword-ad/document
func (s *DocumentService) Create(ctx context.Context, adAccountID int64, d NewDocument) error {
	if adAccountID == 0 {
		return errors.New("keywordad: ad account ID is required")
	}
	if d.BizChannelID == 0 || d.Type == "" || d.RequiredDocumentID == 0 {
		return errors.New("keywordad: biz channel ID, type and required document ID are required")
	}
	if (d.FileURL == "") == (d.File == nil) {
		return errors.New("keywordad: set exactly one of file URL and file")
	}
	c := call{method: http.MethodPost, path: "/v1/documents", adAccountID: adAccountID}
	if d.File != nil {
		name := d.File.Filename
		if name == "" {
			name = "document"
		}
		c.multipart = &kakao.Multipart{
			Fields: url.Values{
				"bizChannelId":       {strconv.FormatInt(d.BizChannelID, 10)},
				"type":               {string(d.Type)},
				"requiredDocumentId": {strconv.FormatInt(d.RequiredDocumentID, 10)},
			},
			Files: []kakao.MultipartFile{{
				Field: "file", Filename: name, ContentType: d.File.ContentType, Data: d.File.Data,
			}},
		}
	} else {
		c.body = struct {
			BizChannelID       int64              `json:"bizChannelId"`
			Type               ReviewDocumentType `json:"type"`
			RequiredDocumentID int64              `json:"requiredDocumentId"`
			FileURL            string             `json:"fileUrl"`
		}{d.BizChannelID, d.Type, d.RequiredDocumentID, d.FileURL}
	}
	return s.c.do(ctx, c, nil)
}
