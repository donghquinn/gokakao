package kakao

import (
	"errors"
	"net/http"
	"slices"
	"testing"
	"time"
)

func TestSubject(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithAdminKey("admin"))
	srv.HandleJSON("GET /v2/user/me", 200, `{"id":1}`)
	srv.HandleJSON("POST /v1/user/logout", 200, `{"id":1}`)

	if _, err := c.User.Me(t.Context(), AccessToken("AT"), nil); err != nil {
		t.Fatal(err)
	}
	req := srv.LastRequest()
	if req.Header.Get("Authorization") != "Bearer AT" || len(req.Query) != 0 {
		t.Errorf("bearer GET: auth %q, query %v", req.Header.Get("Authorization"), req.Query)
	}

	if _, err := c.User.Me(t.Context(), AsAdmin(1234), nil); err != nil {
		t.Fatal(err)
	}
	req = srv.LastRequest()
	if req.Header.Get("Authorization") != "KakaoAK admin" || req.Query.Get("target_id_type") != "user_id" || req.Query.Get("target_id") != "1234" {
		t.Errorf("admin GET: auth %q, query %v", req.Header.Get("Authorization"), req.Query)
	}

	if _, err := c.User.Logout(t.Context(), AsAdmin(1234)); err != nil {
		t.Fatal(err)
	}
	req = srv.LastRequest()
	if f := req.Form(); f.Get("target_id_type") != "user_id" || f.Get("target_id") != "1234" || len(req.Query) != 0 {
		t.Errorf("admin POST: form %v, query %v", f, req.Query)
	}
	if req.Header.Get("Content-Type") != contentTypeForm {
		t.Errorf("admin POST Content-Type = %q", req.Header.Get("Content-Type"))
	}

	if _, err := c.User.Logout(t.Context(), AccessToken("AT")); err != nil {
		t.Fatal(err)
	}
	if req := srv.LastRequest(); len(req.Body) != 0 {
		t.Errorf("bearer POST sent body %q", req.Body)
	}

	if _, err := c.User.Me(t.Context(), nil, nil); err == nil {
		t.Error("nil subject: want error")
	}
	if _, err := c.User.Me(t.Context(), AccessToken(""), nil); !errors.Is(err, ErrMissingCredential) {
		t.Errorf("empty token: err = %v", err)
	}
	noAdmin, _ := newTestClient(t)
	if _, err := noAdmin.User.Me(t.Context(), AsAdmin(1), nil); !errors.Is(err, ErrMissingCredential) {
		t.Errorf("no admin key: err = %v", err)
	}
}

