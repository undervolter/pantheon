package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type CheckerStats struct {
	SeenCount      int
	SentCount      int64
	ActiveSearches int
	TotalSearches  int
	Interval       int
	StartedAt      time.Time
}

type Checker struct {
	cfg       *Config
	client    *OLXClient
	notifier  *TelegramNotifier
	startedAt time.Time
	seenMu    sync.RWMutex
	seenIDs   map[string]bool
	seenFile  string
	seenFh    *os.File
	seenBuf   *bufio.Writer
	runMu     sync.Mutex // не даём двум проверкам идти одновременно

	// frontier — до какого last_refresh_time мы уже дочитали выдачу по каждому фильтру.
	frontierMu sync.Mutex
	frontiers  map[string]time.Time
}

func NewChecker(cfg *Config, client *OLXClient, notifier *TelegramNotifier) *Checker {
	c := &Checker{
		cfg:       cfg,
		client:    client,
		notifier:  notifier,
		startedAt: time.Now(),
		seenIDs:   make(map[string]bool),
		seenFile:  "seen_ids.txt",
		frontiers: make(map[string]time.Time),
	}
	c.loadSeenIDs()
	c.openSeenWriter()
	return c
}

func (c *Checker) loadSeenIDs() {
	c.seenMu.Lock()
	defer c.seenMu.Unlock()

	f, err := os.Open(c.seenFile)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.Contains(line, "|") {
			c.seenIDs[line] = true
			continue
		}
		if _, err := strconv.ParseInt(line, 10, 64); err == nil {
			c.seenIDs[line] = true
		}
	}
	fmt.Printf("[Cache] Загружено %d записей о просмотренных лотах из %s\n", len(c.seenIDs), c.seenFile)
}

func seenKey(id int64, ownerChatID string) string {
	idStr := strconv.FormatInt(id, 10)
	if ownerChatID == "" {
		return "shared|" + idStr
	}
	return ownerChatID + "|" + idStr
}

func (c *Checker) openSeenWriter() {
	f, err := os.OpenFile(c.seenFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("[Cache] ⚠️ Не удалось открыть %s на запись: %v\n", c.seenFile, err)
		return
	}
	c.seenFh = f
	c.seenBuf = bufio.NewWriterSize(f, 64*1024)
}

func (c *Checker) saveSeenID(id int64, ownerChatID string) {
	key := seenKey(id, ownerChatID)
	c.seenMu.Lock()
	c.seenIDs[key] = true
	w := c.seenBuf
	c.seenMu.Unlock()

	if w != nil {
		_, _ = fmt.Fprintf(w, "%s\n", key)
	}
}

func (c *Checker) flushSeen() {
	c.seenMu.Lock()
	defer c.seenMu.Unlock()
	if c.seenBuf != nil {
		_ = c.seenBuf.Flush()
	}
	if c.seenFh != nil {
		_ = c.seenFh.Sync()
	}
}

const (
	skipReasonTopAd         = "Top ad"
	skipReasonOutdated      = "Outdated"
	skipReasonTooOld        = "Too old"
	skipReasonSpamTitle     = "Spam title"
	skipReasonQueryMismatch = "Query mismatch"
)

func (c *Checker) logAdSkipped(offer Offer, reason, details string) {
	line := fmt.Sprintf("[%s] skipped ad %d, reason: %s",
		time.Now().Format("15:04:05"), offer.ID, reason)
	if details != "" {
		line += " (" + details + ")"
	}
	fmt.Println(line)
}

func (c *Checker) logAdFound(offer Offer, filterName string) {
	line := fmt.Sprintf("[%s] found new ad! id %d | created at: %s | updated at: %s | price: %s | phone: %s | [%s]",
		time.Now().Format("15:04:05"),
		offer.ID,
		offerCreatedAtLabel(offer, true),
		offerUpdatedAtLabel(offer, true),
		sanitizeListingText(offer.GetPrice()),
		offerPhoneLabel(offer),
		sanitizeListingText(filterName),
	)
	if title := sanitizeListingText(offer.Title); title != "" {
		line += " | " + title
	}
	if offer.URL != "" {
		line += " | " + c.hyperlinkLabel(offer.URL, "link")
	}
	fmt.Println(line)
}

func (c *Checker) hyperlinkLabel(url, label string) string {
	if url == "" {
		return ""
	}
	if c.cfg == nil || !c.cfg.GetANSIHyperlinks() || !hyperlinksAllowed() {
		return url
	}
	return osc8Hyperlink(url, label)
}

