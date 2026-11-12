package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"math"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

const (
	OLXApiBase = "https://www.olx.ua/api/v1/offers/"
)

var offerHashRe = regexp.MustCompile(`(?i)-ID([A-Za-z0-9]+)\.html`)

type PriceValue struct {
	Value             float64  `json:"value"`
	Currency          string   `json:"currency"`
	Arranged          bool     `json:"arranged"`
	Budget            bool     `json:"budget"`
	Negotiable        bool     `json:"negotiable"`
	ConvertedValue    *float64 `json:"converted_value"`
	ConvertedCurrency string   `json:"converted_currency"`
	Label             string   `json:"label"`
}

type Param struct {
	Key   string          `json:"key"`
	Name  string          `json:"name"`
	Type  string          `json:"type"`
	Value json.RawMessage `json:"value"`
}

type Photo struct {
	ID       int64  `json:"id"`
	Filename string `json:"filename"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Rotation int    `json:"rotation"`
	Link     string `json:"link"`
}

// URL подставляет размеры в CDN-шаблон OLX: .../image;s={width}x{height}.
func (p Photo) URL() string {
	if p.Link == "" {
		return ""
	}
	w, h := p.Width, p.Height
	if w <= 0 || h <= 0 {
		w, h = 600, 600
	}
	link := strings.ReplaceAll(p.Link, "{width}", strconv.Itoa(w))
	link = strings.ReplaceAll(link, "{height}", strconv.Itoa(h))
	return link
}

type City struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Region struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type District struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Location struct {
	City     City     `json:"city"`
	Region   Region   `json:"region"`
	District District `json:"district"`
}

type OfferUser struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	Created         string `json:"created"`
	OtherAdsEnabled bool   `json:"other_ads_enabled"`
	SellerType      string `json:"seller_type"`
	UUID            string `json:"uuid"`
}

type OfferContact struct {
	Name  string `json:"name"`
	Phone bool   `json:"phone"`
	Chat  bool   `json:"chat"`
}

type Promotion struct {
	TopAd         bool     `json:"top_ad"`
	Highlighted   bool     `json:"highlighted"`
	PremiumAdPage bool     `json:"premium_ad_page"`
	Urgent        bool     `json:"urgent"`
	B2CAdPage     bool     `json:"b2c_ad_page"`
	Options       []string `json:"options"`
}

type Offer struct {
	ID              int64        `json:"id"`
	URL             string       `json:"url"`
	Title           string       `json:"title"`
	CreatedTime     string       `json:"created_time"`
	LastRefreshTime string       `json:"last_refresh_time"`
	PushupTime      *string      `json:"pushup_time"`
	Description     string       `json:"description"`
	Params          []Param      `json:"params"`
	Location        Location     `json:"location"`
	Photos          []Photo      `json:"photos"`
	User            OfferUser    `json:"user"`
	Contact         OfferContact `json:"contact"`
	Status          string       `json:"status"`
	Business        bool         `json:"business"`
	Promotion       Promotion    `json:"promotion"`
}

func (o *Offer) IsTopAd() bool {
	if o.Promotion.TopAd || o.Promotion.Highlighted || o.Promotion.PremiumAdPage {
		return true
	}
	for _, opt := range o.Promotion.Options {
		optLower := strings.ToLower(opt)
		if strings.Contains(optLower, "top") || strings.Contains(optLower, "premium") || strings.Contains(optLower, "highlight") {
			return true
		}
	}
	return false
}

func formatIntWithSpaces(n int64) string {
	if n < 0 {
		return "-" + formatIntWithSpaces(-n)
	}
	in := strconv.FormatInt(n, 10)
	if len(in) <= 3 {
		return in
	}
	var out []byte
	lead := len(in) % 3
	if lead == 0 {
		lead = 3
	}
	out = append(out, in[:lead]...)
	for i := lead; i < len(in); i += 3 {
		out = append(out, ' ')
		out = append(out, in[i:i+3]...)
	}
	return string(out)
}

func (o *Offer) GetPrice() string {
	for _, p := range o.Params {
		if p.Key != "price" {
			continue
		}
		var pv PriceValue
		if err := json.Unmarshal(p.Value, &pv); err != nil {
			continue
		}

		if pv.Value > 0 {
			curr := strings.ToUpper(strings.TrimSpace(pv.Currency))
			valInt := int64(math.Round(pv.Value))
			note := ""
			if pv.Arranged {
				note = " (договірна)"
			}

			switch curr {
			case "USD":
				if pv.ConvertedValue != nil && *pv.ConvertedValue > 0 {
					convInt := int64(math.Round(*pv.ConvertedValue))
					return fmt.Sprintf("$%s (~%s грн.)", formatIntWithSpaces(valInt), formatIntWithSpaces(convInt)) + note
				}
				return fmt.Sprintf("$%s", formatIntWithSpaces(valInt)) + note

			case "EUR":
				if pv.ConvertedValue != nil && *pv.ConvertedValue > 0 {
					convInt := int64(math.Round(*pv.ConvertedValue))
					return fmt.Sprintf("€%s (~%s грн.)", formatIntWithSpaces(valInt), formatIntWithSpaces(convInt)) + note
				}
				return fmt.Sprintf("€%s", formatIntWithSpaces(valInt)) + note

			case "UAH", "":
				// Для цен в гривне от 500 грн добавляем ориентир в долларах по курсу ~45.3
				if valInt >= 500 {
					approxUSD := int64(math.Round(float64(valInt) / 45.3))
					return fmt.Sprintf("%s грн. (~$%s)", formatIntWithSpaces(valInt), formatIntWithSpaces(approxUSD)) + note
				}
				return fmt.Sprintf("%s грн.", formatIntWithSpaces(valInt)) + note

			default:
				return fmt.Sprintf("%s %s", formatIntWithSpaces(valInt), pv.Currency) + note
			}
		}

		if pv.Arranged {
			return "Договірна"
		}
		if pv.Budget {
			return "Безкоштовно"
		}
		if pv.Label != "" {
			return pv.Label
		}
	}
	return "Не вказана"
}

// PriceUAH возвращает числовую цену лота в гривнах.
// Для валютных лотов берёт converted_value (OLX считает его по своему курсу).
// ok=false, если цены нет или она договорная/бюджетная без суммы.
func (o *Offer) PriceUAH() (float64, bool) {
	for _, p := range o.Params {
		if p.Key != "price" {
			continue
		}
		var pv PriceValue
		if err := json.Unmarshal(p.Value, &pv); err != nil {
			continue
		}
		if pv.Value <= 0 {
			return 0, false
		}

		switch strings.ToUpper(strings.TrimSpace(pv.Currency)) {
		case "UAH", "":
			return pv.Value, true
		case "USD", "EUR":
			if pv.ConvertedValue != nil && *pv.ConvertedValue > 0 {
				return *pv.ConvertedValue, true
			}
			return 0, false
		default:
			return 0, false
		}
	}
	return 0, false
}

// ParamKV — пара «характеристика → значение» для карточки.
type ParamKV struct {
	Key   string
	Label string
	Value string
}

// DisplayParams возвращает заполненные характеристики лота в порядке API,
// кроме цены и пустых значений.
func (o *Offer) DisplayParams() []ParamKV {
	out := make([]ParamKV, 0, len(o.Params))
	for _, p := range o.Params {
		if p.Key == "price" || p.Type == "price" {
			continue
		}
		v := paramValueLabel(p.Value)
		if v == "" {
			continue
		}
		label := strings.TrimSpace(p.Name)
		if label == "" {
			label = p.Key
		}
		out = append(out, ParamKV{Key: p.Key, Label: label, Value: v})
	}
	return out
}

// paramValueLabel вытаскивает человекочитаемое значение из произвольного value.
func paramValueLabel(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var v interface{}
	if err := json.Unmarshal(raw, &v); err != nil {
		return ""
	}
	return paramAnyLabel(v)
}

func paramAnyLabel(v interface{}) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(html.UnescapeString(t))
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		if t {
			return "Так"
		}
		return "Ні"
	case []interface{}:
		parts := make([]string, 0, len(t))
		for _, it := range t {
			if s := paramAnyLabel(it); s != "" {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, ", ")
	case map[string]interface{}:
		// Приоритет: label -> value -> key.
		for _, k := range []string{"label", "value", "key"} {
			if x, ok := t[k]; ok {
				if s := paramAnyLabel(x); s != "" {
					return s
				}
			}
		}
	}
	return ""
}

func (o *Offer) ParsedCreatedTime() (time.Time, error) {
	return time.Parse(time.RFC3339, o.CreatedTime)
}

func (o *Offer) ParsedRefreshTime() (time.Time, error) {
	return time.Parse(time.RFC3339, o.LastRefreshTime)
}

func (o *Offer) EffectiveTime() time.Time {
	if rTime, err := o.ParsedRefreshTime(); err == nil && !rTime.IsZero() {
		return rTime
	}
	if cTime, err := o.ParsedCreatedTime(); err == nil && !cTime.IsZero() {
		return cTime
	}
	return time.Time{}
}

// FreshnessTime возвращает время, по которому определяется «новизна» лота.
// preferCreated=true — смотрим на created_time (старые поднятые лоты не считаются новыми),
// preferCreated=false — старое поведение по last_refresh_time.
func (o *Offer) FreshnessTime(preferCreated bool) time.Time {
	if preferCreated {
		if cTime, err := o.ParsedCreatedTime(); err == nil && !cTime.IsZero() {
			return cTime
		}
	}
	return o.EffectiveTime()
}

// createdLabel возвращает created_time в компактном виде для логов.
func createdLabel(o Offer) string {
	if t, err := o.ParsedCreatedTime(); err == nil && !t.IsZero() {
		return t.Format("02.01.2006 15:04")
	}
	return "—"
}

func (o *Offer) GetLocationName() string {
	parts := []string{}
	if o.Location.City.Name != "" {
		parts = append(parts, o.Location.City.Name)
	}
	if o.Location.District.Name != "" {
		parts = append(parts, o.Location.District.Name+" р-н")
	}
	if o.Location.Region.Name != "" {
		parts = append(parts, o.Location.Region.Name)
	}
	if len(parts) == 0 {
		return "Україна"
	}
	return strings.Join(parts, ", ")
}

// GetCityDistrict возвращает компактную связку: город и район (например: Київ, Печерський р-н).
func (o *Offer) GetCityDistrict() string {
	var parts []string
	if o.Location.City.Name != "" {
		parts = append(parts, o.Location.City.Name)
	}
	if o.Location.District.Name != "" {
		dist := o.Location.District.Name
		if !strings.HasSuffix(strings.ToLower(dist), "р-н") && !strings.HasSuffix(strings.ToLower(dist), "район") {
			dist += " р-н"
		}
		parts = append(parts, dist)
	} else if o.Location.Region.Name != "" && o.Location.Region.Name != o.Location.City.Name {
		parts = append(parts, o.Location.Region.Name)
	}
	if len(parts) == 0 {
		return "Україна"
	}
	return strings.Join(parts, ", ")
}

// GetRegionCityDistrict возвращает локацию в формате "Область, Город - Район".
func (o *Offer) GetRegionCityDistrict() string {
	region := strings.TrimSpace(o.Location.Region.Name)
	region = strings.ReplaceAll(region, " область", " обл")
	region = strings.ReplaceAll(region, " Автономна Республіка", " АР")

	city := strings.TrimSpace(o.Location.City.Name)
	district := strings.TrimSpace(o.Location.District.Name)

	if region != "" && city != "" && strings.EqualFold(region, city) {
		region = ""
	}

	var parts []string
	if region != "" {
		parts = append(parts, region)
	}
	if city != "" {
		parts = append(parts, city)
	}
	loc := strings.Join(parts, ", ")

	if district != "" {
		d := district
		lower := strings.ToLower(d)
		if !strings.HasSuffix(lower, "р-н") && !strings.HasSuffix(lower, "район") {
			d += " р-н"
		}
		if loc == "" {
			loc = d
		} else {
			loc += " - " + d
		}
	}

	if loc == "" {
		return "Україна"
	}
	return loc
}

// PhotoURLs возвращает до limit уникальных ссылок на фото с подставленными размерами.
func (o *Offer) PhotoURLs(limit int) []string {
	if limit < 0 {
		return nil
	}
	if limit == 0 {
		limit = 3
	}
	if limit > 10 {
		limit = 10
	}

	urls := make([]string, 0, limit)
	seen := map[string]bool{}
	for _, p := range o.Photos {
		url := p.URL()
		if url == "" || seen[url] {
			continue
		}
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			continue
		}
		seen[url] = true
		urls = append(urls, url)
		if len(urls) >= limit {
			break
		}
	}
	return urls
}

type OffersResponse struct {
	Data []Offer `json:"data"`
}

// APIError — ошибка HTTP от OLX с кодом и телом ответа.
// Нужна, чтобы вызывающий код мог отличить 429 (rate limit) от 401/403 (протухший токен).
type APIError struct {
	Status int
	Body   string
	Method string
	URL    string
}

func (e *APIError) Error() string {
	prefix := ""
	if e.Method != "" || e.URL != "" {
		prefix = strings.TrimSpace(e.Method+" "+e.URL) + " -> "
	}
	return fmt.Sprintf("%sHTTP %d: %s", prefix, e.Status, e.Body)
}

func (e *APIError) IsRateLimited() bool { return e.Status == 429 }
func (e *APIError) IsAuthError() bool   { return e.Status == 401 }
func (e *APIError) IsForbidden() bool   { return e.Status == 403 }

// Detail пытается вытащить человекочитаемую причину из тела ответа OLX.
func (e *APIError) Detail() string {
	var fields map[string]json.RawMessage
	if json.Unmarshal([]byte(e.Body), &fields) == nil {
		// Формат limited-phones: {"error":"invalid_token","error_description":"..."}
		if raw, ok := fields["error"]; ok {
			var msg string
			if json.Unmarshal(raw, &msg) == nil && msg != "" {
				var desc string
				if d, ok := fields["error_description"]; ok {
					_ = json.Unmarshal(d, &desc)
				}
				if desc != "" {
					return msg + ": " + desc
				}
				return msg
			}
			// Формат {"error":{"detail":"..."}}
			var obj struct {
				Detail string `json:"detail"`
			}
			if json.Unmarshal(raw, &obj) == nil && obj.Detail != "" {
				return obj.Detail
			}
		}
		if raw, ok := fields["detail"]; ok {
			var msg string
			if json.Unmarshal(raw, &msg) == nil && msg != "" {
				return msg
			}
		}
	}
	return e.Body
}

func readBodyLimited(r io.Reader, limit int64) string {
	b, _ := io.ReadAll(io.LimitReader(r, limit))
	return strings.TrimSpace(string(b))
}

type OLXClient struct {
	client    tls_client.HttpClient
	userAgent string

	geoOnce    sync.Once
	geoEntries []geoEntry
	geoErr     error
}

func NewOLXClient() (*OLXClient, error) {
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(15),
		tls_client.WithClientProfile(profiles.Chrome_124),
		tls_client.WithRandomTLSExtensionOrder(),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		return nil, fmt.Errorf("tls-client init error: %w", err)
	}

	return &OLXClient{
		client:    client,
		userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	}, nil
}

// FetchOffers возвращает первую страницу выдачи.
func (c *OLXClient) FetchOffers(search SearchConfig, limit int) ([]Offer, error) {
	return c.FetchOffersPage(search, limit, 0)
}

// FetchOffersPage возвращает страницу выдачи с указанным offset.
func (c *OLXClient) FetchOffersPage(search SearchConfig, limit int, offset int) ([]Offer, error) {
	reqURL, err := url.Parse(OLXApiBase)
	if err != nil {
		return nil, err
	}

	q := reqURL.Query()
	if offset < 0 {
		offset = 0
	}
	q.Set("offset", strconv.Itoa(offset))
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}
	q.Set("limit", strconv.Itoa(limit))
	q.Set("sort_by", "created_at:desc") // Новейшие первыми

	if search.Query != "" {
		q.Set("query", search.Query)
	}
	if search.Category > 0 {
		q.Set("category_id", strconv.Itoa(search.Category))
	}
	if search.BrandParam != "" {
		values := search.BrandValues
		if len(values) == 0 && search.BrandValue != "" {
			values = []string{search.BrandValue}
		}
		for i, value := range values {
			if value == "" {
				continue
			}
			q.Set(fmt.Sprintf("%s[%d]", search.BrandParam, i), value)
		}
	}
	if search.MinPrice > 0 {
		q.Set("filter_float_price:from", fmt.Sprintf("%.0f", search.MinPrice))
	}
	if search.MaxPrice > 0 {
		q.Set("filter_float_price:to", fmt.Sprintf("%.0f", search.MaxPrice))
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
		if isReservedFilterKey(key) {
			continue
		}
		if strings.HasPrefix(key, "filter_bool_") {
			if len(values) > 0 && (values[0] == "1" || strings.EqualFold(values[0], "true")) {
				q.Set(key, "1")
			}
			continue
		}
		for i, v := range values {
			if strings.TrimSpace(v) == "" {
				continue
			}
			q.Set(fmt.Sprintf("%s[%d]", key, i), v)
		}
	}
	for key, r := range search.Ranges {
		if isReservedFilterKey(key) {
			continue
		}
		if r.From > 0 {
			q.Set(key+":from", strconv.FormatFloat(r.From, 'f', -1, 64))
		}
		if r.To > 0 {
			q.Set(key+":to", strconv.FormatFloat(r.To, 'f', -1, 64))
		}
	}

	reqURL.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header = http.Header{
		"accept":          {"application/json, text/plain, */*"},
		"accept-language": {"uk-UA,uk;q=0.9,ru;q=0.8,en-US;q=0.7,en;q=0.6"},
		"referer":         {"https://www.olx.ua/"},
		"user-agent":      {c.userAgent},
		http.HeaderOrderKey: {
			"accept",
			"accept-language",
			"referer",
			"user-agent",
		},
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", reqURL.String(), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{Status: resp.StatusCode, Body: readBodyLimited(resp.Body, 2000), Method: "GET", URL: reqURL.String()}
	}

	var offersResp OffersResponse
	if err := json.NewDecoder(resp.Body).Decode(&offersResp); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа OLX: %w", err)
	}

	return offersResp.Data, nil
}

// FindOfferByURL находит объявление по публичной ссылке OLX через поисковый API.
// В ссылке нет числового ID, поэтому ищем по хвосту -ID<hash>.html.
func (c *OLXClient) FindOfferByURL(rawURL string) (Offer, error) {
	hash := ""
	if m := offerHashRe.FindStringSubmatch(rawURL); len(m) == 2 {
		hash = m[1]
	}
	if hash == "" {
		return Offer{}, fmt.Errorf("в ссылке нет ID объявления: %s", rawURL)
	}

	slug := rawURL
	if i := strings.IndexAny(slug, "?#"); i >= 0 {
		slug = slug[:i]
	}
	slug = strings.TrimSuffix(slug, "/")
	if i := strings.LastIndex(slug, "/"); i >= 0 {
		slug = slug[i+1:]
	}
	slug = strings.TrimSuffix(slug, ".html")
	if i := strings.LastIndex(strings.ToUpper(slug), "-ID"); i >= 0 {
		slug = slug[:i]
	}
	slug = strings.ReplaceAll(slug, "-", " ")
	words := strings.Fields(slug)
	if len(words) > 8 {
		words = words[:8]
	}
	if len(words) == 0 {
		return Offer{}, fmt.Errorf("не удалось построить поисковый запрос из ссылки")
	}

	target := strings.ToLower("-ID" + hash + ".")
	var lastErr error

	// Пробуем от полного запроса к более короткому: поиск OLX любит лишние слова.
	for attempt := 0; attempt < 3; attempt++ {
		n := len(words) - attempt*2
		if n < 1 {
			n = 1
		}
		query := strings.Join(words[:n], " ")

		offers, err := c.FetchOffers(SearchConfig{Query: query}, 50)
		if err != nil {
			lastErr = err
			continue
		}
		for _, o := range offers {
			if strings.Contains(strings.ToLower(o.URL), target) {
				return o, nil
			}
		}
		lastErr = fmt.Errorf("объявление ID%s не найдено по запросу %q", hash, query)
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("объявление ID%s не найдено", hash)
	}
	return Offer{}, lastErr
}