func TestUserService_Me(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleFile("GET /v2/user/me", "testdata/user/me.json")

	u, err := c.User.Me(t.Context(), AccessToken("AT"), &MeOptions{
		PropertyKeys:   []string{"kakao_account.email", "properties."},
		SecureResource: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	q := srv.LastRequest().Query
	if q.Get("property_keys") != `["kakao_account.email","properties."]` || q.Get("secure_resource") != "true" {
		t.Errorf("query = %v", q)
	}

	if u.ID != 123456789 || !u.HasSignedUp || u.Properties["nickname"] != "홍길동카톡" || u.ForPartner.UUID != "abc-123" {
		t.Errorf("user = %+v", u)
	}
	if !u.ConnectedAt.Equal(time.Date(2022, 4, 11, 1, 45, 28, 0, time.UTC)) {
		t.Errorf("connected_at = %v", u.ConnectedAt)
	}
	a := u.KakaoAccount
	if a == nil || a.Profile == nil {
		t.Fatalf("account = %+v", a)
	}
	if a.Profile.Nickname != "홍길동" || !a.Profile.IsDefaultImage || a.Profile.ThumbnailImageURL == "" {
		t.Errorf("profile = %+v", a.Profile)
	}
	if a.Email != "sample@sample.com" || !a.IsEmailValid || !a.IsEmailVerified || a.Name != "홍길동" ||
		a.AgeRange != "20~29" || a.Birthday != "1130" || a.BirthdayType != "SOLAR" || a.Birthyear != "2002" ||
		a.Gender != "female" || a.PhoneNumber != "+82 010-1234-5678" || a.CI != "CI" || a.CIAuthenticatedAt.IsZero() {
		t.Errorf("account = %+v", a)
	}
	if a.NameNeedsAgreement || a.EmailNeedsAgreement || a.ProfileNeedsAgreement {
		t.Errorf("needs-agreement flags = %+v", a)
	}
}

func TestUserService_Users(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithAdminKey("admin"))
	srv.HandleJSON("GET /v2/app/users", 200, `[{"id":1,"connected_at":"2022-04-11T01:45:28Z","kakao_account":{"email":"a@b.c"}},{"id":2}]`)

	users, err := c.User.Users(t.Context(), []int64{1, 2}, &UsersOptions{PropertyKeys: []string{"kakao_account.email"}})
	if err != nil {
		t.Fatal(err)
	}
	req := srv.LastRequest()
	if req.Header.Get("Authorization") != "KakaoAK admin" {
		t.Errorf("Authorization = %q", req.Header.Get("Authorization"))
	}
	if q := req.Query; q.Get("target_ids") != "[1,2]" || q.Get("target_id_type") != "user_id" || q.Get("property_keys") != `["kakao_account.email"]` {
		t.Errorf("query = %v", q)
	}
	if len(users) != 2 || users[0].KakaoAccount.Email != "a@b.c" || users[1].ID != 2 {
		t.Errorf("users = %+v", users)
	}

	if _, err := c.User.Users(t.Context(), nil, nil); err == nil {
		t.Error("no IDs: want error")
	}
}

func TestUserService_UserIDs(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithAdminKey("admin"))
	srv.Handle("GET /v1/user/ids", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("from_id") {
		case "":
			_, _ = w.Write([]byte(`{"elements":[1,2],"after_url":"https://kapi.kakao.com/v1/user/ids?limit=2&order=asc&from_id=2","before_url":null}`))
		case "2":
			_, _ = w.Write([]byte(`{"elements":[3],"after_url":null,"before_url":"https://kapi.kakao.com/v1/user/ids?limit=2&order=desc&from_id=3"}`))
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	})

	page, err := c.User.UserIDs(t.Context(), &UserIDsOptions{Limit: 2, FromID: 0, Descending: true})
	if err != nil {
		t.Fatal(err)
	}
	if q := srv.LastRequest().Query; q.Get("limit") != "2" || q.Get("order") != "desc" || q.Has("from_id") {
		t.Errorf("query = %v", q)
	}
	if !slices.Equal(page.Elements, []int64{1, 2}) || page.AfterURL == "" || page.BeforeURL != "" {
		t.Errorf("page = %+v", page)
	}

	var ids []int64
	for id, err := range c.User.AllUserIDs(t.Context(), &UserIDsOptions{Limit: 2}) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	if !slices.Equal(ids, []int64{1, 2, 3}) {
		t.Errorf("ids = %v", ids)
	}

	for id, err := range c.User.AllUserIDs(t.Context(), nil) {
		if err != nil || id != 1 {
			t.Errorf("first id = %d", id)
		}
		break // stopping early must not fetch more pages
	}
}

func TestUserService_AllUserIDs_errors(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithAdminKey("admin"))
	srv.HandleJSON("GET /v1/user/ids", 200, `{"elements":[1],"after_url":"https://kapi.kakao.com/v1/user/ids?from_id=abc"}`)
	var gotErr error
	for _, err := range c.User.AllUserIDs(t.Context(), nil) {
		gotErr = err
	}
	if gotErr == nil {
		t.Error("bad from_id: want error")
	}

	c2, srv2 := newTestClient(t, WithAdminKey("admin"))
	srv2.HandleJSON("GET /v1/user/ids", 429, `{"code":-10,"msg":"limit"}`)
	for _, err := range c2.User.AllUserIDs(t.Context(), nil) {
		gotErr = err
	}
	if !errors.Is(gotErr, ErrRateLimited) {
		t.Errorf("err = %v", gotErr)
	}
}

func TestNextFromID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in      string
		want    int64
		wantErr bool
	}{
		{"", 0, false},
		{"https://kapi.kakao.com/v1/user/ids?from_id=42", 42, false},
		{"https://kapi.kakao.com/v1/user/ids?limit=10", 0, false},
		{"https://kapi.kakao.com/v1/user/ids?from_id=x", 0, true},
		{"://bad", 0, true},
	}
	for _, tt := range tests {
		got, err := nextFromID(tt.in)
		if got != tt.want || (err != nil) != tt.wantErr {
			t.Errorf("nextFromID(%q) = %d, %v", tt.in, got, err)
		}
	}
}

func TestUserService_LogoutUnlink(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /v1/user/logout", 200, `{"id":11}`)
	srv.HandleJSON("POST /v1/user/unlink", 200, `{"id":12}`)
	if id, err := c.User.Logout(t.Context(), AccessToken("AT")); err != nil || id != 11 {
		t.Errorf("Logout = %d, %v", id, err)
	}
	if id, err := c.User.Unlink(t.Context(), AccessToken("AT")); err != nil || id != 12 {
		t.Errorf("Unlink = %d, %v", id, err)
	}
}

