package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"pantheon/core/catalog"
	"pantheon/core/textutil"
)

// Режимы цены фильтра.
const (
	// PriceByRange — цена задаётся границами MinPrice/MaxPrice (или не задана).
	PriceByRange = ""
	// PriceFree — только лоты с меткой OLX «Безкоштовно».
	PriceFree = "free"
)

// Пределы пользовательского ввода.
const (
	MaxNameRunes    = 64
	MaxQueryRunes   = 100
	MaxExcludeWords = 20
	MaxExcludeRunes = 40
	MaxVariants     = 20
)

// FilterRange — диапазон числового фильтра (например, год или пробег).
type FilterRange struct {
	From float64 `json:"from,omitempty"`
	To   float64 `json:"to,omitempty"`
}

// Search — один фильтр выдачи OLX.
type Search struct {
	ID            string                 `json:"id,omitempty"`
	Name          string                 `json:"name"`                     // Название фильтра
	Query         string                 `json:"query"`                    // Поисковая фраза
	QueryVariants []string               `json:"query_variants,omitempty"` // Нормализованные варианты для пост-фильтра
	ExcludeWords  []string               `json:"exclude_words,omitempty"`  // Слова-исключения (совпадение по началу слова)
	MinPrice      float64                `json:"min_price,omitempty"`      // Мин цена (0 если без ограничений)
	MaxPrice      float64                `json:"max_price,omitempty"`      // Макс цена (0 если без ограничений)
	PriceMode     string                 `json:"price_mode,omitempty"`     // "" — по границам, "free" — только бесплатные
	CityID        int                    `json:"city_id,omitempty"`        // ID города (0 для всей Украины)
	CityName      string                 `json:"city_name,omitempty"`      // Название города
	RegionID      int                    `json:"region_id,omitempty"`      // ID области
	DistrictID    int                    `json:"district_id,omitempty"`    // ID района внутри города (0 — весь город)
	DistrictName  string                 `json:"district_name,omitempty"`  // Название района
	Category      int                    `json:"category_id,omitempty"`    // ID категории
	CatName       string                 `json:"category_name,omitempty"`  // Название категории
	BrandParam    string                 `json:"brand_param,omitempty"`    // Параметр бренда в API (например: filter_enum_mobile_phone_manufacturer)
	BrandValue    string                 `json:"brand_value,omitempty"`    // Значение бренда в API (например: 2065)
	BrandName     string                 `json:"brand_name,omitempty"`     // Отображаемое имя бренда (например: Apple)
	BrandValues   []string               `json:"brand_values,omitempty"`   // Мультивыбор брендов
	BrandNames    []string               `json:"brand_names,omitempty"`    // Имена выбранных брендов
	OwnerType     string                 `json:"owner_type,omitempty"`     // private | business | "" (все)
	Filters       map[string][]string    `json:"filters,omitempty"`        // enum/checkbox: ключ фильтра -> значения
	Ranges        map[string]FilterRange `json:"ranges,omitempty"`         // float/input: ключ -> диапазон
	Disabled      bool                   `json:"disabled,omitempty"`       // Поставлен на паузу
	SnoozedUntil  int64                  `json:"snoozed_until,omitempty"`  // Unix: до какого момента фильтр «спит»
	EnabledAt     int64                  `json:"enabled_at,omitempty"`     // Unix: когда фильтр последний раз возобновили
	CreatedAt     int64                  `json:"created_at,omitempty"`     // Unix timestamp создания фильтра
	OwnerChatID   string                 `json:"owner_chat_id,omitempty"`  // chat_id владельца: алерты только ему (пусто — общий фильтр)
}

// Clone возвращает глубокую копию: срезы и map фильтра не делятся с оригиналом.
func (s Search) Clone() Search {
	out := s
	out.QueryVariants = append([]string(nil), s.QueryVariants...)
	out.ExcludeWords = append([]string(nil), s.ExcludeWords...)
	out.BrandValues = append([]string(nil), s.BrandValues...)
	out.BrandNames = append([]string(nil), s.BrandNames...)
	if s.Filters != nil {
		out.Filters = make(map[string][]string, len(s.Filters))
		for k, v := range s.Filters {
			out.Filters[k] = append([]string(nil), v...)
		}
	}
	if s.Ranges != nil {
		out.Ranges = make(map[string]FilterRange, len(s.Ranges))
		for k, v := range s.Ranges {
			out.Ranges[k] = v
		}
	}
	return out
}

