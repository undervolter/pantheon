package olx

import (
	"strings"
	"testing"

	"pantheon/core/domain"
)

func TestBuildOffersQueryBasics(t *testing.T) {
	q := BuildOffersQuery(domain.Search{
		Query: " iphone 15 ", Category: 44, CityID: 268, RegionID: 25, DistrictID: 11,
		MinPrice: 1000, MaxPrice: 30000, OwnerType: "private",
	}, 35, 70)

	want := map[string]string{
		"offset": "70", "limit": "35", "sort_by": "created_at:desc",
		"query": "iphone 15", "category_id": "44", "city_id": "268", "district_id": "11",
		"filter_float_price:from": "1000", "filter_float_price:to": "30000", "owner_type": "private",
	}
	for k, v := range want {
		if got := q.Get(k); got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
	if q.Has("region_id") {
		t.Error("region_id must give way to city_id")
	}
}

func TestBuildOffersQueryClampsPaging(t *testing.T) {
	q := BuildOffersQuery(domain.Search{Query: "x"}, 500, -5)
	if q.Get("limit") != "50" || q.Get("offset") != "0" {
		t.Errorf("limit/offset = %s/%s, want 50/0", q.Get("limit"), q.Get("offset"))
	}
	if BuildOffersQuery(domain.Search{Query: "x"}, 0, 0).Get("limit") != "20" {
		t.Error("limit 0 must become the default page size")
	}
}

// Free-only mode is the «Безкоштовно» checkbox of OLX. It must not be
// combined with price bounds, which would exclude what it asks for.
func TestBuildOffersQueryFreeMode(t *testing.T) {
	for _, cat := range []int{0, 44, 108, 3711} {
		q := BuildOffersQuery(domain.Search{Category: cat, PriceMode: domain.PriceFree, MinPrice: 5, MaxPrice: 500}, 20, 0)
		if q.Get("filter_enum_price[0]") != "free" {
			t.Errorf("category %d: free mode must send the checkbox, got %s", cat, q.Encode())
		}
		for key := range q {
			if strings.HasPrefix(key, "filter_float_price") {
				t.Errorf("category %d: free mode must not send %s", cat, key)
			}
		}
	}
	if BuildOffersQuery(domain.Search{Query: "x"}, 20, 0).Has("filter_enum_price[0]") {
		t.Error("an ordinary filter must not send the free checkbox")
	}
}

func TestBuildOffersQueryFiltersAndBrands(t *testing.T) {
	q := BuildOffersQuery(domain.Search{
		Category:   73,
		BrandParam: "filter_enum_digital_camera_manufacturers", BrandValues: []string{"2580", "", "2591"},
		Filters: map[string][]string{
			"filter_enum_state":    {"used", "new"},
			"filter_bool_delivery": {"1"},
			"filter_bool_off":      {"0"},
		},
		Ranges: map[string]domain.FilterRange{"filter_float_motor_year": {From: 2010, To: 2020.5}},
	}, 20, 0)

	if q.Get("filter_enum_digital_camera_manufacturers[0]") != "2580" || q.Get("filter_enum_digital_camera_manufacturers[1]") != "2591" {
		t.Errorf("brand values must be indexed without gaps: %s", q.Encode())
	}
	if q.Get("filter_enum_state[0]") != "used" || q.Get("filter_enum_state[1]") != "new" {
		t.Errorf("enum values: %s", q.Encode())
	}
	if q.Get("filter_bool_delivery") != "1" || q.Has("filter_bool_off") {
		t.Errorf("bool filters: %s", q.Encode())
	}
	if q.Get("filter_float_motor_year:from") != "2010" || q.Get("filter_float_motor_year:to") != "2020.5" {
		t.Errorf("range: %s", q.Encode())
	}
}

// A filter that slipped past validation (an old database row, a bug) still
// must not be able to rewrite the request itself.
func TestBuildOffersQueryDropsForeignKeys(t *testing.T) {
	q := BuildOffersQuery(domain.Search{
		Query:      "x",
		BrandParam: "category_id", BrandValues: []string{"1"},
		Filters: map[string][]string{
			"limit": {"1"}, "sort_by": {"price:asc"}, "offset": {"900"},
			"filter_enum_price": {"exchange"}, "owner_type": {"business"},
			"filter_float_motor_year:from": {"1"},
		},
		Ranges: map[string]domain.FilterRange{
			"filter_float_price": {To: 1}, "limit": {From: 1}, "filter_float_motor_year:from": {From: 2010},
		},
	}, 20, 0)

	if q.Get("limit") != "20" || q.Get("sort_by") != "created_at:desc" || q.Get("offset") != "0" {
		t.Errorf("paging and sorting were overridden: %s", q.Encode())
	}
	for key := range q {
		switch key {
		case "offset", "limit", "sort_by", "query":
		default:
			t.Errorf("unexpected parameter %q reached the request: %s", key, q.Encode())
		}
	}
}
