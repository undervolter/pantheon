// Package config holds what the operator sets up once (token, limits) and the
// small per-chat state the bot changes from its menus (trusted chats, admins,
// language, pause, quiet hours). Filters live in the database, not here.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	_ "time/tzdata" // quiet hours need Europe/Kyiv on hosts without a zone database

	"pantheon/core/domain"
)

const (
	// TokenEnv overrides telegram_bot_token and is never written back to disk.
	TokenEnv = "PANTHEON_BOT_TOKEN"

	defaultTimezone       = "Europe/Kyiv"
	defaultSeenRetention  = 7 * 24 * time.Hour
	defaultMaxFilters     = 50
	minIntervalSeconds    = 3
	defaultMaxAlertAgeMin = 30
)

// ErrLastAdmin is returned when a change would leave the bot without an admin.
var ErrLastAdmin = errors.New("the last admin cannot be removed")

type Config struct {
	TelegramBotToken string   `json:"telegram_bot_token"` // Токен бота от @BotFather
	TelegramChatIDs  []string `json:"telegram_chat_ids"`  // Доверенные чаты: получатели алертов и пользователи пульта
	AdminChatIDs     []string `json:"admin_chat_ids,omitempty"`
	IntervalSeconds  int      `json:"interval_seconds"`  // Интервал проверки в секундах
	NotifyOnStartup  bool     `json:"notify_on_startup"` // Слать ли уведомления о первых найденных

	// Searches — наследие 0.1.x: при первом запуске фильтры переезжают в базу
	// и из файла убираются.
	Searches []domain.Search `json:"searches,omitempty"`

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
	// Максимальный возраст лота в минутах на момент проверки. 0 — без ограничения.
	// Без omitempty: иначе «0» не переживал бы перезапуск и превращался в дефолт.
	MaxAlertAgeMin int `json:"max_alert_age_min"`

	// Характеристики лота в карточке (таблица в сворачиваемом блоке).
	ShowParams bool `json:"show_params"`
	MaxParams  int  `json:"max_params,omitempty"`
	// Ограничение выдачи за один опрос (по умолчанию 20).
	FetchLimit int `json:"fetch_limit,omitempty"`

	// ANSIHyperlinks включает OSC 8-ссылки в консольном логе. Без omitempty по
	// той же причине, что и MaxAlertAgeMin: «false» должен сохраняться.
	ANSIHyperlinks bool `json:"ansi_hyperlinks"`

	// Догрузка страниц при переполнении (backfill). nil — включена.
	Backfill *bool `json:"backfill,omitempty"`
	// DisableBackfill — старое имя той же настройки; при загрузке сливается в Backfill.
	DisableBackfill bool `json:"disable_backfill,omitempty"`

	// DataDir — где лежит pantheon.db. Пусто — рядом с файлом конфига.
	DataDir string `json:"data_dir,omitempty"`
	// Timezone — пояс для тихих часов. Пусто — Europe/Kyiv.
	Timezone string `json:"timezone,omitempty"`
	// SeenRetentionDays — сколько дней помнить просмотренный лот. 0 — 7 дней, меньше нуля — бессрочно.
	SeenRetentionDays int `json:"seen_retention_days,omitempty"`
	// MaxFiltersPerChat — страховочный лимит фильтров на чат. nil — 50, 0 — без лимита.
	MaxFiltersPerChat *int `json:"max_filters_per_chat,omitempty"`

	// Language — язык по умолчанию; Languages — выбранный язык по чатам.
	Language    string            `json:"language,omitempty"`
	Languages   map[string]string `json:"languages,omitempty"`
	PausedChats []string          `json:"paused_chats,omitempty"`
	// QuietHours — тихие часы по чатам в виде "22-8".
	QuietHours map[string]string `json:"quiet_hours,omitempty"`
	// AnnouncedVersion — для какой версии уже показано главное меню при старте.
	AnnouncedVersion string `json:"announced_version,omitempty"`

	mu       sync.RWMutex
	filePath string
	envToken string
	loc      *time.Location
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
	// И не алертим лоты старше 30 минут к моменту проверки. Меньше нельзя:
	// OLX индексирует лот через 10–18 минут после публикации.
	cfg.MaxAlertAgeMin = defaultMaxAlertAgeMin
	// Характеристики в карточке.
	cfg.ShowParams = true
	// 0 = без ограничений (все характеристики)
	cfg.MaxParams = 0
	// По умолчанию ANSI/OSC 8 гиперссылки включены.
	cfg.ANSIHyperlinks = true
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("ошибка парсинга %s: %w", filename, err)
	}

	if cfg.IntervalSeconds <= 0 {
		cfg.IntervalSeconds = 15
	}
	if cfg.IntervalSeconds < minIntervalSeconds {
		cfg.IntervalSeconds = minIntervalSeconds
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
	if cfg.MaxAlertAgeMin < 0 {
		cfg.MaxAlertAgeMin = 0
	}
	if cfg.DisableBackfill {
		off := false
		cfg.Backfill = &off
		cfg.DisableBackfill = false
	}

	// Плейсхолдеры вроде "your_chat_id" и прочий нечисловой мусор — не чаты.
	cfg.TelegramChatIDs = cleanIDs(cfg.TelegramChatIDs)
	cfg.AdminChatIDs = cleanIDs(cfg.AdminChatIDs)
	cfg.PausedChats = cleanIDs(cfg.PausedChats)

	cfg.filePath = filename
	cfg.envToken = strings.TrimSpace(os.Getenv(TokenEnv))
	return &cfg, nil
}

