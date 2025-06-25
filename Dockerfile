# Используем официальный образ Go для сборки
FROM golang:1.20-alpine AS builder

WORKDIR /app

# Копируем модули и заголовочные файлы
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Собираем бинарник
RUN go build -o order-service ./cmd/main.go

# --- Финальный образ ---
FROM alpine:latest

WORKDIR /app

# Копируем собранный бинарник из builder
COPY --from=builder /app/order-service .

# Копируем .env (если используешь)
COPY .env .

# Порт, который слушает сервис
EXPOSE 8080

# Команда запуска
CMD ["./order-service"]