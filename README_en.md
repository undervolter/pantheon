<div align="center">

<h1>pantheon</h1>
<img width="651" height="182" alt="pantheon banner" src="https://github.com/user-attachments/assets/89bcbb96-0329-4bd8-91a7-dd16869881c2" />

<p>
  <a href="./README.md">Русский</a> | <a href="./README_en.md">English</a>
</p>
<p align="center">
  <img src="https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat-square&logo=go" alt="Go">
  <img src="https://img.shields.io/github/v/release/undervolter/pantheon?style=flat-square&color=blue" alt="Release">
  <img src="https://img.shields.io/badge/license-AGPL--3.0-blue?style=flat-square" alt="License">
  <a href="https://t.me/pantheonUpdates"><img src="https://img.shields.io/badge/Telegram-%40pantheonUpdates-2CA5E0?style=flat-square&logo=telegram&logoColor=white" alt="Telegram Updates"></a> 
  <a href="https://www.repostatus.org/#inactive"><img src="https://www.repostatus.org/badges/latest/active.svg" alt="Project Status: Active"></a>
</p>
OLX listing parser & sniper managed via Telegram bot.
</div>

## Features

- **Telegram notifications** whenever a new listing appears
- **Flexible filters** (price ranges, category-specific parameters, cities, exclusion words, etc.)
- **Free-only mode**: catch listings marked «Безкоштовно» in any category
- **Multi-user**: every chat has its own filters, language, pause and quiet hours
- **Rich alerts** (collage, collapsible details table, expandable description)
- Menus in Russian, Ukrainian and English
- Filters, seen listings and the alert journal are kept in SQLite (`pantheon.db`)
- Alerts are not lost: what could not be delivered is sent again after a network failure or a restart
- **Smart filtering** of bumped and promoted ads
- Export and import of filters as a file

## What is new in 0.2.0

- **Free price** — its own mode on the price step, available in every category; typing `0` works too.
- **New control panel** — cards with tables, coloured buttons, a five-step wizard, recent alerts, "today / 7 days" counters, chat pause, quiet hours, filter snooze, exclusion words, export/import, three languages.
- **SQLite instead of `.txt`** — filters and the seen cache live in `pantheon.db`. Data of 0.1.x is moved over on the first start.
- **Access** — the first admin is bound with a one-time code from the console and adds everyone else.
- Many bugs that lost or duplicated alerts are fixed. See the commit history for details.

---
## Frequently Asked Questions (FAQ)

<details>
<summary><b>How is this different from other bots?</b></summary>

Most similar bots rely on Python + BeautifulSoup (BS4). The downside is obvious — the moment OLX tweaks its website layout, scraping breaks, not to mention frequent 403 / 429 errors.

**pantheon uses the OLX mobile API, so it does not have that problem.**
The bot also supports several users at once and has a filter ownership system.

</details>

<details>
<summary><b>Why does the bot send ads with a delay? Is the developer dumb?</b></summary>

OLX has two timestamps — the creation date (when the seller hits "Publish") and the moment the listing enters the search results. The gap is usually ~10 minutes, sometimes 15–20.

That window includes:

1) Content moderation

2) Search indexing

For reasons known only to OLX, listings are indexed in waves of dozens at a time every 5–10 minutes (sometimes longer).

**The bot cannot speed up OLX's indexing!** It fetches new listings as soon as OLX exposes them.

You can verify how fresh an alert is by its view count on OLX. If the counter is 0 or 1, you are the first person looking at it.

Because of the same gap, the "Max listing age" setting should not go below 20 minutes: otherwise the bot treats nearly everything OLX shows it as old. The default is 30.
</details>

<details>
<summary><b>What is the difference between standard and Rich alerts?</b></summary>

You can see the visual difference below:

  <details>    
    <summary>Rich Alert</summary>
    <img width="499" height="411" alt="Rich Alert preview" src="https://github.com/user-attachments/assets/532d4423-b226-4d06-9dac-cd6015e8f841" />
  </details>
  
  <details>    
    <summary>Standard Alert</summary>
    <img width="430" height="389" alt="Standard Alert preview" src="https://github.com/user-attachments/assets/072f2a6f-c0b4-44d7-87d1-cf723a989d30" />
  </details>

The style is switched in Settings → Card style. If the Bot API server does not know rich messages, or rejected a particular card, the bot sends the standard one by itself.
</details>

<details>
<summary><b>Why is there no real estate?</b></summary>

The «Нерухомість» section is not served by the open-core edition: it is not offered among the categories, links and files with it are refused, and real-estate listings are skipped even by filters without a category.
</details>

## Known Limitations
- In groups Telegram only passes commands to a bot by default. To type wizard input (query, price, city) right in a group, turn off privacy mode for the bot in @BotFather, or create filters in a private chat.
- Rich cards and coloured buttons depend on the Bot API version. On an older server the bot falls back to ordinary messages.
- An unfinished wizard does not survive a restart of the bot.
 
