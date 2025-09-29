# Multi-stage build for Go application
FROM golang:1.23.4-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod file
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o claude-code-proxy ./cmd/claude-code-proxy

# Final stage - minimal image
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1001 appgroup && \
    adduser -D -u 1001 -G appgroup appuser

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/claude-code-proxy .

# Change ownership to non-root user
RUN chown appuser:appgroup claude-code-proxy

# Switch to non-root user
USER appuser

# Expose port (adjust if your app uses a different port)
EXPOSE 8080

# Run the binary
CMD ["./claude-code-proxy"]