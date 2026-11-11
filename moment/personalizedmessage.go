package moment

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	kakao "github.com/donghquinn/gokakao"
)

// ReceiverType is how a personalized message receiver is identified
// (발송 유형).
type ReceiverType string

// ReceiverType values.
const (
	// ReceiverAppUserID is the app user ID of the app linked to the
	// template's KakaoTalk Channel profile.
	ReceiverAppUserID   ReceiverType = "APP_USER_ID"
	ReceiverPhoneNumber ReceiverType = "PHONE_NUMBER" // 전화번호
)

// PersonalMessageStatus is the result of sending a personalized message
// (발송 결과).
type PersonalMessageStatus string

// PersonalMessageStatus values.
const (
	PersonalMessageSucceeded PersonalMessageStatus = "SUCCEEDED" // 발송 성공
	PersonalMessageFailed    PersonalMessageStatus = "FAILED"    // 발송 실패
)

// TestPersonalMessage is a test send of a personalized message template.
type TestPersonalMessage struct {
	PhoneNumber string `json:"phoneNumber"`
	// Variables maps the template's variable names to values, e.g.
	// {"user_name1": "홍길동"}.
	Variables map[string]string `json:"variables"`
}

// SendTest sends a personalized message template to a phone number for
// testing; "[테스트 발송]" is added to the copy. Kakao allows one call per
// minute per user and ad account.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/personalized-msg
func (s *PersonalizedMessageService) SendTest(ctx context.Context, adAccountID, creativeID int64, msg TestPersonalMessage) error {
	if creativeID == 0 || msg.PhoneNumber == "" {
		return errors.New("moment: creative ID and phone number are required")
	}
	if msg.Variables == nil {
		msg.Variables = map[string]string{}
	}
	return s.c.do(ctx, call{method: http.MethodPost, path: idPath("/v4/messages/creatives/%d/sendTestPersonalMessage", creativeID), adAccountID: adAccountID, body: msg}, nil)
}

// PersonalMessage is one personalized message to send
// (PersonalMessageSendRequest).
type PersonalMessage struct {
	// MessageSerialNumber identifies the message, up to 39 characters in
	// the form "yyyyMMdd-{creativeID}-{unique ID}".
	MessageSerialNumber string       `json:"messageSerialNumber"`
	ReceiverType        ReceiverType `json:"receiverType"`
	// ReceiverKey is the app user ID or phone number.
	ReceiverKey string `json:"receiverKey"`
	// Variables maps the template's variable names to values. Images must
	// be URLs from [PersonalizedMessageService.UploadImages].
	Variables map[string]string `json:"variables"`
}

func (m *PersonalMessage) validate() error {
	if m.MessageSerialNumber == "" || m.ReceiverType == "" || m.ReceiverKey == "" {
		return errors.New("moment: message serial number, receiver type and receiver key are required")
	}
	if len(m.MessageSerialNumber) > 39 {
		return errors.New("moment: message serial number is at most 39 characters")
	}
	if m.Variables == nil {
		m.Variables = map[string]string{}
	}
	return nil
}

// PersonalMessageResult is the result of a personalized message send.
type PersonalMessageResult struct {
	// RequestID identifies the send for [PersonalizedMessageService.Status];
	// for a single send it is the message serial number.
	RequestID           string                `json:"requestId"`
	MessageSerialNumber string                `json:"messageSerialNumber"`
	Status              PersonalMessageStatus `json:"status"`
	// StatusReason details the result in [PersonalizedMessageService.Status].
	StatusReason string     `json:"statusReason"`
	SendAt       kakao.Time `json:"sendAt"`
}

