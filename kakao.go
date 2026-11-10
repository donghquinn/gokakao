// Package kakao is a client for the Kakao Developers REST APIs
// (https://developers.kakao.com/docs/ko).
//
// A [Client] holds the app credentials (REST API key, admin key, client
// secret) and is safe for concurrent use. Create one per app and share it.
//
//	c, err := kakao.NewClient(
//		kakao.WithRESTKey(os.Getenv("KAKAO_REST_KEY")),
//		kakao.WithAdminKey(os.Getenv("KAKAO_ADMIN_KEY")),
//	)
//
// Errors returned by the Kakao API are reported as [*Error] (kapi/dapi) or
// [*OAuthError] (kauth). Use [errors.Is] with the sentinel values such as
// [ErrInvalidToken] or [ErrRateLimited] to branch on common failures.
package kakao

import (
	"errors"
	"net/http"
	"time"
)

// Version is the library version reported in the User-Agent header.
const Version = "0.0.0-dev"

// Host identifies one of the Kakao API hosts.
type Host int

// Kakao API hosts.
const (
	hostUnknown Host = iota
	// HostAuth is the OAuth server, https://kauth.kakao.com.
	HostAuth
	// HostAPI is the main API server, https://kapi.kakao.com.
	HostAPI
	// HostDAPI serves Daum Search and Kakao Map, https://dapi.kakao.com.
	HostDAPI

	numHosts
)

var defaultBaseURLs = [numHosts]string{
	HostAuth: "https://kauth.kakao.com",
	HostAPI:  "https://kapi.kakao.com",
	HostDAPI: "https://dapi.kakao.com",
}

func (h Host) String() string {
	switch h {
	case HostAuth:
		return "kauth"
	case HostAPI:
		return "kapi"
	case HostDAPI:
		return "dapi"
	default:
		return "unknown"
	}
}

// Client is a Kakao API client. It is safe for concurrent use.
type Client struct {
	httpClient   *http.Client
	baseURLs     [numHosts]string
	restKey      string
	adminKey     string
	clientSecret string
	userAgent    string
	retry        *RetryPolicy
	maxBodySize  int64
	now          func() time.Time

	// Auth implements the Kakao Login OAuth 2.0 and OpenID Connect flows.
	Auth *AuthService
	// User calls the Kakao Login user APIs.
	User *UserService
	// Talk calls the KakaoTalk Message and Social APIs.
	Talk *TalkService
	// Channel calls the KakaoTalk Channel APIs.
	Channel *ChannelService
	// Calendar calls the Talk Calendar API.
	Calendar *CalendarService
	// Push calls the Push Notification API.
	Push *PushService
	// Search calls the Daum Search API.
	Search *SearchService
	// Map calls the Kakao Map REST API.
	Map *MapService
}

// Option configures a [Client].
type Option func(*Client) error

// NewClient returns a Client configured by opts.
func NewClient(opts ...Option) (*Client, error) {
	c := &Client{
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		baseURLs:    defaultBaseURLs,
		userAgent:   "gokakao/" + Version,
		maxBodySize: 16 << 20,
		now:         time.Now,
	}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	c.Auth = &AuthService{c: c}
	c.User = &UserService{c: c}
	c.Talk = &TalkService{c: c}
	c.Channel = &ChannelService{c: c}
	c.Calendar = &CalendarService{c: c}
	c.Push = &PushService{c: c}
	c.Search = &SearchService{c: c}
	c.Map = &MapService{c: c}
	return c, nil
}

// WithRESTKey sets the app's REST API key. It is used as the OAuth client_id
// and as the "KakaoAK" credential for Daum Search, Kakao Map and others.
func WithRESTKey(key string) Option {
	return func(c *Client) error {
		c.restKey = key
		return nil
	}
}

// WithAdminKey sets the app's admin key. Keep it server-side only.
func WithAdminKey(key string) Option {
	return func(c *Client) error {
		c.adminKey = key
		return nil
	}
}

// WithClientSecret sets the OAuth client secret. It is required for token
// requests when client secret is enabled for the app, which is the default
// for new REST API keys.
func WithClientSecret(secret string) Option {
	return func(c *Client) error {
		c.clientSecret = secret
		return nil
	}
}

// WithHTTPClient sets the HTTP client used for requests. The default client
// has a 30 second timeout.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) error {
		if hc == nil {
			return errors.New("kakao: nil http.Client")
		}
		c.httpClient = hc
		return nil
	}
}

// WithBaseURL overrides the base URL of host h, e.g. to point at a test
// server. The URL must not have a trailing slash.
func WithBaseURL(h Host, baseURL string) Option {
	return func(c *Client) error {
		if h <= hostUnknown || h >= numHosts {
			return errors.New("kakao: unknown host")
		}
		if baseURL == "" {
			return errors.New("kakao: empty base URL")
		}
		c.baseURLs[h] = baseURL
		return nil
	}
}

// WithUserAgent prepends ua to the default User-Agent.
func WithUserAgent(ua string) Option {
	return func(c *Client) error {
		c.userAgent = ua + " " + c.userAgent
		return nil
	}
}

// WithRetry enables retries of idempotent requests on transient failures.
// Retries are off by default. Requests with side effects, such as message
// sends, are never retried.
func WithRetry(p RetryPolicy) Option {
	return func(c *Client) error {
		p.setDefaults()
		c.retry = &p
		return nil
	}
}
