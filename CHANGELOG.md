# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project uses
[Semantic Versioning](https://semver.org/).

## [Unreleased]

### Changed

API review before v1.0.0 (breaking against earlier development commits):

- Security events: `EventDetail` → `SecurityEventDetail`, `EventSubject` →
  `SecurityEventSubject`, and the event constants are typed
  `SecurityEventType` values named `SecurityEvent…`. `SecurityEvent.AppID`
  is an `int64`.
- Webhook callbacks return an `error`, and handlers accept options such as
  `WithErrorHandler`. A callback error makes Kakao retry security events
  (503); unlink callbacks still reply 200 as Kakao requires.
- `template.Template` can no longer be implemented outside the package.
- Push: `Device.UUID`/`PushToken.UUID` → `UserID int64`; `Tokens` takes a
  slice; `Deregister` takes `*DeregisterOptions`; `APNsEnv` and `FCMPriority`
  types. `webhook.PushFailure.UserID` and `webhook.UnlinkEvent.AppID` are
  `int64`.
- Calendar: `EventTime.StartAt`/`EndAt` are `time.Time` (`FormatEventTime`
  removed); `UpdateEvent`, `UpdateGuestEvent` and `DeleteEvent` take changes
  plus `*…Options`; `PublicEvents` takes `PublicEventsQuery`; `Task.TaskID` →
  `Task.ID` with a `TaskStatus` type; the list filter is `TaskStatusFilter`
  and `TasksQuery.Conditions`.
- Map: `TransformCoord` takes `*CoordOptions`; `StaticMap` takes
  `StaticMapParams` by value with an `ImageFormat` type.
- `Talk.Profiles` takes `*ProfilesOptions`; `UserIDPage` → `UserIDsPage`;
  `AuthService.TokenSource` takes `*TokenSourceOptions`;
  `ShippingAddress.UpdatedAt` and `webhook.ChannelEvent.UpdatedAt` are
  `kakao.Time`, which now also decodes Unix seconds.

### Added

- Advertising modules `github.com/donghquinn/gokakao/keywordad` (all 65 Kakao
  Keyword Ad endpoints) and `github.com/donghquinn/gokakao/moment` (all 145
  Kakao Moment endpoints), each on its own v0.x version.
- Core support for them: `HostMoment`, `HostKeywordAd`, `Client.Do` and
  `Request` (custom headers, multipart bodies with per-part headers),
  `Error.Extras` for detailed ad API errors, and the lenient `Int64` and `ID`
  types. `Time` now also parses "+0000" offsets and offset-less timestamps
  (read as KST).
- CI: API compatibility check on pull requests, nightly live tests, and
  GitHub releases on version tags.
- Examples for every service.

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
