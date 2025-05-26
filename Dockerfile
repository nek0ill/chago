FROM golang:1.24.3 AS builder

WORKDIR /app
COPY . .
RUN go build -o encrypted-chat .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/encrypted-chat .
COPY --from=builder /app/internal/crypto /app/internal/crypto
COPY --from=builder /app/internal/chat /app/internal/chat

EXPOSE 8080
CMD ["./encrypted-chat"]
