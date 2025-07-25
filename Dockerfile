FROM golang:1.22-alpine3.20 AS builder

ENV GOTOOLCHAIN=auto

RUN apk add --no-cache upx

WORKDIR /app

COPY go.* ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/main

RUN upx --best --lzma ./server

FROM alpine:latest AS runner

WORKDIR /app

COPY --from=builder /app/server ./

USER 1000:1000

EXPOSE 8080

CMD ["./server"]