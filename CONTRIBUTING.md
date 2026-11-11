# Contributing

**English** | [한국어](./CONTRIBUTING.ko.md)

Thanks for helping out. Issues and pull requests are welcome.

## Development

```sh
make test   # go test -race -shuffle=on ./...
make lint   # golangci-lint run
make vet
make fuzz   # fuzz the error parser for 30s
```

`make` targets cover the core module. For the ad modules, run the Go
commands in their directories, e.g. `cd keywordad && go test -race ./...`.

### Modules

The repository holds three modules: the core (`.`), `keywordad` and `moment`.
The ad modules require a released core version. To change the core and an ad
module together, use a local workspace (not committed):

```sh
go work init . ./keywordad ./moment
```

Release the core first, then bump the ad modules' `require` line and tag them
as `keywordad/vX.Y.Z` and `moment/vX.Y.Z`.

## Guidelines

- **Standard library only.** The root package must not gain dependencies.
  Optional integrations (e.g. `golang.org/x/oauth2`) go in their own subpackage.
- **One file per Kakao product or doc page** (`search.go`, `maps.go`,
  `user.go`, ...), with a matching `_test.go`.
- **Link the source.** Every exported method's doc comment links the Kakao
  doc section it implements.
- **Test against the fake server.** Use `internal/kakaotest`. Assert the
  method, path, auth header and encoded body, and decode a response copied
  from the Kakao docs into `testdata/`.
- **Never retry side effects.** Only mark a non-GET request `idempotent`
  when repeating it is harmless.
- **Never log credentials.** Keys and tokens must not appear in errors or logs.
- **Keep docs bilingual.** `README.md`, `CONTRIBUTING.md` and the module
  READMEs have Korean versions (`*.ko.md`). Update both in the same pull
  request, with the same headings and identical code blocks. Doc comments and
  `CHANGELOG.md` stay English-only.

## API conventions

New methods must follow these so the API stays predictable:

- **Signature order:** `ctx`, then the credential (`accessToken string` or
  `Subject`), then required arguments, then optional ones.
- **Required inputs** are positional for one or two simple values; otherwise
  a value struct named `XQuery` (reads), `NewX` (creates) or `XChanges`
  (updates).
- **Optional inputs** go in a nil-able `*XOptions` struct as the last
  argument, even when there is only one option today, so options can be added
  without breaking callers.
- **Updates** use pointer fields to keep omitted values and a
  `Clear []string` field to send JSON nulls. Where Kakao's update replaces
  the whole object (common in the ad APIs), say so in the doc comment and add
  a `Changes()` method on the response type that returns the current values.
- **IDs:** numeric Kakao IDs are `int64` in requests and `kakao.Int64` in
  responses, which also decodes numeric strings. Opaque IDs (friend UUIDs,
  event IDs) are `string`; use `kakao.ID` when Kakao sends one as either a
  string or a number.
- **Times:** response timestamps are `kakao.Time`; request times are
  `time.Time`, sent in UTC.
- **Enums:** request-side enum strings get a named type with constants.
- **Pagination:** page types are named `XPage`; add an `AllX` iterator
  (`iter.Seq2[T, error]`) when Kakao pages the results.

## Commits

Use [Conventional Commits](https://www.conventionalcommits.org/)
(`feat(search): add book search`, `fix: ...`) so the changelog can be generated.
