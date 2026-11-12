package main

import (
	_ "embed"
	"encoding/json"
	"sort"
)

//go:embed filters_meta.json
var filtersMetaRaw []byte

// FilterValue — вариант значения фильтра (например, цвет или модель).
type FilterValue struct {
	Label string
	Value string
}

// FilterDef — определение фильтра/характеристики из metadata/filters.
type FilterDef struct {
	Key        string
	Label      string
	Type       string
	Unit       string
	Order      int
	Values     []FilterValue
	Categories []int
}

type rawFilterValue [2]string

type rawFilterDef struct {
	Key        string           `json:"k"`
	Label      string           `json:"l"`
	Type       string           `json:"t"`
	Unit       string           `json:"u"`
	Order      int              `json:"o"`
	Values     []rawFilterValue `json:"v"`
	Categories []int            `json:"c"`
}

// AllFilters — все определения фильтров OLX.
var AllFilters []FilterDef

var categoryFilters map[int][]int

func init() {
	var payload struct {
		Filters []rawFilterDef `json:"f"`
	}
	if err := json.Unmarshal(filtersMetaRaw, &payload); err != nil {
		return
	}

	AllFilters = make([]FilterDef, 0, len(payload.Filters))
	categoryFilters = make(map[int][]int)

	for _, raw := range payload.Filters {
		def := FilterDef{
			Key:        raw.Key,
			Label:      raw.Label,
			Type:       raw.Type,
			Unit:       raw.Unit,
			Order:      raw.Order,
			Categories: raw.Categories,
		}
		if len(raw.Values) > 0 {
			def.Values = make([]FilterValue, 0, len(raw.Values))
			for _, v := range raw.Values {
				def.Values = append(def.Values, FilterValue{Label: v[0], Value: v[1]})
			}
		}

		idx := len(AllFilters)
		AllFilters = append(AllFilters, def)
		for _, c := range raw.Categories {
			categoryFilters[c] = append(categoryFilters[c], idx)
		}
	}
}

// FiltersForCategory возвращает все фильтры, применимые к категории.
func FiltersForCategory(categoryID int) []FilterDef {
	idxs, ok := categoryFilters[categoryID]
	if !ok {
		return nil
	}
	out := make([]FilterDef, 0, len(idxs))
	for _, i := range idxs {
		out = append(out, AllFilters[i])
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Order == out[j].Order {
			return out[i].Key < out[j].Key
		}
		return out[i].Order < out[j].Order
	})
	return out
}

// HasFilterCategory сообщает, есть ли для категории хотя бы один фильтр.
func HasFilterCategory(categoryID int) bool {
	return len(categoryFilters[categoryID]) > 0
}
