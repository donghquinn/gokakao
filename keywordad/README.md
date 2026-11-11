# gokakao/keywordad

**English** | [한국어](./README.ko.md)

[![Go Reference](https://pkg.go.dev/badge/github.com/donghquinn/gokakao/keywordad.svg)](https://pkg.go.dev/github.com/donghquinn/gokakao/keywordad)

A Go client for the [Kakao Keyword Ad (카카오 키워드광고) API](https://developers.kakao.com/docs/ko/keyword-ad/common),
built on [gokakao](../README.md). It is a separate module on v0.x, so changes in
the Keyword Ad API never force a new major version of the core library.

```sh
go get github.com/donghquinn/gokakao/keywordad
```

## Usage

Get a business token with Business Auth (`kakao.Client.Business`), then:

```go
kc, _ := kakao.NewClient(kakao.WithRESTKey(restKey), kakao.WithClientSecret(secret))
ka := keywordad.New(kc, businessToken)

accounts, err := ka.AdAccounts.List(ctx, nil)
campaigns, err := ka.Campaigns.List(ctx, adAccountID, nil)
report, err := ka.Reports.Campaigns(ctx, adAccountID, keywordad.ReportQuery{...})
```

Calls scoped to an ad account take its ID, which is sent as the `adAccountId` header.
Errors are `*kakao.Error`; Keyword Ad's detailed error is in `Extras`:

```go
var e *kakao.Error
if errors.As(err, &e) && e.Extras != nil {
	log.Println(e.Extras.DetailCode, e.Extras.DetailMsg) // e.g. 103005 없는 비즈채널입니다.
}
```

## Coverage

All 65 endpoints: ad accounts, biz channels, campaigns, ad groups, keywords,
creatives, creative links, images, review documents, Pixel & SDK, business info
and reports.

Note that several Keyword Ad `PUT` updates replace the whole object: fields you
leave out are reset to their defaults. Those methods say so, and the matching
types have a `Changes()` method that starts from the current values:

```go
g, _ := ka.AdGroups.Get(ctx, adAccountID, adGroupID)
changes := g.Changes()
changes.Name = "새 이름"
_, err := ka.AdGroups.Update(ctx, adAccountID, adGroupID, changes)
```

Kakao rate-limits many endpoints (e.g. one create per 5 seconds); the limits are
in each method's documentation.
