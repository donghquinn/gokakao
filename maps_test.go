package kakao

import (
	"errors"
	"net/http"
	"slices"
	"testing"
)

func TestPointAndRect_String(t *testing.T) {
	t.Parallel()
	if got := (Point{X: 127.1105, Y: 37.3955}).String(); got != "127.1105,37.3955" {
		t.Errorf("Point = %q", got)
	}
	r := Rect{Min: Point{X: 127.0, Y: 37.5}, Max: Point{X: 127.1, Y: 37.6}}
	if got := r.String(); got != "127,37.5,127.1,37.6" {
		t.Errorf("Rect = %q", got)
	}
}

func TestMapService_SearchAddress(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithRESTKey("rest"))
	srv.HandleFile("GET /v2/local/search/address.json", "testdata/map/address.json")

	res, err := c.Map.SearchAddress(t.Context(), "전북 삼성동 100", &AddressSearchOptions{AnalyzeType: AnalyzeExact, Page: 2, Size: 30})
	if err != nil {
		t.Fatal(err)
	}
	req := srv.LastRequest()
	if req.Header.Get("Authorization") != "KakaoAK rest" {
		t.Errorf("Authorization = %q", req.Header.Get("Authorization"))
	}
	if q := req.Query; q.Get("query") != "전북 삼성동 100" || q.Get("analyze_type") != "exact" || q.Get("page") != "2" || q.Get("size") != "30" {
		t.Errorf("query = %v", q)
	}
	if res.Meta.TotalCount != 4 || !res.Meta.IsEnd || len(res.Documents) != 2 {
		t.Fatalf("res = %+v", res)
	}
	d := res.Documents[0]
	if d.AddressType != AddressTypeRegionAddr || d.X != 126.99597295767953 || d.Y != 35.97664845766847 {
		t.Errorf("doc = %+v", d)
	}
	a := d.Address
	if a == nil || a.Region3DepthHName != "삼성동" || a.HCode != "4514069000" || a.BCode != "4514013400" ||
		a.Mountain || a.MainAddressNo != "100" || a.SubAddressNo != "" || a.X == 0 {
		t.Errorf("address = %+v", a)
	}
	ra := d.RoadAddress
	if ra == nil || ra.RoadName != "망산길" || ra.Underground || ra.MainBuildingNo != "11" || ra.SubBuildingNo != "17" || ra.ZoneNo != "54547" {
		t.Errorf("road address = %+v", ra)
	}
	if d2 := res.Documents[1]; d2.Address != nil || d2.RoadAddress != nil || d2.AddressType != AddressTypeRegion {
		t.Errorf("null addresses: %+v", d2)
	}

	if _, err := c.Map.SearchAddress(t.Context(), "", nil); !errors.Is(err, errEmptyQuery) {
		t.Errorf("empty query: err = %v", err)
	}
}

func TestMapService_CoordToRegionCode(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithRESTKey("rest"))
	srv.HandleFile("GET /v2/local/geo/coord2regioncode.json", "testdata/map/coord2regioncode.json")

	regions, err := c.Map.CoordToRegionCode(t.Context(), Point{X: 127.1086228, Y: 37.4012191}, &CoordOptions{InputCoord: WGS84, OutputCoord: WTM})
	if err != nil {
		t.Fatal(err)
	}
	if q := srv.LastRequest().Query; q.Get("x") != "127.1086228" || q.Get("y") != "37.4012191" || q.Get("input_coord") != "WGS84" || q.Get("output_coord") != "WTM" {
		t.Errorf("query = %v", q)
	}
	if len(regions) != 2 {
		t.Fatalf("got %d regions", len(regions))
	}
	r := regions[0]
	if r.RegionType != RegionTypeLegal || r.Code != "4113510900" || r.Region2DepthName != "성남시 분당구" || r.X != 127.10459896729914 {
		t.Errorf("region = %+v", r)
	}
	if regions[1].RegionType != RegionTypeAdministrative {
		t.Errorf("region[1] = %+v", regions[1])
	}

	if _, err := c.Map.CoordToRegionCode(t.Context(), Point{X: 1, Y: 2}, nil); err != nil {
		t.Fatal(err)
	}
	if q := srv.LastRequest().Query; len(q) != 2 {
		t.Errorf("nil options query = %v", q)
	}
}

