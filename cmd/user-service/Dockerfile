FROM golang:1.25-alpine
WORKDIR /app
COPY ../../go.mod ./
RUN go mod download

COPY ../../ .
RUN go build -o user-service ./cmd/user-service/main.go
CMD ["./user-service"]