// IsShared сообщает, что у фильтра нет владельца: алерты идут всем чатам.
func (s Search) IsShared() bool { return s.OwnerChatID == "" }

// FreeOnly сообщает, что фильтр ловит только бесплатные лоты.
func (s Search) FreeOnly() bool { return s.PriceMode == PriceFree }

// Snoozed сообщает, что фильтр отложен и ещё не проснулся.
func (s Search) Snoozed(now time.Time) bool {
	return s.SnoozedUntil > now.Unix()
}

// Hidden сообщает, что фильтр относится к разделу, который бот не обслуживает.
func (s Search) Hidden() bool { return catalog.IsRealEstate(s.Category) }

// Active сообщает, что фильтр сейчас нужно опрашивать.
func (s Search) Active(now time.Time) bool {
	return !s.Disabled && !s.Snoozed(now) && !s.Hidden()
}

// Brands возвращает выбранные бренды одной строкой.
func (s Search) Brands() string {
	if len(s.BrandNames) > 0 {
		return strings.Join(s.BrandNames, ", ")
	}
	return s.BrandName
}

// HasCondition сообщает, что фильтр хоть чем-то ограничивает выдачу. Без
// этого он превращается в поток всех объявлений OLX.
func (s Search) HasCondition() bool {
	return strings.TrimSpace(s.Query) != "" || s.Category > 0 || s.FreeOnly()
}

// Ошибки проверки фильтра. Интерфейс переводит их по типу.
var (
	ErrNoCondition = errors.New("filter has no query, category or free-only mode")
	ErrRealEstate  = errors.New("real estate categories are not served")
	ErrBadOwner    = errors.New("unknown seller type")
	ErrBadPrice    = errors.New("unknown price mode")
)

// BadKeyError — параметр выдачи, который нельзя отправлять в OLX.
type BadKeyError struct{ Key string }

func (e *BadKeyError) Error() string { return fmt.Sprintf("filter key %q is not allowed", e.Key) }

var filterKeyRe = regexp.MustCompile(`^filter_(enum|float|int|bool)_[a-z0-9_]+$`)

// ValidFilterKey сообщает, что ключ — обычный параметр фильтра OLX, а не
// служебный параметр запроса (limit, sort_by, category_id…), который изменил
// бы саму выдачу.
func ValidFilterKey(key string) bool {
	return filterKeyRe.MatchString(key) && !catalog.IsReservedFilterKey(key)
}

