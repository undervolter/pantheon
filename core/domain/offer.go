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
