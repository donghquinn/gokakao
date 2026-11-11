# 기여하기

[English](./CONTRIBUTING.md) | **한국어**

기여해 주셔서 감사합니다. 이슈와 풀 리퀘스트를 환영합니다.

## 개발

```sh
make test   # go test -race -shuffle=on ./...
make lint   # golangci-lint run
make vet
make fuzz   # fuzz the error parser for 30s
```

`make` 타깃은 코어 모듈만 다룹니다. 광고 모듈은 해당 디렉터리에서 Go 명령을 직접 실행합니다(예: `cd keywordad && go test -race ./...`).

### 모듈

이 저장소에는 코어(`.`), `keywordad`, `moment` 세 모듈이 있습니다.
광고 모듈은 릴리스된 코어 버전을 의존성으로 사용합니다. 코어와 광고 모듈을 함께 수정하려면 로컬 워크스페이스를 사용합니다(커밋하지 않습니다).

```sh
go work init . ./keywordad ./moment
```

코어를 먼저 릴리스한 다음, 광고 모듈의 `require` 줄을 올리고 `keywordad/vX.Y.Z`, `moment/vX.Y.Z` 형식으로 태그를 붙입니다.

## 가이드라인

- **표준 라이브러리만 사용합니다.** 루트 패키지에 의존성을 추가하면 안 됩니다.
  선택적 연동(예: `golang.org/x/oauth2`)은 별도의 하위 패키지에 둡니다.
- **카카오 제품 또는 문서 페이지마다 파일 하나**(`search.go`, `maps.go`,
  `user.go`, ...)를 두고, 같은 이름의 `_test.go`를 함께 둡니다.
- **출처를 연결합니다.** 공개 메서드의 문서 주석에는 구현한 카카오 문서 링크를 넣습니다.
- **가짜 서버로 테스트합니다.** `internal/kakaotest`를 사용합니다. 메서드, 경로,
  인증 헤더, 인코딩된 본문을 검증하고, 카카오 문서에서 복사한 응답을 `testdata/`에 두고
  디코딩해 확인합니다.
- **부수 효과가 있는 요청은 재시도하지 않습니다.** 반복해도 무해한 경우에만 GET 이외의
  요청에 `idempotent`를 지정합니다.
- **인증 정보를 로그에 남기지 않습니다.** 키와 토큰이 에러나 로그에 나타나면 안 됩니다.
- **문서는 두 언어로 유지합니다.** `README.md`, `CONTRIBUTING.md`와 모듈 README에는
  한국어판(`*.ko.md`)이 있습니다. 같은 풀 리퀘스트에서 두 언어를 함께 수정하고, 제목 구성과
  코드 블록을 똑같이 유지합니다. 문서 주석과 `CHANGELOG.md`는 영어로만 작성합니다.

## API 규칙

API를 예측 가능하게 유지하기 위해 새 메서드는 반드시 다음 규칙을 따라야 합니다.

- **시그니처 순서:** `ctx`, 인증 정보(`accessToken string` 또는 `Subject`),
  필수 인자, 선택 인자 순입니다.
- **필수 입력**은 단순한 값 한두 개라면 위치 인자로 받고, 그보다 많으면
  `XQuery`(조회), `NewX`(생성), `XChanges`(수정)라는 이름의 값 구조체로 받습니다.
- **선택 입력**은 마지막 인자인 nil 허용 `*XOptions` 구조체에 둡니다. 지금은 옵션이
  하나뿐이더라도 이렇게 해야 호출하는 코드를 깨뜨리지 않고 옵션을 추가할 수 있습니다.
- **수정**은 포인터 필드로 지정하지 않은 값을 유지하고, `Clear []string` 필드로
  JSON null을 보냅니다. 카카오의 수정 API가 객체 전체를 교체하는 경우(광고 API에서 흔함)에는
  문서 주석에 그 사실을 밝히고, 응답 타입에 현재 값을 돌려주는 `Changes()` 메서드를 추가합니다.
- **ID:** 숫자형 카카오 ID는 요청에서 `int64`, 응답에서 `kakao.Int64`를 사용합니다.
  `kakao.Int64`는 숫자 문자열도 디코딩합니다. 불투명한 ID(친구 UUID, 일정 ID)는
  `string`이며, 카카오가 문자열과 숫자 중 어느 쪽으로든 보내는 ID에는 `kakao.ID`를 사용합니다.
- **시간:** 응답의 타임스탬프는 `kakao.Time`, 요청 시간은 `time.Time`이며 UTC로 보냅니다.
- **열거형:** 요청에 쓰이는 열거형 문자열에는 상수가 딸린 이름 있는 타입을 만듭니다.
- **페이지네이션:** 페이지 타입 이름은 `XPage`로 하고, 카카오가 결과를 페이지로 나누는
  경우 `AllX` 이터레이터(`iter.Seq2[T, error]`)를 추가합니다.

## 커밋

변경 이력을 생성할 수 있도록 [Conventional Commits](https://www.conventionalcommits.org/)
(`feat(search): add book search`, `fix: ...`) 형식을 사용합니다.
