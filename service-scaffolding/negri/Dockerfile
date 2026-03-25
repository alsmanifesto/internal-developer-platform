# Multi-stage build for Go service
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o /app/server ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/server .

EXPOSE 8080
CMD ["./server"]
