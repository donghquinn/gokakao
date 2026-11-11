package kakao

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// Error is an error response from kapi.kakao.com or dapi.kakao.com.
//
// kapi returns {"code": -401, "msg": "..."}; dapi returns
// {"errorType": "...", "message": "..."}. Both are mapped onto Error.
type Error struct {
	// HTTPStatus is the HTTP status code of the response.
	HTTPStatus int
	// Code is the Kakao error code, e.g. -401. It is 0 for dapi errors.
	Code int `json:"code"`
	// Msg is the human-readable message.
	Msg string `json:"msg"`
	// Type is the dapi errorType, e.g. "MissingParameter".
	Type string `json:"errorType"`
	// APIType names the API that failed, when Kakao reports it.
	APIType string `json:"api_type"`
	// RequiredScopes lists the consent items the user must agree to.
	// It is set for code -402.
	RequiredScopes []string `json:"required_scopes"`
	// AllowedScopes lists the consent items available to the app.
	AllowedScopes []string `json:"allowed_scopes"`
	// Extras holds the detailed error of the Kakao Moment and Keyword Ad
	// APIs.
	Extras *ErrorExtras `json:"extras"`
}

// ErrorExtras is the detailed error of the advertising APIs.
type ErrorExtras struct {
	// DetailCode is the API-specific error, e.g. 21006 (ad account not
	// found).
	DetailCode int    `json:"detailCode"`
	DetailMsg  string `json:"detailMsg"`
	Path       string `json:"path"`
	Timestamp  string `json:"timestamp"`
}

func (e *Error) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "kakao: HTTP %d", e.HTTPStatus)
	if e.Code != 0 {
		fmt.Fprintf(&b, ", code %d", e.Code)
	}
	if e.Type != "" {
		fmt.Fprintf(&b, ", %s", e.Type)
	}
	if e.Msg != "" {
		fmt.Fprintf(&b, ": %s", e.Msg)
	}
	if e.Extras != nil && (e.Extras.DetailCode != 0 || e.Extras.DetailMsg != "") {
		fmt.Fprintf(&b, " (detail %d: %s)", e.Extras.DetailCode, e.Extras.DetailMsg)
	}
	return b.String()
}

// Is reports whether e matches one of the sentinel errors such as
// [ErrInvalidToken].
func (e *Error) Is(target error) bool {
	s, ok := target.(codeError)
	if !ok {
		return false
	}
	if s == ErrRateLimited && e.HTTPStatus == http.StatusTooManyRequests {
		return true
	}
	return e.Code != 0 && int(s) == e.Code
}

// OAuthError is an error response from kauth.kakao.com.
type OAuthError struct {
	// HTTPStatus is the HTTP status code of the response.
	HTTPStatus int
	// Err is the OAuth 2.0 error, e.g. "invalid_grant".
	Err string `json:"error"`
	// Description is the error_description field.
	Description string `json:"error_description"`
	// KOECode is Kakao's own error code, e.g. "KOE320".
	KOECode string `json:"error_code"`
}

func (e *OAuthError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "kakao: oauth HTTP %d: %s", e.HTTPStatus, e.Err)
	if e.KOECode != "" {
		fmt.Fprintf(&b, " (%s)", e.KOECode)
	}
	if e.Description != "" {
		fmt.Fprintf(&b, ": %s", e.Description)
	}
	return b.String()
}

// Is reports whether e matches a sentinel error. KOE237 (token request
// limit exceeded) matches [ErrRateLimited].
func (e *OAuthError) Is(target error) bool {
	return target == ErrRateLimited && (e.KOECode == "KOE237" || e.HTTPStatus == http.StatusTooManyRequests)
}

// codeError is a sentinel matched against [Error.Code].
type codeError int

func (c codeError) Error() string {
	if msg, ok := codeMessages[c]; ok {
		return fmt.Sprintf("kakao: code %d: %s", int(c), msg)
	}
	return fmt.Sprintf("kakao: code %d", int(c))
}

