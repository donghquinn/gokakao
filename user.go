package kakao

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// UserService calls the Kakao Login user APIs on kapi.kakao.com.
//
// Docs: https://developers.kakao.com/docs/ko/kakaologin/rest-api
type UserService struct {
	c *Client
}

// Subject selects whose account a user API call acts on, and how it
// authenticates. Use [AccessToken] to act as the logged-in user or
// [AsAdmin] to act on any linked user with the admin key.
type Subject interface {
	apply(r *request) error
}

// AccessToken returns a Subject that authenticates with the user's access
// token.
func AccessToken(token string) Subject {
	return accessTokenSubject(token)
}

// AsAdmin returns a Subject that authenticates with the admin key and
// targets the user with the given user ID.
func AsAdmin(userID int64) Subject {
	return adminSubject(userID)
}

type accessTokenSubject string

func (s accessTokenSubject) apply(r *request) error {
	if s == "" {
		return fmt.Errorf("%w: access token", ErrMissingCredential)
	}
	r.auth, r.token = authBearer, string(s)
	return nil
}

type adminSubject int64

func (s adminSubject) apply(r *request) error {
	r.auth = authAdminKey
	params := r.query
	if r.method != http.MethodGet {
		if r.form == nil {
			r.form = url.Values{}
		}
		params = r.form
	} else if params == nil {
		r.query = url.Values{}
		params = r.query
	}
	params.Set("target_id_type", "user_id")
	params.Set("target_id", strconv.FormatInt(int64(s), 10))
	return nil
}

// do sends r on behalf of sub.
func (s *UserService) do(ctx context.Context, sub Subject, r *request, out any) error {
	if sub == nil {
		return errors.New("kakao: nil Subject")
	}
	if err := sub.apply(r); err != nil {
		return err
	}
	return s.c.do(ctx, r, out)
}

// User is a Kakao user as returned by [UserService.Me] and
// [UserService.Users].
type User struct {
	// ID is the user ID (회원번호).
	ID int64 `json:"id"`
	// HasSignedUp reports whether the user is linked to the app. It is only
	// returned for apps that link users manually; see [UserService.Signup].
	HasSignedUp bool `json:"has_signed_up"`
	// ConnectedAt is when the user was linked to the app.
	ConnectedAt Time `json:"connected_at"`
	// SynchedAt is when the user was last synced by Kakao Sync.
	SynchedAt Time `json:"synched_at"`
	// GroupUserToken identifies the user across the apps of a group.
	GroupUserToken string `json:"group_user_token"`
	// Properties are the app's user properties.
	Properties map[string]string `json:"properties"`
	// KakaoAccount is nil when no account fields were returned.
	KakaoAccount *Account `json:"kakao_account"`
	// ForPartner holds partner-only fields.
	ForPartner *Partner `json:"for_partner"`
}

// Account is the user's Kakao Account information. A field is empty
// unless the user consented to it; the matching NeedsAgreement field
// reports whether consent can be requested.
type Account struct {
	ProfileNeedsAgreement         bool     `json:"profile_needs_agreement"`
	ProfileNicknameNeedsAgreement bool     `json:"profile_nickname_needs_agreement"`
	ProfileImageNeedsAgreement    bool     `json:"profile_image_needs_agreement"`
	Profile                       *Profile `json:"profile"`

	NameNeedsAgreement bool   `json:"name_needs_agreement"`
	Name               string `json:"name"`

	EmailNeedsAgreement bool `json:"email_needs_agreement"`
	IsEmailValid        bool `json:"is_email_valid"`
	IsEmailVerified     bool `json:"is_email_verified"`
	// Email is the representative email. Check IsEmailValid and
	// IsEmailVerified before relying on it.
	Email string `json:"email"`

	AgeRangeNeedsAgreement bool `json:"age_range_needs_agreement"`
	// AgeRange is e.g. "20~29".
	AgeRange string `json:"age_range"`

	BirthyearNeedsAgreement bool `json:"birthyear_needs_agreement"`
	// Birthyear is "YYYY".
	Birthyear string `json:"birthyear"`

	BirthdayNeedsAgreement bool `json:"birthday_needs_agreement"`
	// Birthday is "MMDD".
	Birthday string `json:"birthday"`
	// BirthdayType is "SOLAR" or "LUNAR".
	BirthdayType string `json:"birthday_type"`
	IsLeapMonth  bool   `json:"is_leap_month"`

	GenderNeedsAgreement bool `json:"gender_needs_agreement"`
	// Gender is "female" or "male".
	Gender string `json:"gender"`

	PhoneNumberNeedsAgreement bool `json:"phone_number_needs_agreement"`
	// PhoneNumber is e.g. "+82 10-1234-5678".
	PhoneNumber string `json:"phone_number"`

	CINeedsAgreement  bool   `json:"ci_needs_agreement"`
	CI                string `json:"ci"`
	CIAuthenticatedAt Time   `json:"ci_authenticated_at"`
}

