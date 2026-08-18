FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o trading-service .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/trading-service .

EXPOSE 8082

CMD ["./trading-service"]