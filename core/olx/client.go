package olx

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"

	"pantheon/core/catalog"
	"pantheon/core/domain"
)

const (
	OLXApiBase = "https://www.olx.ua/api/v1/offers/"
)

var offerHashRe = regexp.MustCompile(`(?i)-ID([A-Za-z0-9]+)\.html`)

type OffersResponse struct {
	Data []domain.Offer `json:"data"`
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
func (c *OLXClient) FetchOffers(search domain.Search, limit int) ([]domain.Offer, error) {
	return c.FetchOffersPage(search, limit, 0)
}

// FetchOffersPage возвращает страницу выдачи с указанным offset.
func (c *OLXClient) FetchOffersPage(search domain.Search, limit int, offset int) ([]domain.Offer, error) {
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
		if catalog.IsReservedFilterKey(key) {
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
		if catalog.IsReservedFilterKey(key) {
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
func (c *OLXClient) FindOfferByURL(rawURL string) (domain.Offer, error) {
	hash := ""
	if m := offerHashRe.FindStringSubmatch(rawURL); len(m) == 2 {
		hash = m[1]
	}
	if hash == "" {
		return domain.Offer{}, fmt.Errorf("в ссылке нет ID объявления: %s", rawURL)
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
		return domain.Offer{}, fmt.Errorf("не удалось построить поисковый запрос из ссылки")
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

		offers, err := c.FetchOffers(domain.Search{Query: query}, 50)
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
	return domain.Offer{}, lastErr
}
