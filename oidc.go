package kakao

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Issuer is the iss claim of Kakao ID tokens.
const Issuer = "https://kauth.kakao.com"

const (
	// jwksMaxAge is how long fetched signing keys are trusted.
	jwksMaxAge = time.Hour
	// jwksMinRefetch limits refetches triggered by unknown key IDs.
	jwksMinRefetch = time.Minute
	// idTokenLeeway tolerates clock skew in exp and iat checks.
	idTokenLeeway = time.Minute
)

// ErrInvalidIDToken is returned by [AuthService.VerifyIDToken] when the
// token is malformed, badly signed, expired, or for another app.
var ErrInvalidIDToken = errors.New("kakao: invalid ID token")

// IDTokenClaims are the claims of a Kakao ID token.
type IDTokenClaims struct {
	Issuer   string `json:"iss"`
	Audience string `json:"aud"` // the app's REST API key
	Subject  string `json:"sub"` // the user ID
	IssuedAt int64  `json:"iat"` // Unix seconds
	Expiry   int64  `json:"exp"` // Unix seconds
	AuthTime int64  `json:"auth_time"`
	Nonce    string `json:"nonce,omitempty"`
	// Nickname and Picture need the profile (or nickname / profile image)
	// consent item.
	Nickname string `json:"nickname,omitempty"`
	Picture  string `json:"picture,omitempty"`
	// Email is set only when the account has a valid email and the user
	// consented to account_email.
	Email string `json:"email,omitempty"`
}

// VerifyIDToken checks the signature and claims of an ID token and returns
// its claims. It verifies the RS256 signature against Kakao's published
// keys, that iss is [Issuer], that aud is the REST API key, that the token
// is not expired, and, when nonce is non-empty, that it matches.
//
// Signing keys are cached for an hour and refetched when a token names an
// unknown key.
//
// Docs: https://developers.kakao.com/docs/ko/kakaologin/rest-api
func (s *AuthService) VerifyIDToken(ctx context.Context, idToken, nonce string) (*IDTokenClaims, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("%w: not a JWS compact serialization", ErrInvalidIDToken)
	}
	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	if err := decodeSegment(parts[0], &header); err != nil {
		return nil, fmt.Errorf("%w: header: %w", ErrInvalidIDToken, err)
	}
	if header.Alg != "RS256" {
		return nil, fmt.Errorf("%w: unsupported alg %q", ErrInvalidIDToken, header.Alg)
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("%w: signature: %w", ErrInvalidIDToken, err)
	}
	key, err := s.jwks.key(ctx, s.c, header.Kid)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, digest[:], sig); err != nil {
		return nil, fmt.Errorf("%w: bad signature", ErrInvalidIDToken)
	}

	var claims IDTokenClaims
	if err := decodeSegment(parts[1], &claims); err != nil {
		return nil, fmt.Errorf("%w: payload: %w", ErrInvalidIDToken, err)
	}
	now := s.c.now()
	switch {
	case claims.Issuer != Issuer:
		return nil, fmt.Errorf("%w: issuer %q", ErrInvalidIDToken, claims.Issuer)
	case s.c.restKey == "" || claims.Audience != s.c.restKey:
		return nil, fmt.Errorf("%w: audience does not match the REST API key", ErrInvalidIDToken)
	case !now.Before(time.Unix(claims.Expiry, 0).Add(idTokenLeeway)):
		return nil, fmt.Errorf("%w: expired", ErrInvalidIDToken)
	case time.Unix(claims.IssuedAt, 0).After(now.Add(idTokenLeeway)):
		return nil, fmt.Errorf("%w: issued in the future", ErrInvalidIDToken)
	case nonce != "" && claims.Nonce != nonce:
		return nil, fmt.Errorf("%w: nonce mismatch", ErrInvalidIDToken)
	}
	return &claims, nil
}

// IDTokenInfo asks Kakao to decode an ID token. Kakao provides it for
// debugging only and may block other use; verify tokens in production with
// [AuthService.VerifyIDToken].
//
// Docs: https://developers.kakao.com/docs/ko/kakaologin/rest-api
func (s *AuthService) IDTokenInfo(ctx context.Context, idToken string) (*IDTokenClaims, error) {
	var claims IDTokenClaims
	err := s.c.do(ctx, &request{
		method: http.MethodPost, host: HostAuth, path: "/oauth/tokeninfo",
		form: url.Values{"id_token": {idToken}},
	}, &claims)
	if err != nil {
		return nil, err
	}
	return &claims, nil
}

func decodeSegment(seg string, v any) error {
	b, err := base64.RawURLEncoding.DecodeString(seg)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

// jwksCache holds Kakao's ID token signing keys.
type jwksCache struct {
	mu        sync.Mutex
	keys      map[string]*rsa.PublicKey
	fetchedAt time.Time
}

func (j *jwksCache) key(ctx context.Context, c *Client, kid string) (*rsa.PublicKey, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	now := c.now()
	age := now.Sub(j.fetchedAt)
	if k, ok := j.keys[kid]; ok && age < jwksMaxAge {
		return k, nil
	}
	if j.fetchedAt.IsZero() || age >= jwksMinRefetch {
		keys, err := fetchJWKS(ctx, c)
		if err != nil {
			return nil, fmt.Errorf("kakao: fetch ID token keys: %w", err)
		}
		j.keys, j.fetchedAt = keys, now
	}
	if k, ok := j.keys[kid]; ok {
		return k, nil
	}
	return nil, fmt.Errorf("%w: unknown key ID %q", ErrInvalidIDToken, kid)
}

func fetchJWKS(ctx context.Context, c *Client) (map[string]*rsa.PublicKey, error) {
	var set struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			Alg string `json:"alg"`
			Use string `json:"use"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	err := c.do(ctx, &request{method: http.MethodGet, host: HostAuth, path: "/.well-known/jwks.json"}, &set)
	if err != nil {
		return nil, err
	}
	keys := make(map[string]*rsa.PublicKey, len(set.Keys))
	for _, k := range set.Keys {
		if k.Kty != "RSA" || (k.Use != "" && k.Use != "sig") {
			continue
		}
		n, err := base64.RawURLEncoding.DecodeString(k.N)
		if err != nil {
			return nil, fmt.Errorf("key %q: modulus: %w", k.Kid, err)
		}
		e, err := base64.RawURLEncoding.DecodeString(k.E)
		if err != nil || len(e) == 0 || len(e) > 4 {
			return nil, fmt.Errorf("key %q: bad exponent", k.Kid)
		}
		keys[k.Kid] = &rsa.PublicKey{
			N: new(big.Int).SetBytes(n),
			E: int(new(big.Int).SetBytes(bytes.TrimLeft(e, "\x00")).Int64()),
		}
	}
	return keys, nil
}
