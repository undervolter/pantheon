<div align="center">

<h1>pantheon</h1>
<img width="651" height="182" alt="изображение" src="https://github.com/user-attachments/assets/89bcbb96-0329-4bd8-91a7-dd16869881c2" />

<p>
  <a href="./README.md">Русский</a> | <a href="./README_en.md">English</a>
</p>
<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go" alt="Go">
  <img src="https://img.shields.io/github/v/release/undervolter/pantheon?style=flat-square&color=blue" alt="Release">
  <img src="https://img.shields.io/badge/license-AGPL--3.0-blue?style=flat-square" alt="License">
<a href="https://t.me/pantheonUpdates"><img src="https://img.shields.io/badge/Telegram-%40pantheonUpdates-2CA5E0?style=flat-square&logo=telegram&logoColor=white" alt="Telegram Updates"></a> 
  <a href="https://www.repostatus.org/#inactive"><img src="https://www.repostatus.org/badges/latest/active.svg" alt="Project Status: Active"></a>
</p>
Парсер / снайпер объявлений OLX с управлением через тг бота.
</div>

## Что умеет

- Уведомлять вас через бот, если вышло новое обьявление
- Гибкая настройка фильтров (диапазон цен, спец. фильтры, города, и др)
- Работать с несколькими юзерами одновременно
- **Отправлять алерты через Rich (недавнее нововведение телеграма)**
- Кэшировать просмотренные объявления в `seen_ids.txt`
- Фильтровать новоподнятые/платные обьявления

---
## Возможно задаваемые вопросы

<details>
<summary><b>В чём разница между другими ботами?</b></summary>

В основном все подобные боты используют связку Python + BS4 (BeautifulSoup). Минус такой связки очевиден - при малейшем изменении вёрстки, такие боты быстро слепнут, а нередко ловят 403/429.

**pantheon использует мобильный API OLX, поэтому такой проблемы у него нет.**
Также бот умеет работать с несколькими юзерами одновременно и имеет систему владения фильтров.

</details>

<details>
<summary><b>Почему бот отправляет обьявления с задержкой? Кодер даун?</b></summary>
 У OLX есть две даты - дата создания (когда продавец нажал на кнопку "Опубликовать") и дата индексации в поисковую выдачу. Обычно, задержка бывает ~10 минут, в некоторых случаях чуть больше. 
  
  В эти ~10 минут вкладываются:
  
  1) Время модерации
  
  2) Время индексации (в условный Elasticsearch)

  По неизвестной для меня причине, OLX индексирует сразу несколько десятков обьявлений волнами по 5-10 минут (иногда больше). 

  **На это бот влиять не может!** Он только забирает новые обьявления из базы OLX.

  Свежесть обьявления вы можете проверить количеством просмотров на ней. Если счётчик 0 или 1 - вы реально единственный кто посмотрел на обьявление.
</details>

<details>
<summary><b>В чём разница между обычным алертом и Rich алертом?</b></summary>
  Разницу можете наблюдать на фото.

  
  <details>    
    <summary>Rich алерт</summary>
    <img width="499" height="411" alt="изображение" src="https://github.com/user-attachments/assets/532d4423-b226-4d06-9dac-cd6015e8f841" />
  </details>
  
  <details>    
  <summary>Обычный алерт</summary>
  <img width="430" height="389" alt="изображение" src="https://github.com/user-attachments/assets/072f2a6f-c0b4-44d7-87d1-cf723a989d30" />
  </details>

  Иногда бот может присылать обычный алерт вместо Rich. Обычно это происходит когда бот получает ошибку context deadline exceeded (Client.Timeout exceeded while awaiting headers), что (надеюсь) будет исправлено в 0.1.1
  </details>

## Известные проблемы
 - Бот может не отправить фото обьявления в некоторых местах
 - Бот присылает обычный алерт вместо Rich алерта (из-за ошибки context deadline exceeded (Client.Timeout exceeded while awaiting headers))
 
## Быстрый старт

### 1. Настройка конфигурации

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

Распространяется под [AGPL-3.0](LICENSE).
