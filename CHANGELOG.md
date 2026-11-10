# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project uses
[Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

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
