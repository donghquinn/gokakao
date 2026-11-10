# gokakao

[![Go Reference](https://pkg.go.dev/badge/github.com/donghquinn/gokakao.svg)](https://pkg.go.dev/github.com/donghquinn/gokakao)
[![Test](https://github.com/donghquinn/gokakao/actions/workflows/test.yml/badge.svg)](https://github.com/donghquinn/gokakao/actions/workflows/test.yml)

A Go client for the [Kakao Developers](https://developers.kakao.com/docs/ko) REST APIs.

> **Status: early development (v0.x).** The API may change before v1.0.0.
> See [PLAN.md](PLAN.md) for the roadmap.

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
img, err := c.Map.StaticMap(ctx, &kakao.StaticMapParams{Width: 640, Height: 480, Center: &pangyo})
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
| Kakao Login / OpenID Connect | planned (v0.2) |
| KakaoTalk Message, Social, Channel | planned (v0.3) |
| Talk Calendar, Push | planned (v0.4) |
| Webhooks | planned (v0.5) |
| Business Auth | planned (v0.6) |
| Kakao Moment, Keyword Ad | planned (after v1.0) |

## Testing

```sh
make test                                                      # unit tests, no network
KAKAO_REST_KEY=... go test -tags=integration -run Integration . # live API
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[MIT](LICENSE). This project is not affiliated with Kakao Corp.
