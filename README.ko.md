# gokakao

[English](./README.md) | **한국어**

[![Go Reference](https://pkg.go.dev/badge/github.com/donghquinn/gokakao.svg)](https://pkg.go.dev/github.com/donghquinn/gokakao)
[![Go Version](https://img.shields.io/github/go-mod/go-version/donghquinn/gokakao)](https://go.dev/)
[![Test](https://github.com/donghquinn/gokakao/actions/workflows/test.yml/badge.svg)](https://github.com/donghquinn/gokakao/actions/workflows/test.yml)
[![License](https://img.shields.io/github/license/donghquinn/gokakao)](./LICENSE)

[카카오디벨로퍼스](https://developers.kakao.com/docs/ko) REST API를 위한 Go 클라이언트입니다. 카카오 로그인과
OpenID Connect, 카카오톡 메시지·소셜·채널, 톡캘린더, 푸시 알림, 카카오맵, Daum 검색, 비즈니스 인증,
웹훅, 그리고 카카오모먼트와 카카오 키워드광고 API를 지원합니다.

> **상태: v0.9.0, 릴리스 후보.** 카카오디벨로퍼스의 모든 REST API를 구현했고,
> 코어 API는 v1.0.0을 목표로 검토를 마쳤습니다. v1.0.0 태그 전까지는 API가 바뀔 수
> 있습니다. 로드맵은 [PLAN.md](PLAN.md)를 참고하세요.

- 표준 라이브러리 외의 의존성이 없습니다
- 카카오 에러 코드별 타입이 있는 에러(`errors.Is(err, kakao.ErrInvalidToken)`)
- 멱등 요청에 한해 선택적으로 재시도(backoff, jitter 적용)
- 모든 호출에서 `context` 지원
- ID 토큰과 보안 이벤트 토큰(SET) 검증(RS256, JWKS 캐시)

## 모듈

| 모듈 | 버전 | 내용 |
|---|---|---|
| `github.com/donghquinn/gokakao` | v0.9.0 | 코어 클라이언트: 로그인, 사용자, 톡, 채널, 캘린더, 푸시, 지도, 검색, 비즈니스 인증, `template`·`webhook` 패키지 |
| [`github.com/donghquinn/gokakao/keywordad`](keywordad/README.ko.md) | v0.1.0 | 카카오 키워드광고 API(엔드포인트 65개) |
| [`github.com/donghquinn/gokakao/moment`](moment/README.ko.md) | v0.1.0 | 카카오모먼트 API(엔드포인트 145개) |

## 설치

```sh
go get github.com/donghquinn/gokakao
```

Go 1.26 이상이 필요합니다.

## 빠른 시작

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	kakao "github.com/donghquinn/gokakao"
)

func main() {
	c, err := kakao.NewClient(kakao.WithRESTKey(os.Getenv("KAKAO_REST_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	res, err := c.Search.Web(context.Background(), "카카오", &kakao.SearchOptions{Size: 3})
	if err != nil {
		log.Fatal(err)
	}
	for _, d := range res.Documents {
		fmt.Println(d.Title, d.URL)
	}
}
```

## 사용법

`Client`는 앱 키를 보관하며 여러 고루틴(goroutine)에서 동시에 사용해도 안전합니다. 앱마다 하나를 만들어 공유합니다.

```go
import kakao "github.com/donghquinn/gokakao"

c, err := kakao.NewClient(
	kakao.WithRESTKey(os.Getenv("KAKAO_REST_KEY")),
	kakao.WithAdminKey(os.Getenv("KAKAO_ADMIN_KEY")),       // optional, server-side only
	kakao.WithClientSecret(os.Getenv("KAKAO_CLIENT_SECRET")), // optional
)
```

사용자 액세스 토큰과 어드민 키를 모두 받는 API는 `Subject`를 인자로 받습니다.
사용자로서 호출하려면 `kakao.AccessToken(token)`을, 앱에 연결된 임의의 사용자에 대해 어드민으로 호출하려면 `kakao.AsAdmin(userID)`를 사용합니다.

### 카카오 로그인

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

전체 서버 예제는 [examples/login](examples/login/main.go)을 참고하세요.

| 메서드 | 카카오 API |
|---|---|
| `Auth.AuthCodeURL`, `Auth.Exchange`, `Auth.Refresh`, `Auth.LogoutURL` | 인가 코드 받기, 토큰 받기/갱신, 카카오계정과 함께 로그아웃 |
| `Auth.VerifyIDToken`, `Auth.IDTokenInfo` | ID 토큰 유효성 검증, ID 토큰 정보 조회 |
| `User.Me`, `User.Users`, `User.UserIDs`, `User.AllUserIDs` | 사용자 정보 (단건/여러 명), 사용자 목록 |
| `User.Logout`, `User.Unlink`, `User.TokenInfo`, `User.Signup` | 로그아웃, 연결 끊기, 토큰 정보, 수동 연결 |
| `User.UpdateProfile`, `User.ShippingAddress` | 사용자 프로퍼티 저장, 배송지 |
| `User.Scopes`, `User.RevokeScopes` | 동의 내역 조회, 동의 철회 |
| `User.ServiceTerms`, `User.RevokeServiceTerms`, `User.AgreeServiceTerms` | 서비스 약관 (카카오싱크) |
| `User.OIDCUserInfo` | OIDC 사용자 정보 |

### 카카오톡 메시지, 소셜, 채널

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

친구에게는 한 번에 최대 5명까지 보낼 수 있으며, 일부 친구에게만 실패한 경우 에러가 아니라 `res.FailureInfo`로 결과가 돌아옵니다.

| 메서드 | 카카오 API |
|---|---|
| `Talk.SendToMe`, `SendCustomToMe`, `SendScrapToMe` | 나에게 보내기 (기본 / 사용자 정의 / 스크랩) |
| `Talk.SendToFriends`, `SendCustomToFriends`, `SendScrapToFriends` | 친구에게 보내기 |
| `Talk.Profile`, `Talk.Friends`, `Talk.AllFriends`, `Talk.Profiles` | 카카오톡 프로필, 친구 목록, 여러 사용자 프로필 |
| `Channel.Relations`, `Channel.MultiRelations` | 채널 관계 조회 |
| `Channel.CreateUserFile`, `UserFiles`, `AddUsers`, `RemoveUsers` | 고객 관리 (고객파일) |

템플릿: `template.Feed`, `List`, `Commerce`, `Location`, `Text`, `Calendar`.

### 톡캘린더와 푸시 알림

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

수정 API는 지정하지 않은 필드를 그대로 유지합니다. 값을 지우려면 해당 필드 이름을 `Clear`에 넣습니다.

| 서비스 | 메서드 |
|---|---|
| `Calendar` (사용자) | `Calendars`, `CreateCalendar`, `UpdateCalendar`, `DeleteCalendar`, `CreateEvent`, `Events`, `AllEvents`, `Event`, `UpdateEvent`, `DeleteEvent`, `UpdateGuestEvent`, `FollowPublicEvent`, `Subscribe`, `Unsubscribe` |
| `Calendar` (어드민) | `CreatePublicEvent`, `PublicEvents`, `PublicEvent`, `UpdatePublicEvent`, `DeletePublicEvent`, `SubscribableCalendars`, `Holidays` |
| `Calendar` (할 일) | `CreateTask`, `Tasks`, `Task`, `TaskRecords`, `UpdateTask`, `CompleteTask`, `DeleteTask` |
| `Push` (어드민) | `Register`, `Tokens`, `Deregister`, `Send` |

### 웹훅

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

카카오는 3초 안에 응답을 받아야 하므로, 오래 걸리는 작업은 큐 등으로 넘기고 바로 반환해야 합니다. 계정 상태 변경 웹훅에서 콜백이 에러를 반환하면 503으로 응답하여 카카오가 다시 전송합니다.

| 웹훅 | 핸들러 / 파서 | 인증 |
|---|---|---|
| 연결 해제 (unlink) | `UnlinkHandler` / `ParseUnlink` | 대표 어드민 키 헤더 |
| 계정 상태 변경 (SSF/RISC/CAEP) | `SecurityEventHandler` / `Client.Auth.VerifySecurityEvent` | RS256 서명(JWKS) |
| 카카오톡 채널 | `ChannelHandler` / `ParseChannel` | 대표 어드민 키 헤더 |
| 카카오톡 공유 | `ShareHandler` / `ParseShare` | 대표 어드민 키 헤더 |
| 푸시 전송 실패 | `PushFailureHandler` / `ParsePushFailure` | 없음. `return_url`에 비밀 값을 넣어 확인합니다 |

[examples/webhook](examples/webhook/main.go)을 참고하세요.

### 비즈니스 인증

```go
url, err := c.Business.AuthCodeURL(redirectURI, []string{"moment_management", "moment_create"},
	&kakao.BusinessAuthOptions{State: state, ResourceIDs: []string{"moment:*"}})
tok, err := c.Business.Exchange(ctx, code, redirectURI) // no expiry; valid until revoked
info, err := c.Business.TokenInfo(ctx, tok.AccessToken)
_, err = c.Business.Revoke(ctx, tok.AccessToken)
```

비즈니스 토큰은 만료되지 않으며 철회할 때까지 유효합니다. 비즈니스 사용자 정보 조회(`/v1/business/userinfo`)는 2026년 11월 30일에 카카오가 제공을 종료하므로 지원하지 않습니다.

### Daum 검색

```go
res, err := c.Search.Book(ctx, "미움받을 용기", &kakao.BookSearchOptions{Target: kakao.BookTargetTitle})
for _, b := range res.Documents {
	fmt.Println(b.Title, b.Authors, b.SalePrice)
}
```

`Web`, `Video`, `Image`, `Blog`, `Cafe`, `Book`을 사용할 수 있습니다.

### 카카오맵

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

`Point`의 `X`는 경도, `Y`는 위도입니다.

| 메서드 | 카카오 API |
|---|---|
| `SearchAddress` | 주소 검색 (주소 → 좌표) |
| `CoordToRegionCode` | 좌표 → 행정구역정보 |
| `CoordToAddress` | 좌표 → 주소 |
| `TransformCoord` | 좌표계 변환 |
| `SearchKeyword` / `SearchCategory` | 키워드 / 카테고리 장소 검색 |
| `WalkingRoute` / `CyclingRoute` / `TransitRoute` | 도보 / 자전거 / 대중교통 길찾기 |
| `StaticMap` | 정적 지도 이미지 |

### 에러 처리

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

-401은 토큰 갱신이, -402는 추가 동의 요청이 필요하다는 뜻입니다. `ErrRateLimited`는 -10, HTTP 429, KOE237에 모두 일치합니다.

### 광고 API

```go
import "github.com/donghquinn/gokakao/keywordad"

ka := keywordad.New(c, businessToken) // token from c.Business.Exchange
accounts, err := ka.AdAccounts.List(ctx, nil)
```

[keywordad](keywordad/README.ko.md)와 [moment](moment/README.ko.md)를 참고하세요.

## 지원 API

| 제품 | 상태 |
|---|---|
| Daum 검색 | ✅ |
| 카카오맵 (장소 검색, 좌표, 길찾기, 정적 지도) | ✅ |
| 카카오 로그인 / OpenID Connect | ✅ |
| 카카오톡 메시지, 소셜, 채널 | ✅ |
| 톡캘린더 (할 일 포함), 푸시 알림 | ✅ |
| 웹훅 | ✅ |
| 비즈니스 인증 | ✅ |
| 카카오모먼트 (엔드포인트 145개), 키워드광고 (65개) | ✅ 별도 모듈: [moment](moment/README.ko.md), [keywordad](keywordad/README.ko.md) |

## 테스트

```sh
make test                                                      # unit tests, no network
KAKAO_REST_KEY=... go test -tags=integration -run Integration . # live API
```

실제 API 테스트는 읽기 전용 API만 호출합니다. 어드민 API와 사용자 API까지 확인하려면 `KAKAO_ADMIN_KEY`와 `KAKAO_ACCESS_TOKEN`도 설정합니다. 필요한 인증 정보가 없는 테스트는 건너뜁니다.

## 변경 이력

[CHANGELOG.md](CHANGELOG.md)를 참고하세요(영문).

## 기여하기

[CONTRIBUTING.ko.md](CONTRIBUTING.ko.md)를 참고하세요.

## 라이선스

[MIT](LICENSE). 이 프로젝트는 Kakao Corp.와 관련이 없습니다.
