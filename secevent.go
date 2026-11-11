package kakao

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// SecurityEventType is the schema URI of a security event type, the key of
// [SecurityEvent.Events].
type SecurityEventType string

// Security event types sent by the account status webhook (계정 상태 변경
// 웹훅).
//
//nolint:gosec // event type URIs, not credentials
const (
	// SecurityEventTokensRevoked is sent when all of a user's tokens are revoked,
	// and with TokenClass "business" when all business tokens are revoked.
	SecurityEventTokensRevoked                   SecurityEventType = "https://schemas.openid.net/secevent/oauth/event-type/tokens-revoked"
	SecurityEventUserLinked                      SecurityEventType = "https://schemas.openid.net/secevent/oauth/event-type/user-linked"
	SecurityEventUserUnlinked                    SecurityEventType = "https://schemas.openid.net/secevent/oauth/event-type/user-unlinked"
	SecurityEventUserScopeConsent                SecurityEventType = "https://schemas.openid.net/secevent/oauth/event-type/user-scope-consent"
	SecurityEventUserScopeWithdraw               SecurityEventType = "https://schemas.openid.net/secevent/oauth/event-type/user-scope-withdraw"
	SecurityEventBusinessTokenIssued             SecurityEventType = "https://schemas.openid.net/secevent/oauth/event-type/token-issued"
	SecurityEventBusinessTokenRevoked            SecurityEventType = "https://schemas.openid.net/secevent/oauth/event-type/token-revoked"
	SecurityEventAccountCredentialChangeRequired SecurityEventType = "https://schemas.openid.net/secevent/risc/event-type/account-credential-change-required"
	SecurityEventAccountDisabled                 SecurityEventType = "https://schemas.openid.net/secevent/risc/event-type/account-disabled"
	SecurityEventAccountEnabled                  SecurityEventType = "https://schemas.openid.net/secevent/risc/event-type/account-enabled"
	SecurityEventAccountPurged                   SecurityEventType = "https://schemas.openid.net/secevent/risc/event-type/account-purged"
	SecurityEventCredentialCompromise            SecurityEventType = "https://schemas.openid.net/secevent/risc/event-type/credential-compromise"
	SecurityEventIdentifierChanged               SecurityEventType = "https://schemas.openid.net/secevent/risc/event-type/identifier-changed"
	SecurityEventIdentifierRecycled              SecurityEventType = "https://schemas.openid.net/secevent/risc/event-type/identifier-recycled"
	SecurityEventSessionsRevoked                 SecurityEventType = "https://schemas.openid.net/secevent/risc/event-type/sessions-revoked"
	SecurityEventAssuranceLevelChange            SecurityEventType = "https://schemas.openid.net/secevent/caep/event-type/assurance-level-change"
	SecurityEventCredentialChange                SecurityEventType = "https://schemas.openid.net/secevent/caep/event-type/credential-change"
	SecurityEventUserProfileChanged              SecurityEventType = "https://schemas.kakao.com/platevent/kakao/event-type/user-profile-changed"
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
	AppID int64 `json:"-"`
	// JTI uniquely identifies the SET; use it to drop duplicate deliveries.
	JTI string `json:"jti"`
	// TransactionID identifies the webhook request.
	TransactionID string `json:"txn"`
	// TimeOfEvent is when the event happened, in Unix seconds.
	TimeOfEvent int64 `json:"toe"`
	// IssuedAt is when the SET was issued, in Unix seconds.
	IssuedAt int64 `json:"iat"`
	// Events maps event type URIs, such as [SecurityEventUserUnlinked], to their
	// details. A SET usually carries one event.
	Events map[SecurityEventType]SecurityEventDetail `json:"events"`
}

// UnmarshalJSON implements [json.Unmarshaler]. Kakao documents the
// transaction claim as "txm"; both "txn" and "txm" are accepted. app_id may
// be a string or a number.
func (e *SecurityEvent) UnmarshalJSON(b []byte) error {
	type alias SecurityEvent
	aux := struct {
		*alias
		Txm   string          `json:"txm"`
		AppID json.RawMessage `json:"app_id"`
	}{alias: (*alias)(e)}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	if e.TransactionID == "" {
		e.TransactionID = aux.Txm
	}
	if len(aux.AppID) > 0 && string(aux.AppID) != "null" {
		id, err := strconv.ParseInt(strings.Trim(string(aux.AppID), `"`), 10, 64)
		if err != nil {
			return fmt.Errorf("kakao: app_id: %w", err)
		}
		e.AppID = id
	}
	return nil
}

// SecurityEventDetail holds the fields of one security event. Which fields are
// set depends on the event type; see the Kakao docs.
type SecurityEventDetail struct {
	// Subject identifies the user, business token or identifier the event
	// is about.
	Subject *SecurityEventSubject `json:"subject"`
	// Reason explains tokens-revoked ("issuer" or "user"), user-unlinked
	// (e.g. "UNLINK_FROM_APPS") and account-disabled events.
	Reason string `json:"reason"`
	// Scope lists consent item IDs, space-separated, for scope events.
	Scope string `json:"scope"`
	// TokenSubject, TokenID and TokenClass describe business token events.
	// TokenClass is "business".
	TokenSubject *SecurityEventSubject `json:"token_subject"`
	TokenID      string                `json:"token_id"`
	TokenClass   string                `json:"token_class"`
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

// SecurityEventSubject identifies what a security event is about.
type SecurityEventSubject struct {
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