// Sentinel errors for common Kakao error codes. Match them with [errors.Is]:
//
//	if errors.Is(err, kakao.ErrInvalidToken) { ... }
//
// Inspect [*Error] with [errors.As] for codes without a sentinel.
const (
	ErrInternal           codeError = -1    // internal error; retry later
	ErrInvalidParameter   codeError = -2    // bad or missing parameter
	ErrFeatureDisabled    codeError = -3    // feature or consent item not enabled for the app
	ErrAccountRestricted  codeError = -4    // account is restricted
	ErrPermissionDenied   codeError = -5    // app lacks permission for the API
	ErrNotAllowed         codeError = -6    // action not allowed
	ErrServiceUnavailable codeError = -7    // service maintenance or internal error
	ErrInvalidHeader      codeError = -8    // bad request header
	ErrAPIDiscontinued    codeError = -9    // API no longer provided
	ErrRateLimited        codeError = -10   // quota exceeded; also matches HTTP 429 and KOE237
	ErrPaidQuotaExceeded  codeError = -11   // paid API limit exceeded
	ErrAppRestricted      codeError = -12   // app or developer is restricted
	ErrDormantApp         codeError = -13   // app is dormant
	ErrNotLinked          codeError = -101  // user is not linked to the app
	ErrAlreadyLinked      codeError = -102  // user is already linked
	ErrAccountNotFound    codeError = -103  // dormant or non-existent account
	ErrPropertyNotFound   codeError = -201  // user property not registered
	ErrInvalidToken       codeError = -401  // invalid or expired key or token
	ErrConsentRequired    codeError = -402  // user consent required; see Error.RequiredScopes
	ErrUnderAgeRestricted codeError = -406  // user under 14 is restricted
	ErrNotTalkUser        codeError = -501  // user is not a KakaoTalk user
	ErrMaintenance        codeError = -9798 // scheduled maintenance
)

var codeMessages = map[codeError]string{
	ErrInternal:           "internal error",
	ErrInvalidParameter:   "invalid parameter",
	ErrFeatureDisabled:    "feature not enabled",
	ErrAccountRestricted:  "account restricted",
	ErrPermissionDenied:   "permission denied",
	ErrNotAllowed:         "not allowed",
	ErrServiceUnavailable: "service unavailable",
	ErrInvalidHeader:      "invalid header",
	ErrAPIDiscontinued:    "API discontinued",
	ErrRateLimited:        "rate limited",
	ErrPaidQuotaExceeded:  "paid quota exceeded",
	ErrAppRestricted:      "app restricted",
	ErrDormantApp:         "dormant app",
	ErrNotLinked:          "user not linked",
	ErrAlreadyLinked:      "user already linked",
	ErrAccountNotFound:    "account not found",
	ErrPropertyNotFound:   "property not found",
	ErrInvalidToken:       "invalid token",
	ErrConsentRequired:    "consent required",
	ErrUnderAgeRestricted: "under-age restricted",
	ErrNotTalkUser:        "not a KakaoTalk user",
	ErrMaintenance:        "maintenance",
}

// ErrMissingCredential is returned when a request needs a credential that
// the [Client] was not configured with.
var ErrMissingCredential = errors.New("kakao: missing credential")

// parseError builds an error from a non-2xx response body.
func parseError(h Host, status int, body []byte) error {
	if h == HostAuth {
		e := &OAuthError{HTTPStatus: status}
		if json.Unmarshal(body, e) == nil && (e.Err != "" || e.KOECode != "") {
			return e
		}
	}
	e := &Error{HTTPStatus: status}
	var dapi struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(body, e) != nil {
		e.Msg = truncate(string(body), 200)
		return e
	}
	if e.Msg == "" && json.Unmarshal(body, &dapi) == nil {
		e.Msg = dapi.Message
	}
	if e.Code == 0 && e.Msg == "" && e.Type == "" {
		e.Msg = truncate(string(body), 200)
	}
	return e
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
