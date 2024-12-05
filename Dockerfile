# Build stage
FROM golang:1.22-alpine AS build-stage

# Set working directory
WORKDIR /app

# Copy all files from the current directory to the container /app
COPY . .

# Download dependencies
RUN go mod download

# Build the application binary
RUN CGO_ENABLED=0 GOOS=linux go build -o ./bin/backend-monitoring-v1-api  .

# Deploy the application binary
FROM alpine:3.13 AS build-stage-release

# Set current working directory to /app
WORKDIR /app

# Copy the application binary from the build stage to the current stage
COPY --from=build-stage /app/bin/backend-monitoring-v1-api .

EXPOSE 8081

# Command to run the application
CMD ["./backend-monitoring-v1-api"]
