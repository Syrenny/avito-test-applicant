# Step 1: Modules caching
FROM golang:alpine AS modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Builder
FROM golang:alpine AS builder
COPY --from=modules /go/pkg /go/pkg
COPY . /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -tags migrate -o /bin/app ./cmd/app

# Step 3: Final
FROM debian:stable-slim

RUN apt-get update -qq && \
    apt-get install -y --no-install-recommends \
    curl=8.17.0 && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/config /config
COPY --from=builder /app/migrations /migrations
COPY --from=builder /bin/app /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
CMD ["/app"]