func osc8Hyperlink(uri, label string) string {
	return "\x1b]8;;" + sanitizeOSCField(uri) + "\x1b\\" +
		sanitizeOSCField(label) + "\x1b]8;;\x1b\\"
}

func sanitizeOSCField(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case 0x1b, 0x07, '\n', '\r':
			return -1
		}
		return r
	}, s)
}

func hyperlinksAllowed() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("PANTHEON_ANSI_LINKS"))) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	}
	return !strings.EqualFold(os.Getenv("TERM"), "dumb")
}

func offerCreatedAtLabel(o Offer, withSeconds bool) string {
	t, err := o.ParsedCreatedTime()
	if err != nil || t.IsZero() {
		return "unknown"
	}
	if withSeconds {
		return t.Format("02.01.2006 15:04:05")
	}
	return t.Format("02.01.2006 15:04")
}

func offerUpdatedAtLabel(o Offer, withSeconds bool) string {
	t, err := o.ParsedRefreshTime()
	if err != nil || t.IsZero() {
		return "unknown"
	}
	if withSeconds {
		return t.Format("02.01.2006 15:04:05")
	}
	return t.Format("02.01.2006 15:04")
}

func offerPhoneLabel(o Offer) string {
	if o.Contact.Phone {
		return "yes"
	}
	return "no"
}

func topAdDetail(o Offer) string {
	switch {
	case o.Promotion.TopAd:
		return "top_ad"
	case o.Promotion.Highlighted:
		return "highlighted"
	case o.Promotion.PremiumAdPage:
		return "premium_ad_page"
	default:
		return "promoted"
	}
}

func (c *Checker) Close() {
	c.runMu.Lock()
	c.runMu.Unlock()

	c.flushSeen()
	c.seenMu.Lock()
	defer c.seenMu.Unlock()
	if c.seenFh != nil {
		_ = c.seenFh.Close()
		c.seenFh = nil
		c.seenBuf = nil
	}
}

func (c *Checker) isSeen(id int64, ownerChatID string) bool {
	idStr := strconv.FormatInt(id, 10)
	c.seenMu.RLock()
	defer c.seenMu.RUnlock()

	if c.seenIDs[idStr] {
		return true
	}
	return c.seenIDs[seenKey(id, ownerChatID)]
}

func (c *Checker) GetStats() CheckerStats {
	c.seenMu.RLock()
	seenCount := len(c.seenIDs)
	c.seenMu.RUnlock()

	searches := c.cfg.GetSearches()
	activeCount := 0
	for _, s := range searches {
		if !s.Disabled {
			activeCount++
		}
	}

	return CheckerStats{
		SeenCount:      seenCount,
		SentCount:      c.notifier.GetSentCount(),
		ActiveSearches: activeCount,
		TotalSearches:  len(searches),
		Interval:       c.cfg.GetInterval(),
		StartedAt:      c.startedAt,
	}
}

func (c *Checker) GetSentCount() int64 {
	return c.notifier.GetSentCount()
}

func (c *Checker) Start() {
	active := c.cfg.GetActiveSearches()
	fmt.Printf("🚀 Монитор OLX запущен в %s. Активных задач: %d (всего: %d). Интервал: %d сек.\n\n",
		c.startedAt.Format("15:04:05 02.01.2006"), len(active), len(c.cfg.GetSearches()), c.cfg.GetInterval())

	// Первый прогон (наполняем кэш и отсекаем старые)
	c.checkAll(true)

	for {
		interval := c.cfg.GetInterval()
		if interval < 3 {
			interval = 3
		}
		time.Sleep(time.Duration(interval) * time.Second)
		c.checkAll(false)
	}
}

func (c *Checker) CheckNow() {
	c.checkAll(false)
}

type searchFetchResult struct {
	offers []Offer
	err    error
}

type pendingOffer struct {
	offer      Offer
	searchName string
}

func searchKey(s SearchConfig) string {
	if s.ID != "" {
		return s.ID
	}
	return fmt.Sprintf("%s|%s|%d|%s|%d|%d", s.OwnerChatID, s.Name, s.Category, s.Query, s.CityID, s.DistrictID)
}

func (c *Checker) getFrontier(key string) time.Time {
	c.frontierMu.Lock()
	defer c.frontierMu.Unlock()
	return c.frontiers[key]
}

