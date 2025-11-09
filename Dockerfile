# ---- build stage ----
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o url-shortener ./cmd/api

# ---- runtime stage ----
FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/url-shortener /app/url-shortener

ENV GIN_MODE=release
EXPOSE 8080

CMD ["/app/url-shortener"]
