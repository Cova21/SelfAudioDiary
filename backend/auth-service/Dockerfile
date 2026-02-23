# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Копируем go.mod и go.sum ПЕРВЫМИ (кэш слоёв Docker)
COPY go.mod go.sum ./

# Копируем vendor - не нужен интернет при сборке!
COPY vendor ./vendor

# Копируем локальные модули (gen/*)
COPY internal ./internal

# Копируем код сервиса
COPY auth-service ./auth-service

# Сборка с -mod=vendor - использует локальный vendor/
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o /app/app ./auth-service/cmd/main.go

# Runtime stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/app ./app

EXPOSE 50051 8081

CMD ["./app"]
