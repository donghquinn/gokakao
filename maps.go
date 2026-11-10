package kakao

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// MapService calls the Kakao Map REST API: address and place search,
// coordinate conversion, routing and static map images.
// It authenticates with the REST API key.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomap/rest-api
type MapService struct {
	c *Client
}

// Point is a coordinate. X is the longitude and Y the latitude in WGS84,
// or the easting and northing in a projected [CoordSystem].
type Point struct {
	X, Y float64
}

func (p Point) String() string {
	return formatFloat(p.X) + "," + formatFloat(p.Y)
}

// Rect is a search rectangle given by two opposite corners.
type Rect struct {
	Min, Max Point
}

func (r Rect) String() string {
	return r.Min.String() + "," + r.Max.String()
}

// CoordSystem is a coordinate reference system.
type CoordSystem string

// Coordinate systems. Each API accepts a subset; see its documentation.
const (
	WGS84      CoordSystem = "WGS84"
	WCONGNAMUL CoordSystem = "WCONGNAMUL"
	CONGNAMUL  CoordSystem = "CONGNAMUL"
	WTM        CoordSystem = "WTM"
	TM         CoordSystem = "TM"
	KTM        CoordSystem = "KTM"
	UTM        CoordSystem = "UTM"
	BESSEL     CoordSystem = "BESSEL"
	WKTM       CoordSystem = "WKTM"
	WUTM       CoordSystem = "WUTM"
)

// AnalyzeType controls how address search matches building names.
type AnalyzeType string

// Address analysis types.
const (
	// AnalyzeSimilar also matches partial building names (default).
	AnalyzeSimilar AnalyzeType = "similar"
	// AnalyzeExact matches building names exactly.
	AnalyzeExact AnalyzeType = "exact"
)

// AddressSearchOptions are the optional parameters of
// [MapService.SearchAddress].
type AddressSearchOptions struct {
	AnalyzeType AnalyzeType
	Page        int // 1-45
	Size        int // 1-30
}

// AddressType classifies an address search result.
type AddressType string

// Address types.
const (
	AddressTypeRegion     AddressType = "REGION"      // place name
	AddressTypeRoad       AddressType = "ROAD"        // road name
	AddressTypeRegionAddr AddressType = "REGION_ADDR" // lot-number address
	AddressTypeRoadAddr   AddressType = "ROAD_ADDR"   // road-name address
)

// Address is a lot-number (지번) address.
type Address struct {
	AddressName      string `json:"address_name"`
	Region1DepthName string `json:"region_1depth_name"` // province or city
	Region2DepthName string `json:"region_2depth_name"` // district
	Region3DepthName string `json:"region_3depth_name"` // legal dong
	// Region3DepthHName is the administrative dong. Address search only.
	Region3DepthHName string `json:"region_3depth_h_name"`
	HCode             string `json:"h_code"` // administrative code. Address search only.
	BCode             string `json:"b_code"` // legal code. Address search only.
	Mountain          YesNo  `json:"mountain_yn"`
	MainAddressNo     string `json:"main_address_no"`
	SubAddressNo      string `json:"sub_address_no"`
	X                 Float  `json:"x"` // address search only
	Y                 Float  `json:"y"` // address search only
}

// RoadAddress is a road-name (도로명) address.
type RoadAddress struct {
	AddressName      string `json:"address_name"`
	Region1DepthName string `json:"region_1depth_name"`
	Region2DepthName string `json:"region_2depth_name"`
	Region3DepthName string `json:"region_3depth_name"`
	RoadName         string `json:"road_name"`
	Underground      YesNo  `json:"underground_yn"`
	MainBuildingNo   string `json:"main_building_no"`
	SubBuildingNo    string `json:"sub_building_no"`
	BuildingName     string `json:"building_name"`
	ZoneNo           string `json:"zone_no"` // 5-digit postal code
	X                Float  `json:"x"`       // address search only
	Y                Float  `json:"y"`       // address search only
}

// AddressDocument is an address search result.
type AddressDocument struct {
	AddressName string      `json:"address_name"`
	AddressType AddressType `json:"address_type"`
	X           Float       `json:"x"`
	Y           Float       `json:"y"`
	// Address is nil when there is no lot-number address.
	Address *Address `json:"address"`
	// RoadAddress is nil when there is no road-name address.
	RoadAddress *RoadAddress `json:"road_address"`
}

