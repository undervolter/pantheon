package domain

import (
	"encoding/json"
	"fmt"
	"html"
	"math"
	"strconv"
	"strings"
	"time"

	"pantheon/core/textutil"
)

// PriceValue — значение параметра "price" из ответа OLX.
type PriceValue struct {
	Value float64 `json:"value"`
	// Type — "price", "arranged", "free" или "exchange". У бесплатных и
	// обменных лотов Value равен нулю, отличает их только это поле.
	Type              string   `json:"type"`
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

// MaxPhotoSide — предел длинной стороны фото в алерте. OLX отдаёт оригиналы
// до 3000×4000, а Telegram должен успеть скачать их сам, до десяти штук на
// сообщение; крупные фото срывали отправку карточки.
const MaxPhotoSide = 1280

// URL подставляет размеры в CDN-шаблон OLX: .../image;s={width}x{height}.
// Размер вписывается в MaxPhotoSide с сохранением пропорций.
func (p Photo) URL() string {
	if p.Link == "" {
		return ""
	}
	w, h := p.Width, p.Height
	if w <= 0 || h <= 0 {
		w, h = 600, 600
	}
	if long := max(w, h); long > MaxPhotoSide {
		w = max(1, w*MaxPhotoSide/long)
		h = max(1, h*MaxPhotoSide/long)
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

// OfferCategory — категория лота из ответа OLX. Type — раздел верхнего уровня
// ("electronics", "services", "real_estate"…), ID — листовая категория.
type OfferCategory struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
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
	ID              int64         `json:"id"`
	URL             string        `json:"url"`
	Title           string        `json:"title"`
	CreatedTime     string        `json:"created_time"`
	LastRefreshTime string        `json:"last_refresh_time"`
	PushupTime      *string       `json:"pushup_time"`
	Description     string        `json:"description"`
	Params          []Param       `json:"params"`
	Category        OfferCategory `json:"category"`
	Location        Location      `json:"location"`
	Photos          []Photo       `json:"photos"`
	User            OfferUser     `json:"user"`
	Contact         OfferContact  `json:"contact"`
	Status          string        `json:"status"`
	Business        bool          `json:"business"`
	Promotion       Promotion     `json:"promotion"`
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

// roundPrice превращает цену в целое число единиц. OLX присылает float;
// битое или абсурдное значение не должно переполнить преобразование.
func roundPrice(v float64) int64 {
	const limit = 1e15
	switch {
	case math.IsNaN(v):
		return 0
	case v > limit:
		return int64(limit)
	case v < -limit:
		return -int64(limit)
	}
	return int64(math.Round(v))
}

func formatIntWithSpaces(n int64) string {
	if n < 0 {
		if n == math.MinInt64 {
			n++ // у MinInt64 нет положительной пары: -n осталось бы тем же числом
		}
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

// PriceKind — что продавец указал в поле цены.
type PriceKind int

const (
	PriceNone     PriceKind = iota // цены нет вовсе (услуги, вакансии) или она нераспознана
	PriceAmount                    // сумма, возможно с пометкой «договорная»
	PriceIsFree                    // «Безкоштовно»
	PriceExchange                  // «Обмін»
	PriceArranged                  // «Договірна» без суммы
)

// PriceWords — слова для текста цены на языке получателя.
type PriceWords struct {
	UAH            string // "грн."
	NegotiableNote string // "договірна" — пометка рядом с суммой
	Negotiable     string // "Договірна" — вместо суммы
	Free           string // "Безкоштовно"
	Exchange       string // "Обмін"
	Unspecified    string // "Не вказана"
}

// DefaultPriceWords — украинские слова, как на самом OLX. Используются в логе.
var DefaultPriceWords = PriceWords{
	UAH:            "грн.",
	NegotiableNote: "договірна",
	Negotiable:     "Договірна",
	Free:           "Безкоштовно",
	Exchange:       "Обмін",
	Unspecified:    "Не вказана",
}

func (o *Offer) priceValue() (PriceValue, bool) {
	for _, p := range o.Params {
		if p.Key != "price" {
			continue
		}
		var pv PriceValue
		if err := json.Unmarshal(p.Value, &pv); err != nil {
			continue
		}
		return pv, true
	}
	return PriceValue{}, false
}

// PriceKind определяет вид цены. Основной признак — поле type; флаги и
// подпись остаются запасным путём для ответов, где type не пришёл.
func (o *Offer) PriceKind() PriceKind {
	pv, ok := o.priceValue()
	if !ok {
		return PriceNone
	}
	switch strings.ToLower(strings.TrimSpace(pv.Type)) {
	case "free":
		return PriceIsFree
	case "exchange":
		return PriceExchange
	}
	if pv.Value > 0 {
		return PriceAmount
	}
	if pv.Arranged || strings.EqualFold(pv.Type, "arranged") {
		return PriceArranged
	}
	switch strings.ToLower(strings.TrimSpace(pv.Label)) {
	case "безкоштовно", "бесплатно":
		return PriceIsFree
	case "обмін", "обмен":
		return PriceExchange
	}
	return PriceNone
}

// IsFree сообщает, что лот отдают бесплатно.
func (o *Offer) IsFree() bool { return o.PriceKind() == PriceIsFree }

// FormatPrice возвращает цену строкой на языке получателя.
func (o *Offer) FormatPrice(w PriceWords) string {
	pv, _ := o.priceValue()
	switch o.PriceKind() {
	case PriceIsFree:
		return w.Free
	case PriceExchange:
		return w.Exchange
	case PriceArranged:
		return w.Negotiable
	case PriceNone:
		if label := strings.TrimSpace(pv.Label); label != "" {
			return label
		}
		return w.Unspecified
	}

	curr := strings.ToUpper(strings.TrimSpace(pv.Currency))
	valInt := roundPrice(pv.Value)
	note := ""
	if pv.Arranged {
		note = " (" + w.NegotiableNote + ")"
	}

	switch curr {
	case "USD":
		if pv.ConvertedValue != nil && *pv.ConvertedValue > 0 {
			convInt := roundPrice(*pv.ConvertedValue)
			return fmt.Sprintf("$%s (~%s %s)", formatIntWithSpaces(valInt), formatIntWithSpaces(convInt), w.UAH) + note
		}
		return fmt.Sprintf("$%s", formatIntWithSpaces(valInt)) + note

	case "EUR":
		if pv.ConvertedValue != nil && *pv.ConvertedValue > 0 {
			convInt := roundPrice(*pv.ConvertedValue)
			return fmt.Sprintf("€%s (~%s %s)", formatIntWithSpaces(valInt), formatIntWithSpaces(convInt), w.UAH) + note
		}
		return fmt.Sprintf("€%s", formatIntWithSpaces(valInt)) + note

	case "UAH", "":
		// Для цен в гривне от 500 грн добавляем ориентир в долларах по курсу ~45.3
		if valInt >= 500 {
			approxUSD := int64(math.Round(float64(valInt) / 45.3))
			return fmt.Sprintf("%s %s (~$%s)", formatIntWithSpaces(valInt), w.UAH, formatIntWithSpaces(approxUSD)) + note
		}
		return fmt.Sprintf("%s %s", formatIntWithSpaces(valInt), w.UAH) + note

	default:
		return fmt.Sprintf("%s %s", formatIntWithSpaces(valInt), pv.Currency) + note
	}
}

// GetPrice возвращает цену украинскими словами OLX — для лога и журнала.
func (o *Offer) GetPrice() string { return o.FormatPrice(DefaultPriceWords) }

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

// SearchText — заголовок и описание одной строкой для пост-фильтров.
func (o *Offer) SearchText() string {
	return o.Title + " " + textutil.StripHTMLTags(o.Description)
}
