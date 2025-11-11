FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY ../../go.mod ../../go.sum ./
RUN go mod download

COPY ../../cmd/user-service ./cmd/user-service
COPY ../../iternal/user ./iternal/user
COPY ../../iternal/pkg ./iternal/pkg
COPY ../../api/proto ./api/proto
COPY ../../iternal/middleware ./iternal/middleware
COPY ../../iternal/auth ./iternal/auth


RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o user-service ./cmd/user-service/main.go


FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/user-service .

COPY ../../iternal/user/migrations ./iternal/user/migrations

CMD ["./user-service"]
