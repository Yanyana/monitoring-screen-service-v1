# Stage 1: Build
FROM golang:1.20 AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire project
COPY . .

# Build the application
RUN go build -o service .

# Stage 2: Runtime
FROM debian:bullseye-slim

# Set timezone to ensure logs have correct timestamps (optional)
ENV TZ=UTC

# Install minimal dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/service .

# Expose port used by the service
EXPOSE 8081

# Set environment variables (optional)
ENV REDIS_ADDR=redis:6379
ENV REDIS_PASS=

# Command to run the application
CMD ["./service"]