<div align="center">

<h1>pantheon</h1>
<img width="651" height="182" alt="pantheon banner" src="https://github.com/user-attachments/assets/89bcbb96-0329-4bd8-91a7-dd16869881c2" />

<p>
  <a href="./README.md">Русский</a> | <a href="./README_en.md">English</a>
</p>
<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go" alt="Go">
  <img src="https://img.shields.io/github/v/release/undervolter/pantheon?style=flat-square&color=blue" alt="Release">
  <img src="https://img.shields.io/github/license/undervolter/pantheon?style=flat-square" alt="License">
  <a href="https://t.me/pantheonUpdates"><img src="https://img.shields.io/badge/Telegram-%40pantheonUpdates-2CA5E0?style=flat-square&logo=telegram&logoColor=white" alt="Telegram Updates"></a> 
  <a href="https://www.repostatus.org/#inactive"><img src="https://www.repostatus.org/badges/latest/active.svg" alt="Project Status: Active"></a>
</p>
High-performance OLX listing parser & sniper managed via Telegram bot.
</div>

## Features

- **Instant Telegram notifications** whenever a new listing appears
- **Flexible filter configuration** (price ranges, specific parameters, cities, etc.)
- **Multi-user support** with filter ownership and access control
- **Rich Alerts** (leveraging Telegram's expandable blockquotes and photo albums)
- **Persistent cache** of processed listings in `seen_ids.txt`
- **Smart filtering** of bumped, promoted, or reposted ads

---
## Frequently Asked Questions (FAQ)

<details>
<summary><b>How is this different from other bots?</b></summary>

Most similar bots rely on Python + BeautifulSoup (BS4). The downside is obvious — the moment OLX tweaks its website layout, scraping breaks instantly, not to mention frequent 403 Forbidden / 429 Too Many Requests errors from Cloudflare.

**pantheon utilizes the private OLX mobile API, completely avoiding web layout breakages.**
Additionally, the bot natively supports multiple users at the same time and provides a granular filter ownership system.

</details>

<details>
<summary><b>Why does the bot send ads with a delay? Is the developer dumb?</b></summary>

OLX has two distinct timestamps — creation date (when the seller hits "Publish") and the search indexing date (when it actually enters public search results). Usually, this delay is around ~10 minutes, sometimes slightly longer.

This ~10-minute window includes:

1) Content moderation

2) Search indexation (into their internal search engine / Elasticsearch)

For reasons known only to OLX, listings are indexed in waves of dozens at a time every 5–10 minutes (sometimes longer).

**The bot cannot speed up OLX's internal indexing!** It instantly fetches new listings as soon as OLX actually exposes them in search results.

You can verify how fresh an alert is by looking at its view count on OLX. If the counter is 0 or 1, you are genuinely the first person looking at it.
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

Occasionally, the bot may fall back to sending a standard alert instead of Rich. This typically occurs when a network request times out (`context deadline exceeded (Client.Timeout exceeded while awaiting headers)`), which will hopefully be resolved in v0.1.1.
</details>

## Known Issues
- Photos may occasionally fail to attach in certain network edge-cases
- Occasional fallback to standard alert format due to network timeout (`context deadline exceeded (Client.Timeout exceeded while awaiting headers)`)
 
## Quick Start

### 1. Configuration

Copy `config.example.json` to `config.json` and configure:
- `telegram_bot_token`: your bot token from `@BotFather`
- `telegram_chat_ids`: list of chat IDs authorized to receive alerts

### 2. Running

```bash
go run .
```

Or build a standalone binary:

```bash
go build -o pantheon .
./pantheon
```

On first launch, send `/start` to your bot in Telegram. An interactive control panel will open, allowing you to configure filters or paste search URLs directly from OLX.ua.

---

## Command-Line Arguments

- `-config <path>` — path to configuration file (default: `config.json`).
- `-once` — run a single check cycle for all filters and exit.
- `-test-alert` — send a test Rich alert to Telegram to verify appearance.
- `-test-url <url>` — parse a single listing by direct OLX URL and send alert.
- `-version` — display current version.

---

## License

Distributed under the [AGPL-3.0](LICENSE) license.
