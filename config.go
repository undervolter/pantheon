package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// FilterRange — диапазон числового фильтра (например, год или пробег).
type FilterRange struct {
	From float64 `json:"from,omitempty"`
	To   float64 `json:"to,omitempty"`
}

type SearchConfig struct {
	ID            string                 `json:"id,omitempty"`
	Name          string                 `json:"name"`                     // Название фильтра
	Query         string                 `json:"query"`                    // Поисковая фраза
	QueryVariants []string               `json:"query_variants,omitempty"` // Нормализованные варианты для пост-фильтра
	MinPrice      float64                `json:"min_price,omitempty"`      // Мин цена (0 если без ограничений)
	MaxPrice      float64                `json:"max_price,omitempty"`      // Макс цена (0 если без ограничений)
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
	CreatedAt     int64                  `json:"created_at,omitempty"`     // Unix timestamp создания фильтра
	OwnerChatID   string                 `json:"owner_chat_id,omitempty"`  // chat_id владельца: алерты только ему (пусто — общий фильтр)
}

type Config struct {
	TelegramBotToken string         `json:"telegram_bot_token"` // Токен бота от @BotFather
	TelegramChatIDs  []string       `json:"telegram_chat_ids"`  // Список chat_id
	IntervalSeconds  int            `json:"interval_seconds"`   // Интервал проверки в секундах
	NotifyOnStartup  bool           `json:"notify_on_startup"`  // Слать ли уведомления о первых найденных
	Searches         []SearchConfig `json:"searches"`           // Список поисковых задач

	// Производительность: сколько поисков тянуть параллельно.
	Concurrency int `json:"concurrency,omitempty"`

	// Фото в алертах: отправлять ли альбом и сколько фото максимум.
	SendPhotos bool `json:"send_photos"`
	MaxPhotos  int  `json:"max_photos,omitempty"`
	// Rich-алерты (Bot API 10.1+): таблица, коллаж, expandable-описание.
	RichAlerts bool `json:"rich_alerts"`
	// Фильтрация «свежеподнятых» старых лотов: смотреть на created_time,
	// а не на last_refresh_time. true (по умолчанию) — старые поднятые лоты не алертятся.
	SkipBumpedOld bool `json:"skip_bumped_old"`
	// Максимальный возраст лота в минутах на момент проверки.
	// 0 — без ограничения.
	MaxAlertAgeMin int `json:"max_alert_age_min,omitempty"`

	// Характеристики лота в карточке (таблица в сворачиваемом блоке).
	ShowParams bool `json:"show_params"`
	MaxParams  int  `json:"max_params,omitempty"`
	// Ограничение выдачи за один опрос (по умолчанию 20).
	FetchLimit int `json:"fetch_limit,omitempty"`

	// ANSIHyperlinks включает OSC 8-ссылки в консольном логе.
	ANSIHyperlinks bool `json:"ansi_hyperlinks,omitempty"`

	mu       sync.RWMutex `json:"-"`
	filePath string       `json:"-"`
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения %s: %w", filename, err)
	}

	var cfg Config
	// По умолчанию фото в алертах включены.
	cfg.SendPhotos = true
	cfg.MaxPhotos = 3
	cfg.RichAlerts = true
	// По умолчанию не алертим старые лоты, которые просто подняли.
	cfg.SkipBumpedOld = true
	// И не алертим лоты старше 20 минут к моменту проверки.
	cfg.MaxAlertAgeMin = 20
	// Характеристики в карточке.
	cfg.ShowParams = true
	cfg.MaxParams = 15
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("ошибка парсинга %s: %w", filename, err)
	}

	if cfg.IntervalSeconds <= 0 {
		cfg.IntervalSeconds = 15
	}
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 4
	}
	if cfg.FetchLimit <= 0 {
		cfg.FetchLimit = 20
	}
	// OLX не принимает limit > 50 (HTTP 400).
	if cfg.FetchLimit > 50 {
		cfg.FetchLimit = 50
	}
	if cfg.MaxPhotos <= 0 {
		cfg.MaxPhotos = 3
	}
	if cfg.MaxPhotos > 10 {
		cfg.MaxPhotos = 10
	}
	cfg.filePath = filename

	return &cfg, nil
}

// PerfSettings возвращает настройки параллелизма одним снимком (без гонок).
func (c *Config) PerfSettings() (concurrency, fetchLimit int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Concurrency, c.FetchLimit
}

func (c *Config) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.saveLocked()
}

// saveLocked пишет конфиг на диск. Вызывать только под c.mu (Lock или RLock).
func (c *Config) saveLocked() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.filePath, data, 0644)
}

func (c *Config) GetSearches() []SearchConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()

	res := make([]SearchConfig, len(c.Searches))
	copy(res, c.Searches)
	return res
}

func (c *Config) GetActiveSearches() []SearchConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()

	res := make([]SearchConfig, 0, len(c.Searches))
	for _, s := range c.Searches {
		if !s.Disabled {
			res = append(res, s)
		}
	}
	return res
}

func (c *Config) AddSearch(s SearchConfig) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if s.CreatedAt == 0 {
		s.CreatedAt = time.Now().Unix()
	}
	c.Searches = append(c.Searches, s)
	return c.saveLocked()
}

