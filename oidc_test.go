package kakao

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/donghquinn/gokakao/internal/kakaotest"
)

var testKeys = sync.OnceValue(func() [2]*rsa.PrivateKey {
	var keys [2]*rsa.PrivateKey
	for i := range keys {
		k, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			panic(err)
		}
		keys[i] = k
	}
	return keys
})

func b64(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

func signJWT(t *testing.T, key *rsa.PrivateKey, header, claims any) string {
	t.Helper()
	h, _ := json.Marshal(header)
	c, _ := json.Marshal(claims)
	input := b64(h) + "." + b64(c)
	sum := sha256.Sum256([]byte(input))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, sum[:])
	if err != nil {
		t.Fatal(err)
	}
	return input + "." + b64(sig)
}

func jwksJSON(keys map[string]*rsa.PrivateKey) string {
	var parts []string
	for kid, k := range keys {
		parts = append(parts, fmt.Sprintf(`{"kid":%q,"kty":"RSA","alg":"RS256","use":"sig","n":%q,"e":%q}`,
			kid, b64(k.N.Bytes()), b64(big.NewInt(int64(k.E)).Bytes())))
	}
	return `{"keys":[` + strings.Join(parts, ",") + `]}`
}

// oidcFixture is a client whose kauth host serves a JWKS with key "k1".
type oidcFixture struct {
	c        *Client
	srv      *kakaotest.Server
	now      time.Time
	jwksHits *atomic.Int32
	jwks     *atomic.Value // string
}

func newOIDCFixture(t *testing.T) *oidcFixture {
	t.Helper()
	c, srv := newTestClient(t, WithRESTKey("rest"))
	f := &oidcFixture{c: c, srv: srv, now: time.Unix(1_700_000_000, 0), jwksHits: new(atomic.Int32), jwks: new(atomic.Value)}
	fixedNow(c, f.now)
	f.jwks.Store(jwksJSON(map[string]*rsa.PrivateKey{"k1": testKeys()[0]}))
	srv.Handle("GET /.well-known/jwks.json", func(w http.ResponseWriter, _ *http.Request) {
		f.jwksHits.Add(1)
		_, _ = w.Write([]byte(f.jwks.Load().(string)))
	})
	return f
}

func (f *oidcFixture) claims() map[string]any {
	return map[string]any{
		"iss": Issuer, "aud": "rest", "sub": "166959",
		"iat": f.now.Unix() - 10, "exp": f.now.Unix() + 3600, "auth_time": f.now.Unix() - 10,
		"nonce": "n1", "nickname": "Ryan", "picture": "http://x/img.jpg", "email": "ryan@kakao.com",
	}
}

func TestAuthService_VerifyIDToken(t *testing.T) {
	t.Parallel()
	f := newOIDCFixture(t)
	key := testKeys()[0]
	hdr := map[string]string{"alg": "RS256", "kid": "k1", "typ": "JWT"}

	tok := signJWT(t, key, hdr, f.claims())
	claims, err := f.c.Auth.VerifyIDToken(t.Context(), tok, "n1")
	if err != nil {
		t.Fatal(err)
	}
	want := IDTokenClaims{
		Issuer: Issuer, Audience: "rest", Subject: "166959", IssuedAt: f.now.Unix() - 10, Expiry: f.now.Unix() + 3600,
		AuthTime: f.now.Unix() - 10, Nonce: "n1", Nickname: "Ryan", Picture: "http://x/img.jpg", Email: "ryan@kakao.com",
	}
	if *claims != want {
		t.Errorf("claims = %+v", claims)
	}
	if _, err := f.c.Auth.VerifyIDToken(t.Context(), tok, ""); err != nil {
		t.Errorf("empty nonce skips the check: %v", err)
	}
	if f.jwksHits.Load() != 1 {
		t.Errorf("JWKS fetched %d times, want 1 (cached)", f.jwksHits.Load())
	}

	with := func(k string, v any) string {
		c := f.claims()
		c[k] = v
		return signJWT(t, key, hdr, c)
	}
	parts := strings.Split(tok, ".")
	tests := []struct {
		name  string
		token string
		nonce string
	}{
		{"wrong nonce", tok, "other"},
		{"wrong issuer", with("iss", "https://evil.example"), ""},
		{"wrong audience", with("aud", "other-app"), ""},
		{"expired", with("exp", f.now.Unix()-120), ""},
		{"issued in future", with("iat", f.now.Unix()+600), ""},
		{"tampered payload", parts[0] + "." + b64([]byte(`{"iss":"https://kauth.kakao.com","aud":"rest","sub":"1","exp":9999999999}`)) + "." + parts[2], ""},
		{"signed by other key", signJWT(t, testKeys()[1], hdr, f.claims()), ""},
		{"alg none", signJWT(t, key, map[string]string{"alg": "none", "kid": "k1"}, f.claims()), ""},
		{"alg HS256", signJWT(t, key, map[string]string{"alg": "HS256", "kid": "k1"}, f.claims()), ""},
		{"two parts", parts[0] + "." + parts[1], ""},
		{"bad header", "!!!." + parts[1] + "." + parts[2], ""},
		{"bad signature encoding", parts[0] + "." + parts[1] + ".!!!", ""},
		{"bad payload JSON", parts[0] + "." + b64([]byte("{")) + "." + parts[2], ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := f.c.Auth.VerifyIDToken(t.Context(), tt.token, tt.nonce); !errors.Is(err, ErrInvalidIDToken) {
				t.Errorf("err = %v, want ErrInvalidIDToken", err)
			}
		})
	}
}

