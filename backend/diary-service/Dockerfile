# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod ./
COPY go.sum ./

# Copy internal directory first (needed for go mod to work with replace directives)
COPY internal ./internal

# Download dependencies
RUN go mod download

# Copy service-specific code
COPY diary-service ./diary-service

# Set GOPRIVATE to prevent Go from trying to fetch our internal modules from the internet
ENV GOPRIVATE=github.com/voice-diary/backend

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod -v -o /app/app ./diary-service/cmd/main.go

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/app ./app

# Expose ports
EXPOSE 50052 8082

# Run the application
CMD ["./app"]
