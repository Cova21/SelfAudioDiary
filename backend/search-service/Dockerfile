# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod ./
COPY go.sum .//

# Download dependencies (this will create go.sum)
RUN go mod download && go mod verify

# Copy internal directory (pkg and pre-generated protobuf)
COPY internal ./internal

# Copy service-specific code
COPY search-service ./search-service

# Set GOPRIVATE and build
ENV GOPRIVATE=github.com/voice-diary/backend
RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod -v -o /app/app ./search-service/cmd/main.go

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/app ./app

# Expose ports
EXPOSE 50054 8084

# Run the application
CMD ["./app"]
