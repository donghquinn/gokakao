# gokakao/moment

[English](./README.md) | **한국어**

[![Go Reference](https://pkg.go.dev/badge/github.com/donghquinn/gokakao/moment.svg)](https://pkg.go.dev/github.com/donghquinn/gokakao/moment)

[gokakao](../README.ko.md)를 기반으로 한 [카카오모먼트 API](https://developers.kakao.com/docs/ko/kakaomoment/common)용 Go 클라이언트입니다.
v0.x 버전의 별도 모듈이므로, 카카오모먼트 API가 바뀌어도 코어 라이브러리의 메이저 버전을 올릴 필요가 없습니다.

```sh
go get github.com/donghquinn/gokakao/moment
```

## 사용법

비즈니스 인증(`kakao.Client.Business`)으로 비즈니스 토큰을 발급받은 다음 사용합니다.

```go
kc, _ := kakao.NewClient(kakao.WithRESTKey(restKey), kakao.WithClientSecret(secret))
m := moment.New(kc, businessToken)

accounts, err := m.AdAccounts.List(ctx, nil)
campaigns, err := m.Campaigns.List(ctx, adAccountID, nil)
report, err := m.Reports.Campaigns(ctx, adAccountID, []int64{campaignID}, moment.ReportQuery{...})
```

광고계정 단위의 호출은 광고계정 ID를 인자로 받으며, 이 값은 `adAccountId` 헤더로 전송됩니다.
에러는 `*kakao.Error`이며, 카카오모먼트의 상세 에러는 `Extras`에 들어 있습니다.

```go
var e *kakao.Error
if errors.As(err, &e) && e.Extras != nil {
	log.Println(e.Extras.DetailCode, e.Extras.DetailMsg) // e.g. 21006 존재하지 않는 광고계정입니다.
}
```

## 지원 범위

엔드포인트 145개 전부를 지원합니다. 광고계정, 캠페인, 디스플레이·메시지·개인화 메시지
광고그룹, 타겟 조회, 소재(디스플레이, 메시지, 개인화 메시지), 비즈보드 랜딩, 보고서, 메시지,
메시지광고(최적화 메시지광고 포함), 애드뷰, 비즈니스폼, 픽셀 & SDK, 오디언스, 광고반응타겟,
고객데이터, 고객파일, 친구그룹이 포함됩니다.

카카오모먼트 API를 사용하려면 앱에 사용 권한이 있어야 합니다(사용 권한 신청).

카카오모먼트의 일부 `PUT` 수정 API는 객체 전체를 교체합니다. 지정하지 않은 필드는 기본값으로
초기화됩니다. 해당 메서드의 문서에 이 사실이 적혀 있으며, 대응하는 타입에는 현재 값에서
시작하는 `Changes()` 메서드가 있습니다.

```go
c, _ := m.Campaigns.Get(ctx, adAccountID, campaignID)
changes := c.Changes()
changes.Name = "새 이름"
_, err := m.Campaigns.Update(ctx, adAccountID, campaignID, changes)
```

카카오는 많은 엔드포인트에 호출 횟수 제한을 둡니다(예: 생성은 5초에 1회). 제한은 각 메서드의
문서에 적혀 있습니다.
