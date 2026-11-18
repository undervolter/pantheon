package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"time"

	"pantheon/core/bot"
	"pantheon/core/checker"
	"pantheon/core/config"
	"pantheon/core/domain"
	"pantheon/core/i18n"
	"pantheon/core/lock"
	"pantheon/core/logx"
	"pantheon/core/olx"
	"pantheon/core/store"
	"pantheon/core/telegram"
	"pantheon/core/version"
)

func main() {
	initTerminal()

	configPath := flag.String("config", "config.json", "Путь к файлу конфигурации")
	once := flag.Bool("once", false, "Выполнить один прогон и выйти (для проверки)")
	tgUI := flag.Bool("tg", true, "Включить Telegram-пульт (управление ботом)")
	testAlert := flag.Bool("test-alert", false, "Отправить тестовый алерт в Telegram и выйти")
	testURL := flag.String("test-url", "", "Отправить алерт по ссылке на объявление OLX и выйти")
	showVersion := flag.Bool("version", false, "Показать версию и выйти")
	flag.Parse()

	if *showVersion {
		fmt.Printf("pantheon open-core %s\n", version.AppVersion)
		return
	}
	os.Exit(run(*configPath, *once, *tgUI, *testAlert, *testURL))
}

// Exit codes. exitSetup means "starting again will not help": the service
// unit lists it in RestartPreventExitStatus so that a missing config or a
// revoked token does not turn into a restart loop.
const (
	exitOK      = 0
	exitRuntime = 1
	exitSetup   = 2
)

// failSetup reports a problem the operator has to fix: the config, the token,
// a second copy already running.
func failSetup(format string, args ...any) int {
	fmt.Fprintln(os.Stderr, logx.Redact(fmt.Sprintf(format, args...)))
	return exitSetup
}

// failRuntime reports a start that may well succeed next time (a busy disk,
// a database error). A stop requested meanwhile is not a failure at all.
func failRuntime(ctx context.Context, format string, args ...any) int {
	if ctx.Err() != nil {
		return exitOK
	}
	fmt.Fprintln(os.Stderr, logx.Redact(fmt.Sprintf(format, args...)))
	return exitRuntime
}

