package kakao_test

import (
	"context"
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

func ExampleSearchService_Book() {
	c, err := kakao.NewClient(kakao.WithRESTKey(os.Getenv("KAKAO_REST_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	res, err := c.Search.Book(context.Background(), "미움받을 용기", &kakao.BookSearchOptions{
		Target: kakao.BookTargetTitle,
		Size:   5,
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, b := range res.Documents {
		fmt.Println(b.Title, b.Authors, b.SalePrice)
	}
}

func ExampleMapService_SearchKeyword() {
	c, err := kakao.NewClient(kakao.WithRESTKey(os.Getenv("KAKAO_REST_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	res, err := c.Map.SearchKeyword(context.Background(), "카페", &kakao.PlaceSearchOptions{
		Center: &kakao.Point{X: 127.1086228, Y: 37.4012191}, // Pangyo station
		Radius: 500,
		Sort:   kakao.PlaceSortDistance,
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, p := range res.Documents {
		fmt.Printf("%s (%.0fm)\n", p.PlaceName, p.Distance)
	}
}

func ExampleMapService_StaticMap() {
	c, err := kakao.NewClient(kakao.WithRESTKey(os.Getenv("KAKAO_REST_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	img, err := c.Map.StaticMap(context.Background(), &kakao.StaticMapParams{
		Width: 640, Height: 480,
		Markers: []kakao.Marker{{Location: kakao.Point{X: 127.1105, Y: 37.3955}}},
	})
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("map.png", img.Data, 0o600); err != nil {
		log.Fatal(err)
	}
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
