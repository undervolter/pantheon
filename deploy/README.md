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

Скопируйте бинарник и файлы конфигурации:

```bash
sudo install -m 0755 pantheon /opt/pantheon/pantheon
sudo cp config.example.json /opt/pantheon/config.json
sudo cp filters_meta.json /opt/pantheon/filters_meta.json
sudo chown -R pantheon:pantheon /opt/pantheon
```

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
