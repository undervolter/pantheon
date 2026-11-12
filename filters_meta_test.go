package main

import (
	"encoding/json"
	"testing"
)

func TestFiltersMetaLoaded(t *testing.T) {
	if len(AllFilters) < 600 {
		t.Fatalf("AllFilters = %d, want >= 600", len(AllFilters))
	}

	defs := FiltersForCategory(1115) // ВАЗ, легкові авто
	if len(defs) < 15 {
		t.Fatalf("FiltersForCategory(1115) = %d, want >= 15", len(defs))
	}

	hasOwner := false
	hasBody := false
	for _, d := range defs {
		switch d.Key {
		case "owner_type":
			hasOwner = true
		case "filter_enum_car_body":
			hasBody = true
		}
	}
	if !hasOwner {
		t.Error("category 1115 has no owner_type filter")
	}
	if !hasBody {
		t.Error("category 1115 has no car_body filter")
	}
}

func TestFilterValues(t *testing.T) {
	for _, d := range FiltersForCategory(1115) {
		if d.Key != "filter_enum_color" {
			continue
		}
		found := false
		for _, v := range d.Values {
			if v.Label == "Зелений" || v.Value == "green" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("color filter has no green value, got %d values", len(d.Values))
		}
		return
	}
	t.Error("color filter not found for category 1115")
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
