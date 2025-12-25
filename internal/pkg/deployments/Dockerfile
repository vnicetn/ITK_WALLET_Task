FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# Этап 2: Финальный образ (минимальный)
FROM alpine:latest

# Устанавливаем ca-certificates для HTTPS запросов (если нужно)
RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/main .

COPY config.env .

EXPOSE 8080

CMD ["./main"]
