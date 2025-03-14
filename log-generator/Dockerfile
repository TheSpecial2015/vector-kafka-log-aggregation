# Stage 1: Build the Go binary
FROM golang:1.23 AS builder
WORKDIR /app
COPY main.go .
RUN go mod init loggen
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o loggen main.go

# Stage 2: Run in Alpine with SSH and Python
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/loggen .

# Install OpenSSH and Python
RUN apk add --no-cache openssh-server python3 && \
    mkdir -p /var/log/hadoop /var/run/sshd && \
    echo "root:password" | chpasswd && \
    ssh-keygen -A

# Configure SSH
RUN sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config && \
    sed -i 's/#PasswordAuthentication yes/PasswordAuthentication yes/' /etc/ssh/sshd_config

# Entry script
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

EXPOSE 22
ENTRYPOINT ["/app/entrypoint.sh"]
