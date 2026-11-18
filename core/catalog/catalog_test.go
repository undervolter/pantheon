package catalog

import (
	"strings"
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

func TestDeepCategoryNavigation(t *testing.T) {
	cam := FindCategory(73)
	if cam == nil {
		t.Fatal("category 73 (digital cameras) not found")
	}
	if cam.ParentID != 429 {
		t.Errorf("category 73 parent = %d, want 429", cam.ParentID)
	}

	bySlug := FindCategoryBySlug("tsifrovye-fotoapparaty")
	if bySlug == nil || bySlug.ID != 73 {
		t.Fatalf("FindCategoryBySlug = %+v", bySlug)
	}

	crumb := CategoryBreadcrumb(73)
	if !strings.Contains(crumb, "Електроніка") || !strings.Contains(crumb, "Фото") {
		t.Errorf("breadcrumb = %q, want Электроника/Фото/...", crumb)
	}
	if !strings.Contains(crumb, "→") {
		t.Errorf("breadcrumb must contain path separator: %q", crumb)
	}
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
		var found *CityItem
		for i := range TopCities {
			if TopCities[i].Name == w.Name {
				found = &TopCities[i]
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

func TestRentCategoryFilters(t *testing.T) {
	cat := FindCategory(1760) // Довгострокова оренда квартир
	if cat == nil {
		t.Fatal("rent category 1760 not found in categories.go")
	}
	if !HasFilterCategory(1760) {
		t.Fatal("rent category has no filters metadata")
	}

	// Проверяем, что до категории можно дойти из TopCategories.
	root := 1760
	for i := 0; i < 10; i++ {
		c := FindCategory(root)
		if c == nil || c.ParentID == 0 {
			break
		}
		root = c.ParentID
	}
	found := false
	for _, tc := range TopCategories {
		if tc.ID == root {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("root category %d is not in TopCategories", root)
	}
}