// Profile is the user's Kakao Account profile.
type Profile struct {
	Nickname          string `json:"nickname"`
	ThumbnailImageURL string `json:"thumbnail_image_url"`
	ProfileImageURL   string `json:"profile_image_url"`
	IsDefaultImage    bool   `json:"is_default_image"`
	IsDefaultNickname bool   `json:"is_default_nickname"`
}

// Partner holds partner-only user fields.
type Partner struct {
	UUID string `json:"uuid"`
}

// MeOptions are the optional parameters of [UserService.Me].
type MeOptions struct {
	// PropertyKeys limits the response to these fields, e.g.
	// "kakao_account.email" or "properties." for all properties.
	PropertyKeys []string
	// SecureResource returns image URLs over HTTPS.
	SecureResource bool
}

// Me returns the user's information.
//
// Docs: https://developers.kakao.com/docs/ko/kakaologin/rest-api
func (s *UserService) Me(ctx context.Context, sub Subject, opts *MeOptions) (*User, error) {
	q := url.Values{}
	if opts != nil {
		if err := setJSON(q, "property_keys", opts.PropertyKeys); err != nil {
			return nil, err
		}
		if opts.SecureResource {
			q.Set("secure_resource", "true")
		}
	}
	var u User
	if err := s.do(ctx, sub, &request{method: http.MethodGet, host: HostAPI, path: "/v2/user/me", query: q}, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

// UsersOptions are the optional parameters of [UserService.Users].
type UsersOptions struct {
	// PropertyKeys limits the response fields. With PropertyKeys at most
	// 20 users can be requested at once, otherwise 100.
	PropertyKeys []string
}

// Users returns the information of several users. It uses the admin key.
//
// Docs: https://developers.kakao.com/docs/ko/kakaologin/rest-api
func (s *UserService) Users(ctx context.Context, userIDs []int64, opts *UsersOptions) ([]User, error) {
	if len(userIDs) == 0 {
		return nil, errors.New("kakao: no user IDs")
	}
	q := url.Values{"target_id_type": {"user_id"}}
	if err := setJSON(q, "target_ids", userIDs); err != nil {
		return nil, err
	}
	if opts != nil {
		if err := setJSON(q, "property_keys", opts.PropertyKeys); err != nil {
			return nil, err
		}
	}
	var users []User
	err := s.c.do(ctx, &request{method: http.MethodGet, host: HostAPI, path: "/v2/app/users", auth: authAdminKey, query: q}, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// UserIDsOptions are the optional parameters of [UserService.UserIDs].
type UserIDsOptions struct {
	// Limit is the page size, 1-100 (default 100).
	Limit int
	// FromID starts the page after this user ID.
	FromID int64
	// Descending lists the newest users first.
	Descending bool
}

// UserIDPage is a page of user IDs.
type UserIDPage struct {
	Elements  []int64 `json:"elements"`
	BeforeURL string  `json:"before_url"`
	// AfterURL is empty on the last page.
	AfterURL string `json:"after_url"`
}

// UserIDs returns a page of the IDs of the users linked to the app. It uses
// the admin key. Kakao allows 100 calls per minute per app.
//
// Docs: https://developers.kakao.com/docs/ko/kakaologin/rest-api
func (s *UserService) UserIDs(ctx context.Context, opts *UserIDsOptions) (*UserIDPage, error) {
	q := url.Values{}
	if opts != nil {
		setInt(q, "limit", opts.Limit)
		if opts.FromID != 0 {
			q.Set("from_id", strconv.FormatInt(opts.FromID, 10))
		}
		if opts.Descending {
			q.Set("order", "desc")
		}
	}
	var page UserIDPage
	err := s.c.do(ctx, &request{method: http.MethodGet, host: HostAPI, path: "/v1/user/ids", auth: authAdminKey, query: q}, &page)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// AllUserIDs iterates over the IDs of all users linked to the app, fetching
// pages as needed. Iteration stops at the first error, which is yielded.
func (s *UserService) AllUserIDs(ctx context.Context, opts *UserIDsOptions) iter.Seq2[int64, error] {
	return func(yield func(int64, error) bool) {
		o := UserIDsOptions{}
		if opts != nil {
			o = *opts
		}
		for {
			page, err := s.UserIDs(ctx, &o)
			if err != nil {
				yield(0, err)
				return
			}
			for _, id := range page.Elements {
				if !yield(id, nil) {
					return
				}
			}
			next, err := nextFromID(page.AfterURL)
			if err != nil {
				yield(0, err)
				return
			}
			if next == 0 || len(page.Elements) == 0 {
				return
			}
			o.FromID = next
		}
	}
}

// nextFromID extracts from_id from a page's after_url. It returns 0 when
// there is no next page.
func nextFromID(afterURL string) (int64, error) {
	if afterURL == "" {
		return 0, nil
	}
	u, err := url.Parse(afterURL)
	if err != nil {
		return 0, fmt.Errorf("kakao: parse after_url: %w", err)
	}
	v := u.Query().Get("from_id")
	if v == "" {
		return 0, nil
	}
	id, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("kakao: parse after_url from_id: %w", err)
	}
	return id, nil
}

type idResponse struct {
	ID int64 `json:"id"`
}

// Logout expires the user's tokens. It returns the user ID.
//
// Docs: https://developers.kakao.com/docs/ko/kakaologin/rest-api
func (s *UserService) Logout(ctx context.Context, sub Subject) (int64, error) {
	var res idResponse
	err := s.do(ctx, sub, &request{method: http.MethodPost, host: HostAPI, path: "/v1/user/logout"}, &res)
	return res.ID, err
}

// Unlink unlinks the user from the app, withdrawing their consent and
// expiring their tokens. It returns the user ID.
//
// Docs: https://developers.kakao.com/docs/ko/kakaologin/rest-api
func (s *UserService) Unlink(ctx context.Context, sub Subject) (int64, error) {
	var res idResponse
	err := s.do(ctx, sub, &request{method: http.MethodPost, host: HostAPI, path: "/v1/user/unlink"}, &res)
	return res.ID, err
}

// AccessTokenInfo describes an access token.
type AccessTokenInfo struct {
	// ID is the user ID.
	ID int64 `json:"id"`
	// ExpiresIn is the remaining lifetime in seconds.
	ExpiresIn int   `json:"expires_in"`
	AppID     int64 `json:"app_id"`
}

// TokenInfo returns information about an access token. Use it to check
// that a token is valid and belongs to the app.
//
// Docs: https://developers.kakao.com/docs/ko/kakaologin/rest-api
func (s *UserService) TokenInfo(ctx context.Context, accessToken string) (*AccessTokenInfo, error) {
	var info AccessTokenInfo
	err := s.do(ctx, AccessToken(accessToken), &request{method: http.MethodGet, host: HostAPI, path: "/v1/user/access_token_info"}, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// UpdateProfile saves user properties (up to 255 characters each). The
// keys must be registered in the app settings. It returns the user ID.
//
// Docs: https://developers.kakao.com/docs/ko/kakaologin/rest-api
func (s *UserService) UpdateProfile(ctx context.Context, accessToken string, properties map[string]string) (int64, error) {
	if len(properties) == 0 {
		return 0, errors.New("kakao: no properties")
	}
	return s.postProperties(ctx, accessToken, "/v1/user/update_profile", properties)
}

// Signup links the user to an app that does not link users automatically,
// optionally saving user properties. It returns the user ID.
//
// Docs: https://developers.kakao.com/docs/ko/kakaologin/rest-api
func (s *UserService) Signup(ctx context.Context, accessToken string, properties map[string]string) (int64, error) {
	return s.postProperties(ctx, accessToken, "/v1/user/signup", properties)
}

func (s *UserService) postProperties(ctx context.Context, accessToken, path string, properties map[string]string) (int64, error) {
	form := url.Values{}
	if len(properties) > 0 {
		if err := setJSON(form, "properties", properties); err != nil {
			return 0, err
		}
	}
	var res idResponse
	err := s.do(ctx, AccessToken(accessToken), &request{method: http.MethodPost, host: HostAPI, path: path, form: form}, &res)
	return res.ID, err
}

// ShippingAddressResult is the response of [UserService.ShippingAddress].
type ShippingAddressResult struct {
	UserID int64 `json:"user_id"`
	// NeedsAgreement reports that the user must consent to the
	// shipping_address item before the address can be read.
	NeedsAgreement bool `json:"shipping_address_needs_agreement"`
	// Address is nil when the user has no address or has not consented.
	Address *ShippingAddress `json:"shipping_address"`
}

// ShippingAddress is a user's shipping address.
type ShippingAddress struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
	// UpdatedAt is in Unix seconds.
	UpdatedAt int64 `json:"updated_at"`
	// Type is "OLD" (lot-number) or "NEW" (road-name).
	Type                 string `json:"type"`
	BaseAddress          string `json:"base_address"`
	DetailAddress        string `json:"detail_address"`
	ReceiverName         string `json:"receiver_name"`
	ReceiverPhoneNumber1 string `json:"receiver_phone_number1"`
	ReceiverPhoneNumber2 string `json:"receiver_phone_number2"`
	ZoneNumber           string `json:"zone_number"`
	ZipCode              string `json:"zip_code"`
}

// ShippingAddress returns one of the user's shipping addresses: the one
// with addressID, or the default address when addressID is 0.
//
// Docs: https://developers.kakao.com/docs/ko/kakaologin/rest-api
func (s *UserService) ShippingAddress(ctx context.Context, sub Subject, addressID int64) (*ShippingAddressResult, error) {
	q := url.Values{}
	if addressID != 0 {
		q.Set("address_id", strconv.FormatInt(addressID, 10))
	}
	var res ShippingAddressResult
	if err := s.do(ctx, sub, &request{method: http.MethodGet, host: HostAPI, path: "/v2/user/shipping_address", query: q}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Scopes is the consent status of a user's consent items.
type Scopes struct {
	// ID is the user ID.
	ID     int64   `json:"id"`
	Scopes []Scope `json:"scopes"`
}

// Scope is a consent item.
type Scope struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	// Type is "PRIVACY" (personal information) or "SERVICE" (access
	// permission).
	Type string `json:"type"`
	// Using reports whether the app currently uses the item.
	Using  bool `json:"using"`
	Agreed bool `json:"agreed"`
	// Revocable reports whether consent can be withdrawn with
	// [UserService.RevokeScopes]. Only set for agreed items.
	Revocable bool `json:"revocable"`
}

// Scopes returns the user's consent status for the app's consent items,
// or only for scopeIDs when given.
//
// Docs: https://developers.kakao.com/docs/ko/kakaologin/rest-api
func (s *UserService) Scopes(ctx context.Context, sub Subject, scopeIDs ...string) (*Scopes, error) {
	q := url.Values{}
	if err := setJSON(q, "scopes", scopeIDs); err != nil {
		return nil, err
	}
	var res Scopes
	if err := s.do(ctx, sub, &request{method: http.MethodGet, host: HostAPI, path: "/v2/user/scopes", query: q}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// RevokeScopes withdraws the user's consent to scopeIDs, which must be
// revocable. It returns the updated consent status.
//
// Docs: https://developers.kakao.com/docs/ko/kakaologin/rest-api
func (s *UserService) RevokeScopes(ctx context.Context, sub Subject, scopeIDs ...string) (*Scopes, error) {
	if len(scopeIDs) == 0 {
		return nil, errors.New("kakao: no scopes to revoke")
	}
	form := url.Values{}
	if err := setJSON(form, "scopes", scopeIDs); err != nil {
		return nil, err
	}
	var res Scopes
	if err := s.do(ctx, sub, &request{method: http.MethodPost, host: HostAPI, path: "/v2/user/revoke/scopes", form: form}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// ServiceTermsResult selects which service terms [UserService.ServiceTerms]
// lists.
type ServiceTermsResult string

// Service terms lists.
const (
	// AgreedServiceTerms lists the terms the user agreed to (default).
	AgreedServiceTerms ServiceTermsResult = "agreed_service_terms"
	// AppServiceTerms lists all terms enabled for the app.
	AppServiceTerms ServiceTermsResult = "app_service_terms"
)

// ServiceTermsOptions are the optional parameters of
// [UserService.ServiceTerms].
type ServiceTermsOptions struct {
	Result ServiceTermsResult
	// Tags limits the list to these terms tags.
	Tags []string
}

// ServiceTerms is a user's service terms status.
type ServiceTerms struct {
	// ID is the user ID.
	ID           int64         `json:"id"`
	ServiceTerms []ServiceTerm `json:"service_terms"`
}

// ServiceTerm is the user's status for one service term.
type ServiceTerm struct {
	Tag       string `json:"tag"`
	Required  bool   `json:"required"`
	Agreed    bool   `json:"agreed"`
	Revocable bool   `json:"revocable"`
	// AgreedAt is when the user last agreed.
	AgreedAt Time `json:"agreed_at"`
	// AgreedBy is "KAUTH" (Kakao Sync sign-up screen) or "KAPI" (the
	// agree API).
	AgreedBy string `json:"agreed_by"`
}

// ServiceTerms returns the user's service terms status. Kakao Sync only.
//
// Docs: https://developers.kakao.com/docs/ko/kakaologin/rest-api
func (s *UserService) ServiceTerms(ctx context.Context, sub Subject, opts *ServiceTermsOptions) (*ServiceTerms, error) {
	q := url.Values{}
	if opts != nil {
		setString(q, "result", string(opts.Result))
		setString(q, "tags", strings.Join(opts.Tags, ","))
	}
	var res ServiceTerms
	if err := s.do(ctx, sub, &request{method: http.MethodGet, host: HostAPI, path: "/v2/user/service_terms", query: q}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// ServiceTermsChange is the result of revoking or agreeing to service
// terms.
type ServiceTermsChange struct {
	// ID is the user ID.
	ID int64 `json:"id"`
	// Revoked lists the terms whose consent was withdrawn.
	Revoked []ServiceTerm `json:"revoked_service_terms"`
	// Agreed lists the terms newly agreed to; terms already agreed to
	// before the request are not included.
	Agreed []ServiceTerm `json:"agreed_service_terms"`
}

// RevokeServiceTerms withdraws the user's consent to the revocable service
// terms with tags. Kakao Sync only.
//
// Docs: https://developers.kakao.com/docs/ko/kakaologin/rest-api
func (s *UserService) RevokeServiceTerms(ctx context.Context, sub Subject, tags ...string) (*ServiceTermsChange, error) {
	return s.changeServiceTerms(ctx, sub, "/v2/user/revoke/service_terms", tags)
}

// AgreeServiceTerms records that the user agreed to the service terms with
// tags, after the service obtained consent itself. Kakao Sync only.
//
// Docs: https://developers.kakao.com/docs/ko/kakaologin/rest-api
func (s *UserService) AgreeServiceTerms(ctx context.Context, sub Subject, tags ...string) (*ServiceTermsChange, error) {
	return s.changeServiceTerms(ctx, sub, "/v2/user/upgrade/service_terms", tags)
}

func (s *UserService) changeServiceTerms(ctx context.Context, sub Subject, path string, tags []string) (*ServiceTermsChange, error) {
	if len(tags) == 0 {
		return nil, errors.New("kakao: no service terms tags")
	}
	form := url.Values{"tags": {strings.Join(tags, ",")}}
	var res ServiceTermsChange
	if err := s.do(ctx, sub, &request{method: http.MethodPost, host: HostAPI, path: path, form: form}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// OIDCUserInfo is the OpenID Connect user info.
type OIDCUserInfo struct {
	// Subject is the user ID.
	Subject       string `json:"sub"`
	Name          string `json:"name"`
	Nickname      string `json:"nickname"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Gender        string `json:"gender"`
	// Birthdate is "YYYY-MM-DD", "0000-MM-DD" or "YYYY" depending on the
	// consented items.
	Birthdate           string `json:"birthdate"`
	PhoneNumber         string `json:"phone_number"`
	PhoneNumberVerified bool   `json:"phone_number_verified"`
}

// OIDCUserInfo returns the OpenID Connect user info for an access token.
// For other fields, use [UserService.Me].
//
// Docs: https://developers.kakao.com/docs/ko/kakaologin/rest-api
func (s *UserService) OIDCUserInfo(ctx context.Context, accessToken string) (*OIDCUserInfo, error) {
	var info OIDCUserInfo
	err := s.do(ctx, AccessToken(accessToken), &request{method: http.MethodGet, host: HostAPI, path: "/v1/oidc/userinfo"}, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// setJSON sets key to the JSON encoding of v unless v is an empty slice or
// map.
func setJSON[T any](q url.Values, key string, v T) error {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("kakao: encode %s: %w", key, err)
	}
	if s := string(b); s != "null" && s != "[]" && s != "{}" {
		q.Set(key, s)
	}
	return nil
}
