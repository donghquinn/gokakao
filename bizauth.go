package kakao

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
)

// BusinessService implements Business Auth (비즈니스 인증), which issues
// business tokens for the Kakao Moment and Keyword Ad APIs. The app must be
// a biz app with permission for business consent items.
//
// Business user info (/v1/business/userinfo) is not implemented; Kakao ends
// it on 2026-11-30.
//
// Docs: https://developers.kakao.com/docs/ko/business-auth/rest-api
type BusinessService struct {
	c *Client
}

// BusinessAuthOptions are the optional parameters of
// [BusinessService.AuthCodeURL].
type BusinessAuthOptions struct {
	// State is returned unchanged to the redirect URI; use it against CSRF.
	State string
	// ResourceIDs pre-select business assets as "ScopeGroup:ResourceID",
	// e.g. "moment:*" for all Kakao Moment ad accounts or
	// "keyword:12345". Required with moment_create or keyword_create,
	// which accept only "*". When empty, the user picks on the consent
	// screen.
	ResourceIDs []string
}

// AuthCodeURL returns the URL of the business consent screen, which is
// shown on every request. scopes are business consent item IDs, e.g.
// "moment_management"; biz_account_email cannot be requested alone.
//
// Docs: https://developers.kakao.com/docs/ko/business-auth/rest-api
func (s *BusinessService) AuthCodeURL(redirectURI string, scopes []string, opts *BusinessAuthOptions) (string, error) {
	if s.c.restKey == "" {
		return "", fmt.Errorf("%w: REST API key (use WithRESTKey)", ErrMissingCredential)
	}
	if len(scopes) == 0 {
		return "", errors.New("kakao: business scopes are required")
	}
	if len(scopes) == 1 && scopes[0] == "biz_account_email" {
		return "", errors.New("kakao: biz_account_email cannot be requested alone")
	}
	var resourceIDs []string
	if opts != nil {
		resourceIDs = opts.ResourceIDs
	}
	if len(resourceIDs) == 0 && (slices.Contains(scopes, "moment_create") || slices.Contains(scopes, "keyword_create")) {
		return "", errors.New("kakao: moment_create and keyword_create need ResourceIDs such as \"moment:*\"")
	}
	q := url.Values{
		"client_id":     {s.c.restKey},
		"response_type": {"code"},
		"redirect_uri":  {redirectURI},
		"scope":         {strings.Join(scopes, ",")},
	}
	if opts != nil {
		setString(q, "state", opts.State)
		for _, id := range opts.ResourceIDs {
			q.Add("resource_ids", id)
		}
	}
	return s.c.baseURLs[HostAuth] + "/oauth/business/authorize?" + q.Encode(), nil
}

// BusinessToken is a business access token. It has no expiry or refresh
// token; it stays valid until revoked.
type BusinessToken struct {
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
	// Scope lists the granted business consent items, space-separated.
	Scope string `json:"scope"`
}

// Exchange trades a business authorization code for a business token.
// Each consent creates a new token, so avoid exchanging codes for the same
// consent repeatedly.
//
// Docs: https://developers.kakao.com/docs/ko/business-auth/rest-api
func (s *BusinessService) Exchange(ctx context.Context, code, redirectURI string) (*BusinessToken, error) {
	if s.c.restKey == "" {
		return nil, fmt.Errorf("%w: REST API key (use WithRESTKey)", ErrMissingCredential)
	}
	if code == "" {
		return nil, errors.New("kakao: empty authorization code")
	}
	form := url.Values{
		"grant_type":   {"authorization_code"},
		"client_id":    {s.c.restKey},
		"code":         {code},
		"redirect_uri": {redirectURI},
	}
	setString(form, "client_secret", s.c.clientSecret)
	var t BusinessToken
	err := s.c.do(ctx, &request{method: http.MethodPost, host: HostAuth, path: "/oauth/business/token", form: form}, &t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// BusinessTokenInfo describes a business token.
type BusinessTokenInfo struct {
	// ID is the business token ID.
	ID int64 `json:"id"`
	// TokenUserID is the user's business auth user ID.
	TokenUserID   string         `json:"token_user_id"`
	BizAgreements []BizAgreement `json:"biz_agreements"`
	// Status is "ACTIVE".
	Status    string `json:"status"`
	UpdatedAt Time   `json:"updated_at"`
}

// BizAgreement is the user's consent within one business consent group.
type BizAgreement struct {
	// Group is "personal", "moment" or "keyword".
	Group string   `json:"group"`
	Scope []string `json:"scope"`
	// AdAccountIDs are the accessible ad accounts; "*" means all of the
	// user's accounts in the group.
	AdAccountIDs     []string `json:"ad_account_ids"`
	ChannelPublicIDs []string `json:"channel_public_ids"`
}

// TokenInfo returns information about a business token, verifying that it
// is valid.
//
// Docs: https://developers.kakao.com/docs/ko/business-auth/rest-api
func (s *BusinessService) TokenInfo(ctx context.Context, businessToken string) (*BusinessTokenInfo, error) {
	var info BusinessTokenInfo
	err := s.c.doAs(ctx, AccessToken(businessToken), &request{method: http.MethodGet, host: HostAPI, path: "/v1/business/tokeninfo"}, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// RevokedBusinessToken identifies a revoked business token.
type RevokedBusinessToken struct {
	ID          int64  `json:"id"`
	TokenUserID string `json:"token_user_id"`
}

// Revoke expires a business token.
//
// Docs: https://developers.kakao.com/docs/ko/business-auth/rest-api
func (s *BusinessService) Revoke(ctx context.Context, businessToken string) (*RevokedBusinessToken, error) {
	var res RevokedBusinessToken
	err := s.c.doAs(ctx, AccessToken(businessToken), &request{method: http.MethodPost, host: HostAPI, path: "/v1/business/revoke"}, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
