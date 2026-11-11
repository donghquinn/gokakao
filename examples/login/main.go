// Command login is a minimal Kakao Login web server using OpenID Connect
// and PKCE.
//
// Register http://localhost:8080/callback as a redirect URI in the app
// settings, enable OpenID Connect, then run:
//
//	KAKAO_REST_KEY=... KAKAO_CLIENT_SECRET=... go run ./examples/login
//
// and open http://localhost:8080/login.
package main

import (
	"crypto/rand"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	kakao "github.com/donghquinn/gokakao"
)

const redirectURI = "http://localhost:8080/callback"

func main() {
	c, err := kakao.NewClient(
		kakao.WithRESTKey(os.Getenv("KAKAO_REST_KEY")),
		kakao.WithClientSecret(os.Getenv("KAKAO_CLIENT_SECRET")),
	)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /login", func(w http.ResponseWriter, r *http.Request) {
		state, nonce, verifier := rand.Text(), rand.Text(), kakao.GenerateVerifier()
		for name, v := range map[string]string{"state": state, "nonce": nonce, "verifier": verifier} {
			// Secure follows the request so the example works on plain-HTTP localhost.
			http.SetCookie(w, &http.Cookie{ //nolint:gosec // see above
				Name: name, Value: v, Path: "/", MaxAge: 600,
				HttpOnly: true, SameSite: http.SameSiteLaxMode,
				Secure: r.TLS != nil,
			})
		}
		u, err := c.Auth.AuthCodeURL(redirectURI, &kakao.AuthCodeOptions{
			State:         state,
			Nonce:         nonce,
			CodeChallenge: kakao.S256Challenge(verifier),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, u, http.StatusFound)
	})

	mux.HandleFunc("GET /callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if e := q.Get("error"); e != "" {
			http.Error(w, "login failed: "+e+": "+q.Get("error_description"), http.StatusUnauthorized)
			return
		}
		if c, err := r.Cookie("state"); err != nil || c.Value != q.Get("state") {
			http.Error(w, "state mismatch", http.StatusBadRequest)
			return
		}
		verifier, _ := r.Cookie("verifier")
		nonce, _ := r.Cookie("nonce")
		if verifier == nil || nonce == nil {
			http.Error(w, "login session expired", http.StatusBadRequest)
			return
		}

		tok, err := c.Auth.Exchange(r.Context(), q.Get("code"), redirectURI, &kakao.ExchangeOptions{CodeVerifier: verifier.Value})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		claims, err := c.Auth.VerifyIDToken(r.Context(), tok.IDToken, nonce.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		user, err := c.User.Me(r.Context(), kakao.AccessToken(tok.AccessToken), nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(map[string]any{"id_token": claims, "user": user})
	})

	srv := &http.Server{Addr: ":8080", Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	log.Println("open http://localhost:8080/login")
	log.Fatal(srv.ListenAndServe())
}
