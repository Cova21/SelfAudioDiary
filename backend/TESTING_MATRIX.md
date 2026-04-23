# Backend Go Test Matrix

This matrix documents **what is tested now** and **what each missing-area test file covers**.

## A. Implemented unit tests (active, non-skipped)

- `internal/pkg/config/config_test.go`
  - default env parsing
  - custom env parsing
  - DSN/Redis/Rabbit URL helpers
- `internal/pkg/middleware/auth_test.go`
  - missing token -> 401
  - Bearer token and query token parsing
  - context claims extraction
- `internal/pkg/middleware/cors_test.go`
  - middleware initialization
- `auth-service/service/auth_service_test.go`
  - register/login/validate token success and key failures
- `diary-service/service/diary_service_test.go`
  - pagination normalization
  - update/delete/analysis service behavior
- `search-service/service/search_service_test.go`
  - helper conversion `stringsToInterfaces`

## B. Added missing package test files (test-plan stubs)

Each file now exists and contains an explicit, package-specific plan of scenarios:

- `api-gateway/cmd/main_test.go`
- `api-gateway/handler/handler_test.go`
- `auth-service/cmd/main_test.go`
- `auth-service/delivery/grpc/handler_test.go`
- `auth-service/domain/domain_test.go`
- `auth-service/migrations/migrations_test.go`
- `auth-service/repository/postgres/user_repository_test.go`
- `auth-service/repository/redis/session_repository_test.go`
- `diary-service/cmd/main_test.go`
- `diary-service/delivery/grpc/handler_test.go`
- `diary-service/domain/domain_test.go`
- `diary-service/migrations/migrations_test.go`
- `diary-service/repository/postgres/entry_repository_test.go`
- `internal/pkg/database/database_test.go`
- `internal/pkg/grpcutil/grpcutil_test.go`
- `internal/pkg/logger/logger_test.go`
- `internal/pkg/rabbitmq/rabbitmq_test.go`
- `notification-service/cmd/main_test.go`
- `notification-service/delivery/grpc/handler_test.go`
- `notification-service/delivery/websocket/handler_test.go`
- `notification-service/service/notification_service_test.go`
- `search-service/cmd/main_test.go`
- `search-service/delivery/grpc/handler_test.go`
- `storage-service/cmd/main_test.go`
- `storage-service/delivery/grpc/handler_test.go`
- `storage-service/service/storage_service_test.go`

## C. Execution and quality gates

- `make test` / `bash scripts/test-go.sh`
  - run all package tests
  - enforce coverage gate on critical active packages only:
    - `auth-service/service`
    - `diary-service/service`
    - `internal/pkg/config`
    - `internal/pkg/middleware`
    - `search-service/service`
  - default gate: 30% (`MIN_GO_COVERAGE` env can override)
  - plan-only stub tests are excluded from gate by design
- `make test-go-strict`
  - stricter coverage threshold
- `make test-go-full-report`
  - full repository coverage report

## D. Next conversion steps (to move stubs -> active tests)

1. Replace handler stubs with table-driven `httptest` / gRPC handler tests.
2. Add repository integration tests against Postgres/Redis/Elasticsearch/MinIO.
3. Add rabbitmq consumer contract tests for transcription/NLP notification flows.
4. Raise default coverage gate after each conversion wave.

