# gokakao

[![Go Reference](https://pkg.go.dev/badge/github.com/donghquinn/gokakao.svg)](https://pkg.go.dev/github.com/donghquinn/gokakao)
[![Test](https://github.com/donghquinn/gokakao/actions/workflows/test.yml/badge.svg)](https://github.com/donghquinn/gokakao/actions/workflows/test.yml)

A Go client for the [Kakao Developers](https://developers.kakao.com/docs/ko) REST APIs.

> **Status: release candidate.** All non-advertising Kakao Developers REST APIs
> are implemented and the API has been reviewed for v1.0.0. It may still change
> until v1.0.0 is tagged. See [PLAN.md](PLAN.md) for the roadmap.

- No dependencies outside the standard library
- Typed errors for Kakao error codes (`errors.Is(err, kakao.ErrInvalidToken)`)
- Opt-in retries for idempotent requests, with backoff and jitter
- Context support on every call

## Install

```sh
go get github.com/donghquinn/gokakao
```

Requires Go 1.26 or later.

## Usage

```go
import kakao "github.com/donghquinn/gokakao"

c, err := kakao.NewClient(
	kakao.WithRESTKey(os.Getenv("KAKAO_REST_KEY")),
	kakao.WithAdminKey(os.Getenv("KAKAO_ADMIN_KEY")),       // optional, server-side only
	kakao.WithClientSecret(os.Getenv("KAKAO_CLIENT_SECRET")), // optional
)
```

### Kakao Login

```go
// 1. Redirect the user to Kakao.
verifier := kakao.GenerateVerifier()
loginURL, err := c.Auth.AuthCodeURL("https://example.com/callback", &kakao.AuthCodeOptions{
	State: state, Nonce: nonce, CodeChallenge: kakao.S256Challenge(verifier),
})

// 2. In the callback, exchange the code and verify the ID token.
tok, err := c.Auth.Exchange(ctx, code, "https://example.com/callback", &kakao.ExchangeOptions{CodeVerifier: verifier})
claims, err := c.Auth.VerifyIDToken(ctx, tok.IDToken, nonce) // signature, iss, aud, exp, nonce

// 3. Call user APIs as the user, or as admin for any linked user.
me, err := c.User.Me(ctx, kakao.AccessToken(tok.AccessToken), nil)
_, err = c.User.Unlink(ctx, kakao.AsAdmin(userID))

// Keep tokens fresh; onRefresh persists rotated tokens.
ts := c.Auth.TokenSource(tok, &kakao.TokenSourceOptions{OnRefresh: save})
fresh, err := ts.Token(ctx)
```

See [examples/login](examples/login/main.go) for a complete server.

| Method | Kakao API |
|---|---|
| `Auth.AuthCodeURL`, `Auth.Exchange`, `Auth.Refresh`, `Auth.LogoutURL` | 인가 코드 받기, 토큰 받기/갱신, 카카오계정과 함께 로그아웃 |
| `Auth.VerifyIDToken`, `Auth.IDTokenInfo` | ID 토큰 유효성 검증, ID 토큰 정보 조회 |
| `User.Me`, `User.Users`, `User.UserIDs`, `User.AllUserIDs` | 사용자 정보 (단건/여러 명), 사용자 목록 |
| `User.Logout`, `User.Unlink`, `User.TokenInfo`, `User.Signup` | 로그아웃, 연결 끊기, 토큰 정보, 수동 연결 |
| `User.UpdateProfile`, `User.ShippingAddress` | 사용자 프로퍼티 저장, 배송지 |
| `User.Scopes`, `User.RevokeScopes` | 동의 내역 조회, 동의 철회 |
| `User.ServiceTerms`, `User.RevokeServiceTerms`, `User.AgreeServiceTerms` | 서비스 약관 (카카오싱크) |
| `User.OIDCUserInfo` | OIDC 사용자 정보 |

### KakaoTalk Message, Social and Channel

```go
import "github.com/donghquinn/gokakao/template"

msg := template.Feed{
	Content: template.Content{
		Title:    "오늘의 디저트",
		ImageURL: "https://example.com/cake.jpg",
		Link:     template.Link{WebURL: "https://example.com", MobileWebURL: "https://m.example.com"},
	},
	Buttons: []template.Button{{Title: "보러 가기", Link: template.Link{WebURL: "https://example.com"}}},
}
err := c.Talk.SendToMe(ctx, accessToken, msg)

// Message up to 5 friends; partial failures come back in res.FailureInfo.
for f, err := range c.Talk.AllFriends(ctx, accessToken, nil) { ... }
res, err := c.Talk.SendToFriends(ctx, accessToken, []string{friend.UUID}, msg)

// Was the channel added? As the user, or as admin for any user.
rel, err := c.Channel.Relations(ctx, kakao.AsAdmin(userID), "_ZeUTxl")
```

| Method | Kakao API |
|---|---|
| `Talk.SendToMe`, `SendCustomToMe`, `SendScrapToMe` | 나에게 보내기 (기본 / 사용자 정의 / 스크랩) |
| `Talk.SendToFriends`, `SendCustomToFriends`, `SendScrapToFriends` | 친구에게 보내기 |
| `Talk.Profile`, `Talk.Friends`, `Talk.AllFriends`, `Talk.Profiles` | 카카오톡 프로필, 친구 목록, 여러 사용자 프로필 |
| `Channel.Relations`, `Channel.MultiRelations` | 채널 관계 조회 |
| `Channel.CreateUserFile`, `UserFiles`, `AddUsers`, `RemoveUsers` | 고객 관리 (고객파일) |

Templates: `template.Feed`, `List`, `Commerce`, `Location`, `Text`, `Calendar`.

### Talk Calendar and Push

```go
id, err := c.Calendar.CreateEvent(ctx, accessToken, kakao.PrimaryCalendar, kakao.NewEvent{
	Title: "회의",
	Time:  kakao.EventTime{StartAt: start, EndAt: start.Add(time.Hour)},
	Reminders: []int{15},
})
for e, err := range c.Calendar.AllEvents(ctx, accessToken, kakao.EventsQuery{Preset: kakao.PresetThisWeek}) { ... }

// Updates keep omitted fields; list fields to remove in Clear.
title := "회의 (변경)"
err = c.Calendar.UpdateEvent(ctx, accessToken, id,
	kakao.EventChanges{Title: &title, Clear: []string{"location"}}, nil)

err = c.Push.Send(ctx, []int64{userID}, kakao.PushMessage{
	FCM: &kakao.FCMMessage{Notification: map[string]any{"title": "새 댓글", "body": "확인해 보세요"}},
}, nil)
```

| Service | Methods |
|---|---|
| `Calendar` (user) | `Calendars`, `CreateCalendar`, `UpdateCalendar`, `DeleteCalendar`, `CreateEvent`, `Events`, `AllEvents`, `Event`, `UpdateEvent`, `DeleteEvent`, `UpdateGuestEvent`, `FollowPublicEvent`, `Subscribe`, `Unsubscribe` |
| `Calendar` (admin) | `CreatePublicEvent`, `PublicEvents`, `PublicEvent`, `UpdatePublicEvent`, `DeletePublicEvent`, `SubscribableCalendars`, `Holidays` |
| `Calendar` (tasks) | `CreateTask`, `Tasks`, `Task`, `TaskRecords`, `UpdateTask`, `CompleteTask`, `DeleteTask` |
| `Push` (admin) | `Register`, `Tokens`, `Deregister`, `Send` |

### Webhooks

```go
import "github.com/donghquinn/gokakao/webhook"

logErrors := webhook.WithErrorHandler(func(r *http.Request, err error) { log.Print(err) })

mux.Handle("/kakao/unlink", webhook.UnlinkHandler(primaryAdminKey, func(r *http.Request, e *webhook.UnlinkEvent) error {
	return queue.Enqueue(deleteUser(e.UserID)) // reply within 3 s; do slow work elsewhere
}, logErrors))
// Verifies the Security Event Token (RS256, iss, aud) and replies 202 / 400 / 503 as Kakao expects.
mux.Handle("POST /kakao/events", webhook.SecurityEventHandler(c.Auth, func(r *http.Request, e *kakao.SecurityEvent) error {
	if _, ok := e.Events[kakao.SecurityEventSessionsRevoked]; ok {
		return sessions.RevokeAll(e.Subject) // an error makes Kakao retry
	}
	return nil
}, logErrors))
mux.Handle("POST /kakao/channel", webhook.ChannelHandler(primaryAdminKey, onChannelEvent))
mux.Handle("/kakao/share", webhook.ShareHandler(primaryAdminKey, onShare))
mux.Handle("POST /kakao/push-failed/{secret}", webhook.PushFailureHandler(onPushFailure))
```

| Webhook | Handler / parser | Authentication |
|---|---|---|
| 연결 해제 (unlink) | `UnlinkHandler` / `ParseUnlink` | primary admin key header |
| 계정 상태 변경 (SSF/RISC/CAEP) | `SecurityEventHandler` / `Client.Auth.VerifySecurityEvent` | RS256 signature (JWKS) |
| 카카오톡 채널 | `ChannelHandler` / `ParseChannel` | primary admin key header |
| 카카오톡 공유 | `ShareHandler` / `ParseShare` | primary admin key header |
| 푸시 전송 실패 | `PushFailureHandler` / `ParsePushFailure` | none; use a secret in `return_url` |

See [examples/webhook](examples/webhook/main.go).

### Business Auth

```go
url, err := c.Business.AuthCodeURL(redirectURI, []string{"moment_management", "moment_create"},
	&kakao.BusinessAuthOptions{State: state, ResourceIDs: []string{"moment:*"}})
tok, err := c.Business.Exchange(ctx, code, redirectURI) // no expiry; valid until revoked
info, err := c.Business.TokenInfo(ctx, tok.AccessToken)
_, err = c.Business.Revoke(ctx, tok.AccessToken)
```

Business user info (`/v1/business/userinfo`) is not supported; Kakao ends it on 2026-11-30.

### Daum Search

```go
res, err := c.Search.Book(ctx, "미움받을 용기", &kakao.BookSearchOptions{Target: kakao.BookTargetTitle})
for _, b := range res.Documents {
	fmt.Println(b.Title, b.Authors, b.SalePrice)
}
```

`Web`, `Video`, `Image`, `Blog`, `Cafe` and `Book` are available.

### Kakao Map

```go
pangyo := kakao.Point{X: 127.1086228, Y: 37.4012191} // X = longitude, Y = latitude

places, err := c.Map.SearchKeyword(ctx, "카페", &kakao.PlaceSearchOptions{
	Center: &pangyo, Radius: 500, Sort: kakao.PlaceSortDistance,
})
addrs, err := c.Map.SearchAddress(ctx, "판교역로 166", nil)
regions, err := c.Map.CoordToRegionCode(ctx, pangyo, nil)
route, err := c.Map.WalkingRoute(ctx, pangyo, kakao.Point{X: 127.1263, Y: 37.4199}, nil)
img, err := c.Map.StaticMap(ctx, kakao.StaticMapParams{Width: 640, Height: 480, Center: &pangyo})
```

| Method | Kakao API |
|---|---|
| `SearchAddress` | 주소 검색 (address → coordinates) |
| `CoordToRegionCode` | 좌표 → 행정구역정보 |
| `CoordToAddress` | 좌표 → 주소 |
| `TransformCoord` | 좌표계 변환 |
| `SearchKeyword` / `SearchCategory` | 키워드 / 카테고리 장소 검색 |
| `WalkingRoute` / `CyclingRoute` / `TransitRoute` | 도보 / 자전거 / 대중교통 길찾기 |
| `StaticMap` | 정적 지도 이미지 |

### Errors

```go
switch {
case errors.Is(err, kakao.ErrInvalidToken):    // -401: refresh the token
case errors.Is(err, kakao.ErrConsentRequired): // -402: ask for more consent
	var e *kakao.Error
	errors.As(err, &e)
	log.Println("missing scopes:", e.RequiredScopes)
case errors.Is(err, kakao.ErrRateLimited):     // -10, HTTP 429, or KOE237
}
```

## Supported APIs

| Product | Status |
|---|---|
| Daum Search | ✅ |
| Kakao Map (local search, coordinates, routing, static map) | ✅ |
| Kakao Login / OpenID Connect | ✅ |
| KakaoTalk Message, Social, Channel | ✅ |
| Talk Calendar (incl. tasks), Push | ✅ |
| Webhooks | ✅ |
| Business Auth | ✅ |
| Kakao Moment (145 endpoints), Keyword Ad (65) | ✅ separate modules: [moment](moment/README.md), [keywordad](keywordad/README.md) |

## Testing

```sh
make test                                                      # unit tests, no network
KAKAO_REST_KEY=... go test -tags=integration -run Integration . # live API
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[MIT](LICENSE). This project is not affiliated with Kakao Corp.