func (c *Checker) setFrontier(key string, t time.Time) {
	if t.IsZero() {
		return
	}
	c.frontierMu.Lock()
	defer c.frontierMu.Unlock()
	if cur := c.frontiers[key]; t.After(cur) {
		c.frontiers[key] = t
	}
}

const defaultStartupGrace = 15 * time.Minute

func alertStartupGrace(maxAlertAgeMin int) time.Duration {
	grace := defaultStartupGrace
	if maxAlertAgeMin > 0 {
		if d := time.Duration(maxAlertAgeMin) * time.Minute; d > grace {
			grace = d
		}
	}
	return grace
}

func pageOldestRefresh(offers []Offer) time.Time {
	var oldest time.Time
	for _, o := range offers {
		if o.IsTopAd() {
			continue
		}
		t, err := o.ParsedRefreshTime()
		if err != nil || t.IsZero() {
			continue
		}
		if oldest.IsZero() || t.Before(oldest) {
			oldest = t
		}
	}
	return oldest
}

func pageNewestRefresh(offers []Offer) time.Time {
	var newest time.Time
	for _, o := range offers {
		if o.IsTopAd() {
			continue
		}
		t, err := o.ParsedRefreshTime()
		if err != nil || t.IsZero() {
			continue
		}
		if newest.IsZero() || t.After(newest) {
			newest = t
		}
	}
	return newest
}

func pageReached(offers []Offer, prev time.Time) bool {
	oldest := pageOldestRefresh(offers)
	if oldest.IsZero() {
		return true
	}
	return !oldest.After(prev)
}

const maxBackfillPages = 3

func (c *Checker) fetchOffersWithBackfill(search SearchConfig, limit int) ([]Offer, error) {
	key := searchKey(search)
	prev := c.getFrontier(key)

	first, err := c.client.FetchOffersPage(search, limit, 0)
	if err != nil {
		return nil, err
	}
	if len(first) == 0 {
		return first, nil
	}

	newest := pageNewestRefresh(first)
	if prev.IsZero() || pageReached(first, prev) {
		if prev.IsZero() {
			fmt.Printf("[%s] 📌 '%s': фронт выдачи %s\n",
				time.Now().Format("15:04:05"), search.Name, newest.Format("15:04:05"))
		}
		c.setFrontier(key, newest)
		return first, nil
	}

	all := first
	extra := 0
	reached := false
	for page := 1; page <= maxBackfillPages; page++ {
		next, err := c.client.FetchOffersPage(search, limit, page*limit)
		if err != nil {
			break
		}
		if len(next) == 0 {
			reached = true
			break
		}
		all = append(all, next...)
		extra += len(next)
		if t := pageNewestRefresh(next); t.After(newest) {
			newest = t
		}
		if pageReached(next, prev) {
			reached = true
			break
		}
	}

	c.setFrontier(key, newest)

	if extra > 0 {
		fmt.Printf("[%s] ↕️ '%s': догружено %d лотов (страница переполнена)\n",
			time.Now().Format("15:04:05"), search.Name, extra)
	}
	if !reached {
		fmt.Printf("[%s] ⚠️ '%s': не дошли до прошлого фронта за %d доп. страниц — возможна потеря\n",
			time.Now().Format("15:04:05"), search.Name, maxBackfillPages)
	}
	return all, nil
}

