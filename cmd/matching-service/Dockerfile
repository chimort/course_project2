FROM golang:1.25-alpine

WORKDIR /app
COPY . .

RUN go mod download

WORKDIR /app/cmd/matching-service

RUN go build -o matching-service .

CMD ["./matching-service"]
