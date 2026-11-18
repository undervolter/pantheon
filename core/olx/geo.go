package olx

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	http "github.com/bogdanfinn/fhttp"
)

type GeoRegion struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	NormalizedName string `json:"normalized_name"`
}

type GeoCity struct {
	ID             int    `json:"id"`
	RegionID       int    `json:"region_id"`
	Name           string `json:"name"`
	NormalizedName string `json:"normalized_name"`
	Municipality   string `json:"municipality"`
	HasDistricts   bool   `json:"has_districts"`
}

type GeoDistrict struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	NormalizedName string `json:"normalized_name"`
}

type GeoLocationSuggestion struct {
	City         GeoCity   `json:"city"`
	Region       GeoRegion `json:"region"`
	Municipality struct {
		Name string `json:"name"`
	} `json:"municipality"`
}

// getGeoJSON дергает geo-encoder API OLX с теми же browser-like заголовками,
// что и остальные запросы.
func (c *OLXClient) getGeoJSON(path string, out interface{}) error {
	reqURL := "https://www.olx.ua" + path
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return err
	}
	req.Header = c.headers()

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("GET %s: %w", reqURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &APIError{Status: resp.StatusCode, Body: readBodyLimited(resp.Body, 2000), Method: "GET", URL: reqURL}
	}

	if err := json.NewDecoder(io.LimitReader(resp.Body, 2*1024*1024)).Decode(out); err != nil {
		return fmt.Errorf("decode %s: %w", reqURL, err)
	}
	return nil
}

