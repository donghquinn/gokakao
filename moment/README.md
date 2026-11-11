# gokakao/moment

**English** | [한국어](./README.ko.md)

[![Go Reference](https://pkg.go.dev/badge/github.com/donghquinn/gokakao/moment.svg)](https://pkg.go.dev/github.com/donghquinn/gokakao/moment)

A Go client for the [Kakao Moment (카카오모먼트) API](https://developers.kakao.com/docs/ko/kakaomoment/common),
built on [gokakao](../README.md). It is a separate module on v0.x, so changes in
the Kakao Moment API never force a new major version of the core library.

```sh
go get github.com/donghquinn/gokakao/moment
```

## Usage

Get a business token with Business Auth (`kakao.Client.Business`), then:

```go
kc, _ := kakao.NewClient(kakao.WithRESTKey(restKey), kakao.WithClientSecret(secret))
m := moment.New(kc, businessToken)

accounts, err := m.AdAccounts.List(ctx, nil)
campaigns, err := m.Campaigns.List(ctx, adAccountID, nil)
report, err := m.Reports.Campaigns(ctx, adAccountID, []int64{campaignID}, moment.ReportQuery{...})
```

Calls scoped to an ad account take its ID, which is sent as the `adAccountId` header.
Errors are `*kakao.Error`; Kakao Moment's detailed error is in `Extras`:

```go
var e *kakao.Error
if errors.As(err, &e) && e.Extras != nil {
	log.Println(e.Extras.DetailCode, e.Extras.DetailMsg) // e.g. 21006 존재하지 않는 광고계정입니다.
}
```

## Coverage

All 145 endpoints: ad accounts, campaigns, display/message/personalized
message ad groups, targeting lookups, creatives (display, message,
personalized message), bizboard landings, reports, messages, message ads
(including optimized message ads), ad views, business forms, Pixel & SDK,
audiences, ad reaction targets, customer data, customer files and friend
groups.

The app needs permission for the Kakao Moment API (사용 권한 신청).

Note that several Kakao Moment `PUT` updates replace the whole object: fields you
leave out are reset to their defaults. Those methods say so, and the matching
types have a `Changes()` method that starts from the current values:

```go
c, _ := m.Campaigns.Get(ctx, adAccountID, campaignID)
changes := c.Changes()
changes.Name = "새 이름"
_, err := m.Campaigns.Update(ctx, adAccountID, campaignID, changes)
```

Kakao rate-limits many endpoints (e.g. one create per 5 seconds); the limits are
in each method's documentation.