## Quick Start

### 1. Configuration

Copy `config.example.json` to `config.json` and set:
- `telegram_bot_token`: your bot token from `@BotFather` (or set the `PANTHEON_BOT_TOKEN` environment variable — then the token is not stored in the file)

`telegram_chat_ids` may stay empty.

### 2. Running

```bash
go run .
```

Or build a standalone binary:

```bash
go build -o pantheon .
./pantheon
```

### 3. Binding the admin

If the config lists no chat id, the bot prints a one-time code to the console on start:

```
  This bot has no admin yet. To become one, send it this command in a private chat:
      /start a1b2c3d4
  or open  https://t.me/your_bot?start=a1b2c3d4
```

Send that command to the bot in a private chat — you become the admin and the control panel opens. Merely writing to the bot or pressing one of its buttons is not enough: access is never granted that way.

If `telegram_chat_ids` is filled in beforehand, the first user on the list becomes the admin (the bot prints who that is). Use the `admin_chat_ids` key to name someone else.

Everyone else is added by the admin: `/addchat <id>` for a user or a channel, `/bind` inside a group. Anyone can learn their `chat_id` by sending `/start` to the bot.

---

## Commands

| Command | What it does |
|---|---|
| `/start` | Main menu |
| `/filters` | My filters |
| `/add` | New filter |
| `/history` | Recent alerts |
| `/settings` | Settings |
| `/pause`, `/resume` | Pause and resume alerts in this chat |
| `/lang` | Language |
| `/export` | Filters as a file; send a file to the bot to import |
| `/cancel` | Cancel the current action |
| `/chatid`, `/version` | chat_id of this chat, bot version |

Admin only: `/chats`, `/bind`, `/unbind`, `/addchat <id>`, `/delchat <id>`, `/check` (check the results now).

A link to an OLX search page sent to the bot becomes a filter.

---

## Command-Line Arguments

- `-config <path>` — path to the configuration file (default: `config.json`).
- `-once` — run a single check cycle for all filters, wait for the alerts to go out, and exit.
- `-test-alert` — send a test alert to Telegram to check its appearance.
- `-test-url <url>` — find a listing by its OLX URL and send an alert for it.
- `-tg=false` — run without the Telegram control panel (monitoring and alerts only).
- `-version` — display the current version.

---

## Configuration File

Filters are no longer kept in `config.json` — they live in the database and are managed from the bot. The file holds the token, the chats and the settings; the bot writes back what is changed from its menus.

| Key | Default | Meaning |
|---|---|---|
| `telegram_bot_token` | — | Bot token |
| `telegram_chat_ids` | `[]` | Trusted chats: they use the control panel and receive the alerts of shared filters |
| `admin_chat_ids` | first user | Admins |
| `interval_seconds` | `15` | Poll interval, at least 3 |
| `fetch_limit` | `20` | Listings per request, at most 50 |
| `concurrency` | `4` | How many filters are requested at once |
| `backfill` | `true` | Load further pages when the first one overflowed |
| `notify_on_startup` | `false` | Send alerts for the first pass after a start |
| `max_alert_age_min` | `30` | Max listing age in minutes, `0` — no limit |
| `skip_bumped_old` | `true` | Do not treat old bumped listings as new |
| `rich_alerts` | `true` | Rich card instead of plain text |
| `send_photos`, `max_photos` | `true`, `3` | Photos in alerts, up to 10 |
| `show_params`, `max_params` | `true`, `0` | Details in the card, `0` — all |
| `language` | `ru` | Default language: `ru`, `uk`, `en` |
| `timezone` | `Europe/Kyiv` | Time zone for quiet hours |
| `data_dir` | next to the config | Where `pantheon.db` lives |
| `seen_retention_days` | `7` | How many days a seen listing is remembered |
| `max_filters_per_chat` | `50` | Filter limit per chat, `0` — no limit |
| `ansi_hyperlinks` | `true` | Clickable links in the console log |

`languages`, `paused_chats`, `quiet_hours` and `announced_version` are maintained by the bot.

---

## Upgrading from 0.1.x

Nothing to do: on the first start of 0.2.0

1. the filters from `config.json` move to `pantheon.db`, and the original file is kept as `config.json.bak`;
2. `seen_ids.txt` is imported into the database (the file itself is left in place);
3. real-estate filters are carried over switched off;
4. if no admin is set, the first user in `telegram_chat_ids` becomes one — the bot prints this to the console on start. In 0.1.x that list was filled by whoever wrote to the bot first, so check that it is you; if not, stop the bot and put your own id into `admin_chat_ids`.

Rolling back to 0.1.x: put `config.json.bak` back in place of `config.json`.

---

## License

Distributed under the [AGPL-3.0](LICENSE) license.
