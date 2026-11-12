package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	configPath := flag.String("config", "config.json", "Путь к файлу конфигурации")
	once := flag.Bool("once", false, "Выполнить один прогон и выйти (для проверки)")
	tgUI := flag.Bool("tg", true, "Включить Telegram-пульт (управление ботом)")
	testAlert := flag.Bool("test-alert", false, "Отправить тестовый rich-алерт в Telegram и выйти")
	testURL := flag.String("test-url", "", "Отправить rich-алерт по ссылке на объявление OLX и выйти")
	showVersion := flag.Bool("version", false, "Показать версию и выйти")
	flag.Parse()

	if *showVersion {
		fmt.Printf("pantheon open-core %s\n", appVersion)
		return
	}

	cfg, err := LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка загрузки конфига: %v\n", err)
		os.Exit(1)
	}

	if len(cfg.Searches) == 0 {
		fmt.Println("Предупреждение: список поиска пуст. Добавьте поисковые задачи в config.json!")
	}

	if *testAlert {
		notifier := NewTelegramNotifier(cfg.TelegramBotToken, cfg.TelegramChatIDs)
		notifier.cfg = cfg
		notifier.SendTestAlert()
		return
	}

	if *testURL != "" {
		client, err := NewOLXClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка создания TLS клиента: %v\n", err)
			os.Exit(1)
		}
		notifier := NewTelegramNotifier(cfg.TelegramBotToken, cfg.TelegramChatIDs)
		notifier.cfg = cfg
		if err := notifier.SendTestAlertURL(client, *testURL); err != nil {
			fmt.Fprintf(os.Stderr, "❌ %v\n", err)
			os.Exit(1)
		}
		return
	}

	client, err := NewOLXClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка создания TLS клиента: %v\n", err)
		os.Exit(1)
	}
	notifier := NewTelegramNotifier(cfg.TelegramBotToken, cfg.TelegramChatIDs)
	checker := NewChecker(cfg, client, notifier)
	defer checker.Close()

	botUI := NewBotUI(cfg, checker)
	notifier.SetUI(botUI, cfg)

	if *once {
		fmt.Println("⚡ Тестовый одиночный прогон...")
		checker.CheckNow()
		fmt.Printf("✅ Готово. В кэше: %d\n", checker.GetStats().SeenCount)
		return
	}

	// Telegram-пульт кнопками (опционально)
	if *tgUI {
		go notifier.StartPolling(checker)
	}

	// Красивый выход по Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nОстановка монитора OLX...")
		checker.Close()
		os.Exit(0)
	}()

	// Основной цикл мониторинга.
	checker.Start()
}
