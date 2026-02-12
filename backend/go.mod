module github.com/voice-diary/backend

go 1.22

require (
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/mux v1.8.1
	github.com/gorilla/websocket v1.5.1
	github.com/lib/pq v1.10.9
	github.com/minio/minio-go/v7 v7.0.66
	github.com/olivere/elastic/v7 v7.0.32
	github.com/rabbitmq/amqp091-go v1.9.0
	github.com/redis/go-redis/v9 v9.4.0
	github.com/rs/cors v1.10.1
	github.com/voice-diary/backend/internal/gen/auth v0.0.0-00010101000000-000000000000
	github.com/voice-diary/backend/internal/gen/diary v0.0.0-00010101000000-000000000000
	github.com/voice-diary/backend/internal/gen/notification v0.0.0-00010101000000-000000000000
	github.com/voice-diary/backend/internal/gen/search v0.0.0-00010101000000-000000000000
	github.com/voice-diary/backend/internal/gen/storage v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.26.0
	golang.org/x/crypto v0.27.0
	google.golang.org/grpc v1.68.1
	google.golang.org/protobuf v1.36.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.4 // indirect
	github.com/klauspost/cpuid/v2 v2.2.6 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/minio/sha256-simd v1.0.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rs/xid v1.5.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/stretchr/testify v1.8.3 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
)

// Replace directives to use local code instead of trying to fetch from GitHub
replace github.com/voice-diary/backend => ./

// Replace for generated protobuf packages (each in its own subdirectory)
replace github.com/voice-diary/backend/internal/gen/auth => ./internal/gen/auth

replace github.com/voice-diary/backend/internal/gen/diary => ./internal/gen/diary

replace github.com/voice-diary/backend/internal/gen/storage => ./internal/gen/storage

replace github.com/voice-diary/backend/internal/gen/search => ./internal/gen/search

replace github.com/voice-diary/backend/internal/gen/notification => ./internal/gen/notification
