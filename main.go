package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"pantheon/core/bot"
	"pantheon/core/checker"
	"pantheon/core/config"
	"pantheon/core/olx"
	"pantheon/core/version"
)

func main() {
	initTerminal()

	configPath := flag.String("config", "config.json", "Путь к файлу конфигурации")
	once := flag.Bool("once", false, "Выполнить один прогон и выйти (для проверки)")
	tgUI := flag.Bool("tg", true, "Включить Telegram-пульт (управление ботом)")
	testAlert := flag.Bool("test-alert", false, "Отправить тестовый rich-алерт в Telegram и выйти")
	testURL := flag.String("test-url", "", "Отправить rich-алерт по ссылке на объявление OLX и выйти")
	showVersion := flag.Bool("version", false, "Показать версию и выйти")
	flag.Parse()

	if *showVersion {
		fmt.Printf("pantheon open-core %s\n", version.AppVersion)
		return
	}

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка загрузки конфига: %v\n", err)
		os.Exit(1)
	}

	if len(cfg.Searches) == 0 {
		fmt.Println("Предупреждение: список поиска пуст. Добавьте поисковые задачи в config.json!")
	}

	if *testAlert {
		notifier := bot.NewTelegramNotifier(cfg.TelegramBotToken, cfg.TelegramChatIDs)
		notifier.SetUI(nil, cfg)
		notifier.SendTestAlert()
		return
	}

	if *testURL != "" {
		client, err := olx.NewOLXClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка создания TLS клиента: %v\n", err)
			os.Exit(1)
		}
		notifier := bot.NewTelegramNotifier(cfg.TelegramBotToken, cfg.TelegramChatIDs)
		notifier.SetUI(nil, cfg)
		if err := notifier.SendTestAlertURL(client, *testURL); err != nil {
			fmt.Fprintf(os.Stderr, "❌ %v\n", err)
			os.Exit(1)
		}
		return
	}

	client, err := olx.NewOLXClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка создания TLS клиента: %v\n", err)
		os.Exit(1)
	}
	notifier := bot.NewTelegramNotifier(cfg.TelegramBotToken, cfg.TelegramChatIDs)
	chk := checker.NewChecker(cfg, client, notifier)
	defer chk.Close()

	botUI := bot.NewBotUI(cfg, chk)
	notifier.SetUI(botUI, cfg)

	if *once {
		fmt.Println("⚡ Тестовый одиночный прогон...")
		chk.CheckNow()
		fmt.Printf("✅ Готово. В кэше: %d\n", chk.GetStats().SeenCount)
		return
	}

	// Telegram-пульт кнопками (опционально)
	if *tgUI {
		go notifier.StartPolling(chk)
	}

	// Красивый выход по Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nОстановка монитора OLX...")
		chk.Close()
		os.Exit(0)
	}()

	// Основной цикл мониторинга.
	chk.Start()
}
