package keywordad

import (
	"net/http"
	"testing"
)

func TestBusinessInfoService_Get(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /openapi/v1/companies", 200, `{"id":"000","businessRegistrationNumber":"120-81-47521","name":"주식회사 사업자","representationName":"사업자대표","businessCategory":"서비스","businessItem":"서비스","address":{"zipCode":"13494","baseAddress":"경기 성남시 분당구 판교역로 235","detailAddress":"H스퀘어"}}`)

	b, err := c.BusinessInfo.Get(t.Context(), 1111111111, "120-81-47521")
	if err != nil {
		t.Fatal(err)
	}
	req := assertRequest(t, srv, "1111111111")
	if req.Method != http.MethodGet || req.Path != "/openapi/v1/companies" || req.Query.Get("bizNum") != "120-81-47521" || len(req.Query) != 1 {
		t.Errorf("request = %s %s?%v", req.Method, req.Path, req.Query)
	}
	want := BusinessInfo{
		BusinessRegistrationNumber: "120-81-47521", Name: "주식회사 사업자", RepresentationName: "사업자대표",
		BusinessCategory: "서비스", BusinessItem: "서비스",
		Address: Address{ZipCode: "13494", BaseAddress: "경기 성남시 분당구 판교역로 235", DetailAddress: "H스퀘어"},
	}
	if *b != want {
		t.Errorf("business = %+v", b)
	}

	if _, err := c.BusinessInfo.Get(t.Context(), 1111111111, ""); err == nil {
		t.Error("missing number: want error")
	}
}
