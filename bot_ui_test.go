package main

import "testing"

func TestRenderCityResults(t *testing.T) {
	ui := &BotUI{}
	results := []GeoLocationSuggestion{
		{
			City:   GeoCity{ID: 221, Name: "Кременчук"},
			Region: GeoRegion{ID: 15, Name: "Полтавська область"},
		},
		{
			City:   GeoCity{ID: 284, Name: "Полтава"},
			Region: GeoRegion{ID: 15, Name: "Полтавська область"},
		},
	}

	_, km := ui.RenderCityResults("Кременчук", results)
	if km == nil {
		t.Fatal("no keyboard")
	}

	var data []string
	for _, row := range km.InlineKeyboard {
		for _, b := range row {
			data = append(data, b.CallbackData)
		}
	}

	hasCity1 := false
	hasCity2 := false
	hasRegion := 0
	for _, d := range data {
		switch d {
		case "add_city_pick:221:15":
			hasCity1 = true
		case "add_city_pick:284:15":
			hasCity2 = true
		case "add_region_pick:15":
			hasRegion++
		}
	}
	if !hasCity1 || !hasCity2 {
		t.Errorf("city buttons missing: %v", data)
	}
	if hasRegion != 1 {
		t.Errorf("region button must be unique, got %d: %v", hasRegion, data)
	}
}

func TestRenderCityStepHasManualButton(t *testing.T) {
	ui := &BotUI{}
	_, km := ui.RenderCityStep(SearchConfig{})
	if km == nil {
		t.Fatal("no keyboard")
	}
	found := false
	for _, row := range km.InlineKeyboard {
		for _, b := range row {
			if b.CallbackData == "add_city_manual" {
				found = true
			}
		}
	}
	if !found {
		t.Error("city step has no manual city button")
	}
}

func TestRenderCityResultsRegionOnly(t *testing.T) {
	ui := &BotUI{}
	results := []GeoLocationSuggestion{
		{Region: GeoRegion{ID: 15, Name: "Полтавська область"}},
	}
	_, km := ui.RenderCityResults("Полтавська", results)
	if km == nil {
		t.Fatal("no keyboard")
	}
	found := false
	for _, row := range km.InlineKeyboard {
		for _, b := range row {
			if b.CallbackData == "add_region_pick:15" {
				found = true
			}
		}
	}
	if !found {
		t.Error("region-only result must produce an add_region_pick button")
	}
}

func TestRenderPriceStepHasRangeButton(t *testing.T) {
	ui := &BotUI{}
	_, km := ui.RenderPriceStep(SearchConfig{})
	if km == nil {
		t.Fatal("no keyboard")
	}
	found := false
	for _, row := range km.InlineKeyboard {
		for _, b := range row {
			if b.CallbackData == "add_price_range" {
				found = true
			}
		}
	}
	if !found {
		t.Error("price step has no range button")
	}
}

func TestParsePriceRange(t *testing.T) {
	checks := []struct {
		in       string
		min, max float64
		ok       bool
	}{
		{"5000-20000", 5000, 20000, true},
		{"5 000 – 20 000", 5000, 20000, true},
		{"от 5000 до 20000", 5000, 20000, true},
		{"від 5000 до 20000", 5000, 20000, true},
		{"до 20000", 0, 20000, true},
		{"от 5000", 5000, 0, true},
		{"5000", 0, 5000, true},
		{"1000-2000 грн", 1000, 2000, true},
		{"-20000", 0, 20000, true},
		{"5000-", 5000, 0, true},
		{"грн", 0, 0, false},
		{"abc", 0, 0, false},
	}

	for _, c := range checks {
		gotMin, gotMax, ok := parsePriceRange(c.in)
		if ok != c.ok || gotMin != c.min || gotMax != c.max {
			t.Errorf("parsePriceRange(%q) = (%v, %v, %v), want (%v, %v, %v)",
				c.in, gotMin, gotMax, ok, c.min, c.max, c.ok)
		}
	}
}

func TestRenderDistrictStep(t *testing.T) {
	ui := &BotUI{}
	districts := []GeoDistrict{
		{ID: 11, Name: "Печерський"},
		{ID: 13, Name: "Подільський"},
	}
	_, km := ui.RenderDistrictStep("Київ", districts, 268, 25)
	if km == nil {
		t.Fatal("no keyboard")
	}
	foundDistrict := false
	foundWhole := false
	for _, row := range km.InlineKeyboard {
		for _, b := range row {
			if b.CallbackData == "add_district_pick:11:268:25" {
				foundDistrict = true
			}
			if b.CallbackData == "add_city_whole:268:25" {
				foundWhole = true
			}
		}
	}
	if !foundDistrict {
		t.Error("district button missing")
	}
	if !foundWhole {
		t.Error("whole-city button missing")
	}
}
