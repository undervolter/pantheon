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

func offerWithPrice(t *testing.T, value string) Offer {
	t.Helper()
	var params []Param
	raw := `[{"key":"price","name":"Ціна","type":"price","value":` + value + `}]`
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		t.Fatal(err)
	}
	return Offer{Params: params}
}

// The shapes below are copied from live OLX responses: a free listing has
// value 0 and budget=false, only "type" tells it from a missing price.
func TestPriceKindFromLiveShapes(t *testing.T) {
	cases := []struct {
		name  string
		value string
		kind  PriceKind
		text  string
	}{
		{"free", `{"value":0,"type":"free","arranged":false,"budget":false,"currency":"UAH","label":"Безкоштовно"}`, PriceIsFree, "Безкоштовно"},
		{"exchange", `{"value":0,"type":"exchange","arranged":false,"budget":false,"currency":"UAH","label":"Обмін"}`, PriceExchange, "Обмін"},
		{"amount", `{"value":4500,"type":"price","currency":"UAH","label":"4 500 грн."}`, PriceAmount, "4 500 грн. (~$99)"},
		{"arranged with amount", `{"value":4000,"type":"arranged","arranged":true,"currency":"UAH"}`, PriceAmount, "4 000 грн. (~$88) (договірна)"},
		{"arranged without amount", `{"value":0,"arranged":true,"label":"Договірна"}`, PriceArranged, "Договірна"},
		{"free by label only", `{"value":0,"label":"Безкоштовно"}`, PriceIsFree, "Безкоштовно"},
		{"budget flag is not free", `{"value":0,"budget":true,"label":""}`, PriceNone, "Не вказана"},
	}
	for _, c := range cases {
		o := offerWithPrice(t, c.value)
		if got := o.PriceKind(); got != c.kind {
			t.Errorf("%s: PriceKind = %d, want %d", c.name, got, c.kind)
		}
		if got := o.GetPrice(); got != c.text {
			t.Errorf("%s: GetPrice = %q, want %q", c.name, got, c.text)
		}
	}

	none := Offer{}
	if none.PriceKind() != PriceNone || none.IsFree() {
		t.Error("an offer without a price param has no price")
	}
}

func TestFormatPriceUsesTheGivenWords(t *testing.T) {
	ru := PriceWords{UAH: "грн", NegotiableNote: "договорная", Negotiable: "Договорная", Free: "Бесплатно", Exchange: "Обмен", Unspecified: "Не указана"}
	free := offerWithPrice(t, `{"value":0,"type":"free","label":"Безкоштовно"}`)
	if got := free.FormatPrice(ru); got != "Бесплатно" {
		t.Errorf("free in ru = %q", got)
	}
	usd := offerWithPrice(t, `{"value":100,"type":"price","currency":"USD","converted_value":4530}`)
	if got := usd.FormatPrice(ru); got != "$100 (~4 530 грн)" {
		t.Errorf("usd in ru = %q", got)
	}
}

func TestPhotoURLIsCappedForTelegram(t *testing.T) {
	// 3024×4032 is the largest original seen in a live sample; Telegram has to
	// download up to ten of these itself.
	p := Photo{Width: 3024, Height: 4032, Link: "https://x/image;s={width}x{height}"}
	if got := p.URL(); got != "https://x/image;s=960x1280" {
		t.Errorf("portrait URL = %q, want the long side capped at %d", got, MaxPhotoSide)
	}
	p = Photo{Width: 4096, Height: 3072, Link: "https://x/image;s={width}x{height}"}
	if got := p.URL(); got != "https://x/image;s=1280x960" {
		t.Errorf("landscape URL = %q", got)
	}
}

func TestOfferCategoryIsParsed(t *testing.T) {
	var o Offer
	if err := json.Unmarshal([]byte(`{"id":1,"category":{"id":1758,"type":"real_estate"}}`), &o); err != nil {
		t.Fatal(err)
	}
	if o.Category.ID != 1758 || o.Category.Type != "real_estate" {
		t.Errorf("category = %+v", o.Category)
	}
}