// Send sends a personalized message template to one receiver and waits for
// the result. Kakao allows 10 calls per second.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/personalized-msg
func (s *PersonalizedMessageService) Send(ctx context.Context, adAccountID, creativeID int64, msg PersonalMessage) (*PersonalMessageResult, error) {
	if creativeID == 0 {
		return nil, errors.New("moment: creative ID is required")
	}
	if err := msg.validate(); err != nil {
		return nil, err
	}
	var res PersonalMessageResult
	if err := s.c.do(ctx, call{method: http.MethodPost, path: idPath("/v4/messages/creatives/%d/sendPersonalMessage", creativeID), adAccountID: adAccountID, body: msg}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// PersonalMessageBatch is a bulk send of a personalized message template.
type PersonalMessageBatch struct {
	// RequestID identifies the send, up to 39 characters in the form
	// "yyyyMMdd-{creativeID}-{unique ID}"; use a new one for each send.
	RequestID string `json:"requestId"`
	// Receivers are 2~100 messages.
	Receivers []PersonalMessage `json:"receivers"`
}

// SendBulk sends a personalized message template to 2~100 receivers and
// returns the request ID. Kakao validates the request at once and sends
// asynchronously; check the results with [PersonalizedMessageService.Status].
// Kakao allows 5 calls per second.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/personalized-msg
func (s *PersonalizedMessageService) SendBulk(ctx context.Context, adAccountID, creativeID int64, batch PersonalMessageBatch) (string, error) {
	if creativeID == 0 || batch.RequestID == "" {
		return "", errors.New("moment: creative ID and request ID are required")
	}
	if len(batch.RequestID) > 39 {
		return "", errors.New("moment: request ID is at most 39 characters")
	}
	if n := len(batch.Receivers); n < 2 || n > 100 {
		return "", errors.New("moment: a bulk send needs 2 to 100 receivers")
	}
	receivers := make([]PersonalMessage, len(batch.Receivers))
	for i, m := range batch.Receivers {
		if err := m.validate(); err != nil {
			return "", fmt.Errorf("%w (receiver %d)", err, i)
		}
		receivers[i] = m
	}
	batch.Receivers = receivers
	var res struct {
		RequestID string `json:"requestId"`
	}
	if err := s.c.do(ctx, call{method: http.MethodPost, path: idPath("/v4/messages/creatives/%d/sendPersonalMessages", creativeID), adAccountID: adAccountID, body: batch}, &res); err != nil {
		return "", err
	}
	return res.RequestID, nil
}

// PersonalMessageSendStatus is the state of a personalized message send
// request.
type PersonalMessageSendStatus struct {
	// Completed reports whether every send was attempted; Results is empty
	// until then.
	Completed bool                    `json:"completed"`
	Results   []PersonalMessageResult `json:"results"`
}

// Status returns the results of a send request made in the last 7 days.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/personalized-msg
func (s *PersonalizedMessageService) Status(ctx context.Context, adAccountID, creativeID int64, requestID string) (*PersonalMessageSendStatus, error) {
	if creativeID == 0 || requestID == "" {
		return nil, errors.New("moment: creative ID and request ID are required")
	}
	path := idPath("/v4/messages/creatives/%d/statuses/", creativeID) + url.PathEscape(requestID)
	var res PersonalMessageSendStatus
	if err := s.c.do(ctx, call{method: http.MethodGet, path: path, adAccountID: adAccountID}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// UploadedImage is an image uploaded for personalized messages.
type UploadedImage struct {
	// DownloadURL is the value for image variables. It is not permanent;
	// Kakao may delete the image without notice.
	DownloadURL      string `json:"downloadUrl"`
	OriginalFileName string `json:"originalFileName"`
}

// InvalidReason is why an image upload failed.
type InvalidReason struct {
	Reason      string `json:"reason"`
	Description string `json:"description"`
}

// InvalidFile is an image that failed to upload.
type InvalidFile struct {
	FileName       string          `json:"fileName"`
	InvalidReasons []InvalidReason `json:"invalidReasons"`
}

// ImageUploadResult is the result of [PersonalizedMessageService.UploadImages].
type ImageUploadResult struct {
	SuccessFiles []UploadedImage `json:"successFiles"`
	InvalidFiles []InvalidFile   `json:"invalidFiles"`
}

// UploadImages uploads 1~100 images for personalized message image
// variables: JPG or PNG up to 10 MB, ideally 800x400 (2:1), 800x800 (1:1)
// or 800x600 (4:3). Upload only images to be sent, close to the send time.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomoment/personalized-msg
func (s *PersonalizedMessageService) UploadImages(ctx context.Context, adAccountID int64, files []UploadFile) (*ImageUploadResult, error) {
	if n := len(files); n < 1 || n > 100 {
		return nil, errors.New("moment: upload 1 to 100 images")
	}
	m := &kakao.Multipart{}
	for i := range files {
		p, err := files[i].part("files[" + strconv.Itoa(i) + "]")
		if err != nil {
			return nil, err
		}
		m.Files = append(m.Files, p)
	}
	var res ImageUploadResult
	if err := s.c.do(ctx, call{method: http.MethodPost, path: "/v4/messages/personal/images/upload", adAccountID: adAccountID, multipart: m}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}
