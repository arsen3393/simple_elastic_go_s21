package types

type Place struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
}

type Limits struct {
	Size int `json:"size"`
	From int `json:"from"`
}

type Sort struct {
	Geo GeoDistance `json:"_geo_distance"`
}

type Query struct {
	Size int  `json:"size"`
	Sort Sort `json:"sort"`
}

func NewQuery(NewLat, NewLon float64) Query {
	var s Sort
	s.SetGeo(NewLat, NewLon)
	return Query{
		Size: 3,
		Sort: s,
	}
}

func (s *Sort) SetGeo(NewLat float64, NewLong float64) {
	s.Geo = GeoDistance{
		Order:          "asc",
		Unit:           "km",
		Mode:           "min",
		IgnoreUnmapped: true,
		DistanceType:   "arc",
		Location: Location{
			Lat:  NewLat,
			Long: NewLong,
		},
	}
}

type GeoDistance struct {
	Location       Location `json:"location"`
	Order          string   `json:"order"`
	Unit           string   `json:"unit"`
	Mode           string   `json:"mode"`
	IgnoreUnmapped bool     `json:"ignore_unmapped"`
	DistanceType   string   `json:"distance_type"`
}

type Location struct {
	Lat  float64 `json:"lat"`
	Long float64 `json:"lon"`
}

type Response struct {
	Name   string  `json:"name"`
	Places []Place `json:"places"`
}

func NewResponse(places []Place) Response {
	return Response{
		Name:   "places",
		Places: places,
	}
}