func TestMapService_CoordToAddress(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithRESTKey("rest"))
	srv.HandleFile("GET /v2/local/geo/coord2address.json", "testdata/map/coord2address.json")

	got, err := c.Map.CoordToAddress(t.Context(), Point{X: 127.423084873712, Y: 37.0789561558879}, &CoordOptions{InputCoord: WGS84, OutputCoord: WTM})
	if err != nil {
		t.Fatal(err)
	}
	q := srv.LastRequest().Query
	if q.Get("input_coord") != "WGS84" || q.Has("output_coord") {
		t.Errorf("query = %v (output_coord must not be sent)", q)
	}
	if len(got) != 1 || got[0].Address == nil || got[0].RoadAddress == nil {
		t.Fatalf("got %+v", got)
	}
	if got[0].Address.Region3DepthName != "죽산면 죽산리" || got[0].RoadAddress.BuildingName != "무지개아파트" || got[0].RoadAddress.ZoneNo != "17519" {
		t.Errorf("got %+v / %+v", got[0].Address, got[0].RoadAddress)
	}
}

func TestMapService_TransformCoord(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithRESTKey("rest"))
	srv.HandleFile("GET /v2/local/geo/transcoord.json", "testdata/map/transcoord.json")

	p, err := c.Map.TransformCoord(t.Context(), Point{X: 160710.37729270622, Y: -4388.879299157299}, WTM, WGS84)
	if err != nil {
		t.Fatal(err)
	}
	if q := srv.LastRequest().Query; q.Get("x") != "160710.37729270622" || q.Get("y") != "-4388.879299157299" || q.Get("input_coord") != "WTM" || q.Get("output_coord") != "WGS84" {
		t.Errorf("query = %v", q)
	}
	if p != (Point{X: 126.57740680000002, Y: 33.453357700000005}) {
		t.Errorf("point = %+v", p)
	}

	empty, srv2 := newTestClient(t, WithRESTKey("rest"))
	srv2.HandleJSON("GET /v2/local/geo/transcoord.json", 200, `{"meta":{"total_count":0},"documents":[]}`)
	if _, err := empty.Map.TransformCoord(t.Context(), Point{}, "", ""); err == nil {
		t.Error("no documents: want error")
	}
	if q := srv2.LastRequest().Query; q.Has("input_coord") || q.Has("output_coord") {
		t.Errorf("empty systems sent: %v", q)
	}
}

func TestMapService_SearchKeyword(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithRESTKey("rest"))
	srv.HandleFile("GET /v2/local/search/keyword.json", "testdata/map/keyword.json")

	res, err := c.Map.SearchKeyword(t.Context(), "카카오프렌즈", &PlaceSearchOptions{
		Category: CategoryCafe,
		Center:   &Point{X: 127.06283102249932, Y: 37.514322572335935},
		Radius:   20000,
		Rect:     &Rect{Min: Point{X: 127, Y: 37.5}, Max: Point{X: 127.1, Y: 37.6}},
		Page:     1,
		Size:     15,
		Sort:     PlaceSortDistance,
	})
	if err != nil {
		t.Fatal(err)
	}
	q := srv.LastRequest().Query
	for k, want := range map[string]string{
		"query": "카카오프렌즈", "category_group_code": "CE7", "x": "127.06283102249932", "y": "37.514322572335935",
		"radius": "20000", "rect": "127,37.5,127.1,37.6", "page": "1", "size": "15", "sort": "distance",
	} {
		if got := q.Get(k); got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}
	if sn := res.Meta.SameName; sn == nil || sn.Keyword != "카카오프렌즈" || len(sn.Region) != 0 {
		t.Errorf("same_name = %+v", sn)
	}
	if res.Meta.TotalCount != 14 || !res.Meta.IsEnd {
		t.Errorf("meta = %+v", res.Meta)
	}
	p := res.Documents[0]
	if p.ID != "26338954" || p.PlaceName != "카카오프렌즈 코엑스점" || p.Distance != 418 || p.Phone != "02-6002-1880" ||
		p.X != 127.05902969025047 || p.Y != 37.51207412593136 || p.CategoryGroupCode != "" || p.RoadAddressName != "서울 강남구 영동대로 513" {
		t.Errorf("place = %+v", p)
	}

	if _, err := c.Map.SearchKeyword(t.Context(), "", nil); !errors.Is(err, errEmptyQuery) {
		t.Errorf("empty query: err = %v", err)
	}
}

