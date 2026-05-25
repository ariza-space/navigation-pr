FROM golang:1.22-alpine AS builder

WORKDIR /src

RUN apk add --no-cache build-base

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/navigation .

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S -G app app \
    && mkdir -p /app/data \
    && chown -R app:app /app/data

COPY --from=builder /out/navigation /app/navigation
COPY index.html /app/index.html

USER app
EXPOSE 8080
VOLUME ["/app/data"]

ENTRYPOINT ["/app/navigation", "-data", "/app/data"]
