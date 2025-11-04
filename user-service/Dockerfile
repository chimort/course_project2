# -------- Stage 1: Builder --------
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Копируем только зависимости для кеширования
COPY go.mod ./
RUN go mod download

# Копируем весь код
COPY . .

# Собираем бинарь
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o user-service .

# -------- Stage 2: Final image --------
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /root/

COPY --from=builder /app/user-service .
COPY --from=builder /app/static ./static

EXPOSE 8080

CMD ["./user-service"]