// SearchAddress converts an address to coordinates.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomap/rest-api
func (s *MapService) SearchAddress(ctx context.Context, query string, opts *AddressSearchOptions) (*SearchResult[AddressDocument], error) {
	if query == "" {
		return nil, errEmptyQuery
	}
	q := url.Values{"query": {query}}
	if opts != nil {
		setString(q, "analyze_type", string(opts.AnalyzeType))
		setInt(q, "page", opts.Page)
		setInt(q, "size", opts.Size)
	}
	var res SearchResult[AddressDocument]
	if err := s.get(ctx, "/v2/local/search/address.json", q, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// CoordOptions selects the coordinate systems of a conversion.
// Empty fields default to WGS84.
type CoordOptions struct {
	// InputCoord is the system of the given point: WGS84, WCONGNAMUL,
	// CONGNAMUL, WTM or TM.
	InputCoord CoordSystem
	// OutputCoord is the system of the returned coordinates.
	// Ignored by [MapService.CoordToAddress].
	OutputCoord CoordSystem
}

// RegionType distinguishes administrative and legal regions.
type RegionType string

// Region types.
const (
	RegionTypeAdministrative RegionType = "H" // 행정동
	RegionTypeLegal          RegionType = "B" // 법정동
)

// Region is an administrative or legal region.
type Region struct {
	RegionType       RegionType `json:"region_type"`
	AddressName      string     `json:"address_name"`
	Region1DepthName string     `json:"region_1depth_name"`
	Region2DepthName string     `json:"region_2depth_name"`
	Region3DepthName string     `json:"region_3depth_name"`
	// Region4DepthName is the 리 name, set only for legal regions.
	Region4DepthName string `json:"region_4depth_name"`
	Code             string `json:"code"`
	X                Float  `json:"x"`
	Y                Float  `json:"y"`
}

// CoordToRegionCode returns the administrative and legal regions that
// contain p.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomap/rest-api
func (s *MapService) CoordToRegionCode(ctx context.Context, p Point, opts *CoordOptions) ([]Region, error) {
	q := pointQuery(p)
	if opts != nil {
		setString(q, "input_coord", string(opts.InputCoord))
		setString(q, "output_coord", string(opts.OutputCoord))
	}
	var res SearchResult[Region]
	if err := s.get(ctx, "/v2/local/geo/coord2regioncode.json", q, &res); err != nil {
		return nil, err
	}
	return res.Documents, nil
}

// CoordAddress is the address at a coordinate.
type CoordAddress struct {
	// Address is nil when there is no lot-number address.
	Address *Address `json:"address"`
	// RoadAddress is nil when there is no road-name address.
	RoadAddress *RoadAddress `json:"road_address"`
}

// CoordToAddress returns the address at p. The result is empty when no
// address is found.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomap/rest-api
func (s *MapService) CoordToAddress(ctx context.Context, p Point, opts *CoordOptions) ([]CoordAddress, error) {
	q := pointQuery(p)
	if opts != nil {
		setString(q, "input_coord", string(opts.InputCoord))
	}
	var res SearchResult[CoordAddress]
	if err := s.get(ctx, "/v2/local/geo/coord2address.json", q, &res); err != nil {
		return nil, err
	}
	return res.Documents, nil
}

// TransformCoord converts p from opts.InputCoord to opts.OutputCoord, which
// may be any [CoordSystem]. Empty systems default to WGS84.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomap/rest-api
func (s *MapService) TransformCoord(ctx context.Context, p Point, opts *CoordOptions) (Point, error) {
	q := pointQuery(p)
	if opts != nil {
		setString(q, "input_coord", string(opts.InputCoord))
		setString(q, "output_coord", string(opts.OutputCoord))
	}
	var res SearchResult[struct {
		X Float `json:"x"`
		Y Float `json:"y"`
	}]
	if err := s.get(ctx, "/v2/local/geo/transcoord.json", q, &res); err != nil {
		return Point{}, err
	}
	if len(res.Documents) == 0 {
		return Point{}, errors.New("kakao: transcoord returned no coordinates")
	}
	d := res.Documents[0]
	return Point{X: float64(d.X), Y: float64(d.Y)}, nil
}

// CategoryGroup is a place category group code.
type CategoryGroup string

// Place category groups.
const (
	CategoryMart          CategoryGroup = "MT1" // 대형마트
	CategoryConvenience   CategoryGroup = "CS2" // 편의점
	CategoryKindergarten  CategoryGroup = "PS3" // 어린이집, 유치원
	CategorySchool        CategoryGroup = "SC4" // 학교
	CategoryAcademy       CategoryGroup = "AC5" // 학원
	CategoryParking       CategoryGroup = "PK6" // 주차장
	CategoryGasStation    CategoryGroup = "OL7" // 주유소, 충전소
	CategorySubwayStation CategoryGroup = "SW8" // 지하철역
	CategoryBank          CategoryGroup = "BK9" // 은행
	CategoryCulture       CategoryGroup = "CT1" // 문화시설
	CategoryRealEstate    CategoryGroup = "AG2" // 중개업소
	CategoryPublicOffice  CategoryGroup = "PO3" // 공공기관
	CategoryAttraction    CategoryGroup = "AT4" // 관광명소
	CategoryLodging       CategoryGroup = "AD5" // 숙박
	CategoryRestaurant    CategoryGroup = "FD6" // 음식점
	CategoryCafe          CategoryGroup = "CE7" // 카페
	CategoryHospital      CategoryGroup = "HP8" // 병원
	CategoryPharmacy      CategoryGroup = "PM9" // 약국
)

// PlaceSort orders place search results.
type PlaceSort string

// Place sort orders.
const (
	PlaceSortAccuracy PlaceSort = "accuracy" // default
	PlaceSortDistance PlaceSort = "distance" // requires Center
)

// PlaceSearchOptions are the optional parameters of the place searches.
type PlaceSearchOptions struct {
	// Category filters keyword search by category group. Ignored by
	// [MapService.SearchCategory], which takes the group as an argument.
	Category CategoryGroup
	// Center is the reference point for Radius, distance sorting and the
	// Distance field of results.
	Center *Point
	// Radius limits results to this many meters from Center (0-20000).
	Radius int
	// Rect limits results to a rectangle.
	Rect *Rect
	Page int // 1-45
	Size int // 1-15
	Sort PlaceSort
}

// SameName is the region analysis of a keyword search query.
type SameName struct {
	// Region lists the regions recognized in the query.
	Region []string `json:"region"`
	// Keyword is the query without the region.
	Keyword string `json:"keyword"`
	// SelectedRegion is the region used for the search.
	SelectedRegion string `json:"selected_region"`
}

// PlaceSearchMeta describes a page of place search results.
type PlaceSearchMeta struct {
	SearchMeta
	// SameName is the query analysis. It is nil for category search.
	SameName *SameName `json:"same_name"`
}

// PlaceSearchResult is a page of place search results.
type PlaceSearchResult struct {
	Meta      PlaceSearchMeta `json:"meta"`
	Documents []Place         `json:"documents"`
}

// Place is a place search result.
type Place struct {
	ID                string        `json:"id"`
	PlaceName         string        `json:"place_name"`
	CategoryName      string        `json:"category_name"`
	CategoryGroupCode CategoryGroup `json:"category_group_code"`
	CategoryGroupName string        `json:"category_group_name"`
	Phone             string        `json:"phone"`
	AddressName       string        `json:"address_name"`
	RoadAddressName   string        `json:"road_address_name"`
	X                 Float         `json:"x"`
	Y                 Float         `json:"y"`
	PlaceURL          string        `json:"place_url"`
	// Distance is the distance from the search center in meters, or 0 when
	// no center was given.
	Distance Float `json:"distance"`
}

// SearchKeyword searches places by keyword.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomap/rest-api
func (s *MapService) SearchKeyword(ctx context.Context, query string, opts *PlaceSearchOptions) (*PlaceSearchResult, error) {
	if query == "" {
		return nil, errEmptyQuery
	}
	q := url.Values{"query": {query}}
	if opts != nil {
		setString(q, "category_group_code", string(opts.Category))
	}
	return s.searchPlaces(ctx, "/v2/local/search/keyword.json", q, opts)
}

// SearchCategory searches places in a category group.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomap/rest-api
func (s *MapService) SearchCategory(ctx context.Context, group CategoryGroup, opts *PlaceSearchOptions) (*PlaceSearchResult, error) {
	if group == "" {
		return nil, errors.New("kakao: empty category group")
	}
	q := url.Values{"category_group_code": {string(group)}}
	return s.searchPlaces(ctx, "/v2/local/search/category.json", q, opts)
}

func (s *MapService) searchPlaces(ctx context.Context, path string, q url.Values, opts *PlaceSearchOptions) (*PlaceSearchResult, error) {
	if opts != nil {
		if opts.Center != nil {
			q.Set("x", formatFloat(opts.Center.X))
			q.Set("y", formatFloat(opts.Center.Y))
		}
		setInt(q, "radius", opts.Radius)
		if opts.Rect != nil {
			q.Set("rect", opts.Rect.String())
		}
		setInt(q, "page", opts.Page)
		setInt(q, "size", opts.Size)
		setString(q, "sort", string(opts.Sort))
	}
	var res PlaceSearchResult
	if err := s.get(ctx, path, q, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// RouteMode is the path preference of walking and cycling routes.
type RouteMode string

// Route modes.
const (
	RouteBroadFirst RouteMode = "BROAD_FIRST" // wide roads first; walking default
	RouteBikeOnly   RouteMode = "BIKE_ONLY"   // bike lanes first; cycling default
	RouteShortest   RouteMode = "SHORTEST"
	RouteAccessible RouteMode = "ACCESSIBLE" // comfortable path
)

// RouteOptions are the optional parameters of walking and cycling routes.
type RouteOptions struct {
	// Via lists up to 5 waypoints.
	Via []Point
	// ViaNames names the waypoints, in order.
	ViaNames  []string
	StartName string // default "출발"
	EndName   string // default "도착"
	// InputCoord and OutputCoord are WGS84 (default), WTM, TM or WCONGNAMUL.
	InputCoord  CoordSystem
	OutputCoord CoordSystem
	Mode        RouteMode
}

// Route is a walking or cycling route.
type Route struct {
	Properties RouteProperties `json:"properties"`
	Legs       []RouteLeg      `json:"legs"`
}

// RouteProperties summarizes a [Route].
type RouteProperties struct {
	TotalDistance int    `json:"totalDistance"` // meters
	TotalTime     int    `json:"totalTime"`     // seconds
	LandingURL    string `json:"landingUrl"`    // Kakao Map link
}

// RouteLeg is the part of a route between two waypoints.
type RouteLeg struct {
	Properties RouteLegProperties `json:"properties"`
	Steps      []RouteStep        `json:"steps"`
}

// RouteLegProperties summarizes a [RouteLeg].
type RouteLegProperties struct {
	Distance int `json:"distance"` // meters
	Time     int `json:"time"`     // seconds
}

// RouteStep is one instruction of a route.
type RouteStep struct {
	Properties RouteStepProperties `json:"properties"`
	Path       Path                `json:"path"`
}

// RouteStepProperties describes a [RouteStep].
type RouteStepProperties struct {
	Distance int    `json:"distance"` // meters
	Guidance string `json:"guidance"`
	Time     int    `json:"time"` // seconds
	X        Float  `json:"x"`    // step start
	Y        Float  `json:"y"`
}

// Path is a polyline of [x, y] points.
type Path struct {
	Points [][2]float64 `json:"points"`
}

// RouteStatusError reports a routing request that found no route, such as
// when the start and end are the same point.
type RouteStatusError struct {
	// Status is the routing status, e.g. "SAME_POINT" or "NO_RESULTS".
	Status string
}

func (e *RouteStatusError) Error() string {
	return "kakao: no route: " + e.Status
}

// WalkingRoute finds a walking route from start to end.
// It returns a [*RouteStatusError] when no route is found.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomap/rest-api
func (s *MapService) WalkingRoute(ctx context.Context, start, end Point, opts *RouteOptions) (*Route, error) {
	return s.route(ctx, "/v2/routing/walk", start, end, opts)
}

// CyclingRoute finds a bicycle route from start to end.
// It returns a [*RouteStatusError] when no route is found.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomap/rest-api
func (s *MapService) CyclingRoute(ctx context.Context, start, end Point, opts *RouteOptions) (*Route, error) {
	return s.route(ctx, "/v2/routing/bicycle", start, end, opts)
}

func (s *MapService) route(ctx context.Context, path string, start, end Point, opts *RouteOptions) (*Route, error) {
	q := routeQuery(start, end)
	if opts != nil {
		if len(opts.Via) > 0 {
			xs := make([]string, len(opts.Via))
			ys := make([]string, len(opts.Via))
			for i, p := range opts.Via {
				xs[i], ys[i] = formatFloat(p.X), formatFloat(p.Y)
			}
			q.Set("via_x", strings.Join(xs, ","))
			q.Set("via_y", strings.Join(ys, ","))
		}
		if len(opts.ViaNames) > 0 {
			q.Set("v_name", strings.Join(opts.ViaNames, ","))
		}
		setString(q, "s_name", opts.StartName)
		setString(q, "e_name", opts.EndName)
		setString(q, "input_coord", string(opts.InputCoord))
		setString(q, "output_coord", string(opts.OutputCoord))
		setString(q, "route_mode", string(opts.Mode))
	}
	var res struct {
		Status string `json:"status"`
		Route  *Route `json:"route"`
	}
	if err := s.get(ctx, path, q, &res); err != nil {
		return nil, err
	}
	if res.Status != "OK" || res.Route == nil {
		return nil, &RouteStatusError{Status: res.Status}
	}
	return res.Route, nil
}

// TransitRouteOptions are the optional parameters of
// [MapService.TransitRoute].
type TransitRouteOptions struct {
	StartName   string // default "출발"
	EndName     string // default "도착"
	InputCoord  CoordSystem
	OutputCoord CoordSystem
}

// TransitRoutes are public transit route options between two points.
type TransitRoutes struct {
	Properties TransitSummary `json:"properties"`
	Routes     []TransitRoute `json:"routes"`
}

// TransitSummary counts the routes in [TransitRoutes] by type.
type TransitSummary struct {
	Total        int    `json:"total"`
	Bus          int    `json:"bus"`
	Subway       int    `json:"subway"`
	BusAndSubway int    `json:"busAndSubway"`
	LandingURL   string `json:"landingURL"` // Kakao Map link
}

// TransitRoute is one public transit route.
type TransitRoute struct {
	Properties TransitRouteProperties `json:"properties"`
	Steps      []TransitStep          `json:"steps"`
}

// TransitRouteProperties summarizes a [TransitRoute].
type TransitRouteProperties struct {
	Type          string `json:"type"`          // BUS, SUBWAY or BUS_AND_SUBWAY
	TotalDistance int    `json:"totalDistance"` // meters
	TotalTime     int    `json:"totalTime"`     // seconds
	Transfers     int    `json:"transfers"`
	Fare          Fare   `json:"fare"`
}

// Fare is a transit fare in KRW.
type Fare struct {
	Value int `json:"value"`
	Min   int `json:"min"`
	Max   int `json:"max"`
}

// TransitStep is one leg of a transit route.
type TransitStep struct {
	Properties TransitStepProperties `json:"properties"`
	Path       Path                  `json:"path"`
}

// TransitStepProperties describes a [TransitStep].
type TransitStepProperties struct {
	Guidance string           `json:"guidance"`
	Type     string           `json:"type"`     // BUS, SUBWAY or WALKING
	Distance int              `json:"distance"` // meters
	Time     int              `json:"time"`     // seconds
	Stops    []TransitStop    `json:"stops"`
	Vehicles []TransitVehicle `json:"vehicles"`
}

// TransitStop is a stop on a transit step.
type TransitStop struct {
	Name string `json:"name"`
}

// TransitVehicle is a bus or train line on a transit step.
type TransitVehicle struct {
	Type string `json:"type"` // e.g. "간선", "마을"; values may change
	Name string `json:"name"`
}

// TransitRoute finds public transit routes from start to end.
// It returns a [*RouteStatusError] when no route is found.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomap/rest-api
func (s *MapService) TransitRoute(ctx context.Context, start, end Point, opts *TransitRouteOptions) (*TransitRoutes, error) {
	q := routeQuery(start, end)
	if opts != nil {
		setString(q, "s_name", opts.StartName)
		setString(q, "e_name", opts.EndName)
		setString(q, "input_coord", string(opts.InputCoord))
		setString(q, "output_coord", string(opts.OutputCoord))
	}
	var res struct {
		Status string `json:"status"`
		TransitRoutes
	}
	if err := s.get(ctx, "/v2/routing/publictraffic", q, &res); err != nil {
		return nil, err
	}
	if res.Status != "OK" {
		return nil, &RouteStatusError{Status: res.Status}
	}
	return &res.TransitRoutes, nil
}

// Marker is a static map marker.
type Marker struct {
	Location Point
	// ExcludeFromBounds leaves the marker out when the map center and zoom
	// level are computed automatically.
	ExcludeFromBounds bool
}

// LogoPosition places the Kakao logo on a static map.
type LogoPosition string

// Logo positions.
const (
	LogoBottomLeft  LogoPosition = "BOTTOM_LEFT" // default
	LogoBottomRight LogoPosition = "BOTTOM_RIGHT"
	LogoTopLeft     LogoPosition = "TOP_LEFT"
	LogoTopRight    LogoPosition = "TOP_RIGHT"
)

// ImageFormat is a static map image format.
type ImageFormat string

// Image formats.
const (
	ImagePNG  ImageFormat = "png" // default
	ImageJPEG ImageFormat = "jpg"
)

// StaticMapParams are the parameters of [MapService.StaticMap].
// Width and Height are required, and so is at least one of Center or
// Markers.
type StaticMapParams struct {
	Width, Height int // pixels, up to 2048x1024
	Center        *Point
	Markers       []Marker // up to 5
	Format        ImageFormat
	LogoPos       LogoPosition
	// Scale is 1 or 2 (default). At 2 the image is twice Width x Height.
	Scale int
	// Level is the zoom level, 1-15 (default 3).
	Level int
	// Coord is WGS84 (default), WCONGNAMUL, WTM or TM.
	Coord CoordSystem
}

// Image is a binary image.
type Image struct {
	ContentType string // e.g. "image/png"
	Data        []byte
}

// StaticMap renders a map image.
//
// Docs: https://developers.kakao.com/docs/ko/kakaomap/rest-api
func (s *MapService) StaticMap(ctx context.Context, p StaticMapParams) (*Image, error) {
	if p.Width <= 0 || p.Height <= 0 {
		return nil, errors.New("kakao: static map width and height are required")
	}
	if p.Center == nil && len(p.Markers) == 0 {
		return nil, errors.New("kakao: static map needs a center or markers")
	}
	q := url.Values{"size": {fmt.Sprintf("%dx%d", p.Width, p.Height)}}
	if p.Center != nil {
		q.Set("center", p.Center.String())
	}
	for _, m := range p.Markers {
		q.Add("markers", "location:"+m.Location.String()+"|option:"+strconv.FormatBool(m.ExcludeFromBounds))
	}
	setString(q, "format", string(p.Format))
	setString(q, "logo_pos", string(p.LogoPos))
	setInt(q, "scale", p.Scale)
	setInt(q, "lv", p.Level)
	setString(q, "coord", string(p.Coord))
	var raw rawBody
	if err := s.get(ctx, "/v2/maps/staticmap", q, &raw); err != nil {
		return nil, err
	}
	return &Image{ContentType: raw.ContentType, Data: raw.Data}, nil
}

func (s *MapService) get(ctx context.Context, path string, q url.Values, out any) error {
	return s.c.do(ctx, &request{method: http.MethodGet, host: HostDAPI, path: path, auth: authRESTKey, query: q}, out)
}

func pointQuery(p Point) url.Values {
	return url.Values{"x": {formatFloat(p.X)}, "y": {formatFloat(p.Y)}}
}

func routeQuery(start, end Point) url.Values {
	return url.Values{
		"start_x": {formatFloat(start.X)},
		"start_y": {formatFloat(start.Y)},
		"end_x":   {formatFloat(end.X)},
		"end_y":   {formatFloat(end.Y)},
	}
}
