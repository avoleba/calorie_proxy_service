FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/proxy ./cmd/proxy

# Этап 2: Финальный образ
FROM alpine:latest

# Добавляем CA сертификаты для HTTPS запросов
RUN apk --no-cache add ca-certificates tzdata

# Создаем непривилегированного пользователя
RUN addgroup -g 1000 -S app && \
    adduser -u 1000 -S app -G app

# Копируем бинарник из этапа сборки
COPY --from=builder --chown=app:app /app/bin/proxy /app/proxy

# Копируем .env файл (опционально, если есть)
# COPY --chown=app:app .env /app/.env

# Переключаемся на непривилегированного пользователя
USER app:app

# Устанавливаем рабочую директорию
WORKDIR /app

# Открываем порт
EXPOSE 8080

# Запускаем приложение
CMD ["./proxy"]