// ValidChatID сообщает, что строка — настоящий chat_id Telegram: ненулевое целое.
func ValidChatID(id string) bool {
	n, err := strconv.ParseInt(strings.TrimSpace(id), 10, 64)
	return err == nil && n != 0
}

// IsGroupID сообщает, что chat_id принадлежит группе или каналу (они отрицательные).
func IsGroupID(id string) bool {
	return strings.HasPrefix(strings.TrimSpace(id), "-")
}

func cleanIDs(ids []string) []string {
	out := make([]string, 0, len(ids))
	seen := map[string]bool{}
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if !ValidChatID(id) || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}

func contains(list []string, id string) bool {
	for _, v := range list {
		if v == id {
			return true
		}
	}
	return false
}

func without(list []string, id string) []string {
	out := list[:0:0]
	for _, v := range list {
		if v != id {
			out = append(out, v)
		}
	}
	return out
}

// Token возвращает токен бота: из окружения, если он там задан, иначе из файла.
func (c *Config) Token() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.envToken != "" {
		return c.envToken
	}
	return strings.TrimSpace(c.TelegramBotToken)
}

// Validate проверяет то, без чего запускаться бессмысленно.
func (c *Config) Validate() error {
	token := c.Token()
	if token == "" {
		return fmt.Errorf("не задан токен бота: впишите telegram_bot_token в конфиг или задайте %s", TokenEnv)
	}
	if !strings.Contains(token, ":") {
		return errors.New("telegram_bot_token не похож на токен от @BotFather (ожидается вид 123456789:AA…)")
	}
	return nil
}

// PerfSettings возвращает настройки параллелизма одним снимком (без гонок).
func (c *Config) PerfSettings() (concurrency, fetchLimit int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Concurrency, c.FetchLimit
}

// SetFilePath задаёт путь к файлу конфига на диске.
func (c *Config) SetFilePath(p string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.filePath = p
}

// FilePath возвращает путь, по которому конфиг читается и пишется.
func (c *Config) FilePath() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.filePath
}

// DataPath возвращает путь к файлу состояния (база, lock) в каталоге данных.
func (c *Config) DataPath(name string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	dir := strings.TrimSpace(c.DataDir)
	if dir == "" {
		dir = filepath.Dir(c.filePath)
	}
	return filepath.Join(dir, name)
}

func (c *Config) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.saveLocked()
}

// saveLocked пишет конфиг на диск атомарно: во временный файл и rename.
// Вызывать только под c.mu.Lock — два писателя делили бы один tmp-файл.
func (c *Config) saveLocked() error {
	if c.filePath == "" {
		return nil
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	tmp := c.filePath + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, c.filePath); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

// ---------- доверенные чаты и админы ----------

// GetTelegramChatIDs возвращает доверенные чаты: получателей общих алертов.
func (c *Config) GetTelegramChatIDs() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return append([]string(nil), c.TelegramChatIDs...)
}

// IsTrusted сообщает, что chat_id есть в списке доверенных.
func (c *Config) IsTrusted(chatID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return contains(c.TelegramChatIDs, strings.TrimSpace(chatID))
}

// IsAdmin сообщает, что пользователь может менять общие настройки и состав чатов.
func (c *Config) IsAdmin(chatID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return contains(c.AdminChatIDs, strings.TrimSpace(chatID))
}

// Admins возвращает список админов.
func (c *Config) Admins() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return append([]string(nil), c.AdminChatIDs...)
}