func (c *Checker) checkAll(isInitial bool) {
	if !c.runMu.TryLock() {
		if !isInitial {
			fmt.Printf("[%s] ⏭ Проверка уже выполняется, пропускаю внеочередной запуск\n", time.Now().Format("15:04:05"))
		}
		return
	}
	defer c.runMu.Unlock()
	defer c.flushSeen()

	activeSearches := c.cfg.GetActiveSearches()
	if len(activeSearches) == 0 {
		return
	}

	concurrency, fetchLimit := c.cfg.PerfSettings()
	if concurrency < 1 {
		concurrency = 1
	}
	skipBumpedOld := c.cfg.GetSkipBumpedOld()
	maxAlertAgeMin := c.cfg.GetMaxAlertAgeMin()

	// 1. Параллельно тянем выдачу по всем фильтрам.
	results := make([]searchFetchResult, len(activeSearches))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for i, search := range activeSearches {
		wg.Add(1)
		go func(i int, search SearchConfig) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			offers, err := c.fetchOffersWithBackfill(search, fetchLimit)
			results[i] = searchFetchResult{offers: offers, err: err}
		}(i, search)
	}
	wg.Wait()

	// 2. Разбираем результаты в порядке фильтров.
	for i, search := range activeSearches {
		res := results[i]
		if res.err != nil {
			fmt.Printf("[%s] Ошибка проверки '%s': %v\n", time.Now().Format("15:04:05"), search.Name, res.err)
			continue
		}

		newCount := 0
		cachedCount := 0
		querySkipped := 0
		var pending []*pendingOffer

		variants := search.QueryVariants
		if len(variants) == 0 && strings.TrimSpace(search.Query) != "" {
			if norm := NormalizeQuery(search.Query); norm != "" {
				variants = []string{norm}
			}
		}

		for _, offer := range res.offers {
			if c.isSeen(offer.ID, search.OwnerChatID) {
				continue
			}
			c.saveSeenID(offer.ID, search.OwnerChatID)
			cachedCount++

			// 1. Игнорируем ТОП / платное продвижение (мусор и накрутка)
			if offer.IsTopAd() {
				c.logAdSkipped(offer, skipReasonTopAd, topAdDetail(offer))
				continue
			}

			// 2. Проверяем «возраст» лота.
			checkTime := offer.FreshnessTime(skipBumpedOld)
			if !checkTime.IsZero() {
				minTime := c.startedAt.Add(-alertStartupGrace(maxAlertAgeMin))
				if search.CreatedAt > 0 {
					filterTime := time.Unix(search.CreatedAt, 0).Add(-defaultStartupGrace)
					if filterTime.After(minTime) {
						minTime = filterTime
					}
				}

				if checkTime.Before(minTime) {
					details := fmt.Sprintf("created: %s | updated: %s",
						offerCreatedAtLabel(offer, false), offerUpdatedAtLabel(offer, false))
					if skipBumpedOld {
						if refreshed, err := offer.ParsedRefreshTime(); err == nil && refreshed.After(minTime) {
							details = "bumped old | " + details
						}
					}
					c.logAdSkipped(offer, skipReasonOutdated, details)
					continue
				}

				if maxAlertAgeMin > 0 {
					if age := time.Since(checkTime); age > time.Duration(maxAlertAgeMin)*time.Minute {
						c.logAdSkipped(offer, skipReasonTooOld, fmt.Sprintf(
							"age: %s | limit: %dm | created: %s",
							formatModerationDelay(age), maxAlertAgeMin, offerCreatedAtLabel(offer, false)))
						continue
					}
				}
			}

			// 3. Нормализуем заголовок и спам-токены.
			offer.Title = sanitizeListingText(offer.Title)
			if isLikelySpamTitle(offer.Title) {
				c.logAdSkipped(offer, skipReasonSpamTitle, "title: "+truncateRunes(offer.Title, 60))
				continue
			}

			// 4. Пост-фильтр по запросу.
			if !OfferMatchesQuery(&offer, variants) {
				querySkipped++
				details := "filter: " + search.Name
				if q := strings.TrimSpace(search.Query); q != "" {
					details = fmt.Sprintf("query: %s | filter: %s", q, search.Name)
				}
				c.logAdSkipped(offer, skipReasonQueryMismatch, details)
				continue
			}

			if !isInitial || c.cfg.NotifyOnStartup {
				pending = append(pending, &pendingOffer{offer: offer, searchName: search.Name})
			}
		}

		// 5. Отправляем уведомления.
		for _, it := range pending {
			offer := it.offer
			newCount++

			c.logAdFound(offer, search.Name)
			c.notifier.NotifyNewOffer(offer, search.Name, search.OwnerChatID)
		}

		if isInitial {
			fmt.Printf("[%s] Инициализация '%s': проверено %d лотов (новых в кэш: %d)\n",
				time.Now().Format("15:04:05"), search.Name, len(res.offers), cachedCount)
		} else if newCount > 0 {
			fmt.Printf("[%s] Поиск '%s': найдено %d новых объявлений!\n",
				time.Now().Format("15:04:05"), search.Name, newCount)
		}
		if querySkipped > 0 {
			fmt.Printf("[%s] 🎯 '%s': отсеяно %d лотов — не совпали с запросом\n",
				time.Now().Format("15:04:05"), search.Name, querySkipped)
		}
	}
}
