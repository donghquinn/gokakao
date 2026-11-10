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
| Daum Search | planned (v0.1) |
| Kakao Map (Local) | planned (v0.1) |
| Kakao Login / OpenID Connect | planned (v0.2) |
| KakaoTalk Message, Social, Channel | planned (v0.3) |
| Talk Calendar, Push | planned (v0.4) |
| Webhooks | planned (v0.5) |
| Business Auth | planned (v0.6) |
| Kakao Moment, Keyword Ad | planned (after v1.0) |

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[MIT](LICENSE). This project is not affiliated with Kakao Corp.