func TestAuthService_VerifyIDToken_payloadNotJSON(t *testing.T) {
	t.Parallel()
	f := newOIDCFixture(t)
	hdr, _ := json.Marshal(map[string]string{"alg": "RS256", "kid": "k1"})
	input := b64(hdr) + "." + b64([]byte("not json"))
	sum := sha256.Sum256([]byte(input))
	sig, _ := rsa.SignPKCS1v15(rand.Reader, testKeys()[0], crypto.SHA256, sum[:])
	if _, err := f.c.Auth.VerifyIDToken(t.Context(), input+"."+b64(sig), ""); !errors.Is(err, ErrInvalidIDToken) {
		t.Errorf("err = %v", err)
	}
}

func TestAuthService_VerifyIDToken_keyRotation(t *testing.T) {
	t.Parallel()
	f := newOIDCFixture(t)
	hdr2 := map[string]string{"alg": "RS256", "kid": "k2"}
	tok2 := signJWT(t, testKeys()[1], hdr2, f.claims())

	// Prime the cache with k1 only.
	if _, err := f.c.Auth.VerifyIDToken(t.Context(), signJWT(t, testKeys()[0], map[string]string{"alg": "RS256", "kid": "k1"}, f.claims()), ""); err != nil {
		t.Fatal(err)
	}
	// Kakao publishes k2. Within a minute of the last fetch, an unknown
	// kid does not trigger a refetch.
	f.jwks.Store(jwksJSON(map[string]*rsa.PrivateKey{"k1": testKeys()[0], "k2": testKeys()[1]}))
	if _, err := f.c.Auth.VerifyIDToken(t.Context(), tok2, ""); !errors.Is(err, ErrInvalidIDToken) {
		t.Errorf("err = %v, want unknown kid", err)
	}
	if f.jwksHits.Load() != 1 {
		t.Errorf("JWKS fetched %d times, want 1", f.jwksHits.Load())
	}
	// After a minute, the unknown kid triggers a refetch.
	fixedNow(f.c, f.now.Add(2*time.Minute))
	if _, err := f.c.Auth.VerifyIDToken(t.Context(), tok2, ""); err != nil {
		t.Errorf("after refetch: %v", err)
	}
	if f.jwksHits.Load() != 2 {
		t.Errorf("JWKS fetched %d times, want 2", f.jwksHits.Load())
	}
	// Cached keys expire after an hour.
	fixedNow(f.c, f.now.Add(2*time.Hour))
	c := f.claims()
	c["exp"] = f.now.Add(3 * time.Hour).Unix()
	if _, err := f.c.Auth.VerifyIDToken(t.Context(), signJWT(t, testKeys()[1], hdr2, c), ""); err != nil {
		t.Fatal(err)
	}
	if f.jwksHits.Load() != 3 {
		t.Errorf("JWKS fetched %d times, want 3", f.jwksHits.Load())
	}
}

