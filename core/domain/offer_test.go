package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestOfferFilters(t *testing.T) {
	top := Offer{Promotion: Promotion{TopAd: true}}
	if !top.IsTopAd() {
		t.Error("top ad must be filtered")
	}
	normal := Offer{Promotion: Promotion{Options: []string{"bundle_optimum"}}}
	if normal.IsTopAd() {
		t.Error("bundle_optimum must not be treated as top ad")
	}
}

func TestPhotoURL(t *testing.T) {
	p := Photo{
		Width:  750,
		Height: 1000,
		Link:   "https://ireland.apollo.olxcdn.com/v1/files/abc-UA/image;s={width}x{height}",
	}
	want := "https://ireland.apollo.olxcdn.com/v1/files/abc-UA/image;s=750x1000"
	if got := p.URL(); got != want {
		t.Errorf("URL() = %q, want %q", got, want)
	}

	// Если размеров нет — подставляем дефолтные.
	p = Photo{Link: "https://x/image;s={width}x{height}"}
	if got := p.URL(); got != "https://x/image;s=600x600" {
		t.Errorf("default URL() = %q", got)
	}

	if got := (Photo{}).URL(); got != "" {
		t.Errorf("empty photo URL = %q", got)
	}
}

func TestOfferPhotoURLs(t *testing.T) {
	o := Offer{Photos: []Photo{
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
	var params []Param
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		t.Fatal(err)
	}
	o := Offer{Params: params}
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
	var params []Param
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		t.Fatal(err)
	}
	o := Offer{Params: params}
	if got := o.GetPrice(); got != "Договірна" {
		t.Errorf("GetPrice() = %q, want Договірна", got)
	}
}

func TestDisplayParams(t *testing.T) {
	raw := `[
		{"key":"price","name":"Ціна","type":"price","value":{"value":7500,"label":"7 500 грн."}},
		{"key":"motor_year","name":"Рік випуску","type":"input","value":{"value":"1986","label":"1986"}},
		{"key":"color","name":"Колір","type":"select","value":{"key":"green","label":"Зелений"}},
		{"key":"condition","name":"Технічний стан","type":"checkboxes","value":[{"key":"1","label":"На ходу"},{"key":"2","label":"Технічно справна"}]},
		{"key":"empty","name":"Пусто","type":"input","value":null}
	]`

	var params []Param
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		t.Fatal(err)
	}
	o := Offer{Params: params}
	got := o.DisplayParams()

	if len(got) != 3 {
		t.Fatalf("DisplayParams = %d, want 3: %+v", len(got), got)
	}
	if got[0].Label != "Рік випуску" || got[0].Value != "1986" {
		t.Errorf("param[0] = %+v", got[0])
	}
	if got[1].Label != "Колір" || got[1].Value != "Зелений" {
		t.Errorf("param[1] = %+v", got[1])
	}
	if got[2].Value != "На ходу, Технічно справна" {
		t.Errorf("param[2] = %+v", got[2])
	}
}
