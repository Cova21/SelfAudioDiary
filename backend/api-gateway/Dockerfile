# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Download dependencies (this will create go.sum)
RUN go mod download && go mod verify

# Copy internal directory (pkg and pre-generated protobuf)
COPY internal ./internal

# Copy service-specific code
COPY api-gateway ./api-gateway

# Set GOPRIVATE and build
ENV GOPRIVATE=github.com/voice-diary/backend
RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod -v -o /app/app ./api-gateway/cmd/main.go

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/app ./app

# Expose port
EXPOSE 8080

# Run the application
CMD ["./app"]