func TestAuthService_VerifyIDToken_jwksErrors(t *testing.T) {
	t.Parallel()
	hdr := map[string]string{"alg": "RS256", "kid": "k1"}
	tests := []struct {
		name string
		jwks string
	}{
		{"bad modulus", `{"keys":[{"kid":"k1","kty":"RSA","use":"sig","n":"!!!","e":"AQAB"}]}`},
		{"bad exponent", `{"keys":[{"kid":"k1","kty":"RSA","use":"sig","n":"AQAB","e":"!!!"}]}`},
		{"empty exponent", `{"keys":[{"kid":"k1","kty":"RSA","use":"sig","n":"AQAB","e":""}]}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := newOIDCFixture(t)
			f.jwks.Store(tt.jwks)
			_, err := f.c.Auth.VerifyIDToken(t.Context(), signJWT(t, testKeys()[0], hdr, f.claims()), "")
			if err == nil || !strings.Contains(err.Error(), "fetch ID token keys") {
				t.Errorf("err = %v", err)
			}
		})
	}

	t.Run("non-signing and non-RSA keys are skipped", func(t *testing.T) {
		t.Parallel()
		f := newOIDCFixture(t)
		f.jwks.Store(`{"keys":[{"kid":"k1","kty":"EC","use":"sig"},{"kid":"k1","kty":"RSA","use":"enc","n":"AQAB","e":"AQAB"}]}`)
		_, err := f.c.Auth.VerifyIDToken(t.Context(), signJWT(t, testKeys()[0], hdr, f.claims()), "")
		if !errors.Is(err, ErrInvalidIDToken) || !strings.Contains(err.Error(), "unknown key ID") {
			t.Errorf("err = %v", err)
		}
	})

	t.Run("server error", func(t *testing.T) {
		t.Parallel()
		c, srv := newTestClient(t, WithRESTKey("rest"))
		srv.HandleJSON("GET /.well-known/jwks.json", 500, ``)
		_, err := c.Auth.VerifyIDToken(t.Context(), signJWT(t, testKeys()[0], hdr, map[string]any{}), "")
		if err == nil || !strings.Contains(err.Error(), "fetch ID token keys") {
			t.Errorf("err = %v", err)
		}
	})
}

func TestAuthService_VerifyIDToken_noRESTKey(t *testing.T) {
	t.Parallel()
	f := newOIDCFixture(t)
	f.c.restKey = ""
	c := f.claims()
	c["aud"] = ""
	tok := signJWT(t, testKeys()[0], map[string]string{"alg": "RS256", "kid": "k1"}, c)
	if _, err := f.c.Auth.VerifyIDToken(t.Context(), tok, ""); !errors.Is(err, ErrInvalidIDToken) {
		t.Errorf("err = %v, want audience mismatch", err)
	}
}

func TestAuthService_IDTokenInfo(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithRESTKey("rest"))
	srv.HandleJSON("POST /oauth/tokeninfo", 200, `{"iss":"https://kauth.kakao.com","aud":"rest","sub":"166959","iat":1647183250,"exp":1647190450,"nonce":"N","auth_time":1647183250,"nickname":"Ryan","picture":"http://sample.com/xxx/yyy/img_110x110.jpg","email":"sample@sample.com"}`)
	claims, err := c.Auth.IDTokenInfo(t.Context(), "IDT")
	if err != nil {
		t.Fatal(err)
	}
	if f := srv.LastRequest().Form(); f.Get("id_token") != "IDT" {
		t.Errorf("form = %v", f)
	}
	if claims.Subject != "166959" || claims.Expiry != 1647190450 || claims.Email != "sample@sample.com" {
		t.Errorf("claims = %+v", claims)
	}

	c2, srv2 := newTestClient(t)
	srv2.HandleJSON("POST /oauth/tokeninfo", 400, `{"error":"invalid_token","error_description":"bad","error_code":"KOE400"}`)
	var oe *OAuthError
	if _, err := c2.Auth.IDTokenInfo(t.Context(), "x"); !errors.As(err, &oe) || oe.KOECode != "KOE400" {
		t.Errorf("err = %v", err)
	}
}
