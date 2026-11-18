package olx

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"pantheon/core/catalog"
	"pantheon/core/domain"
)

const (
	defaultPageSize = 20
	// OLX rejects limit > 50 with HTTP 400.
	maxPageSize = 50
)

// BuildOffersQuery turns a filter into the query string of the offers API.
//
// Only ordinary OLX filter parameters reach the request: a key that is not
// one (limit, sort_by, category_id smuggled in through an imported file or a
// pasted link) is dropped here even if it got past validation.
func BuildOffersQuery(search domain.Search, limit, offset int) url.Values {
	q := url.Values{}
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = defaultPageSize
	}
	if limit > maxPageSize {
		limit = maxPageSize
	}
	q.Set("offset", strconv.Itoa(offset))
	q.Set("limit", strconv.Itoa(limit))
	// Despite the name this sorts by last_refresh_time, newest first.
	q.Set("sort_by", "created_at:desc")

	if query := strings.TrimSpace(search.Query); query != "" {
		q.Set("query", query)
	}
	if search.Category > 0 {
		q.Set("category_id", strconv.Itoa(search.Category))
	}

	if search.BrandParam != "" && catalog.IsBrandParam(search.BrandParam) {
		values := search.BrandValues
		if len(values) == 0 && search.BrandValue != "" {
			values = []string{search.BrandValue}
		}
		setIndexed(q, search.BrandParam, values)
	}

	if search.FreeOnly() {
		// The «Безкоштовно» checkbox of OLX. It works in every category and
		// without one, also where the site itself does not show the checkbox.
		q.Set("filter_enum_price[0]", "free")
	} else {
		if search.MinPrice > 0 {
			q.Set("filter_float_price:from", fmt.Sprintf("%.0f", search.MinPrice))
		}
		if search.MaxPrice > 0 {
			q.Set("filter_float_price:to", fmt.Sprintf("%.0f", search.MaxPrice))
		}
	}

	if search.CityID > 0 {
		q.Set("city_id", strconv.Itoa(search.CityID))
	} else if search.RegionID > 0 {
		q.Set("region_id", strconv.Itoa(search.RegionID))
	}
	if search.DistrictID > 0 {
		q.Set("district_id", strconv.Itoa(search.DistrictID))
	}
	if search.OwnerType == "private" || search.OwnerType == "business" {
		q.Set("owner_type", search.OwnerType)
	}

	for key, values := range search.Filters {
		if !domain.ValidFilterKey(key) || key == search.BrandParam {
			continue
		}
		if strings.HasPrefix(key, "filter_bool_") {
			if len(values) > 0 && (values[0] == "1" || strings.EqualFold(values[0], "true")) {
				q.Set(key, "1")
			}
			continue
		}
		setIndexed(q, key, values)
	}
	for key, r := range search.Ranges {
		if !domain.ValidFilterKey(key) {
			continue
		}
		if r.From > 0 {
			q.Set(key+":from", strconv.FormatFloat(r.From, 'f', -1, 64))
		}
		if r.To > 0 {
			q.Set(key+":to", strconv.FormatFloat(r.To, 'f', -1, 64))
		}
	}
	return q
}

// setIndexed writes key[0], key[1]… without gaps for skipped empty values.
func setIndexed(q url.Values, key string, values []string) {
	i := 0
	for _, v := range values {
		if v = strings.TrimSpace(v); v == "" {
			continue
		}
		q.Set(fmt.Sprintf("%s[%d]", key, i), v)
		i++
	}
}
