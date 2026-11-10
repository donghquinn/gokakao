# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project uses
[Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- Business Auth: `Client.Business` with `AuthCodeURL`, `Exchange`,
  `TokenInfo` and `Revoke`. Business user info is omitted because Kakao ends
  it on 2026-11-30.
- `webhook` package: handlers and parsers for the unlink, account status
  (Security Event Token), KakaoTalk Channel, KakaoTalk Share and push failure
  webhooks, replying with the status codes Kakao requires.
- `AuthService.VerifySecurityEvent` and `SecurityEvent` with constants for all
  OAUTH, RISC, CAEP and KAKAO event types.
- `examples/webhook`, a webhook receiver.
- Talk Calendar: `Client.Calendar` with user calendars, events (with
  `AllEvents` paging and `Clear` for removing fields), channel public events,
  subscription calendars, holidays and tasks.
- Push Notification v2: `Client.Push` with `Register`, `Tokens`, `Deregister`
  and `Send` (APNs and FCM payloads, bypass).
- `template` package: Feed, List, Commerce, Location, Text and Calendar
  message templates.
- KakaoTalk Message and Social: `Client.Talk` with `SendToMe`,
  `SendCustomToMe`, `SendScrapToMe`, `SendToFriends`, `SendCustomToFriends`,
  `SendScrapToFriends`, `Profile`, `Friends`, `AllFriends` and `Profiles`.
- KakaoTalk Channel: `Client.Channel` with `Relations`, `MultiRelations`,
  `CreateUserFile`, `UserFiles`, `AddUsers` and `RemoveUsers`.
- Kakao Login: `Client.Auth` with `AuthCodeURL` (incl. PKCE), `Exchange`,
  `Refresh`, `LogoutURL`, `VerifyIDToken` (RS256 + JWKS cache), `IDTokenInfo`
  and a refreshing `TokenSource`.
- Kakao Login user APIs: `Client.User` with `Me`, `Users`, `UserIDs`,
  `AllUserIDs`, `Logout`, `Unlink`, `TokenInfo`, `UpdateProfile`, `Signup`,
  `ShippingAddress`, `Scopes`, `RevokeScopes`, `ServiceTerms`,
  `RevokeServiceTerms`, `AgreeServiceTerms` and `OIDCUserInfo`.
- `Subject` (`AccessToken`, `AsAdmin`) to choose user-token or admin-key calls.
- `examples/login`, a Kakao Login + OIDC + PKCE web server.
- Daum Search: `Client.Search` with `Web`, `Video`, `Image`, `Blog`, `Cafe`
  and `Book`.
- Kakao Map: `Client.Map` with `SearchAddress`, `CoordToRegionCode`,
  `CoordToAddress`, `TransformCoord`, `SearchKeyword`, `SearchCategory`,
  `WalkingRoute`, `CyclingRoute`, `TransitRoute` and `StaticMap`.
- `Time`, `Float` and `YesNo` types that tolerate Kakao's mixed JSON encodings.
- `Client` with functional options for the REST API key, admin key, client
  secret, HTTP client, base URLs, User-Agent and retries.
- `Error` and `OAuthError` types, and sentinel errors for common Kakao error codes.
- Opt-in retries for idempotent requests, with exponential backoff, full jitter
  and `Retry-After` support.
