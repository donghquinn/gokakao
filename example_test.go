package kakao_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	kakao "github.com/donghquinn/gokakao"
	"github.com/donghquinn/gokakao/template"
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
	img, err := c.Map.StaticMap(context.Background(), kakao.StaticMapParams{
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

func ExampleAuthService_AuthCodeURL() {
	c, _ := kakao.NewClient(kakao.WithRESTKey("REST_API_KEY"))
	verifier := kakao.GenerateVerifier() // keep it in the user's session
	u, err := c.Auth.AuthCodeURL("https://example.com/callback", &kakao.AuthCodeOptions{
		State:         "random-state",
		Nonce:         "random-nonce",
		CodeChallenge: kakao.S256Challenge(verifier),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(u[:len("https://kauth.kakao.com/oauth/authorize")])
	// Output: https://kauth.kakao.com/oauth/authorize
}

func ExampleAuthService_TokenSource() {
	c, _ := kakao.NewClient(kakao.WithRESTKey(os.Getenv("KAKAO_REST_KEY")))
	var stored *kakao.Token // loaded from your database

	ts := c.Auth.TokenSource(stored, &kakao.TokenSourceOptions{
		OnRefresh: func(*kakao.Token) { /* save the token */ },
	})
	tok, err := ts.Token(context.Background()) // refreshed when about to expire
	if err != nil {
		log.Fatal(err)
	}
	me, err := c.User.Me(context.Background(), kakao.AccessToken(tok.AccessToken), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(me.ID)
}

func ExampleUserService_Me() {
	c, _ := kakao.NewClient(kakao.WithAdminKey(os.Getenv("KAKAO_ADMIN_KEY")))
	ctx := context.Background()

	// As the logged-in user...
	me, err := c.User.Me(ctx, kakao.AccessToken("USER_ACCESS_TOKEN"), &kakao.MeOptions{
		PropertyKeys: []string{"kakao_account.email"},
	})
	if err != nil {
		log.Fatal(err)
	}
	// ...or as admin for any linked user.
	same, err := c.User.Me(ctx, kakao.AsAdmin(me.ID), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(same.KakaoAccount.Email)
}

func ExampleTalkService_SendToMe() {
	c, _ := kakao.NewClient()
	err := c.Talk.SendToMe(context.Background(), "USER_ACCESS_TOKEN", template.Text{
		Text: "주문이 완료되었습니다.",
		Link: template.Link{WebURL: "https://example.com/orders/1"},
	})
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleTalkService_SendToFriends() {
	c, _ := kakao.NewClient()
	ctx := context.Background()
	var uuids []string
	for f, err := range c.Talk.AllFriends(ctx, "USER_ACCESS_TOKEN", nil) {
		if err != nil {
			log.Fatal(err)
		}
		if len(uuids) < 5 {
			uuids = append(uuids, f.UUID)
		}
	}
	res, err := c.Talk.SendToFriends(ctx, "USER_ACCESS_TOKEN", uuids, template.Feed{
		Content: template.Content{Title: "초대합니다", Link: template.Link{WebURL: "https://example.com"}},
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range res.FailureInfo { // partial failures are not errors
		fmt.Println(f.Code, f.ReceiverUUIDs)
	}
}

func ExampleChannelService_Relations() {
	c, _ := kakao.NewClient(kakao.WithAdminKey(os.Getenv("KAKAO_ADMIN_KEY")))
	rel, err := c.Channel.Relations(context.Background(), kakao.AsAdmin(1234), "_ZeUTxl")
	if err != nil {
		log.Fatal(err)
	}
	for _, ch := range rel.Channels {
		fmt.Println(ch.ChannelPublicID, ch.Relation == kakao.RelationAdded)
	}
}

func ExampleCalendarService_CreateEvent() {
	c, _ := kakao.NewClient()
	start := time.Date(2026, 12, 1, 10, 0, 0, 0, time.FixedZone("KST", 9*3600))
	id, err := c.Calendar.CreateEvent(context.Background(), "USER_ACCESS_TOKEN", kakao.PrimaryCalendar, kakao.NewEvent{
		Title:     "주간 회의",
		Time:      kakao.EventTime{StartAt: start, EndAt: start.Add(time.Hour), TimeZone: "Asia/Seoul"},
		RRule:     "FREQ=WEEKLY;COUNT=4",
		Reminders: []int{15},
	})
	if err != nil {
		log.Fatal(err)
	}
	// Recurring events need a scope when changed; Clear removes a field.
	title := "주간 회의 (장소 변경)"
	err = c.Calendar.UpdateEvent(context.Background(), "USER_ACCESS_TOKEN", id,
		kakao.EventChanges{Title: &title, Clear: []string{"location"}},
		&kakao.EventUpdateOptions{RecurUpdateType: kakao.RecurAll})
	if err != nil {
		log.Fatal(err)
	}
}

func ExamplePushService_Send() {
	c, _ := kakao.NewClient(kakao.WithAdminKey(os.Getenv("KAKAO_ADMIN_KEY")))
	err := c.Push.Send(context.Background(), []int64{1234}, kakao.PushMessage{
		APNs: &kakao.APNsMessage{Message: "새 댓글이 달렸습니다.", ReturnURL: "https://example.com/push-failed/SECRET"},
		FCM: &kakao.FCMMessage{
			Priority:     kakao.FCMHigh,
			Notification: map[string]any{"title": "새 댓글", "body": "확인해 보세요"},
			ReturnURL:    "https://example.com/push-failed/SECRET",
		},
	}, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleBusinessService_AuthCodeURL() {
	c, _ := kakao.NewClient(kakao.WithRESTKey("REST_API_KEY"))
	u, err := c.Business.AuthCodeURL("https://example.com/biz/callback",
		[]string{"moment_management", "moment_create"},
		&kakao.BusinessAuthOptions{State: "random-state", ResourceIDs: []string{"moment:*"}})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(strings.Contains(u, "resource_ids=moment%3A%2A"))
	// Output: true
}
