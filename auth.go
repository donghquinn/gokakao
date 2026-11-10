package kakao

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// AuthService implements the Kakao Login OAuth 2.0 and OpenID Connect
// flows on kauth.kakao.com. It uses the REST API key as the client ID and
// sends the client secret when one is configured.
//
// Docs: https://developers.kakao.com/docs/ko/kakaologin/rest-api
type AuthService struct {
	c    *Client
	jwks jwksCache
}

// Prompt changes how the authorization page interacts with the user.
type Prompt string

// Prompts.
const (
	// PromptLogin asks for credentials even when the user is logged in.
	PromptLogin Prompt = "login"
	// PromptNone fails with error=consent_required instead of showing UI.
	PromptNone Prompt = "none"
	// PromptCreate shows the Kakao Account sign-up page first.
	PromptCreate Prompt = "create"
	// PromptSelectAccount lets the user choose among logged-in accounts.
	PromptSelectAccount Prompt = "select_account"
)

// AuthCodeOptions are the optional parameters of [AuthService.AuthCodeURL].
type AuthCodeOptions struct {
	// Scopes requests additional consent items. Include "openid" to
	// receive an ID token when requesting extra scopes.
	Scopes []string
	// Prompts are sent comma-separated.
	Prompts []Prompt
	// LoginHint pre-fills the account ID on the login page.
	LoginHint string
	// ServiceTerms lists the service terms tags to ask consent for.
	// Kakao Sync only.
	ServiceTerms []string
	// State is returned unchanged to the redirect URI. Use it to prevent
	// CSRF; generate it per request.
	State string
	// Nonce is copied into the ID token. Check it with
	// [AuthService.VerifyIDToken].
	Nonce string
	// CodeChallenge is a PKCE S256 challenge, see [S256Challenge].
	// Pass the matching verifier to [AuthService.Exchange].
	CodeChallenge string
}

// AuthCodeURL returns the URL to send the user to for login and consent.
// After login Kakao redirects to redirectURI with a code (or error) and
// the state.
//
// Docs: https://developers.kakao.com/docs/ko/kakaologin/rest-api
func (s *AuthService) AuthCodeURL(redirectURI string, opts *AuthCodeOptions) (string, error) {
	if s.c.restKey == "" {
		return "", fmt.Errorf("%w: REST API key (use WithRESTKey)", ErrMissingCredential)
	}
	q := url.Values{
		"client_id":     {s.c.restKey},
		"redirect_uri":  {redirectURI},
		"response_type": {"code"},
	}
	if opts != nil {
		setString(q, "scope", strings.Join(opts.Scopes, ","))
		prompts := make([]string, len(opts.Prompts))
		for i, p := range opts.Prompts {
			prompts[i] = string(p)
		}
		setString(q, "prompt", strings.Join(prompts, ","))
		setString(q, "login_hint", opts.LoginHint)
		setString(q, "service_terms", strings.Join(opts.ServiceTerms, ","))
		setString(q, "state", opts.State)
		setString(q, "nonce", opts.Nonce)
		if opts.CodeChallenge != "" {
			q.Set("code_challenge", opts.CodeChallenge)
			q.Set("code_challenge_method", "S256")
		}
	}
	return s.c.baseURLs[HostAuth] + "/oauth/authorize?" + q.Encode(), nil
}

// LogoutURL returns the URL that logs the user out of their Kakao Account
// and then redirects to logoutRedirectURI with state. The redirect URI must
// be registered in the app settings.
//
// Docs: https://developers.kakao.com/docs/ko/kakaologin/rest-api
func (s *AuthService) LogoutURL(logoutRedirectURI, state string) (string, error) {
	if s.c.restKey == "" {
		return "", fmt.Errorf("%w: REST API key (use WithRESTKey)", ErrMissingCredential)
	}
	q := url.Values{"client_id": {s.c.restKey}, "logout_redirect_uri": {logoutRedirectURI}}
	setString(q, "state", state)
	return s.c.baseURLs[HostAuth] + "/oauth/logout?" + q.Encode(), nil
}

// Token is a Kakao Login token set.
type Token struct {
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
	// IDToken is set when OpenID Connect is enabled. Verify it with
	// [AuthService.VerifyIDToken] before trusting its claims.
	IDToken string `json:"id_token,omitempty"`
	// ExpiresIn is the access token lifetime in seconds.
	ExpiresIn int `json:"expires_in"`
	// RefreshToken is always returned by Exchange. Refresh returns a new
	// one only when the old one expires within a month.
	RefreshToken          string `json:"refresh_token,omitempty"`
	RefreshTokenExpiresIn int    `json:"refresh_token_expires_in,omitempty"`
	// Scope lists the granted scopes, space-separated.
	Scope string `json:"scope,omitempty"`

	// Expiry is when the access token expires, computed on receipt.
	// A zero Expiry means unknown; the token is then assumed valid.
	Expiry time.Time `json:"expiry,omitzero"`
	// RefreshTokenExpiry is when the refresh token expires, computed on
	// receipt.
	RefreshTokenExpiry time.Time `json:"refresh_token_expiry,omitzero"`
}

