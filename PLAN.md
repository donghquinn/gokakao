# gokakao — Implementation Plan

A Go client library for the server-side Kakao Developers REST APIs
(https://developers.kakao.com/docs/ko).

## 1. Scope

### In scope (REST APIs a Go server can call)

| # | Product | Host | Main auth | Endpoints |
|---|---|---|---|---|
| 1 | Kakao Login (OAuth 2.0 / OIDC) | kauth, kapi | REST key + client secret, user token, admin key | authorize URL, token, refresh, logout, unlink, token info, user/me, app users, user ids, update profile, shipping address, consents/terms |
| 2 | KakaoTalk Message | kapi | user token | send to self / to friends × default / custom / scrap template |
| 3 | KakaoTalk Social | kapi | user token, admin key | my profile, friends list, multi-user profiles |
| 4 | KakaoTalk Channel | kapi | user token, admin key, REST key | channel relationship (single / multi), customer files (create, list, add users, remove users) |
| 5 | Talk Calendar | kapi | user token, admin key | calendars CRUD, events CRUD, public events, follow, subscribable calendars, holidays |
| 6 | Push Notification v2 | kapi | admin key | register, tokens, deregister, send |
| 7 | Kakao Map (Local) | dapi | REST key | address search, coord→region, coord→address, transcoord, keyword search, category search, walk/bicycle/transit routing, static map |
| 8 | Daum Search | dapi | REST key | web, vclip, image, blog, book (v3), cafe |
| 9 | Business Auth | kauth, kapi | business token | authorize URL, token, token info, revoke |
| 10 | Webhooks (inbound) | — | primary admin key / signed JWT | unlink, KakaoTalk channel, KakaoTalk share, account status (SSF: RISC/CAEP/OAuth SET) |
| 11 | Kakao Moment (ads) | apis.moment.kakao.com | business token | ~150 endpoints in ~20 groups |
| 12 | Kakao Keyword Ad | api.keywordad.kakao.com | business token | accounts, campaigns, ad groups, keywords, creatives, reports, … |

### Out of scope

- **No REST API, SDK only:** sending KakaoTalk Share, the JS/Android/iOS map SDKs, Kakao Navi SDKs, the JS-only login and picker flows.
- **Ending soon:** `GET /v1/business/userinfo` ends on 2026-11-30, and `GET /v1/user/shipping_address` is marked for removal. Use the v2 shipping address API instead.
- **Old versions:** Push v1, and Talk Channel relationship v1 (replaced by v2).
- **Kakao Navi REST:** it is hosted at developers.kakaomobility.com, which is a separate product. It could be a later add-on.
- **Translate, Vision, Pose, Pay, Cert:** they are no longer in the docs index.

## 2. Key design decisions

1. **Module path:** `github.com/donghquinn/gokakao`. The current `go.mod` says `goakakao`, which is a typo, so fix it before anyone imports the module.
2. **`go` directive:** set it to the oldest Go release that is still supported (`go 1.26`), not `1.27.3`, so more people can adopt the library.
3. **Zero runtime dependencies:** stdlib only (`net/http`, `encoding/json`, `crypto/rsa`, `crypto/x509`). The JWT and JWKS checks for OIDC `id_token` and SSF SET are small enough to write ourselves, so we avoid the risk of pulling in a jwt library.
4. **Package layout**

   ```
   gokakao/                 package kakao: Client, options, Error, transport
   ├── oauth.go user.go message.go social.go channel.go calendar.go push.go
   ├── local.go search.go bizauth.go
   ├── template/            message template builders (feed, list, location, commerce, text, calendar)
   ├── webhook/             http.Handlers and verifiers for inbound callbacks
   ├── moment/              Kakao Moment ads API (large; separate package)
   ├── keywordad/           Keyword Ad API (large; separate package)
   ├── internal/kakaotest/  test server, fixture loader
   └── examples/            runnable login server, send-to-me, local search, webhook receiver
   ```
   Platform APIs live in one package with a service per product, like go-github (`c.Local.SearchAddress(...)`). The ads APIs are big and change on their own schedule, so they get subpackages.

5. **Auth model.** App credentials live on the `Client`. A user API that can run either as the user or as the admin takes a `Subject`: `kakao.AccessToken(tok)` sends a Bearer token, and `kakao.AsAdmin(userID)` sends the admin key plus `target_id_type`/`target_id`. Endpoints that accept only one credential take it directly.
   ```go
   c, _ := kakao.NewClient(kakao.WithRESTKey(restKey), kakao.WithAdminKey(adminKey))
   c.Map.SearchKeyword(ctx, "판교 카페", nil)            // REST key
   c.User.Me(ctx, kakao.AccessToken(tok), nil)          // as the user
   c.User.Unlink(ctx, kakao.AsAdmin(userID))            // as admin
   c.User.Users(ctx, ids, nil)                          // admin only
   ```

6. **Transport core** (`do(ctx, req, out)`):
   - Encodes form, query, or JSON bodies per endpoint. `template_object` and the calendar `event` are JSON strings inside form fields.
   - Decodes errors into `*kakao.Error{HTTPStatus, Code, Msg, RequiredScopes}`, with sentinel helpers such as `errors.Is(err, kakao.ErrInvalidToken)` (-401), `ErrConsentRequired` (-402), and `ErrRateLimited` (-10 or HTTP 429). OAuth errors (`KOE...`, `error`/`error_description`) get their own `*OAuthError`.
   - Reports partial message-send failures (`failure_info` in a 200 response) as typed result data, not as an error.
   - Opt-in retries only, applied to idempotent GETs on -1, -7, -9798, 429, or 5xx, with backoff and jitter. Never retries sends.
   - Sets a `User-Agent: gokakao/<version>` header. The base URL can be overridden for tests.
7. **Pagination:** use Go 1.23 iterators (`iter.Seq2[T, error]`) for friends, user IDs, calendar events, and search pages, alongside the plain page calls.
8. **Token handling:** a `TokenSource`-style helper that refreshes on expiry and respects the KOE237 limits (20 access tokens per 10 min, 30 refreshes per hour). Optional `golang.org/x/oauth2` interop through an exported `kakao.Endpoint`, in an `oauth2kakao/` subpackage so the core keeps zero dependencies.

## 3. Phases and releases

| Phase | Deliverable | Release |
|---|---|---|
| **0. Foundation** ✅ | Fix the module path. Add LICENSE (MIT or Apache-2.0), README, CONTRIBUTING. Build the transport core, the error types, and the `kakaotest` fake server. CI: test with `-race` on the last two Go versions, golangci-lint, govulncheck, coverage. | — |
| **1. Quick wins** ✅ | Daum Search, Kakao Map Local (address, coord, transcoord, keyword, category), static map, routing. These use only the REST key, so they test the core early. | v0.1.0 |
| **2. Kakao Login** ✅ | Authorize/logout URL builders with `state`, `nonce`, and `prompt`. Token exchange and refresh. OIDC `id_token` verification (JWKS cache, RS256, iss/aud/nonce checks). Token info, me, update profile, logout, unlink, app users, user ID iterator, shipping address v2, consents and terms. TokenSource. Example login server. | v0.2.0 |
| **3. Talk** ✅ | `template/` builders for all default template types. Message send to self and to friends (default, custom, scrap). Social: profile, friends iterator, multi-profiles. Channel: relationship v2 single and multi, customer files. | v0.3.0 |
| **4. Calendar + Push** | Talk Calendar: user calendars and events, public events, follow, subscribable calendars, holidays. Push v2: register, tokens, deregister, send, plus typed `for_apns` / `for_fcm` payloads. | v0.4.0 |
| **5. Webhooks** | `webhook.UnlinkHandler`, `ChannelHandler`, `ShareHandler` (all verify `Authorization: KakaoAK <primary key>`). An SSF handler that verifies the SET JWT against `/.well-known/ssf-configuration` JWKS, parses typed events (RISC, CAEP, OAuth, KAKAO), and returns 202 on success and 400 on failure. Handlers answer within 3 s by handing work off. | v0.5.0 |
| **6. Business Auth** | Business authorize URL, token, token info, revoke. Shared by the ads packages. | v0.6.0 |
| **7. Hardening → v1.0** | Live integration suite against a real test app, gated by `KAKAO_*` environment variables. Godoc examples for every service. API review and freeze. `apidiff` in CI. | **v1.0.0** |
| **8. Ads** | `moment/` and `keywordad/`, built group by group, honoring each endpoint's rate limit (a per-endpoint limiter). Possibly separate Go modules so their changes don't force major version bumps on the core. | v1.x (or `moment/v0.x`) |

## 4. Testing strategy

- **Unit:** `httptest` fake server per service. Request assertions cover method, path, auth header, and the encoded body. Response fixtures are copied from the doc examples into `testdata/`.
- **Error mapping:** a table-driven test for every documented code.
- **JWT/JWKS:** use self-generated RSA keys for the happy path, wrong aud/iss, expiry, `kid` rotation, and bad algorithm cases.
- **Live (opt-in):** `go test -tags=integration`, using a dedicated Kakao test app. Nightly CI job with secrets. Only safe calls (search, local, me, send-to-me).
- **Targets:** at least 85% coverage for the core and services. Fuzz the error and response decoders.

## 5. Release and maintenance

- Semantic versioning with git tags. Stay on v0.x until phase 7. release-please or a hand-written CHANGELOG. pkg.go.dev picks up tags automatically, and goreleaser isn't needed for a library.
- Dependabot for GitHub Actions only, since there are no Go dependencies.
- Docs: a README quickstart per product, a table mapping each Kakao doc page to a Go method, and a deprecation policy that follows Kakao's 제공 종료 notices.
- Add a `SECURITY.md`. Never log tokens or keys, and redact `Authorization` in debug hooks.

## 6. Open items to resolve during implementation

- Fetch the exact paths and params for **consents, terms, the calendar holiday API, the IP list API**, and the **Moment/Keyword Ad auth headers**. These were not captured in the first docs pass.
- Confirm whether `/v2/api/talk/memo/send` (custom template to self) is the current path.
- Error code -10 is listed as HTTP 400 on one page and 429 on another. Handle both.
- Decide whether to include Kakao Navi (kakaomobility).