func TestUserService_TokenInfo(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /v1/user/access_token_info", 200, `{"id":123456789,"expires_in":7199,"app_id":1234}`)
	info, err := c.User.TokenInfo(t.Context(), "AT")
	if err != nil {
		t.Fatal(err)
	}
	if *info != (AccessTokenInfo{ID: 123456789, ExpiresIn: 7199, AppID: 1234}) {
		t.Errorf("info = %+v", info)
	}
	srv.HandleJSON("GET /v2/user/me", 401, `{"msg":"this access token does not exist","code":-401}`)
	if _, err := c.User.Me(t.Context(), AccessToken("bad"), nil); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("err = %v", err)
	}
}

func TestUserService_UpdateProfileAndSignup(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("POST /v1/user/update_profile", 200, `{"id":7}`)
	srv.HandleJSON("POST /v1/user/signup", 200, `{"id":8}`)

	id, err := c.User.UpdateProfile(t.Context(), "AT", map[string]string{"grade": "gold"})
	if err != nil || id != 7 {
		t.Fatalf("UpdateProfile = %d, %v", id, err)
	}
	if f := srv.LastRequest().Form(); f.Get("properties") != `{"grade":"gold"}` {
		t.Errorf("form = %v", f)
	}
	if _, err := c.User.UpdateProfile(t.Context(), "AT", nil); err == nil {
		t.Error("no properties: want error")
	}

	id, err = c.User.Signup(t.Context(), "AT", nil)
	if err != nil || id != 8 {
		t.Fatalf("Signup = %d, %v", id, err)
	}
	if f := srv.LastRequest().Form(); f.Has("properties") {
		t.Errorf("form = %v", f)
	}
}

func TestUserService_ShippingAddress(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /v2/user/shipping_address", 200, `{"user_id":1,"shipping_address_needs_agreement":false,"shipping_address":{"id":300,"name":"집","is_default":true,"updated_at":1658214426,"type":"NEW","base_address":"경기 성남시 분당구 판교역로 166","detail_address":"카카오","receiver_name":"홍길동","receiver_phone_number1":"010-0000-0000","zone_number":"13529"}}`)

	res, err := c.User.ShippingAddress(t.Context(), AccessToken("AT"), 300)
	if err != nil {
		t.Fatal(err)
	}
	if srv.LastRequest().Query.Get("address_id") != "300" {
		t.Errorf("query = %v", srv.LastRequest().Query)
	}
	a := res.Address
	if res.UserID != 1 || res.NeedsAgreement || a == nil || a.ID != 300 || !a.IsDefault || a.UpdatedAt != 1658214426 || a.Type != "NEW" || a.ZoneNumber != "13529" {
		t.Errorf("res = %+v, address = %+v", res, a)
	}
	if _, err := c.User.ShippingAddress(t.Context(), AccessToken("AT"), 0); err != nil {
		t.Fatal(err)
	}
	if srv.LastRequest().Query.Has("address_id") {
		t.Error("address_id sent for default address")
	}
}

func TestUserService_Scopes(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithAdminKey("admin"))
	srv.HandleFile("GET /v2/user/scopes", "testdata/user/scopes.json")
	srv.HandleFile("POST /v2/user/revoke/scopes", "testdata/user/scopes.json")

	res, err := c.User.Scopes(t.Context(), AccessToken("AT"), "account_email", "friends")
	if err != nil {
		t.Fatal(err)
	}
	if q := srv.LastRequest().Query; q.Get("scopes") != `["account_email","friends"]` {
		t.Errorf("query = %v", q)
	}
	if res.ID != 123456789 || len(res.Scopes) != 3 {
		t.Fatalf("res = %+v", res)
	}
	if s := res.Scopes[1]; s != (Scope{ID: "account_email", DisplayName: "Email", Type: "PRIVACY", Using: true, Agreed: true, Revocable: true}) {
		t.Errorf("scope = %+v", s)
	}
	if _, err := c.User.Scopes(t.Context(), AccessToken("AT")); err != nil {
		t.Fatal(err)
	}
	if srv.LastRequest().Query.Has("scopes") {
		t.Error("empty scopes sent")
	}

	if _, err := c.User.RevokeScopes(t.Context(), AsAdmin(5), "account_email"); err != nil {
		t.Fatal(err)
	}
	if f := srv.LastRequest().Form(); f.Get("scopes") != `["account_email"]` || f.Get("target_id") != "5" {
		t.Errorf("form = %v", f)
	}
	if _, err := c.User.RevokeScopes(t.Context(), AsAdmin(5)); err == nil {
		t.Error("no scopes: want error")
	}
}

