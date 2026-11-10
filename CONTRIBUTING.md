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

## Commits

Use [Conventional Commits](https://www.conventionalcommits.org/)
(`feat(search): add book search`, `fix: ...`) so the changelog can be generated.
