# Use official Golang image as base for building
FROM golang:1.23-alpine AS builder

# Set working directory inside the container
WORKDIR /app

# Install build dependencies (SQLite needs gcc)
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Copy go.mod and go.sum first (for better caching)
COPY go.mod go.sum ./

# Download Go dependencies
RUN go mod download

# Copy all source code
COPY . .

# Build the Go application
RUN go build -o forum main.go

# Start fresh with smaller Alpine image for runtime
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache sqlite wget

# Create app directory
WORKDIR /app

# Copy the built binary from builder stage
COPY --from=builder /app/forum .

# Copy frontend files
COPY --from=builder /app/frontend ./frontend

# Copy internals directory (contains database schema)
COPY --from=builder /app/internals ./internals

# Create directory for database
RUN mkdir -p /app/data

# Create non-root user for security
RUN addgroup -g 1000 -S appgroup && \
    adduser -S appuser -u 1000 -G appgroup

# Change ownership of app directory
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port 8080
EXPOSE 8080

# Set environment variables
ENV PORT=8080
ENV DB_PATH=/app/forum.db

# Health check to ensure container is running correctly
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

# Run the application
CMD ["./forum"]