func (c *OLXClient) FetchRegions() ([]GeoRegion, error) {
	var resp struct {
		Data []GeoRegion `json:"data"`
	}
	if err := c.getGeoJSON("/api/v1/geo-encoder/regions/", &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *OLXClient) FetchRegion(regionID int) (*GeoRegion, error) {
	var resp struct {
		Data GeoRegion `json:"data"`
	}
	path := fmt.Sprintf("/api/v1/geo-encoder/regions/%d", regionID)
	if err := c.getGeoJSON(path, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *OLXClient) FetchRegionCities(regionID int) ([]GeoCity, error) {
	var resp struct {
		Data []GeoCity `json:"data"`
	}
	path := fmt.Sprintf("/api/v1/geo-encoder/regions/%d/cities/", regionID)
	if err := c.getGeoJSON(path, &resp); err != nil {
		return nil, err
	}
	for i := range resp.Data {
		resp.Data[i].RegionID = regionID
	}
	return resp.Data, nil
}

func (c *OLXClient) FetchCity(cityID int) (*GeoCity, error) {
	var resp struct {
		Data GeoCity `json:"data"`
	}
	path := fmt.Sprintf("/api/v1/geo-encoder/cities/%d/", cityID)
	if err := c.getGeoJSON(path, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *OLXClient) FetchCityDistricts(cityID int) ([]GeoDistrict, error) {
	var resp struct {
		Data []GeoDistrict `json:"data"`
	}
	path := fmt.Sprintf("/api/v1/geo-encoder/cities/%d/districts/", cityID)
	if err := c.getGeoJSON(path, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *OLXClient) AutocompleteLocations(query string) ([]GeoLocationSuggestion, error) {
	var resp struct {
		Data []GeoLocationSuggestion `json:"data"`
	}
	path := "/api/v1/geo-encoder/location-autocomplete/?query=" + url.QueryEscape(query)
	if err := c.getGeoJSON(path, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// ---------- нечёткий поиск городов ----------

type geoEntry struct {
	City   GeoCity
	Region GeoRegion
}

func (c *OLXClient) WarmGeoIndex() {
	_, _ = c.geoIndex()
}

// geoIndex возвращает список всех городов по областям. Полный список живёт
// в кэше сутки. Неполный (часть областей не ответила) — пару минут: им можно
// пользоваться, но он должен скоро пересобраться, а не жить до перезапуска и
// не строиться заново на каждый запрос (это ~26 обращений к OLX).
func (c *OLXClient) geoIndex() ([]geoEntry, error) {
	c.geoMu.Lock()
	defer c.geoMu.Unlock()
	if len(c.geoEntries) > 0 && time.Since(c.geoBuiltAt) < geoIndexTTL {
		return c.geoEntries, nil
	}
	entries, complete, err := c.buildGeoIndex()
	if err != nil {
		return nil, err
	}
	if len(entries) > 0 {
		c.geoEntries, c.geoBuiltAt = entries, time.Now()
		if !complete {
			c.geoBuiltAt = c.geoBuiltAt.Add(geoPartialTTL - geoIndexTTL)
		}
	}
	return entries, nil
}

// buildGeoIndex собирает список городов. complete=false означает, что часть
// областей не ответила: результатом можно пользоваться, но кэшировать нельзя.
func (c *OLXClient) buildGeoIndex() (entries []geoEntry, complete bool, err error) {
	regions, err := c.FetchRegions()
	if err != nil {
		return nil, false, err
	}

	var mu sync.Mutex
	var firstErr error
	entries = make([]geoEntry, 0, 2048)
	sem := make(chan struct{}, 4)
	var wg sync.WaitGroup

	for _, region := range regions {
		wg.Add(1)
		go func(region GeoRegion) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			cities, err := c.FetchRegionCities(region.ID)
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
				return
			}
			mu.Lock()
			for _, city := range cities {
				entries = append(entries, geoEntry{City: city, Region: region})
			}
			mu.Unlock()
		}(region)
	}
	wg.Wait()

	if len(entries) == 0 {
		if firstErr != nil {
			return nil, false, firstErr
		}
		return nil, false, fmt.Errorf("geo-индекс пуст")
	}
	return entries, firstErr == nil, nil
}

// ResolveSlug ищет город или область по слагу из пути ссылки OLX
// (".../telefony/kiev/"). Возвращает нулевые ID, если слаг не найден.
func (c *OLXClient) ResolveSlug(slug string) (city GeoCity, region GeoRegion, ok bool) {
	slug = strings.ToLower(strings.TrimSpace(slug))
	if slug == "" {
		return GeoCity{}, GeoRegion{}, false
	}
	entries, err := c.geoIndex()
	if err != nil {
		return GeoCity{}, GeoRegion{}, false
	}
	for _, e := range entries {
		if strings.EqualFold(e.City.NormalizedName, slug) {
			return e.City, e.Region, true
		}
	}
	for _, e := range entries {
		if strings.EqualFold(e.Region.NormalizedName, slug) {
			return GeoCity{}, e.Region, true
		}
	}
	return GeoCity{}, GeoRegion{}, false
}

var geoNameReplacer = strings.NewReplacer(
	"ё", "е",
	"і", "и",
	"ї", "и",
	"є", "е",
	"ґ", "г",
	"’", "",
	"'", "",
	"-", " ",
	"_", " ",
)

func normalizeGeoName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = geoNameReplacer.Replace(s)
	return strings.Join(strings.Fields(s), " ")
}

func levenshtein(a, b string) int {
	ar := []rune(a)
	br := []rune(b)
	if len(ar) == 0 {
		return len(br)
	}
	if len(br) == 0 {
		return len(ar)
	}

	prev := make([]int, len(br)+1)
	cur := make([]int, len(br)+1)
	for j := 0; j <= len(br); j++ {
		prev[j] = j
	}

	for i := 1; i <= len(ar); i++ {
		cur[0] = i
		for j := 1; j <= len(br); j++ {
			cost := 1
			if ar[i-1] == br[j-1] {
				cost = 0
			}
			del := prev[j] + 1
			ins := cur[j-1] + 1
			sub := prev[j-1] + cost
			m := del
			if ins < m {
				m = ins
			}
			if sub < m {
				m = sub
			}
			cur[j] = m
		}
		prev, cur = cur, prev
	}
	return prev[len(br)]
}

func maxTypoDistance(query string) int {
	n := len([]rune(query))
	if n >= 9 {
		return 3
	}
	if n >= 5 {
		return 2
	}
	return 1
}

func locationDistance(normalizedQuery string, r GeoLocationSuggestion) int {
	best := levenshtein(normalizedQuery, normalizeGeoName(r.City.Name))
	if normalized := normalizeGeoName(r.City.NormalizedName); normalized != "" {
		if d := levenshtein(normalizedQuery, normalized); d < best {
			best = d
		}
	}
	return best
}

func rankLocations(query string, results []GeoLocationSuggestion) []GeoLocationSuggestion {
	q := normalizeGeoName(query)
	maxD := maxTypoDistance(q)

	type scored struct {
		item GeoLocationSuggestion
		dist int
	}
	var list []scored
	seen := map[int]bool{}

	for _, r := range results {
		if r.City.ID == 0 || seen[r.City.ID] {
			continue
		}
		d := locationDistance(q, r)
		if d <= maxD {
			seen[r.City.ID] = true
			list = append(list, scored{item: r, dist: d})
		}
	}

	sort.SliceStable(list, func(i, j int) bool {
		if list[i].dist != list[j].dist {
			return list[i].dist < list[j].dist
		}
		return list[i].item.City.Name < list[j].item.City.Name
	})

	out := make([]GeoLocationSuggestion, 0, len(list))
	for _, x := range list {
		out = append(out, x.item)
		if len(out) >= 8 {
			break
		}
	}
	return out
}

// SearchLocations ищет город/область: сначала точный autocomplete,
// потом префиксы (для опечаток в конце), потом нечёткий поиск по всему geo-индексу.
func (c *OLXClient) SearchLocations(query string) ([]GeoLocationSuggestion, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return nil, nil
	}

	exact, err := c.AutocompleteLocations(q)
	if err == nil && len(exact) > 0 {
		return exact, nil
	}

	// Префиксный fallback: "креминчуг" -> "крем".
	runes := []rune(q)
	maxLen := len(runes) - 1
	if maxLen > 6 {
		maxLen = 6
	}
	var candidates []GeoLocationSuggestion
	seen := map[int]bool{}
	for l := maxLen; l >= 4; l-- {
		prefix := string(runes[:l])
		res, perr := c.AutocompleteLocations(prefix)
		if perr != nil || len(res) == 0 {
			continue
		}
		for _, r := range res {
			if r.City.ID != 0 && !seen[r.City.ID] {
				seen[r.City.ID] = true
				candidates = append(candidates, r)
			}
		}
	}
	if len(candidates) > 0 {
		if ranked := rankLocations(q, candidates); len(ranked) > 0 {
			return ranked, nil
		}
	}

	return c.FuzzyLocations(q)
}

// FuzzyLocations ищет по всему списку городов/областей через расстояние Левенштейна.
func (c *OLXClient) FuzzyLocations(query string) ([]GeoLocationSuggestion, error) {
	entries, err := c.geoIndex()
	if err != nil {
		return nil, err
	}

	q := normalizeGeoName(query)
	if q == "" {
		return nil, nil
	}
	maxD := maxTypoDistance(q)

	type scored struct {
		item GeoLocationSuggestion
		dist int
	}
	var list []scored

	seenCities := map[int]bool{}
	for _, e := range entries {
		if seenCities[e.City.ID] {
			continue
		}
		r := GeoLocationSuggestion{City: e.City, Region: e.Region}
		d := locationDistance(q, r)
		if d <= maxD {
			seenCities[e.City.ID] = true
			list = append(list, scored{item: r, dist: d})
		}
	}

	// Области: дают кнопку «вся область» даже без конкретного города.
	seenRegions := map[int]bool{}
	for _, e := range entries {
		if seenRegions[e.Region.ID] {
			continue
		}
		seenRegions[e.Region.ID] = true
		d := levenshtein(q, normalizeGeoName(e.Region.Name))
		if d <= maxD {
			list = append(list, scored{
				item: GeoLocationSuggestion{Region: e.Region},
				dist: d,
			})
		}
	}

	sort.SliceStable(list, func(i, j int) bool {
		if list[i].dist != list[j].dist {
			return list[i].dist < list[j].dist
		}
		return list[i].item.City.Name < list[j].item.City.Name
	})

	out := make([]GeoLocationSuggestion, 0, len(list))
	for _, x := range list {
		out = append(out, x.item)
		if len(out) >= 8 {
			break
		}
	}
	return out, nil
}
