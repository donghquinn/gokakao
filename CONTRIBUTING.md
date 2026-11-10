# Contributing

Thanks for helping out. Issues and pull requests are welcome.

## Development

```sh
make test   # go test -race -shuffle=on ./...
make lint   # golangci-lint run
make vet
make fuzz   # fuzz the error parser for 30s
```

## Guidelines

- **Standard library only.** The root package must not gain dependencies.
  Optional integrations (e.g. `golang.org/x/oauth2`) go in their own subpackage.
- **One file per Kakao product** (`search.go`, `local.go`, `user.go`, ...),
  with a matching `_test.go`.
- **Link the source.** Every exported method's doc comment links the Kakao
  doc section it implements.
- **Test against the fake server.** Use `internal/kakaotest`. Assert the
  method, path, auth header and encoded body, and decode a response copied
  from the Kakao docs into `testdata/`.
- **Never retry side effects.** Only mark a non-GET request `idempotent`
  when repeating it is harmless.
- **Never log credentials.** Keys and tokens must not appear in errors or logs.

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
  `Clear []string` field to send JSON nulls.
- **IDs:** numeric Kakao IDs are `int64`, decoded leniently when Kakao sends
  strings. Opaque IDs (friend UUIDs, event IDs) are `string`.
- **Times:** response timestamps are `kakao.Time`; request times are
  `time.Time`, sent in UTC.
- **Enums:** request-side enum strings get a named type with constants.
- **Pagination:** page types are named `XPage`; add an `AllX` iterator
  (`iter.Seq2[T, error]`) when Kakao pages the results.

## Commits

Use [Conventional Commits](https://www.conventionalcommits.org/)
(`feat(search): add book search`, `fix: ...`) so the changelog can be generated.
