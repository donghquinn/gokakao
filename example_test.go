package kakao_test

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	kakao "github.com/donghquinn/gokakao"
)

func ExampleNewClient() {
	c, err := kakao.NewClient(
		kakao.WithRESTKey(os.Getenv("KAKAO_REST_KEY")),
		kakao.WithAdminKey(os.Getenv("KAKAO_ADMIN_KEY")),
		kakao.WithClientSecret(os.Getenv("KAKAO_CLIENT_SECRET")),
		kakao.WithRetry(kakao.RetryPolicy{MaxAttempts: 3, BaseDelay: 200 * time.Millisecond}),
	)
	if err != nil {
		log.Fatal(err)
	}
	_ = c
}

func ExampleError() {
	// An error as returned by a call made with an expired access token.
	var err error = &kakao.Error{HTTPStatus: 401, Code: -401, Msg: "this access token does not exist"}

	switch {
	case errors.Is(err, kakao.ErrInvalidToken):
		fmt.Println("refresh the token and retry")
	case errors.Is(err, kakao.ErrConsentRequired):
		var e *kakao.Error
		errors.As(err, &e)
		fmt.Println("ask the user to agree to", e.RequiredScopes)
	}
	// Output: refresh the token and retry
}
