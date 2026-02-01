# Build stage
FROM golang:1.25.6-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o bot ./cmd/bot

# Runtime stage
FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates wget

COPY --from=builder /app/bot ./bot
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/migrations ./migrations

RUN adduser -D -u 1000 botuser
USER botuser

EXPOSE 8000

ENTRYPOINT ["./bot"]
CMD []
