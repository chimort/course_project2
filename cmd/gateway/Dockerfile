FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY ../../go.mod ../../go.sum ./
RUN go mod download

COPY ../../cmd/gateway ./cmd/gateway
COPY ../../api/proto ./api/proto
COPY ../../iternal/pkg ./iternal/pkg
COPY ../../iternal/gateway ./iternal/gateway
COPY ../../iternal/auth ./iternal/auth

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o gateway ./cmd/gateway/main.go

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/gateway .

CMD ["./gateway"]
