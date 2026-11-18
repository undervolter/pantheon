# Развертывание Pantheon (systemd)

Инструкция по запуску Pantheon в виде фонового сервиса на Linux (Ubuntu / Debian / CentOS).

## 1. Сборка бинарника для Linux

Из корня репозитория:

```bash
GOOS=linux GOARCH=amd64 go build -o pantheon .
```

(Для ARM64 / Raspberry Pi используйте `GOARCH=arm64`.) cgo не нужен: SQLite собран на чистом Go.

## 2. Создание пользователя и директорий

```bash
sudo useradd --system --home /opt/pantheon --shell /usr/sbin/nologin pantheon || true
sudo mkdir -p /opt/pantheon
sudo chown -R pantheon:pantheon /opt/pantheon
```

## 3. Установка файлов

Скопируйте бинарник и пример конфигурации:

```bash
sudo install -m 0755 pantheon /opt/pantheon/pantheon
sudo install -m 0600 -o pantheon -g pantheon config.example.json /opt/pantheon/config.json
```

Справочник фильтров OLX встроен в бинарник, отдельно его копировать не нужно.

Укажите токен бота — в `/opt/pantheon/config.json`:

```bash
sudo nano /opt/pantheon/config.json
```

или отдельным файлом, чтобы токен не лежал в конфиге:

```bash
echo 'PANTHEON_BOT_TOKEN=123456789:AA...' | sudo install -m 0600 -o pantheon -g pantheon /dev/stdin /opt/pantheon/pantheon.env
```

## 4. Настройка systemd-сервиса

```bash
sudo cp deploy/pantheon.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now pantheon
```

## 5. Привязка администратора

Если в `telegram_chat_ids` никого нет, бот печатает в лог одноразовый код. Найдите его:

```bash
sudo journalctl -u pantheon -n 30 --no-pager
```

и отправьте боту в личку строку `/start <код>`. Код действует, пока администратор не привязан; после пяти неверных попыток он меняется — новый снова появится в логе.

## 6. Проверка статуса и логов

```bash
sudo systemctl status pantheon
sudo journalctl -u pantheon -f
```

## Что лежит в /opt/pantheon

| Файл | Что это |
|---|---|
| `config.json` | токен, чаты, настройки; бот дописывает туда изменения из меню |
| `pantheon.db`, `pantheon.db-wal`, `pantheon.db-shm` | фильтры, просмотренные лоты, очередь и журнал алертов |
| `pantheon.lock` | защита от второго экземпляра |
| `config.json.bak` | исходный конфиг 0.1.x, появляется при обновлении |

Для резервной копии остановите сервис и скопируйте `config.json` и `pantheon.db`.

## Обновление с 0.1.x

Замените бинарник и unit-файл, затем `sudo systemctl daemon-reload && sudo systemctl restart pantheon`. Фильтры из `config.json` и кэш из `seen_ids.txt` перенесутся в `pantheon.db` при первом запуске. Откат: остановить сервис, вернуть старый бинарник и `config.json.bak` на место `config.json`.

## Коды завершения

- `0` — штатная остановка;
- `1` — сбой во время работы, systemd перезапустит сервис;
- `2` — ошибка настройки: нет конфига или токена, Telegram отклонил токен, уже запущен второй экземпляр. Сервис не перезапускается, причина — в `journalctl -u pantheon`.