func TestMapService_SearchCategory(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithRESTKey("rest"))
	srv.HandleFile("GET /v2/local/search/category.json", "testdata/map/category.json")

	res, err := c.Map.SearchCategory(t.Context(), CategoryPharmacy, &PlaceSearchOptions{Category: CategoryCafe})
	if err != nil {
		t.Fatal(err)
	}
	if q := srv.LastRequest().Query; q.Get("category_group_code") != "PM9" || len(q["category_group_code"]) != 1 {
		t.Errorf("query = %v (option Category must be ignored)", q)
	}
	if res.Meta.SameName != nil {
		t.Errorf("same_name = %+v, want nil", res.Meta.SameName)
	}
	p := res.Documents[0]
	if p.CategoryGroupCode != CategoryPharmacy || p.CategoryGroupName != "약국" || p.Distance != 0 {
		t.Errorf("place = %+v", p)
	}

	if _, err := c.Map.SearchCategory(t.Context(), "", nil); err == nil {
		t.Error("empty group: want error")
	}
	if _, err := c.Map.SearchCategory(t.Context(), CategoryPharmacy, nil); err != nil {
		t.Fatal(err)
	}
}

func TestMapService_routes(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithRESTKey("rest"))
	srv.HandleFile("GET /v2/routing/walk", "testdata/map/walk.json")
	srv.HandleFile("GET /v2/routing/bicycle", "testdata/map/bicycle.json")
	start := Point{X: 127.11119669891646, Y: 37.394776627382875}
	end := Point{X: 127.12629039752096, Y: 37.4199323570413}

	r, err := c.Map.WalkingRoute(t.Context(), start, end, &RouteOptions{
		Via:         []Point{{X: 127.1, Y: 37.4}, {X: 127.2, Y: 37.5}},
		ViaNames:    []string{"a", "b"},
		StartName:   "판교역",
		EndName:     "시청",
		InputCoord:  WGS84,
		OutputCoord: WTM,
		Mode:        RouteShortest,
	})
	if err != nil {
		t.Fatal(err)
	}
	q := srv.LastRequest().Query
	for k, want := range map[string]string{
		"start_x": "127.11119669891646", "start_y": "37.394776627382875", "end_x": "127.12629039752096", "end_y": "37.4199323570413",
		"via_x": "127.1,127.2", "via_y": "37.4,37.5", "v_name": "a,b", "s_name": "판교역", "e_name": "시청",
		"input_coord": "WGS84", "output_coord": "WTM", "route_mode": "SHORTEST",
	} {
		if got := q.Get(k); got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}
	if r.Properties.TotalDistance != 4025 || r.Properties.TotalTime != 3914 || r.Properties.LandingURL == "" {
		t.Errorf("properties = %+v", r.Properties)
	}
	step := r.Legs[0].Steps[0]
	if r.Legs[0].Properties.Distance != 4025 || step.Properties.Guidance != "판교역 1번 출구까지 역사 내 이동" ||
		step.Properties.X != 127.11119783 || len(step.Path.Points) != 4 || step.Path.Points[0] != [2]float64{127.11119783, 37.39477573} {
		t.Errorf("leg = %+v", r.Legs[0])
	}

	b, err := c.Map.CyclingRoute(t.Context(), start, end, nil)
	if err != nil {
		t.Fatal(err)
	}
	if b.Properties.TotalTime != 1242 {
		t.Errorf("bicycle = %+v", b.Properties)
	}
	if q := srv.LastRequest().Query; len(q) != 4 {
		t.Errorf("nil options query = %v", q)
	}
}

