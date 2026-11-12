package test

import (
	"strings"
	"testing"

	"pantheon/core/bot"
	"pantheon/core/olx"
)

func TestDeepCategoryNavigation(t *testing.T) {
	cam := olx.FindCategory(73)
	if cam == nil {
		t.Fatal("category 73 (digital cameras) not found")
	}
	if cam.ParentID != 429 {
		t.Errorf("category 73 parent = %d, want 429", cam.ParentID)
	}

	bySlug := olx.FindCategoryBySlug("tsifrovye-fotoapparaty")
	if bySlug == nil || bySlug.ID != 73 {
		t.Fatalf("FindCategoryBySlug = %+v", bySlug)
	}

	crumb := olx.CategoryBreadcrumb(73)
	if !strings.Contains(crumb, "Електроніка") || !strings.Contains(crumb, "Фото") {
		t.Errorf("breadcrumb = %q, want Электроника/Фото/...", crumb)
	}
	if !strings.Contains(crumb, "→") {
		t.Errorf("breadcrumb must contain path separator: %q", crumb)
	}

	ui := &bot.BotUI{}

	// Внутри "Фото/Відео" должна быть кнопка "Цифрові фотоапарати".
	_, km := ui.RenderSubcategoryStep(429)
	if !hasCallback(km, "add_subcat:73") {
		t.Error("subcategory step 429 has no add_subcat:73 button")
	}
	if !hasCallback(km, "add_cat_back:37") {
		t.Error("subcategory step 429 has no back-to-37 button")
	}

	// Внутри "Електроніка" должна быть кнопка "Фото/Відео".
	_, km = ui.RenderSubcategoryStep(37)
	if !hasCallback(km, "add_subcat:429") {
		t.Error("subcategory step 37 has no add_subcat:429 button")
	}

	// Для листа 73 должны показываться бренды.
	_, km = ui.RenderBrandStep(73, 0, nil)
	if !hasCallback(km, "add_brand_toggle:73:0:2580") {
		t.Error("brand step 73 has no Canon toggle button")
	}
}

func hasCallback(km *bot.InlineKeyboardMarkup, data string) bool {
	if km == nil {
		return false
	}
	for _, row := range km.InlineKeyboard {
		for _, b := range row {
			if b.CallbackData == data {
				return true
			}
		}
	}
	return false
}

func TestTopCitiesKnownIDs(t *testing.T) {
	want := []struct {
		Name   string
		City   int
		Region int
	}{
		{"📍 Київ", 268, 25},
		{"📍 Львів", 176, 5},
		{"📍 Одеса", 62, 9},
		{"📍 Дніпро", 121, 21},
		{"📍 Харків", 280, 8},
		{"📍 Запоріжжя", 194, 17},
		{"📍 Вінниця", 313, 24},
		{"📍 Полтава", 284, 15},
		{"📍 Івано-Франківськ", 132, 13},
		{"📍 Чернівці", 369, 18},
		{"📍 Черкаси", 145, 12},
		{"📍 Житомир", 324, 6},
		{"📍 Миколаїв", 106, 19},
		{"📍 Хмельницький", 101, 20},
		{"📍 Рівне", 124, 14},
		{"📍 Тернопіль", 362, 11},
		{"📍 Луцьк", 72, 22},
		{"📍 Ужгород", 247, 10},
	}

	for _, w := range want {
		var found *olx.CityItem
		for i := range olx.TopCities {
			if olx.TopCities[i].Name == w.Name {
				found = &olx.TopCities[i]
				break
			}
		}
		if found == nil {
			t.Errorf("город %q не найден в TopCities", w.Name)
			continue
		}
		if found.CityID != w.City || found.RegionID != w.Region {
			t.Errorf("%q: city_id=%d region_id=%d, want %d/%d",
				w.Name, found.CityID, found.RegionID, w.City, w.Region)
		}
	}
}
