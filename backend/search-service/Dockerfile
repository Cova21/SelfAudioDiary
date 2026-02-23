# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
COPY vendor ./vendor
COPY internal ./internal
COPY search-service ./search-service

RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o /app/app ./search-service/cmd/main.go

# Runtime stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/app ./app

EXPOSE 50054 8084

CMD ["./app"]
