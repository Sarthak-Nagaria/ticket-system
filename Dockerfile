# ---- Build stage ----
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Cache dependencies first
COPY go.mod go.sum* ./
RUN go mod download

COPY . .

# Pure-Go SQLite driver (glebarez/sqlite) means we don't need CGO/gcc.
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/ticket-system ./cmd/server

# ---- Runtime stage ----
FROM alpine:3.20

WORKDIR /app

# Certs needed for any outbound HTTPS calls; harmless if unused.
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/ticket-system .
COPY .env.example .env

EXPOSE 8080

ENTRYPOINT ["./ticket-system"]
