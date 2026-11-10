package moment

import (
	"testing"

	kakao "github.com/donghquinn/gokakao"
	"github.com/donghquinn/gokakao/internal/kakaotest"
)

// newTestClient returns a Client whose Kakao Moment host is a fake server.
func newTestClient(t *testing.T) (*Client, *kakaotest.Server) {
	t.Helper()
	srv := kakaotest.NewServer(t)
	kc, err := kakao.NewClient(kakao.WithBaseURL(kakao.HostMoment, srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	return New(kc, "BT"), srv
}

// assertRequest checks the auth and adAccountId headers of the last request.
func assertRequest(t *testing.T, srv *kakaotest.Server, adAccountID string) kakaotest.Request {
	t.Helper()
	req := srv.LastRequest()
	if got := req.Header.Get("Authorization"); got != "Bearer BT" {
		t.Errorf("Authorization = %q", got)
	}
	if got := req.Header.Get("adAccountId"); got != adAccountID {
		t.Errorf("adAccountId = %q, want %q", got, adAccountID)
	}
	return req
}