func TestUserService_ServiceTerms(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithAdminKey("admin"))
	srv.HandleFile("GET /v2/user/service_terms", "testdata/user/service_terms.json")
	srv.HandleJSON("POST /v2/user/revoke/service_terms", 200, `{"id":111111,"revoked_service_terms":[{"tag":"optional_20200616","agreed":false},{"tag":"optional_no_consent","agreed":false}]}`)
	srv.HandleJSON("POST /v2/user/upgrade/service_terms", 200, `{"id":111111,"agreed_service_terms":[{"tag":"service_2020_0218","agreed":true,"agreed_at":"2023-06-29T07:39:56Z","agreed_by":"KAPI"}]}`)

	res, err := c.User.ServiceTerms(t.Context(), AsAdmin(111111), &ServiceTermsOptions{Result: AppServiceTerms, Tags: []string{"a", "b"}})
	if err != nil {
		t.Fatal(err)
	}
	if q := srv.LastRequest().Query; q.Get("result") != "app_service_terms" || q.Get("tags") != "a,b" || q.Get("target_id") != "111111" {
		t.Errorf("query = %v", q)
	}
	if len(res.ServiceTerms) != 3 {
		t.Fatalf("res = %+v", res)
	}
	st := res.ServiceTerms[0]
	if st.Tag != "optional_20200616" || st.Required || !st.Agreed || !st.Revocable || st.AgreedBy != "KAUTH" ||
		!st.AgreedAt.Equal(time.Date(2023, 6, 29, 7, 39, 56, 0, time.UTC)) {
		t.Errorf("term = %+v", st)
	}
	if !res.ServiceTerms[1].AgreedAt.IsZero() {
		t.Errorf("unagreed term has agreed_at: %+v", res.ServiceTerms[1])
	}
	if _, err := c.User.ServiceTerms(t.Context(), AccessToken("AT"), nil); err != nil {
		t.Fatal(err)
	}
	if q := srv.LastRequest().Query; len(q) != 0 {
		t.Errorf("nil options query = %v", q)
	}

	rev, err := c.User.RevokeServiceTerms(t.Context(), AccessToken("AT"), "optional_20200616", "optional_no_consent")
	if err != nil {
		t.Fatal(err)
	}
	if f := srv.LastRequest().Form(); f.Get("tags") != "optional_20200616,optional_no_consent" {
		t.Errorf("form = %v", f)
	}
	if len(rev.Revoked) != 2 || rev.Revoked[0].Agreed || len(rev.Agreed) != 0 {
		t.Errorf("revoke = %+v", rev)
	}

	agr, err := c.User.AgreeServiceTerms(t.Context(), AsAdmin(111111), "service_2020_0218")
	if err != nil {
		t.Fatal(err)
	}
	if f := srv.LastRequest().Form(); f.Get("tags") != "service_2020_0218" || f.Get("target_id") != "111111" {
		t.Errorf("form = %v", f)
	}
	if len(agr.Agreed) != 1 || agr.Agreed[0].AgreedBy != "KAPI" {
		t.Errorf("agree = %+v", agr)
	}

	if _, err := c.User.AgreeServiceTerms(t.Context(), AsAdmin(1)); err == nil {
		t.Error("no tags: want error")
	}
}

func TestUserService_OIDCUserInfo(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t)
	srv.HandleJSON("GET /v1/oidc/userinfo", 200, `{"sub":"123456789","name":"홍길동","nickname":"홍길동","picture":"https://x/p.jpg","email":"sample@sample.com","email_verified":true,"gender":"male","birthdate":"2002-11-30","phone_number":"+82 00-0000-0000","phone_number_verified":true}`)
	info, err := c.User.OIDCUserInfo(t.Context(), "AT")
	if err != nil {
		t.Fatal(err)
	}
	if srv.LastRequest().Header.Get("Authorization") != "Bearer AT" {
		t.Error("missing bearer token")
	}
	want := OIDCUserInfo{
		Subject: "123456789", Name: "홍길동", Nickname: "홍길동", Picture: "https://x/p.jpg", Email: "sample@sample.com",
		EmailVerified: true, Gender: "male", Birthdate: "2002-11-30", PhoneNumber: "+82 00-0000-0000", PhoneNumberVerified: true,
	}
	if *info != want {
		t.Errorf("info = %+v", info)
	}
}

func TestSetJSON(t *testing.T) {
	t.Parallel()
	q := map[string][]string{}
	if err := setJSON(q, "a", []string(nil)); err != nil || len(q) != 0 {
		t.Errorf("nil slice: %v, %v", q, err)
	}
	if err := setJSON(q, "a", map[string]string{}); err != nil || len(q) != 0 {
		t.Errorf("empty map: %v, %v", q, err)
	}
	if err := setJSON(q, "a", func() {}); err == nil {
		t.Error("unencodable: want error")
	}
}
