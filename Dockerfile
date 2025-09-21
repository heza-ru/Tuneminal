# Multi-stage build for Tuneminal
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -extldflags '-static'" -o tuneminal cmd/tuneminal/main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates alsa-lib alsa-utils

# Create non-root user
RUN adduser -D -s /bin/sh tuneminal

# Set working directory
WORKDIR /home/tuneminal

# Copy binary from builder stage
COPY --from=builder /app/tuneminal .

# Copy demo files
COPY --from=builder /app/uploads/demo ./uploads/demo

# Change ownership
RUN chown -R tuneminal:tuneminal /home/tuneminal

# Switch to non-root user
USER tuneminal

# Expose port (if needed for future web features)
EXPOSE 8080

# Set environment variables
ENV PATH="/home/tuneminal:${PATH}"

# Default command
ENTRYPOINT ["./tuneminal"]





