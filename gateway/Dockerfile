# -------- Stage 1: Builder --------
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Копируем зависимости
COPY go.mod ./
RUN go mod download

# Копируем весь код
COPY . .

# Собираем бинарь
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o gateway .

# -------- Stage 2: Final image --------
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /root/

COPY --from=builder /app/gateway .
COPY --from=builder /app/static ./static

EXPOSE 8000

CMD ["./gateway"]