func (c *Config) RemoveSearch(index int) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if index < 0 || index >= len(c.Searches) {
		return fmt.Errorf("неверный индекс фильтра: %d", index)
	}
	c.Searches = append(c.Searches[:index], c.Searches[index+1:]...)
	return c.saveLocked()
}

func (c *Config) ToggleSearch(index int) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if index < 0 || index >= len(c.Searches) {
		return false, fmt.Errorf("неверный индекс фильтра: %d", index)
	}
	c.Searches[index].Disabled = !c.Searches[index].Disabled
	state := !c.Searches[index].Disabled
	return state, c.saveLocked()
}

// SetSearchOwner закрепляет фильтр за конкретным чатом (владельцем алертов).
// Пустой chatID делает фильтр общим: алерты уходят всем получателям.
func (c *Config) SetSearchOwner(index int, chatID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if index < 0 || index >= len(c.Searches) {
		return fmt.Errorf("неверный индекс фильтра: %d", index)
	}
	c.Searches[index].OwnerChatID = chatID
	return c.saveLocked()
}

func (c *Config) SetInterval(sec int) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if sec < 3 {
		sec = 3
	}
	c.IntervalSeconds = sec
	return c.saveLocked()
}

func (c *Config) GetTelegramChatIDs() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	res := make([]string, len(c.TelegramChatIDs))
	copy(res, c.TelegramChatIDs)
	return res
}

func (c *Config) IsAuthorized(chatID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, id := range c.TelegramChatIDs {
		if id == chatID {
			return true
		}
	}
	return false
}

// AddTrustedChatID добавляет chat_id в белый список, если он ещё не там.
func (c *Config) AddTrustedChatID(chatID string, maxIDs int) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, id := range c.TelegramChatIDs {
		if id == chatID {
			return false, nil
		}
	}
	if maxIDs > 0 && len(c.TelegramChatIDs) >= maxIDs {
		return false, fmt.Errorf("достигнут лимит доверенных chat_id (%d)", maxIDs)
	}
	c.TelegramChatIDs = append(c.TelegramChatIDs, chatID)
	return true, c.saveLocked()
}

// RemoveChatID убирает chat_id из списка получателей.
func (c *Config) RemoveChatID(chatID string) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, id := range c.TelegramChatIDs {
		if id == chatID {
			c.TelegramChatIDs = append(c.TelegramChatIDs[:i], c.TelegramChatIDs[i+1:]...)
			return true, c.saveLocked()
		}
	}
	return false, nil
}

// AutoAuthorizeChatID — авто-привязка первых пользователей бота.
func (c *Config) AutoAuthorizeChatID(chatID string, maxValid int) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	valid := make([]string, 0, len(c.TelegramChatIDs))
	for _, id := range c.TelegramChatIDs {
		id = strings.TrimSpace(id)
		if id == "" || id == "YOUR_CHAT_ID_HERE" || id == "HOMIE_CHAT_ID_HERE" {
			continue
		}
		if id == chatID {
			return false, nil
		}
		valid = append(valid, id)
	}
	if maxValid > 0 && len(valid) >= maxValid {
		return false, fmt.Errorf("лимит доверенных пользователей (%d)", maxValid)
	}

	c.TelegramChatIDs = append(c.TelegramChatIDs, chatID)
	return true, c.saveLocked()
}

// GetInterval возвращает текущий интервал опроса.
func (c *Config) GetInterval() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.IntervalSeconds
}

func (c *Config) GetSendPhotos() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.SendPhotos
}

func (c *Config) GetMaxPhotos() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.MaxPhotos
}

func (c *Config) GetRichAlerts() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.RichAlerts
}

func (c *Config) GetSkipBumpedOld() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.SkipBumpedOld
}

func (c *Config) GetMaxAlertAgeMin() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.MaxAlertAgeMin
}

func (c *Config) GetShowParams() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ShowParams
}

func (c *Config) GetMaxParams() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.MaxParams
}

func (c *Config) GetANSIHyperlinks() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ANSIHyperlinks
}

func (c *Config) SetFetchLimit(limit int) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}
	c.FetchLimit = limit
	return c.saveLocked()
}

func (c *Config) SetConcurrency(n int) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if n < 1 {
		n = 1
	}
	c.Concurrency = n
	return c.saveLocked()
}

func (c *Config) SetRichAlerts(enabled bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.RichAlerts = enabled
	return c.saveLocked()
}

func (c *Config) SetSkipBumpedOld(enabled bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.SkipBumpedOld = enabled
	return c.saveLocked()
}

func (c *Config) SetMaxAlertAgeMin(minutes int) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if minutes < 0 {
		minutes = 0
	}
	c.MaxAlertAgeMin = minutes
	return c.saveLocked()
}

func (c *Config) SetShowParams(enabled bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ShowParams = enabled
	return c.saveLocked()
}

func (c *Config) SetMaxParams(n int) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if n < 1 {
		n = 1
	}
	c.MaxParams = n
	return c.saveLocked()
}

func (c *Config) SetANSIHyperlinks(enabled bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ANSIHyperlinks = enabled
	return c.saveLocked()
}

func (c *Config) SetSendPhotos(enabled bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.SendPhotos = enabled
	return c.saveLocked()
}

func (c *Config) SetMaxPhotos(n int) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if n < 1 {
		n = 1
	}
	if n > 10 {
		n = 10
	}
	c.MaxPhotos = n
	return c.saveLocked()
}