func TestMapService_routeStatus(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithRESTKey("rest"))
	srv.HandleJSON("GET /v2/routing/walk", 200, `{"status":"SAME_POINT"}`)
	srv.HandleJSON("GET /v2/routing/publictraffic", 200, `{"status":"NO_RESULTS"}`)
	srv.HandleJSON("GET /v2/routing/bicycle", 400, `{"errorType":"InvalidArgument","message":"bad"}`)

	_, err := c.Map.WalkingRoute(t.Context(), Point{}, Point{}, nil)
	var se *RouteStatusError
	if !errors.As(err, &se) || se.Status != "SAME_POINT" || se.Error() != "kakao: no route: SAME_POINT" {
		t.Errorf("walk err = %v", err)
	}
	_, err = c.Map.TransitRoute(t.Context(), Point{}, Point{}, nil)
	if !errors.As(err, &se) || se.Status != "NO_RESULTS" {
		t.Errorf("transit err = %v", err)
	}
	_, err = c.Map.CyclingRoute(t.Context(), Point{}, Point{}, nil)
	var e *Error
	if !errors.As(err, &e) {
		t.Errorf("bicycle err = %v", err)
	}
}

func TestMapService_TransitRoute(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithRESTKey("rest"))
	srv.HandleFile("GET /v2/routing/publictraffic", "testdata/map/publictraffic.json")

	res, err := c.Map.TransitRoute(t.Context(), Point{X: 127.11119217, Y: 37.39477123}, Point{X: 127.12628814, Y: 37.41993056},
		&TransitRouteOptions{StartName: "출발", EndName: "도착", InputCoord: WGS84, OutputCoord: WGS84})
	if err != nil {
		t.Fatal(err)
	}
	q := srv.LastRequest().Query
	if q.Get("start_x") != "127.11119217" || q.Get("s_name") != "출발" || q.Get("e_name") != "도착" || q.Get("input_coord") != "WGS84" {
		t.Errorf("query = %v", q)
	}
	if res.Properties.Total != 15 || res.Properties.Bus != 9 || res.Properties.BusAndSubway != 5 || res.Properties.LandingURL == "" {
		t.Errorf("properties = %+v", res.Properties)
	}
	route := res.Routes[0]
	if route.Properties.Type != "BUS" || route.Properties.TotalTime != 2115 || route.Properties.Fare.Value != 1350 {
		t.Errorf("route = %+v", route.Properties)
	}
	step := route.Steps[0].Properties
	if step.Type != "BUS" || len(step.Stops) != 2 || step.Stops[0].Name != "판교역동편" ||
		!slices.Equal(step.Vehicles, []TransitVehicle{{Type: "마을", Name: "76"}}) {
		t.Errorf("step = %+v", step)
	}
}

func TestMapService_StaticMap(t *testing.T) {
	t.Parallel()
	c, srv := newTestClient(t, WithRESTKey("rest"))
	srv.Handle("GET /v2/maps/staticmap", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("\x89PNG"))
	})

	img, err := c.Map.StaticMap(t.Context(), &StaticMapParams{
		Width: 640, Height: 480,
		Center: &Point{X: 127.1105, Y: 37.3955},
		Markers: []Marker{
			{Location: Point{X: 127.1105, Y: 37.3955}},
			{Location: Point{X: 127.112, Y: 37.397}, ExcludeFromBounds: true},
		},
		Format: "jpg", LogoPos: LogoTopRight, Scale: 1, Level: 5, Coord: WGS84,
	})
	if err != nil {
		t.Fatal(err)
	}
	if img.ContentType != "image/png" || string(img.Data) != "\x89PNG" {
		t.Errorf("image = %+v", img)
	}
	q := srv.LastRequest().Query
	wantMarkers := []string{"location:127.1105,37.3955|option:false", "location:127.112,37.397|option:true"}
	if !slices.Equal(q["markers"], wantMarkers) {
		t.Errorf("markers = %q", q["markers"])
	}
	for k, want := range map[string]string{
		"size": "640x480", "center": "127.1105,37.3955", "format": "jpg", "logo_pos": "TOP_RIGHT", "scale": "1", "lv": "5", "coord": "WGS84",
	} {
		if got := q.Get(k); got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}

	for name, p := range map[string]*StaticMapParams{
		"nil":       nil,
		"no size":   {Center: &Point{}},
		"no center": {Width: 10, Height: 10},
		"negative":  {Width: -1, Height: 10, Center: &Point{}},
	} {
		if _, err := c.Map.StaticMap(t.Context(), p); err == nil {
			t.Errorf("%s: want error", name)
		}
	}
}
