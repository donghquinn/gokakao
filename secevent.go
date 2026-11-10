package kakao

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// Security event types sent by the account status webhook (계정 상태 변경
// 웹훅), keyed by their schema URI in [SecurityEvent.Events].
//
//nolint:gosec // event type URIs, not credentials
const (
	// EventTokensRevoked is sent when all of a user's tokens are revoked,
	// and with TokenClass "business" when all business tokens are revoked.
	EventTokensRevoked                   = "https://schemas.openid.net/secevent/oauth/event-type/tokens-revoked"
	EventUserLinked                      = "https://schemas.openid.net/secevent/oauth/event-type/user-linked"
	EventUserUnlinked                    = "https://schemas.openid.net/secevent/oauth/event-type/user-unlinked"
	EventUserScopeConsent                = "https://schemas.openid.net/secevent/oauth/event-type/user-scope-consent"
	EventUserScopeWithdraw               = "https://schemas.openid.net/secevent/oauth/event-type/user-scope-withdraw"
	EventBusinessTokenIssued             = "https://schemas.openid.net/secevent/oauth/event-type/token-issued"
	EventBusinessTokenRevoked            = "https://schemas.openid.net/secevent/oauth/event-type/token-revoked"
	EventAccountCredentialChangeRequired = "https://schemas.openid.net/secevent/risc/event-type/account-credential-change-required"
	EventAccountDisabled                 = "https://schemas.openid.net/secevent/risc/event-type/account-disabled"
	EventAccountEnabled                  = "https://schemas.openid.net/secevent/risc/event-type/account-enabled"
	EventAccountPurged                   = "https://schemas.openid.net/secevent/risc/event-type/account-purged"
	EventCredentialCompromise            = "https://schemas.openid.net/secevent/risc/event-type/credential-compromise"
	EventIdentifierChanged               = "https://schemas.openid.net/secevent/risc/event-type/identifier-changed"
	EventIdentifierRecycled              = "https://schemas.openid.net/secevent/risc/event-type/identifier-recycled"
	EventSessionsRevoked                 = "https://schemas.openid.net/secevent/risc/event-type/sessions-revoked"
	EventAssuranceLevelChange            = "https://schemas.openid.net/secevent/caep/event-type/assurance-level-change"
	EventCredentialChange                = "https://schemas.openid.net/secevent/caep/event-type/credential-change"
	EventUserProfileChanged              = "https://schemas.kakao.com/platevent/kakao/event-type/user-profile-changed"
)

// SecurityEvent is a verified Security Event Token (SET) from the account
// status webhook.
type SecurityEvent struct {
	Issuer string `json:"iss"`
	// Audience is the app's primary REST API key.
	Audience string `json:"aud"`
	// Subject is the user ID.
	Subject string `json:"sub"`
	// AppID is the app ID.
	AppID string `json:"app_id"`
	// JTI uniquely identifies the SET; use it to drop duplicate deliveries.
	JTI string `json:"jti"`
	// TransactionID identifies the webhook request.
	TransactionID string `json:"txn"`
	// TimeOfEvent is when the event happened, in Unix seconds.
	TimeOfEvent int64 `json:"toe"`
	// IssuedAt is when the SET was issued, in Unix seconds.
	IssuedAt int64 `json:"iat"`
	// Events maps event type URIs, such as [EventUserUnlinked], to their
	// details. A SET usually carries one event.
	Events map[string]EventDetail `json:"events"`
}

// UnmarshalJSON implements [json.Unmarshaler]. Kakao documents the
// transaction claim as "txm"; both "txn" and "txm" are accepted.
func (e *SecurityEvent) UnmarshalJSON(b []byte) error {
	type alias SecurityEvent
	aux := struct {
		*alias
		Txm string `json:"txm"`
	}{alias: (*alias)(e)}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	if e.TransactionID == "" {
		e.TransactionID = aux.Txm
	}
	return nil
}

