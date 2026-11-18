package olx

import (
	"testing"
)

func TestParseOLXURL(t *testing.T) {
	raw := "https://www.ua/d/uk/elektronika/telefony-i-aksesuary/q-iphone-15/?search[filter_float_price:to]=30000&search[city_id]=268"
	sc, err := ParseOLXURL(raw)
	if err != nil {
		t.Fatal(err)
	}
	if sc.Query != "iphone 15" {
		t.Errorf("Query = %q", sc.Query)
	}
	if sc.Category != 44 || sc.CatName == "" {
		t.Errorf("Category = %d / %q", sc.Category, sc.CatName)
	}
	if sc.CityID != 268 || sc.RegionID != 25 {
		t.Errorf("City = %d / region %d", sc.CityID, sc.RegionID)
	}
	if sc.MaxPrice != 30000 {
		t.Errorf("MaxPrice = %v", sc.MaxPrice)
	}
	if sc.Name != "iphone 15" {
		t.Errorf("Name = %q", sc.Name)
	}
}

func TestParseOLXURLMultiBrand(t *testing.T) {
	raw := "https://www.ua/uk/elektronika/foto-video/tsifrovye-fotoapparaty/" +
		"?search[filter_enum_digital_camera_manufacturers][0]=2580" +
		"&search[filter_enum_digital_camera_manufacturers][1]=2591"

	sc, err := ParseOLXURL(raw)
	if err != nil {
		t.Fatal(err)
	}
	if sc.Category != 73 {
		t.Fatalf("Category = %d, want 73", sc.Category)
	}
	if sc.BrandParam != "filter_enum_digital_camera_manufacturers" {
		t.Errorf("BrandParam = %q", sc.BrandParam)
	}
	if len(sc.BrandValues) != 2 || sc.BrandValues[0] != "2580" || sc.BrandValues[1] != "2591" {
		t.Fatalf("BrandValues = %#v", sc.BrandValues)
	}
	if len(sc.BrandNames) != 2 || sc.BrandNames[0] != "Canon" || sc.BrandNames[1] != "Nikon" {
		t.Fatalf("BrandNames = %#v", sc.BrandNames)
	}
}

func TestParseFilterParam(t *testing.T) {
	cases := []struct {
		in, base, sub string
	}{
		{"search[filter_enum_color][0]", "filter_enum_color", "0"},
		{"filter_enum_color[3]", "filter_enum_color", "3"},
		{"search[filter_float_motor_year][from]", "filter_float_motor_year", "from"},
		{"filter_float_motor_year:to", "filter_float_motor_year", "to"},
		{"filter_enum_color", "filter_enum_color", ""},
	}
	for _, c := range cases {
		base, sub := ParseFilterParam(c.in)
		if base != c.base || sub != c.sub {
			t.Errorf("parseFilterParam(%q) = (%q, %q), want (%q, %q)", c.in, base, sub, c.base, c.sub)
		}
	}
}

func TestParseOLXURLFilters(t *testing.T) {
	u := "https://www.ua/uk/transport/legkovye-avtomobili/vaz/?search%5Bfilter_enum_color%5D%5B0%5D=4&search%5Bfilter_float_motor_year%5D%5Bfrom%5D=2000&search%5Bfilter_float_motor_year%5D%5Bto%5D=2010"
	sc, err := ParseOLXURL(u)
	if err != nil {
		t.Fatal(err)
	}
	colors := sc.Filters["filter_enum_color"]
	if len(colors) != 1 || colors[0] != "4" {
		t.Errorf("color filter = %v, want [4]", colors)
	}
	r, ok := sc.Ranges["filter_float_motor_year"]
	if !ok || r.From != 2000 || r.To != 2010 {
		t.Errorf("year range = %+v, want 2000..2010", r)
	}
}
