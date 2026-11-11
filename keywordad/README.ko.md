# gokakao/keywordad

[English](./README.md) | **한국어**

[![Go Reference](https://pkg.go.dev/badge/github.com/donghquinn/gokakao/keywordad.svg)](https://pkg.go.dev/github.com/donghquinn/gokakao/keywordad)

[gokakao](../README.ko.md)를 기반으로 한 [카카오 키워드광고 API](https://developers.kakao.com/docs/ko/keyword-ad/common)용 Go 클라이언트입니다.
v0.x 버전의 별도 모듈이므로, 키워드광고 API가 바뀌어도 코어 라이브러리의 메이저 버전을 올릴 필요가 없습니다.

```sh
go get github.com/donghquinn/gokakao/keywordad
```

## 사용법

비즈니스 인증(`kakao.Client.Business`)으로 비즈니스 토큰을 발급받은 다음 사용합니다.

```go
kc, _ := kakao.NewClient(kakao.WithRESTKey(restKey), kakao.WithClientSecret(secret))
ka := keywordad.New(kc, businessToken)

accounts, err := ka.AdAccounts.List(ctx, nil)
campaigns, err := ka.Campaigns.List(ctx, adAccountID, nil)
report, err := ka.Reports.Campaigns(ctx, adAccountID, keywordad.ReportQuery{...})
```

광고계정 단위의 호출은 광고계정 ID를 인자로 받으며, 이 값은 `adAccountId` 헤더로 전송됩니다.
에러는 `*kakao.Error`이며, 키워드광고의 상세 에러는 `Extras`에 들어 있습니다.

```go
var e *kakao.Error
if errors.As(err, &e) && e.Extras != nil {
	log.Println(e.Extras.DetailCode, e.Extras.DetailMsg) // e.g. 103005 없는 비즈채널입니다.
}
```

## 지원 범위

엔드포인트 65개 전부를 지원합니다. 광고계정, 비즈채널, 캠페인, 광고그룹, 키워드, 소재,
소재연결, 이미지, 심사서류, 픽셀 & SDK, 사업자 정보, 보고서가 포함됩니다.

키워드광고의 일부 `PUT` 수정 API는 객체 전체를 교체합니다. 지정하지 않은 필드는 기본값으로
초기화됩니다. 해당 메서드의 문서에 이 사실이 적혀 있으며, 대응하는 타입에는 현재 값에서
시작하는 `Changes()` 메서드가 있습니다.

```go
g, _ := ka.AdGroups.Get(ctx, adAccountID, adGroupID)
changes := g.Changes()
changes.Name = "새 이름"
_, err := ka.AdGroups.Update(ctx, adAccountID, adGroupID, changes)
```

카카오는 많은 엔드포인트에 호출 횟수 제한을 둡니다(예: 생성은 5초에 1회). 제한은 각 메서드의
문서에 적혀 있습니다.