// EventDetail holds the fields of one security event. Which fields are
// set depends on the event type; see the Kakao docs.
type EventDetail struct {
	// Subject identifies the user, business token or identifier the event
	// is about.
	Subject *EventSubject `json:"subject"`
	// Reason explains tokens-revoked ("issuer" or "user"), user-unlinked
	// (e.g. "UNLINK_FROM_APPS") and account-disabled events.
	Reason string `json:"reason"`
	// Scope lists consent item IDs, space-separated, for scope events.
	Scope string `json:"scope"`
	// TokenSubject, TokenID and TokenClass describe business token events.
	// TokenClass is "business".
	TokenSubject *EventSubject `json:"token_subject"`
	TokenID      string        `json:"token_id"`
	TokenClass   string        `json:"token_class"`
	// NewValue is the new or recycled phone number or email for identifier
	// events.
	NewValue string `json:"new_value"`
	// CurrentLevel, PreviousLevel ("nist-aal1" or "nist-aal2") and
	// ChangeDirection ("increase" or "decrease") describe assurance level
	// changes.
	CurrentLevel    string `json:"current_level"`
	PreviousLevel   string `json:"previous_level"`
	ChangeDirection string `json:"change_direction"`
	// ChangeType is "update" for credential changes.
	ChangeType string `json:"change_type"`
	// Profile lists the changed consent item IDs, space-separated, for
	// user-profile-changed events. The new values are not included.
	Profile string `json:"profile"`
}

// EventSubject identifies what a security event is about.
type EventSubject struct {
	// SubjectType is "iss-sub", "phone", "account_email" or, for business
	// tokens, "oauth_token".
	SubjectType string `json:"subject_type"`
	Iss         string `json:"iss"`
	Sub         string `json:"sub"`
	// PhoneNumber and AccountEmail are the old identifiers in identifier
	// events.
	PhoneNumber  string `json:"phone_number"`
	AccountEmail string `json:"account_email"`
	// TokenType, TokenIdentifierAlg and Token describe a business token;
	// Token is base64url(sha256(token)).
	TokenType          string `json:"token_type"`
	TokenIdentifierAlg string `json:"token_identifier_alg"`
	Token              string `json:"token"`
}

// SETError reports an invalid Security Event Token. Code is the RFC 8935
// error code to return to Kakao: "invalid_request", "invalid_key",
// "invalid_issuer" or "invalid_audience".
type SETError struct {
	Code        string
	Description string
}

func (e *SETError) Error() string {
	return fmt.Sprintf("kakao: invalid security event token: %s: %s", e.Code, e.Description)
}

// VerifySecurityEvent verifies a Security Event Token from the account
// status webhook and returns its contents. It checks the RS256 signature
// against Kakao's keys, the secevent+jwt type, that iss is [Issuer] and
// that aud is the REST API key.
//
// A token that fails validation returns a [*SETError]; other errors mean
// the signing keys could not be fetched and the delivery should be
// retried. The webhook package's handler maps both to the responses Kakao
// expects.
//
// Docs: https://developers.kakao.com/docs/ko/kakaologin/callback
func (s *AuthService) VerifySecurityEvent(ctx context.Context, token string) (*SecurityEvent, error) {
	payload, err := s.verifyJWS(ctx, token, "secevent+jwt")
	if err != nil {
		var je *jwsError
		if errors.As(err, &je) {
			return nil, &SETError{Code: je.code, Description: je.msg}
		}
		return nil, err
	}
	var ev SecurityEvent
	if err := json.Unmarshal(payload, &ev); err != nil {
		return nil, &SETError{Code: "invalid_request", Description: "payload: " + err.Error()}
	}
	if ev.Issuer != Issuer {
		return nil, &SETError{Code: "invalid_issuer", Description: fmt.Sprintf("issuer %q", ev.Issuer)}
	}
	if s.c.restKey == "" || ev.Audience != s.c.restKey {
		return nil, &SETError{Code: "invalid_audience", Description: "audience does not match the REST API key"}
	}
	return &ev, nil
}
