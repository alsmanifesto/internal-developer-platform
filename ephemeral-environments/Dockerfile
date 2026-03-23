# Stage 1: build the binary
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o ephemeral-env .

# Stage 2: runtime with docker-cli and compose plugin
FROM alpine:latest

RUN apk add --no-cache docker-cli curl bash

# Install docker compose v2 plugin
RUN mkdir -p /usr/local/lib/docker/cli-plugins && \
    curl -SL https://github.com/docker/compose/releases/download/v2.27.0/docker-compose-linux-x86_64 \
    -o /usr/local/lib/docker/cli-plugins/docker-compose && \
    chmod +x /usr/local/lib/docker/cli-plugins/docker-compose

WORKDIR /app
COPY --from=builder /app/ephemeral-env .

ENTRYPOINT ["./ephemeral-env"]
