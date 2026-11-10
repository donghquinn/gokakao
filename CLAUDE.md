# gokakao

Go client library for the Kakao Developers REST APIs. Roadmap: PLAN.md.

## Go skills

Load the `golang-how-to` skill before any Go coding, review, debugging or setup task.

## Conventions

- Root package `kakao`; standard library only.
- Every call goes through `Client.do` (request.go). Pick the credential with `authKind`.
- Tests use `internal/kakaotest` and `newTestClient` (kakao_test.go).
- Run `make test vet` before finishing.
