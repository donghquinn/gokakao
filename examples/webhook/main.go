// Command webhook receives Kakao webhooks.
//
// Register these URLs in the app settings (HTTPS is required in production):
//
//	/kakao/unlink    연결 해제 웹훅
//	/kakao/events    계정 상태 변경 웹훅
//	/kakao/channel   카카오톡 채널 웹훅
//
// and run:
//
//	KAKAO_REST_KEY=... KAKAO_PRIMARY_ADMIN_KEY=... go run ./examples/webhook
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	kakao "github.com/donghquinn/gokakao"
	"github.com/donghquinn/gokakao/webhook"
)

func main() {
	c, err := kakao.NewClient(kakao.WithRESTKey(os.Getenv("KAKAO_REST_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	adminKey := os.Getenv("KAKAO_PRIMARY_ADMIN_KEY")

	// Kakao expects an answer within 3 seconds; hand real work to a queue.
	jobs := make(chan func(), 100)
	go func() {
		for job := range jobs {
			job()
		}
	}()

	mux := http.NewServeMux()
	mux.Handle("/kakao/unlink", webhook.UnlinkHandler(adminKey, func(_ *http.Request, e *webhook.UnlinkEvent) {
		jobs <- func() { log.Printf("unlink: user %d (%s)", e.UserID, e.ReferrerType) }
	}))
	mux.Handle("POST /kakao/events", webhook.SecurityEventHandler(c.Auth, func(_ *http.Request, e *kakao.SecurityEvent) {
		jobs <- func() {
			for typ, d := range e.Events {
				log.Printf("security event %s for user %s: %+v", typ, e.Subject, d)
			}
		}
	}))
	mux.Handle("POST /kakao/channel", webhook.ChannelHandler(adminKey, func(_ *http.Request, e *webhook.ChannelEvent) {
		jobs <- func() { log.Printf("channel %s: %s by %s %s", e.ChannelPublicID, e.Event, e.IDType, e.ID) }
	}))

	srv := &http.Server{Addr: ":8080", Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	log.Fatal(srv.ListenAndServe())
}
