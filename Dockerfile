# --- Build ---
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git
WORKDIR /app
COPY main.go .
RUN go build -o proxy main.go

# --- Run ---
FROM alpine:3.20

WORKDIR /app
COPY --from=builder /app/proxy .
EXPOSE 8080
CMD ["./proxy"]