// ExchangeOptions are the optional parameters of [AuthService.Exchange].
type ExchangeOptions struct {
	// CodeVerifier is the PKCE verifier whose challenge was sent with
	// [AuthCodeOptions.CodeChallenge].
	CodeVerifier string
}

// Exchange trades an authorization code for tokens. redirectURI must equal
// the one used in [AuthService.AuthCodeURL].
//
// Docs: https://developers.kakao.com/docs/ko/kakaologin/rest-api
func (s *AuthService) Exchange(ctx context.Context, code, redirectURI string, opts *ExchangeOptions) (*Token, error) {
	if code == "" {
		return nil, errors.New("kakao: empty authorization code")
	}
	form := url.Values{
		"grant_type":   {"authorization_code"},
		"redirect_uri": {redirectURI},
		"code":         {code},
	}
	if opts != nil {
		setString(form, "code_verifier", opts.CodeVerifier)
	}
	return s.token(ctx, form)
}

// Refresh obtains a new access token. The returned Token carries the new
// refresh token only when Kakao rotated it; otherwise RefreshToken is empty
// and the old refresh token stays valid. [TokenSource] handles this.
//
// Docs: https://developers.kakao.com/docs/ko/kakaologin/rest-api
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*Token, error) {
	if refreshToken == "" {
		return nil, errors.New("kakao: empty refresh token")
	}
	return s.token(ctx, url.Values{"grant_type": {"refresh_token"}, "refresh_token": {refreshToken}})
}

func (s *AuthService) token(ctx context.Context, form url.Values) (*Token, error) {
	if s.c.restKey == "" {
		return nil, fmt.Errorf("%w: REST API key (use WithRESTKey)", ErrMissingCredential)
	}
	form.Set("client_id", s.c.restKey)
	setString(form, "client_secret", s.c.clientSecret)
	var t Token
	err := s.c.do(ctx, &request{method: http.MethodPost, host: HostAuth, path: "/oauth/token", form: form}, &t)
	if err != nil {
		return nil, err
	}
	now := s.c.now()
	if t.ExpiresIn > 0 {
		t.Expiry = now.Add(time.Duration(t.ExpiresIn) * time.Second)
	}
	if t.RefreshTokenExpiresIn > 0 {
		t.RefreshTokenExpiry = now.Add(time.Duration(t.RefreshTokenExpiresIn) * time.Second)
	}
	return &t, nil
}

// GenerateVerifier returns a random PKCE code verifier.
func GenerateVerifier() string {
	return rand.Text() + rand.Text()
}

// S256Challenge returns the PKCE S256 code challenge for verifier.
func S256Challenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// tokenRefreshLeeway is how long before expiry a TokenSource refreshes.
const tokenRefreshLeeway = time.Minute

// TokenSource returns valid access tokens for one user, refreshing them
// when they are about to expire. It is safe for concurrent use.
type TokenSource struct {
	auth      *AuthService
	onRefresh func(*Token)

	mu  sync.Mutex
	tok *Token
}

// TokenSourceOptions are the optional parameters of
// [AuthService.TokenSource].
type TokenSourceOptions struct {
	// OnRefresh is called with every refreshed token so it can be
	// persisted. It runs while the TokenSource is locked and must not call
	// back into it.
	OnRefresh func(*Token)
}

// TokenSource returns a TokenSource that starts from tok.
func (s *AuthService) TokenSource(tok *Token, opts *TokenSourceOptions) *TokenSource {
	t := *tok
	ts := &TokenSource{auth: s, tok: &t}
	if opts != nil {
		ts.onRefresh = opts.OnRefresh
	}
	return ts
}

// Token returns a valid token, refreshing it first if it expires within a
// minute. The returned Token is a copy.
func (ts *TokenSource) Token(ctx context.Context) (*Token, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	now := ts.auth.c.now()
	if ts.tok.AccessToken != "" && (ts.tok.Expiry.IsZero() || now.Add(tokenRefreshLeeway).Before(ts.tok.Expiry)) {
		t := *ts.tok
		return &t, nil
	}
	fresh, err := ts.auth.Refresh(ctx, ts.tok.RefreshToken)
	if err != nil {
		return nil, err
	}
	if fresh.RefreshToken == "" {
		fresh.RefreshToken = ts.tok.RefreshToken
		fresh.RefreshTokenExpiresIn = ts.tok.RefreshTokenExpiresIn
		fresh.RefreshTokenExpiry = ts.tok.RefreshTokenExpiry
	}
	ts.tok = fresh
	if ts.onRefresh != nil {
		ts.onRefresh(fresh)
	}
	t := *fresh
	return &t, nil
}
