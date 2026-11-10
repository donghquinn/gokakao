package kakao

import (
	"net/http"
	"strings"
	"testing"

	"github.com/donghquinn/gokakao/internal/kakaotest"
)

// newTestClient returns a Client with every host pointed at a fake server.
func newTestClient(t *testing.T, opts ...Option) (*Client, *kakaotest.Server) {
	t.Helper()
	srv := kakaotest.NewServer(t)
	base := []Option{
		WithBaseURL(HostAuth, srv.URL),
		WithBaseURL(HostAPI, srv.URL),
		WithBaseURL(HostDAPI, srv.URL),
	}
	c, err := NewClient(append(base, opts...)...)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c, srv
}

func TestNewClient(t *testing.T) {
	t.Parallel()
	hc := &http.Client{}
	c, err := NewClient(
		WithRESTKey("rest"),
		WithAdminKey("admin"),
		WithClientSecret("secret"),
		WithHTTPClient(hc),
		WithUserAgent("myapp/1.0"),
		WithRetry(RetryPolicy{}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if c.restKey != "rest" || c.adminKey != "admin" || c.clientSecret != "secret" {
		t.Errorf("credentials not set: %+v", c)
	}
	if c.httpClient != hc {
		t.Error("http client not set")
	}
	if want := "myapp/1.0 gokakao/" + Version; c.userAgent != want {
		t.Errorf("userAgent = %q, want %q", c.userAgent, want)
	}
	if c.retry == nil || c.retry.MaxAttempts != 3 {
		t.Errorf("retry defaults not applied: %+v", c.retry)
	}
	if c.baseURLs[HostAPI] != "https://kapi.kakao.com" {
		t.Errorf("default kapi URL = %q", c.baseURLs[HostAPI])
	}
}

func TestNewClient_invalidOptions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		opt  Option
		want string
	}{
		{"nil http client", WithHTTPClient(nil), "nil http.Client"},
		{"unknown host", WithBaseURL(Host(99), "http://x"), "unknown host"},
		{"zero host", WithBaseURL(Host(0), "http://x"), "unknown host"},
		{"empty base URL", WithBaseURL(HostAPI, ""), "empty base URL"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewClient(tt.opt)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("err = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestHost_String(t *testing.T) {
	t.Parallel()
	for h, want := range map[Host]string{HostAuth: "kauth", HostAPI: "kapi", HostDAPI: "dapi", 0: "unknown"} {
		if got := h.String(); got != want {
			t.Errorf("Host(%d).String() = %q, want %q", h, got, want)
		}
	}
}
