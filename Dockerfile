# Build stage
FROM golang:1.24.1-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /conduit ./cmd/...

# Final stage
FROM alpine:latest AS runner

# Set working directory
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /conduit .

# Expose application port
EXPOSE 8080

# Run the application
CMD ["./conduit"]