func run(configPath string, once, tgUI, testAlert bool, testURL string) int {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return failSetup("Не найден %s. Скопируйте config.example.json в %s и впишите токен бота.", configPath, configPath)
		}
		return failSetup("Ошибка загрузки конфига: %v", err)
	}
	// Registered before anything can be logged.
	logx.AddSecret(cfg.Token())
	if err := cfg.Validate(); err != nil {
		return failSetup("%v", err)
	}

	tg := telegram.New(cfg.Token())
	client, err := olx.NewOLXClient()
	if err != nil {
		return failSetup("Ошибка создания TLS клиента: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	// After the first signal the default handling is back: the bot is waiting
	// for queued alerts to go out, and a second Ctrl+C ends that wait at once.
	// Whatever was still queued stays in the database for the next start.
	go func() {
		<-ctx.Done()
		stop()
	}()

	if testAlert || testURL != "" {
		return sendTestAlert(ctx, cfg, tg, client, testURL)
	}

	release, err := lock.Acquire(cfg.DataPath("pantheon.lock"))
	if err != nil {
		if errors.Is(err, lock.ErrLocked) {
			return failSetup("Pantheon уже запущен с этим каталогом данных (%s). Второй экземпляр конфликтовал бы с первым за Telegram и базу.", filepath.Dir(cfg.DataPath("pantheon.lock")))
		}
		return failRuntime(ctx, "Не удалось взять блокировку: %v", err)
	}
	defer release()

	db, err := store.Open(cfg.DataPath("pantheon.db"))
	if err != nil {
		if errors.Is(err, store.ErrNewerSchema) {
			return failSetup("База %s создана более новой версией Pantheon: %v", cfg.DataPath("pantheon.db"), err)
		}
		return failRuntime(ctx, "Не удалось открыть базу %s: %v", cfg.DataPath("pantheon.db"), err)
	}
	defer db.Close()

	logx.Print(checker.Banner)
	logx.Print("")
	if err := migrateLegacy(ctx, cfg, db); err != nil {
		return failRuntime(ctx, "Перенос данных 0.1.x не удался: %v", err)
	}

	chk := checker.New(cfg, db, client, tg)
	active, err := chk.ActiveSearches(ctx)
	if err != nil {
		return failRuntime(ctx, "Не удалось прочитать фильтры: %v", err)
	}
	logx.Print(checker.StartMessage(chk.StartedAt(), len(active), cfg.GetInterval()))

	if once {
		logx.Infof("⚡ single run")
		chk.RunOnce(ctx)
		logx.Infof("✅ done")
		return exitOK
	}

	var exit atomic.Int32
	if tgUI {
		panel := bot.New(tg, cfg, db, client, scanner{chk})
		go func() {
			if err := panel.Run(ctx); err != nil {
				// Run returns only when Telegram rejects the bot token.
				logx.Errorf("%v", err)
				exit.Store(exitSetup)
				stop()
			}
		}()
	}

	// Blocks until Ctrl+C or SIGTERM, then waits for queued alerts to go out.
	chk.Run(ctx)
	logx.Infof("pantheon stopped")
	return int(exit.Load())
}

// scanner lets the control panel see and nudge the checker without the two
// packages knowing each other.
type scanner struct{ c *checker.Checker }

func (s scanner) TriggerCheck() { s.c.TriggerCheck() }

func (s scanner) Status() bot.ScanStatus {
	return bot.ScanStatus{Started: s.c.Started(), BlockedUntil: s.c.BlockedUntil()}
}

// migrateLegacy moves the state of a 0.1.x installation into the database:
// the filters kept in config.json and the seen cache of seen_ids.txt. Each
// part runs once; a later start finds its flag and does nothing.
func migrateLegacy(ctx context.Context, cfg *config.Config, db *store.Store) error {
	legacy := cfg.LegacySearches()
	imported, ran, err := db.ImportLegacySearches(ctx, legacy, time.Now())
	if err != nil {
		return err
	}
	switch {
	case ran && imported > 0:
		logx.Infof("📦 %d filters moved from %s to %s", imported, filepath.Base(cfg.FilePath()), filepath.Base(db.Path()))
		if err := cfg.DropLegacySearches(); err != nil {
			// The filters are safe in the database; the file just keeps its copy.
			logx.Warnf("could not rewrite %s without the filters (%v); they are ignored there from now on", cfg.FilePath(), err)
		} else {
			logx.Infof("   the original config is kept as %s.bak", filepath.Base(cfg.FilePath()))
		}
	case !ran && len(legacy) > 0:
		logx.Warnf("%s still lists %d filters; they are ignored — filters live in the database now and are managed from the bot", cfg.FilePath(), len(legacy))
	}

	// 0.1.x wrote seen_ids.txt into the working directory.
	for _, path := range []string{cfg.DataPath("seen_ids.txt"), "seen_ids.txt"} {
		n, ran, err := db.ImportLegacySeen(ctx, path, time.Now())
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if ran {
			logx.Infof("📦 %d seen marks imported from %s (the file is left in place)", n, path)
			break
		}
	}
	return nil
}

// sendTestAlert delivers one card to every trusted chat: a built-in sample,
// or the listing behind an OLX link.
func sendTestAlert(ctx context.Context, cfg *config.Config, tg *telegram.Client, client *olx.OLXClient, rawURL string) int {
	chats := cfg.GetTelegramChatIDs()
	if len(chats) == 0 {
		return failSetup("В конфиге нет chat_id: некому отправлять тестовый алерт.")
	}
	offer := sampleOffer()
	if rawURL != "" {
		found, err := client.FindOfferByURL(rawURL)
		if err != nil {
			return failRuntime(ctx, "❌ %v", err)
		}
		offer = found
		logx.Infof("found listing %d: %s", offer.ID, offer.Title)
	}

	failed := 0
	for _, chatID := range chats {
		lang := i18n.Of(cfg.LangFor(chatID))
		opt := telegram.CardOptions{
			Lang:       lang,
			FilterName: i18n.T(lang, "card.test_filter"),
			Rich:       cfg.GetRichAlerts(),
			ShowParams: cfg.GetShowParams(),
			MaxParams:  cfg.GetMaxParams(),
		}
		if cfg.GetSendPhotos() {
			opt.MaxPhotos = cfg.GetMaxPhotos()
		}
		if _, err := tg.SendOffer(ctx, chatID, offer, opt); err != nil {
			failed++
			logx.Errorf("test alert to %s: %v", chatID, err)
			continue
		}
		logx.Infof("✅ test alert sent to %s", chatID)
	}
	if failed > 0 {
		return exitRuntime
	}
	return exitOK
}

func sampleOffer() domain.Offer {
	now := time.Now()
	price, _ := json.Marshal(map[string]any{"value": 28000, "type": "price", "currency": "UAH", "label": "28 000 грн."})
	return domain.Offer{
		ID:              1160082,
		URL:             "https://www.olx.ua/d/uk/obyavlenie/prodam-iphone-15-pro-max-256-neverlock-ID116hSA.html",
		Title:           "iPhone 15 Pro Max 256GB Neverlock",
		Description:     "Продам свой iPhone 15 Pro Max на 256 ГБ в идеальном состоянии. С первого дня в чехле и с защитным стеклом. Без сколов, без ремонтов, батарея 98%. Полный заводской комплект: коробка, кабель, чек. Продажа при личной встрече в Киеве или отправка OLX Доставкой.",
		CreatedTime:     now.Add(-17 * time.Minute).Format(time.RFC3339),
		LastRefreshTime: now.Format(time.RFC3339),
		Location: domain.Location{
			City:     domain.City{ID: 268, Name: "Київ"},
			Region:   domain.Region{ID: 25, Name: "Київська область"},
			District: domain.District{ID: 11, Name: "Печерський"},
		},
		Params: []domain.Param{{Key: "price", Name: "Ціна", Type: "price", Value: json.RawMessage(price)}},
		Photos: []domain.Photo{
			{Width: 750, Height: 1000, Link: "https://ireland.apollo.olxcdn.com/v1/files/0id9cnxtr6gn2-UA/image;s={width}x{height}"},
			{Width: 750, Height: 1000, Link: "https://ireland.apollo.olxcdn.com/v1/files/srpaym0pxest-UA/image;s={width}x{height}"},
			{Width: 750, Height: 1000, Link: "https://ireland.apollo.olxcdn.com/v1/files/a2rbfbzqxk2a3-UA/image;s={width}x{height}"},
		},
		Contact: domain.OfferContact{Phone: true, Chat: true},
		User:    domain.OfferUser{ID: 1, Name: "Продавец"},
	}
}