// Normalize приводит фильтр к каноническому виду: обрезает строки, убирает
// пустое и дубли, ограничивает размеры. Вызывается перед Validate и записью.
func (s *Search) Normalize() {
	s.Name = textutil.TruncateRunes(strings.Join(strings.Fields(s.Name), " "), MaxNameRunes)
	s.Query = textutil.TruncateRunes(strings.Join(strings.Fields(s.Query), " "), MaxQueryRunes)
	s.OwnerType = strings.TrimSpace(s.OwnerType)
	s.PriceMode = strings.TrimSpace(s.PriceMode)
	s.OwnerChatID = strings.TrimSpace(s.OwnerChatID)
	// 0.1.x хранил названия вместе со значком («📍 Київ», «🌍 Полтавська
	// область»). Значок — оформление, его добавляет экран.
	s.CityName = strings.TrimSpace(strings.TrimLeft(s.CityName, "📍🌍🇺🇦 "))
	s.DistrictName = strings.TrimSpace(s.DistrictName)
	s.CatName = strings.TrimSpace(s.CatName)

	s.QueryVariants = uniqueNonEmpty(s.QueryVariants, MaxVariants, 0)
	s.ExcludeWords = uniqueNonEmpty(s.ExcludeWords, MaxExcludeWords, MaxExcludeRunes)

	if s.MinPrice < 0 {
		s.MinPrice = 0
	}
	if s.MaxPrice < 0 {
		s.MaxPrice = 0
	}
	if s.MaxPrice > 0 && s.MinPrice > s.MaxPrice {
		s.MinPrice, s.MaxPrice = s.MaxPrice, s.MinPrice
	}
	if s.FreeOnly() {
		s.MinPrice, s.MaxPrice = 0, 0
	}

	for key, values := range s.Filters {
		clean := uniqueNonEmpty(values, 0, 0)
		if len(clean) == 0 {
			delete(s.Filters, key)
			continue
		}
		s.Filters[key] = clean
	}
	for key, r := range s.Ranges {
		if r.From <= 0 && r.To <= 0 {
			delete(s.Ranges, key)
		}
	}
	if len(s.Filters) == 0 {
		s.Filters = nil
	}
	if len(s.Ranges) == 0 {
		s.Ranges = nil
	}

	s.BrandValues = uniqueNonEmpty(s.BrandValues, 0, 0)
	if len(s.BrandValues) == 0 && s.BrandValue == "" {
		s.BrandParam, s.BrandName, s.BrandNames = "", "", nil
	}
}

func uniqueNonEmpty(list []string, maxItems, maxRunes int) []string {
	if len(list) == 0 {
		return nil
	}
	out := make([]string, 0, len(list))
	seen := map[string]bool{}
	for _, v := range list {
		v = strings.Join(strings.Fields(v), " ")
		if maxRunes > 0 {
			v = string([]rune(v)[:min(len([]rune(v)), maxRunes)])
		}
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
		if maxItems > 0 && len(out) == maxItems {
			break
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// Validate проверяет фильтр перед сохранением. Одна и та же проверка стоит на
// пути из визарда, из ссылки OLX и из импортированного файла.
func (s Search) Validate() error {
	if s.Hidden() {
		return ErrRealEstate
	}
	if !s.HasCondition() {
		return ErrNoCondition
	}
	switch s.OwnerType {
	case "", "private", "business":
	default:
		return ErrBadOwner
	}
	switch s.PriceMode {
	case PriceByRange, PriceFree:
	default:
		return ErrBadPrice
	}
	for key := range s.Filters {
		if !ValidFilterKey(key) {
			return &BadKeyError{Key: key}
		}
	}
	for key := range s.Ranges {
		if !ValidFilterKey(key) {
			return &BadKeyError{Key: key}
		}
	}
	if s.BrandParam != "" && !catalog.IsBrandParam(s.BrandParam) {
		return &BadKeyError{Key: s.BrandParam}
	}
	return nil
}

// RuleKey — содержательная часть фильтра одной строкой: по ней находятся
// дубли при импорте. Имя, владелец и служебные отметки в ключ не входят.
func (s Search) RuleKey() string {
	c := s.Clone()
	c.Normalize()
	var b strings.Builder
	fmt.Fprintf(&b, "q=%s|cat=%d|city=%d|reg=%d|dist=%d|own=%s|pm=%s|min=%g|max=%g|bp=%s",
		strings.ToLower(c.Query), c.Category, c.CityID, c.RegionID, c.DistrictID, c.OwnerType,
		c.PriceMode, c.MinPrice, c.MaxPrice, c.BrandParam)
	b.WriteString("|bv=" + strings.Join(sortedCopy(append(c.BrandValues, c.BrandValue)), ","))
	b.WriteString("|qv=" + strings.Join(sortedCopy(c.QueryVariants), ","))
	b.WriteString("|ex=" + strings.Join(sortedCopy(c.ExcludeWords), ","))
	for _, key := range sortedKeys(c.Filters) {
		b.WriteString("|" + key + "=" + strings.Join(sortedCopy(c.Filters[key]), ","))
	}
	for _, key := range sortedKeys(c.Ranges) {
		fmt.Fprintf(&b, "|%s=%g..%g", key, c.Ranges[key].From, c.Ranges[key].To)
	}
	return b.String()
}
