FROM golang:1.23-alpine AS builder

# Install git and build dependencies
RUN apk add --no-cache git build-base

# Set working directory
WORKDIR /app

# Copy go files
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

# Create necessary directories
WORKDIR /app
RUN mkdir -p /logs /etc/vector/config.d

# Copy the Vector configuration
COPY vector.toml /etc/vector/

# Copy the log generator binary
COPY --from=builder /app/log-generator /app/

# Copy startup script
COPY start.sh /app/
RUN chmod +x /app/start.sh

# Set entrypoint
ENTRYPOINT ["/app/start.sh"]