// EnsureAdmin один раз назначает админом первого пользователя из списка
// доверенных и записывает это в файл. Права не «переезжают» к другому
// пользователю при удалении чатов, потому что больше не выводятся из порядка.
// Возвращает false, если назначить некого — тогда нужна привязка по коду.
func (c *Config) EnsureAdmin() (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.AdminChatIDs) > 0 {
		return true, nil
	}
	for _, id := range c.TelegramChatIDs {
		if !IsGroupID(id) {
			c.AdminChatIDs = []string{id}
			return true, c.saveLocked()
		}
	}
	return false, nil
}

// BindFirstAdmin привязывает первого админа (после проверки кода привязки).
// Срабатывает только пока админов нет.
func (c *Config) BindFirstAdmin(chatID string) (bool, error) {
	chatID = strings.TrimSpace(chatID)
	if !ValidChatID(chatID) || IsGroupID(chatID) {
		return false, fmt.Errorf("недопустимый chat_id: %q", chatID)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.AdminChatIDs) > 0 {
		return false, nil
	}
	if !contains(c.TelegramChatIDs, chatID) {
		c.TelegramChatIDs = append(c.TelegramChatIDs, chatID)
	}
	c.AdminChatIDs = []string{chatID}
	return true, c.saveLocked()
}

// AddTrustedChatID добавляет chat_id в список доверенных, если он ещё не там.
func (c *Config) AddTrustedChatID(chatID string) (bool, error) {
	chatID = strings.TrimSpace(chatID)
	if !ValidChatID(chatID) {
		return false, fmt.Errorf("недопустимый chat_id: %q", chatID)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if contains(c.TelegramChatIDs, chatID) {
		return false, nil
	}
	c.TelegramChatIDs = append(c.TelegramChatIDs, chatID)
	return true, c.saveLocked()
}

// RemoveChatID убирает chat_id из доверенных вместе с его языком, паузой и
// тихими часами. Последнего админа удалить нельзя.
func (c *Config) RemoveChatID(chatID string) (bool, error) {
	chatID = strings.TrimSpace(chatID)
	c.mu.Lock()
	defer c.mu.Unlock()
	if !contains(c.TelegramChatIDs, chatID) {
		return false, nil
	}
	if contains(c.AdminChatIDs, chatID) && len(c.AdminChatIDs) == 1 {
		return false, ErrLastAdmin
	}
	c.TelegramChatIDs = without(c.TelegramChatIDs, chatID)
	c.AdminChatIDs = without(c.AdminChatIDs, chatID)
	c.PausedChats = without(c.PausedChats, chatID)
	delete(c.Languages, chatID)
	delete(c.QuietHours, chatID)
	return true, c.saveLocked()
}

// ---------- язык, пауза, тихие часы ----------

// LangFor возвращает язык чата: выбранный в меню, иначе язык по умолчанию из
// конфига. Пустая строка — язык не задан.
func (c *Config) LangFor(chatID string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if l := c.Languages[strings.TrimSpace(chatID)]; l != "" {
		return l
	}
	return c.Language
}

// HasLang сообщает, выбирал ли чат язык явно.
func (c *Config) HasLang(chatID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Languages[strings.TrimSpace(chatID)] != ""
}

func (c *Config) SetLang(chatID, lang string) error {
	chatID = strings.TrimSpace(chatID)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Languages == nil {
		c.Languages = map[string]string{}
	}
	if c.Languages[chatID] == lang {
		return nil
	}
	c.Languages[chatID] = lang
	return c.saveLocked()
}

// IsPaused сообщает, что алерты в этот чат поставлены на паузу.
func (c *Config) IsPaused(chatID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return contains(c.PausedChats, strings.TrimSpace(chatID))
}

func (c *Config) SetPaused(chatID string, paused bool) error {
	chatID = strings.TrimSpace(chatID)
	c.mu.Lock()
	defer c.mu.Unlock()
	if contains(c.PausedChats, chatID) == paused {
		return nil
	}
	if paused {
		c.PausedChats = append(c.PausedChats, chatID)
	} else {
		c.PausedChats = without(c.PausedChats, chatID)
	}
	return c.saveLocked()
}

// ParseQuietHours разбирает "22-8" в пару часов.
func ParseQuietHours(v string) (from, to int, ok bool) {
	a, b, found := strings.Cut(strings.TrimSpace(v), "-")
	if !found {
		return 0, 0, false
	}
	from, err1 := strconv.Atoi(strings.TrimSpace(a))
	to, err2 := strconv.Atoi(strings.TrimSpace(b))
	if err1 != nil || err2 != nil || from < 0 || from > 23 || to < 0 || to > 23 || from == to {
		return 0, 0, false
	}
	return from, to, true
}

func (c *Config) GetQuietHours(chatID string) (from, to int, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return ParseQuietHours(c.QuietHours[strings.TrimSpace(chatID)])
}

func (c *Config) SetQuietHours(chatID string, from, to int, on bool) error {
	chatID = strings.TrimSpace(chatID)
	c.mu.Lock()
	defer c.mu.Unlock()
	if !on {
		if _, had := c.QuietHours[chatID]; !had {
			return nil
		}
		delete(c.QuietHours, chatID)
		return c.saveLocked()
	}
	if c.QuietHours == nil {
		c.QuietHours = map[string]string{}
	}
	c.QuietHours[chatID] = fmt.Sprintf("%d-%d", from, to)
	return c.saveLocked()
}

// Location возвращает часовой пояс для тихих часов. Сервер бота часто живёт
// в UTC, поэтому берём не локальное время хоста, а настроенный пояс.
func (c *Config) Location() *time.Location {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.loc != nil {
		return c.loc
	}
	name := strings.TrimSpace(c.Timezone)
	if name == "" {
		name = defaultTimezone
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		loc = time.Local
	}
	c.loc = loc
	return loc
}

// InQuietHours сообщает, попадает ли момент в тихие часы чата. Интервал может
// переходить через полночь ("22-8").
func (c *Config) InQuietHours(chatID string, at time.Time) bool {
	from, to, ok := c.GetQuietHours(chatID)
	if !ok {
		return false
	}
	h := at.In(c.Location()).Hour()
	if from < to {
		return h >= from && h < to
	}
	return h >= from || h < to
}

// MarkAnnounced запоминает, что для этой версии главное меню уже разослано.
// Возвращает true только в первый раз.
func (c *Config) MarkAnnounced(version string) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.AnnouncedVersion == version {
		return false, nil
	}
	c.AnnouncedVersion = version
	return true, c.saveLocked()
}

