FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git build-base

WORKDIR /app

COPY *.go ./
COPY go.mod ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o log-generator

# Final image
FROM alpine:3.17

# Install Vector and dependencies
RUN apk add --no-cache ca-certificates tzdata curl bash jq

# Install Vector using the official script
RUN curl --proto '=https' --tlsv1.2 -sSfL https://sh.vector.dev | bash -s -- -y --prefix /usr/local

# Create Vector Directories
RUN mkdir -p /var/lib/vector && chmod 755 /var/lib/vector

WORKDIR /app
RUN mkdir -p /logs /etc/vector/config.d

COPY vector.toml /etc/vector/

COPY --from=builder /app/log-generator /app/

COPY start.sh /app/
RUN chmod +x /app/start.sh

ENTRYPOINT ["/app/start.sh"]