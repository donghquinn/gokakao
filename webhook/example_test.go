package webhook_test

import (
	"log"
	"net/http"
	"os"

	kakao "github.com/donghquinn/gokakao"
	"github.com/donghquinn/gokakao/webhook"
)

func Example() {
	c, err := kakao.NewClient(kakao.WithRESTKey(os.Getenv("KAKAO_REST_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	adminKey := os.Getenv("KAKAO_PRIMARY_ADMIN_KEY")
	logErrors := webhook.WithErrorHandler(func(r *http.Request, err error) {
		log.Printf("%s: %v", r.URL.Path, err)
	})

	mux := http.NewServeMux()
	mux.Handle("/kakao/unlink", webhook.UnlinkHandler(adminKey, func(_ *http.Request, e *webhook.UnlinkEvent) error {
		log.Printf("user %d unlinked (%s)", e.UserID, e.ReferrerType)
		return nil
	}, logErrors))
	mux.Handle("POST /kakao/events", webhook.SecurityEventHandler(c.Auth, func(_ *http.Request, e *kakao.SecurityEvent) error {
		if _, ok := e.Events[kakao.SecurityEventSessionsRevoked]; ok {
			log.Printf("log out user %s everywhere", e.Subject)
		}
		return nil
	}, logErrors))
	_ = http.ListenAndServe(":8080", mux) //nolint:gosec // example
}