// ---------- лимиты ----------

// SeenRetention возвращает срок хранения просмотренных лотов; 0 — бессрочно.
func (c *Config) SeenRetention() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	switch {
	case c.SeenRetentionDays < 0:
		return 0
	case c.SeenRetentionDays == 0:
		return defaultSeenRetention
	}
	return time.Duration(c.SeenRetentionDays) * 24 * time.Hour
}

// MaxFilters возвращает страховочный лимит фильтров на чат; 0 — без лимита.
func (c *Config) MaxFilters() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.MaxFiltersPerChat == nil {
		return defaultMaxFilters
	}
	if *c.MaxFiltersPerChat < 0 {
		return 0
	}
	return *c.MaxFiltersPerChat
}

// ---------- наследие 0.1.x ----------

// LegacySearches возвращает фильтры, оставшиеся в файле конфига от 0.1.x.
func (c *Config) LegacySearches() []domain.Search {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]domain.Search, 0, len(c.Searches))
	for _, s := range c.Searches {
		out = append(out, s.Clone())
	}
	return out
}

// DropLegacySearches убирает фильтры из файла конфига после их переезда в базу.
// Перед первой правкой исходный файл сохраняется как <config>.bak; уже
// существующая копия не перезаписывается.
func (c *Config) DropLegacySearches() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.Searches) == 0 || c.filePath == "" {
		c.Searches = nil
		return nil
	}
	bak := c.filePath + ".bak"
	if _, err := os.Stat(bak); os.IsNotExist(err) {
		original, err := os.ReadFile(c.filePath)
		if err != nil {
			return err
		}
		if err := os.WriteFile(bak, original, 0o600); err != nil {
			return err
		}
	}
	c.Searches = nil
	return c.saveLocked()
}

// ---------- настройки из меню ----------

func (c *Config) SetInterval(sec int) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if sec < minIntervalSeconds {
		sec = minIntervalSeconds
	}
	c.IntervalSeconds = sec
	return c.saveLocked()
}

// GetInterval возвращает текущий интервал опроса.
func (c *Config) GetInterval() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.IntervalSeconds
}

func (c *Config) GetNotifyOnStartup() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.NotifyOnStartup
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

func (c *Config) GetBackfillEnabled() bool {
	if c == nil {
		return true
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.DisableBackfill {
		return false
	}
	if c.Backfill != nil {
		return *c.Backfill
	}
	return true
}

func (c *Config) SetBackfillEnabled(enabled bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	b := enabled
	c.Backfill = &b
	c.DisableBackfill = false
	return c.saveLocked()
}
