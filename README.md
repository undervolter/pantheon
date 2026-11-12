<div align="center">

<h1>pantheon</h1>
<img width="651" height="182" alt="изображение" src="https://github.com/user-attachments/assets/89bcbb96-0329-4bd8-91a7-dd16869881c2" />

<p>
  <a href="./README.md">Русский</a> | <a href="./README_en.md">English</a>
</p>
<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go" alt="Go">
  <img src="https://img.shields.io/github/v/release/undervolter/pantheon?style=flat-square&color=blue" alt="Release">
  <img src="https://img.shields.io/github/license/undervolter/pantheon?style=flat-square" alt="License">
  <img src="https://img.shields.io/github/downloads/undervolter/pantheon/total?style=flat-square&color=blue" alt="Downloads">
<a href="https://t.me/pantheonUpdates"><img src="https://img.shields.io/badge/Telegram-%40pantheonUpdates-2CA5E0?style=flat-square&logo=telegram&logoColor=white" alt="Telegram Updates"></a> 
  <a href="https://www.repostatus.org/#inactive"><img src="https://www.repostatus.org/badges/latest/active.svg" alt="Project Status: Active"></a>
</p>
</div>

Высокопроизводительный парсер и монитор объявлений OLX с управлением через Telegram-бота.

## Возможности Open-Core

- **TLS Fingerprinting (tls-client)**: имитация реальных браузерных профилей (Chrome) для обхода Cloudflare и базовых WAF.
- **Интерактивный Telegram-пульт**:
  - Мастер создания и редактирования фильтров через инлайн-кнопки прямо в чате.
  - Поддержка категорий, городов, диапазонов цен, специфических фильтров (год, пробег, объем) и вариантов запроса.
  - Разделение прав: личные фильтры чата и общие фильтры.
- **Rich-алерты (Telegram Bot API 10.1+)**:
  - Красивое форматирование с раскрывающимися описаниями (`expandable blockquote`).
  - Альбомы фотографий лотов.
  - Детализированные характеристики и прямые ссылки на лоты.
- **Многопоточность и кэш**:
  - Независимый параллельный опрос фильтров с настраиваемыми лимитами.
  - Дедупликация и персистентный кэш просмотренных объявлений (`seen_ids.txt`).
  - Автоматическая фильтрация рекламных / поднятых объявлений (`skip_bumped_old`).

---

## Быстрый старт

### 1. Настройка конфигурации

Скопируйте `config.example.json` в `config.json`:

```bash
cp config.example.json config.json
```

Укажите в `config.json`:
- `telegram_bot_token`: токен вашего бота от `@BotFather`
- `telegram_chat_ids`: список ID чатов для получения алертов

### 2. Запуск

```bash
go run .
```

Или сборка исполняемого файла:

```bash
go build -o pantheon .
./pantheon
```

При первом запуске отправьте боту в Telegram команду `/start`. Откроется кнопочный пульт управления, где вы сможете добавлять новые фильтры или вставлять прямые ссылки поиска с OLX.ua.

---

## Параметры командной строки

- `-config <path>` — путь к конфигурационному файлу (по умолчанию `config.json`).
- `-once` — выполнить один цикл проверки всех фильтров и завершить работу.
- `-test-alert` — отправить тестовый rich-алерт в Telegram для проверки оформления.
- `-test-url <url>` — распарсить лот по прямой ссылке OLX и прислать алерт в чат.
- `-version` — показать текущую версию.

---

## Лицензия

Pantheon Open-Core распространяется под лицензией [AGPL-3.0](LICENSE).
