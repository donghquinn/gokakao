package moment

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestPersonalizedMessageService_SendTest(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v4/messages/creatives/1/sendTestPersonalMessage", 200, ``)

	err := c.PersonalizedMessages.SendTest(t.Context(), 1, 1, TestPersonalMessage{
		PhoneNumber: "010-1234-1234",
		Variables: map[string]string{
			"image_url1": "https://t1.kakaocdn.net/b2/personalMessage/1/testImage.jpg",
			"user_name1": "테스트유저",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1")
	creativeAssertJSON(t, req, `{
  "phoneNumber": "010-1234-1234",
  "variables": {"image_url1": "https://t1.kakaocdn.net/b2/personalMessage/1/testImage.jpg", "user_name1": "테스트유저"}
}`)

	if err := c.PersonalizedMessages.SendTest(t.Context(), 1, 1, TestPersonalMessage{PhoneNumber: "010"}); err != nil {
		t.Fatal(err)
	}
	creativeAssertJSON(t, srv.LastRequest(), `{"phoneNumber": "010", "variables": {}}`)

	if err := c.PersonalizedMessages.SendTest(t.Context(), 1, 1, TestPersonalMessage{}); err == nil {
		t.Error("missing phone: want error")
	}
	if err := c.PersonalizedMessages.SendTest(t.Context(), 1, 0, TestPersonalMessage{PhoneNumber: "1"}); err == nil {
		t.Error("missing creative: want error")
	}
}

func TestPersonalizedMessageService_Send(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v4/messages/creatives/1/sendPersonalMessage", 200, `{
  "requestId": "20230731-1-001",
  "messageSerialNumber": "20230731-1-send1",
  "status": "SUCCEEDED",
  "sendAt": "2023-07-31 16:00:00"
}`)

	res, err := c.PersonalizedMessages.Send(t.Context(), 1, 1, PersonalMessage{
		MessageSerialNumber: "20230731-1-send1",
		ReceiverType:        ReceiverPhoneNumber,
		ReceiverKey:         "010-1234-1234",
		Variables: map[string]string{
			"image_url1": "https://t1.kakaocdn.net/b2/personalMessage/1/testImage.jpg",
			"user_name1": "테스트유저",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1")
	creativeAssertJSON(t, req, `{
  "messageSerialNumber": "20230731-1-send1",
  "receiverType": "PHONE_NUMBER",
  "receiverKey": "010-1234-1234",
  "variables": {"image_url1": "https://t1.kakaocdn.net/b2/personalMessage/1/testImage.jpg", "user_name1": "테스트유저"}
}`)
	if res.RequestID != "20230731-1-001" || res.MessageSerialNumber != "20230731-1-send1" || res.Status != PersonalMessageSucceeded ||
		!res.SendAt.Equal(time.Date(2023, 7, 31, 16, 0, 0, 0, creativeKST)) {
		t.Errorf("result = %+v", res)
	}

	for name, m := range map[string]PersonalMessage{
		"empty":       {},
		"long serial": {MessageSerialNumber: fmt.Sprintf("%040d", 0), ReceiverType: ReceiverAppUserID, ReceiverKey: "1"},
	} {
		if _, err := c.PersonalizedMessages.Send(t.Context(), 1, 1, m); err == nil {
			t.Errorf("%s: want error", name)
		}
	}
	if _, err := c.PersonalizedMessages.Send(t.Context(), 1, 0, PersonalMessage{}); err == nil {
		t.Error("missing creative: want error")
	}
}

func TestPersonalizedMessageService_SendBulk(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v4/messages/creatives/1/sendPersonalMessages", 200, `{"requestId": "20230731-1-request1"}`)

	id, err := c.PersonalizedMessages.SendBulk(t.Context(), 1, 1, PersonalMessageBatch{
		RequestID: "20230731-1-request1",
		Receivers: []PersonalMessage{
			{
				MessageSerialNumber: "20230731-1-send1", ReceiverType: ReceiverPhoneNumber, ReceiverKey: "010-1234-1234",
				Variables: map[string]string{"image_url1": "https://t1.kakaocdn.net/b2/personalMessage/1/testImage1.jpg", "user_name1": "테스트유저1"},
			},
			{
				MessageSerialNumber: "20230731-1-send2", ReceiverType: ReceiverPhoneNumber, ReceiverKey: "010-4321-4321",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1")
	creativeAssertJSON(t, req, `{
  "requestId": "20230731-1-request1",
  "receivers": [
    {"messageSerialNumber": "20230731-1-send1", "receiverType": "PHONE_NUMBER", "receiverKey": "010-1234-1234",
     "variables": {"image_url1": "https://t1.kakaocdn.net/b2/personalMessage/1/testImage1.jpg", "user_name1": "테스트유저1"}},
    {"messageSerialNumber": "20230731-1-send2", "receiverType": "PHONE_NUMBER", "receiverKey": "010-4321-4321", "variables": {}}
  ]
}`)
	if id != "20230731-1-request1" {
		t.Errorf("request ID = %q", id)
	}

	ok := PersonalMessage{MessageSerialNumber: "s", ReceiverType: ReceiverAppUserID, ReceiverKey: "1"}
	for name, b := range map[string]PersonalMessageBatch{
		"no request ID":   {Receivers: []PersonalMessage{ok, ok}},
		"long request ID": {RequestID: fmt.Sprintf("%040d", 0), Receivers: []PersonalMessage{ok, ok}},
		"one receiver":    {RequestID: "r", Receivers: []PersonalMessage{ok}},
		"101 receivers":   {RequestID: "r", Receivers: make([]PersonalMessage, 101)},
		"bad receiver":    {RequestID: "r", Receivers: []PersonalMessage{ok, {}}},
	} {
		if _, err := c.PersonalizedMessages.SendBulk(t.Context(), 1, 1, b); err == nil {
			t.Errorf("%s: want error", name)
		}
	}
}

func TestPersonalizedMessageService_Status(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v4/messages/creatives/1/statuses/{requestID}", 200, `{
  "completed": true,
  "results": [
    {"status": "SUCCEEDED", "sendAt": "2023-07-31 16:00:00"}
  ]
}`)

	st, err := c.PersonalizedMessages.Status(t.Context(), 7, 1, "20230731-1-request1")
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "7")
	if req.Method != http.MethodGet || req.Path != "/openapi/v4/messages/creatives/1/statuses/20230731-1-request1" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	if !st.Completed || len(st.Results) != 1 || st.Results[0].Status != PersonalMessageSucceeded || st.Results[0].SendAt.Hour() != 16 {
		t.Errorf("status = %+v", st)
	}
	if _, err := c.PersonalizedMessages.Status(t.Context(), 7, 1, ""); err == nil {
		t.Error("missing request ID: want error")
	}
}

func TestPersonalizedMessageService_UploadImages(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /openapi/v4/messages/personal/images/upload", 200, `{
  "successFiles": [
    {"originalFileName": "testImage.jpg", "downloadUrl": "https://t1.kakaocdn.net/b2/personalMessage/1/testImage.jpg"}
  ],
  "invalidFiles": [
    {"fileName": "string", "invalidReasons": [{"reason": "string", "description": "string"}]}
  ]
}`)

	res, err := c.PersonalizedMessages.UploadImages(t.Context(), 5, []UploadFile{
		{Filename: "image.jpg", ContentType: "image/jpeg", Data: []byte("jpg")},
		{Filename: "b.png", Data: []byte("\x89PNG\r\n\x1a\n")},
	})
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "5")
	fields, files := creativeParseMultipart(t, req)
	if len(fields) != 0 || len(files) != 2 || files["files[0]"].Filename != "image.jpg" || files["files[0]"].Data != "jpg" ||
		files["files[1]"].ContentType != "image/png" {
		t.Errorf("fields = %v, files = %v", fields, files)
	}
	if len(res.SuccessFiles) != 1 || res.SuccessFiles[0].DownloadURL != "https://t1.kakaocdn.net/b2/personalMessage/1/testImage.jpg" ||
		res.SuccessFiles[0].OriginalFileName != "testImage.jpg" || len(res.InvalidFiles) != 1 ||
		res.InvalidFiles[0].FileName != "string" || res.InvalidFiles[0].InvalidReasons[0].Description != "string" {
		t.Errorf("result = %+v", res)
	}

	if _, err := c.PersonalizedMessages.UploadImages(t.Context(), 5, nil); err == nil {
		t.Error("no files: want error")
	}
	if _, err := c.PersonalizedMessages.UploadImages(t.Context(), 5, make([]UploadFile, 101)); err == nil {
		t.Error("101 files: want error")
	}
	if _, err := c.PersonalizedMessages.UploadImages(t.Context(), 5, []UploadFile{{Filename: "a"}}); err == nil {
		t.Error("empty file: want error")
	}
}
