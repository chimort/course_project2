FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY ../../go.mod ../../go.sum ./
RUN go mod download

COPY ../../cmd/auth-service ./cmd/auth-service
COPY ../../iternal/auth ./iternal/auth
COPY ../../iternal/pkg ./iternal/pkg
COPY ../../api/proto ./api/proto

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o auth-service ./cmd/auth-service/main.go

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/auth-service .

CMD ["./auth-service"]
