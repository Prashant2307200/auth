FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN apk add --no-cache git
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app ./cmd/main

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget
COPY --from=builder /app /app
EXPOSE 8080
EXPOSE 9090
HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
  CMD wget -q -O- http://127.0.0.1:8080/health/live || exit 1
ENTRYPOINT ["/app"]
