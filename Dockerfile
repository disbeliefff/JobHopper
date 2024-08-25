# Укажите версию Go
ARG GO_VERSION=1.23
FROM golang:${GO_VERSION}-bookworm as builder

# Установите рабочую директорию
WORKDIR /usr/src/app

# Копирование файлов зависимостей
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Копирование всех исходных файлов приложения
COPY . .

# Сборка приложения
RUN go build -v -o /run-app ./cmd/

# Финальный образ на основе минимального образа Debian
FROM debian:bookworm

# Установите корневые сертификаты для поддержки HTTPS-запросов
RUN apt-get update && apt-get install -y ca-certificates

# Копирование скомпилированного приложения из этапа сборки
COPY --from=builder /run-app /usr/local/bin/

# Установите переменные окружения
ENV TELEGRAM_BOT_TOKEN=${TelegramBotToken}
ENV DATABASE_DSN=${DatabaseDSN}

# Запуск приложения 
CMD ["run-app"]
