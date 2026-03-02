FROM golang:1.25-alpine

WORKDIR /app
COPY . .

RUN go mod download

WORKDIR /app/cmd/chat-service

RUN go build -o chat-service .

CMD ["./chat-service"]
