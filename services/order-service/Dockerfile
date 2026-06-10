# Dockerfile for Order Service
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o order-service ./cmd/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/order-service .

# Expose the port
EXPOSE 3003

# Command to run the application
CMD ["./order-service"]