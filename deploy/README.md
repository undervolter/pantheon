# Развертывание Pantheon (systemd)

Инструкция по запуску Pantheon в виде фонового сервиса на Linux (Ubuntu / Debian / CentOS).

## 1. Сборка бинарника для Linux

Из корня репозитория:

```bash
GOOS=linux GOARCH=amd64 go build -o pantheon .
```

(Для ARM64 / Raspberry Pi используйте `GOARCH=arm64`).

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
sudo install -m 0600 config.example.json /opt/pantheon/config.json
sudo chown -R pantheon:pantheon /opt/pantheon
```

Справочник фильтров OLX встроен в бинарник, отдельно его копировать не нужно.

Отредактируйте `/opt/pantheon/config.json`, указав токен вашего Telegram-бота:
```bash
sudo nano /opt/pantheon/config.json
```

## 4. Настройка systemd-сервиса

```bash
sudo cp deploy/pantheon.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now pantheon
```

## 5. Проверка статуса и логов

```bash
sudo systemctl status pantheon
sudo journalctl -u pantheon -f
```
