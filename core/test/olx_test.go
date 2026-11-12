package test

import (
	"encoding/json"
	"strings"
	"testing"

	"pantheon/core/olx"
)

func TestParseOLXURL(t *testing.T) {
	raw := "https://www.olx.ua/d/uk/elektronika/telefony-i-aksesuary/q-iphone-15/?search[filter_float_price:to]=30000&search[city_id]=268"
	sc, err := olx.ParseOLXURL(raw)
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
	raw := "https://www.olx.ua/uk/elektronika/foto-video/tsifrovye-fotoapparaty/" +
		"?search[filter_enum_digital_camera_manufacturers][0]=2580" +
		"&search[filter_enum_digital_camera_manufacturers][1]=2591"

	sc, err := olx.ParseOLXURL(raw)
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

func TestOfferFilters(t *testing.T) {
	top := olx.Offer{Promotion: olx.Promotion{TopAd: true}}
	if !top.IsTopAd() {
		t.Error("top ad must be filtered")
	}
	normal := olx.Offer{Promotion: olx.Promotion{Options: []string{"bundle_optimum"}}}
	if normal.IsTopAd() {
		t.Error("bundle_optimum must not be treated as top ad")
	}
}

func TestPhotoURL(t *testing.T) {
	p := olx.Photo{
		Width:  750,
		Height: 1000,
		Link:   "https://ireland.apollo.olxcdn.com/v1/files/abc-UA/image;s={width}x{height}",
	}
	want := "https://ireland.apollo.olxcdn.com/v1/files/abc-UA/image;s=750x1000"
	if got := p.URL(); got != want {
		t.Errorf("URL() = %q, want %q", got, want)
	}

	// Если размеров нет — подставляем дефолтные.
	p = olx.Photo{Link: "https://x/image;s={width}x{height}"}
	if got := p.URL(); got != "https://x/image;s=600x600" {
		t.Errorf("default URL() = %q", got)
	}

	if got := (olx.Photo{}).URL(); got != "" {
		t.Errorf("empty photo URL = %q", got)
	}
}

func TestOfferPhotoURLs(t *testing.T) {
	o := olx.Offer{Photos: []olx.Photo{
		{Width: 750, Height: 1000, Link: "https://x/1;s={width}x{height}"},
		{Width: 750, Height: 1000, Link: "https://x/1;s={width}x{height}"}, // дубль
		{Width: 750, Height: 1000, Link: "https://x/2;s={width}x{height}"},
		{Width: 750, Height: 1000, Link: "https://x/3;s={width}x{height}"},
	}}

	urls := o.PhotoURLs(2)
	if len(urls) != 2 {
		t.Fatalf("len = %d, want 2", len(urls))
	}
	if urls[0] != "https://x/1;s=750x1000" || urls[1] != "https://x/2;s=750x1000" {
		t.Fatalf("urls = %#v", urls)
	}
}

func TestGetPriceArrangedWithValue(t *testing.T) {
	raw := `[{"key":"price","name":"Ціна","type":"price","value":{"value":4000,"currency":"UAH","arranged":true,"budget":false,"negotiable":true,"label":"4 000 грн."}}]`
	var params []olx.Param
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		t.Fatal(err)
	}
	o := olx.Offer{Params: params}
	got := o.GetPrice()
	if !strings.Contains(got, "4 000") {
		t.Errorf("GetPrice() = %q, want actual amount 4 000", got)
	}
	if !strings.Contains(got, "договірна") {
		t.Errorf("GetPrice() = %q, want negotiable note", got)
	}
}

func TestGetPriceArrangedWithoutValue(t *testing.T) {
	raw := `[{"key":"price","name":"Ціна","type":"price","value":{"value":0,"currency":"UAH","arranged":true,"budget":false,"label":"Договірна"}}]`
	var params []olx.Param
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		t.Fatal(err)
	}
	o := olx.Offer{Params: params}
	if got := o.GetPrice(); got != "Договірна" {
		t.Errorf("GetPrice() = %q, want Договірна", got)
	}
}
