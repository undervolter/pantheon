// Package olx talks to the public OLX.ua API: listing search, the geo
// encoder, and the parser that turns a search URL of the site into a filter.
package olx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"

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

// IsBlocked сообщает, что OLX отказал по адресу, а не по содержанию запроса:
// лимит частоты (429) или антибот-защита (403). Такой отказ касается всех
// фильтров сразу, поэтому ответ на него — пауза для всего опроса.
func IsBlocked(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && (apiErr.Status == 429 || apiErr.Status == 403)
}

// IsBadRequest сообщает, что OLX отверг именно этот запрос (400): проблема в
// параметрах фильтра, остальные фильтры ни при чём.
func IsBadRequest(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.Status == 400
}

func readBodyLimited(r io.Reader, limit int64) string {
	b, _ := io.ReadAll(io.LimitReader(r, limit))
	return strings.TrimSpace(string(b))
}

// geoIndexTTL — как долго живёт собранный список городов.
const geoIndexTTL = 24 * time.Hour

// geoPartialTTL — сколько живёт список, собранный не полностью.
const geoPartialTTL = 2 * time.Minute

type OLXClient struct {
	client    tls_client.HttpClient
	userAgent string

	geoMu      sync.Mutex
	geoEntries []geoEntry
	geoBuiltAt time.Time
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

func (c *OLXClient) headers() http.Header {
	return http.Header{
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
}

// FetchPage возвращает страницу выдачи фильтра. На limit=50 OLX отдаёт до 65
// лотов: сверху идут платные.
func (c *OLXClient) FetchPage(ctx context.Context, search domain.Search, limit, offset int) ([]domain.Offer, error) {
	reqURL := OLXApiBase + "?" + BuildOffersQuery(search, limit, offset).Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header = c.headers()

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", reqURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{Status: resp.StatusCode, Body: readBodyLimited(resp.Body, 2000), Method: "GET", URL: reqURL}
	}

	var offersResp OffersResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 32<<20)).Decode(&offersResp); err != nil {
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

		offers, err := c.FetchPage(context.Background(), domain.Search{Query: query}, 50, 0)
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
