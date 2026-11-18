package olx

import (
	"net/url"
	"sort"
	"strconv"
	"strings"

	"pantheon/core/catalog"
	"pantheon/core/domain"
	"pantheon/core/textutil"
)

// ParseOLXURL разбирает ссылку на поисковую выдачу OLX в SearchConfig.
func ParseOLXURL(rawURL string) (*domain.Search, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	sc := &domain.Search{}

	if query := q.Get("search[query]"); query != "" {
		sc.Query = query
	} else if query := q.Get("q"); query != "" {
		sc.Query = query
	}

	if maxP := q.Get("search[filter_float_price:to]"); maxP != "" {
		if p, err := strconv.ParseFloat(maxP, 64); err == nil {
			sc.MaxPrice = p
		}
	}
	if minP := q.Get("search[filter_float_price:from]"); minP != "" {
		if p, err := strconv.ParseFloat(minP, 64); err == nil {
			sc.MinPrice = p
		}
	}

	if v := firstNonEmpty(q.Get("search[city_id]"), q.Get("city_id")); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			sc.CityID = id
			if city := catalog.FindCity(id); city != nil {
				sc.CityName = city.Name
				sc.RegionID = city.RegionID
			}
		}
	}
	if v := firstNonEmpty(q.Get("search[region_id]"), q.Get("region_id")); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			sc.RegionID = id
		}
	}
	if v := firstNonEmpty(q.Get("search[category_id]"), q.Get("category_id")); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			sc.Category = id
			if cat := catalog.FindCategory(id); cat != nil {
				sc.CatName = cat.Name
			}
		}
	}

	if v := firstNonEmpty(q.Get("search[district_id]"), q.Get("district_id")); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			sc.DistrictID = id
		}
	}

	if v := firstNonEmpty(q.Get("owner_type"), q.Get("search[owner_type]"), q.Get("private_business"), q.Get("search[private_business]")); v != "" {
		switch v {
		case "private", "business":
			sc.OwnerType = v
		}
	}

	for k, vals := range q {
		base, sub := ParseFilterParam(k)
		if base == "" || catalog.IsReservedFilterKey(base) || base == sc.BrandParam {
			continue
		}
		switch {
		case strings.HasPrefix(base, "filter_float_") || strings.HasPrefix(base, "filter_int_"):
			if sc.Ranges == nil {
				sc.Ranges = map[string]domain.FilterRange{}
			}
			r := sc.Ranges[base]
			for _, v := range vals {
				f, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(v), ",", "."), 64)
				if err != nil {
					continue
				}
				if sub == "to" {
					r.To = f
				} else {
					r.From = f
				}
			}
			sc.Ranges[base] = r
		case strings.HasPrefix(base, "filter_enum_") || strings.HasPrefix(base, "filter_bool_"):
			if sc.Filters == nil {
				sc.Filters = map[string][]string{}
			}
			if strings.HasPrefix(base, "filter_bool_") {
				if len(vals) > 0 {
					sc.Filters[base] = []string{"1"}
				}
				continue
			}
			for _, v := range vals {
				v = strings.TrimSpace(v)
				if v == "" || textutil.StringInSlice(sc.Filters[base], v) {
					continue
				}
				sc.Filters[base] = append(sc.Filters[base], v)
			}
		}
	}

	brandVals := map[string][]string{}
	for k, vals := range q {
		cleanK := strings.TrimPrefix(k, "search[")
		cleanK = strings.TrimSuffix(cleanK, "]")
		if idx := strings.Index(cleanK, "]"); idx >= 0 {
			cleanK = cleanK[:idx]
		}
		cleanK = strings.TrimSuffix(cleanK, "]")
		if !strings.HasPrefix(cleanK, "filter_enum_") {
			continue
		}
		for _, v := range vals {
			v = strings.TrimSpace(v)
			if v != "" {
				brandVals[cleanK] = append(brandVals[cleanK], v)
			}
		}
	}

	if len(brandVals) > 0 {
		chosen := ""
		if sc.Category > 0 {
			if info := catalog.GetCategoryBrands(sc.Category); info != nil {
				if _, ok := brandVals[info.ParamName]; ok {
					chosen = info.ParamName
				}
			}
		}
		if chosen == "" {
			for p := range brandVals {
				chosen = p
				break
			}
		}

		sc.BrandParam = chosen
		seen := map[string]bool{}
		for _, id := range brandVals[chosen] {
			if seen[id] {
				continue
			}
			seen[id] = true
			sc.BrandValues = append(sc.BrandValues, id)
		}

		order := map[string]int{}
		nextOrder := 0
		for _, cinfo := range catalog.CategoryBrands {
			if cinfo.ParamName != chosen {
				continue
			}
			for _, b := range cinfo.Brands {
				if _, ok := order[b.ID]; !ok {
					order[b.ID] = nextOrder
					nextOrder++
				}
			}
		}
		sort.SliceStable(sc.BrandValues, func(i, j int) bool {
			oi, iok := order[sc.BrandValues[i]]
			oj, jok := order[sc.BrandValues[j]]
			if iok && jok {
				return oi < oj
			}
			if iok != jok {
				return iok
			}
			return sc.BrandValues[i] < sc.BrandValues[j]
		})

		nameByID := map[string]string{}
		for _, cinfo := range catalog.CategoryBrands {
			if cinfo.ParamName == chosen {
				for _, b := range cinfo.Brands {
					nameByID[b.ID] = b.Name
				}
			}
		}
		for _, id := range sc.BrandValues {
			name := nameByID[id]
			if name == "" {
				name = id
			}
			sc.BrandNames = append(sc.BrandNames, name)
		}

		if len(sc.BrandValues) == 1 {
			sc.BrandValue = sc.BrandValues[0]
			sc.BrandName = sc.BrandNames[0]
		}
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	for _, part := range parts {
		if strings.HasPrefix(part, "q-") {
			sc.Query = strings.TrimPrefix(part, "q-")
			sc.Query = strings.ReplaceAll(sc.Query, "-", " ")
		}
		if cat := catalog.FindCategoryBySlug(part); cat != nil {
			sc.Category = cat.ID
			sc.CatName = cat.Name
		}
	}

	if sc.Query != "" {
		sc.QueryVariants = textutil.QueryVariants(sc.Query)
	}

	if sc.Query != "" {
		sc.Name = sc.Query
	} else if sc.BrandName != "" {
		sc.Name = sc.BrandName
	} else if sc.CatName != "" {
		sc.Name = sc.CatName
	} else {
		sc.Name = "Поиск по ссылке"
	}

	return sc, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// ParseFilterParam разбирает параметры фильтра (filter_enum_..., filter_float_...).
func ParseFilterParam(raw string) (base, sub string) {
	s := raw
	if strings.HasPrefix(s, "search[") {
		s = s[len("search["):]
		if i := strings.Index(s, "]"); i >= 0 {
			base = s[:i]
			rest := s[i+1:]
			if strings.HasPrefix(rest, "[") && strings.HasSuffix(rest, "]") {
				sub = rest[1 : len(rest)-1]
			}
			return base, sub
		}
	}
	if i := strings.Index(s, "["); i > 0 && strings.HasSuffix(s, "]") {
		return s[:i], s[i+1 : len(s)-1]
	}
	if i := strings.Index(s, ":"); i > 0 {
		return s[:i], s[i+1:]
	}
	return s, ""
}
