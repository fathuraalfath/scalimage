# Stage 1: Build compilation
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Compile optimized static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server cmd/server/main.go

# Stage 2: Runtime image
FROM alpine:latest

# Install certs for future integrations (e.g. AWS S3 via HTTPS)
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server .

# Expose server port
EXPOSE 8080

# Execute server
CMD ["./server"]
