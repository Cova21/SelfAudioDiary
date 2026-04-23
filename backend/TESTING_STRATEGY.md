# Go Testing Strategy (Commercial Grade)

This document describes **what** is tested in Go services and **how** tests should be written and executed for production-level quality.

## 1) Quality Goals

- Prevent regressions in business-critical flows (auth, diary lifecycle, storage/search integration points).
- Ensure deterministic behavior for edge cases and failures.
- Enforce coverage gates in CI/CD.
- Keep tests fast enough for every PR.

## 2) Test Types and Scope

### A. Unit Tests (mandatory on every package with business logic)

**Purpose:** Validate logic in isolation using mocks/fakes.

**Current examples:**
- `backend/internal/pkg/config/config_test.go`
  - env parsing defaults/custom values
  - DSN/Redis/Rabbit URL helpers
- `backend/internal/pkg/middleware/auth_test.go`
  - missing token -> 401
  - Bearer/query token parsing
  - context propagation (`user_id`, `email`)
- `backend/internal/pkg/middleware/cors_test.go`
  - middleware initialization
- `backend/auth-service/service/auth_service_test.go`
  - register/login/validate token happy and error paths
- `backend/diary-service/service/diary_service_test.go`
  - pagination normalization
  - update/delete/analysis behavior

### B. Handler Tests (HTTP/gRPC boundary)

**Purpose:** Verify request validation, status codes, serialization contracts.

Recommended:
- table-driven tests with `httptest` for REST handlers.
- gRPC handler tests with mocked services and contract checks.

Target packages:
- `api-gateway/handler`
- `*/delivery/grpc`
- `notification-service/delivery/websocket`

### C. Repository/Adapter Tests

**Purpose:** Validate DB and external adapter behavior.

Recommended split:
- repository unit tests: SQL generation/mapping with mocks.
- integration tests (dockerized dependencies) for:
  - Postgres repositories
  - Redis sessions
  - Elasticsearch indexing/search
  - MinIO storage operations
  - RabbitMQ publishers/consumers

### D. Integration & Contract Tests

**Purpose:** Validate cross-service workflows:
- upload entry -> transcription queued -> transcription consumed -> NLP queued -> analysis persisted.
- auth token issuance/validation across gateway and services.

Run in CI as separate stage (slower).

## 3) Coverage Policy

### Coverage Gate

`scripts/test-go.sh` runs:
- `go test -covermode=atomic -coverprofile=backend-coverage.out`
- excludes vendor and generated code from scope
- computes total coverage
- fails if below threshold (`MIN_GO_COVERAGE`, default `65`)

### Commands

- Default gate:
  - `make test` or `bash scripts/test-go.sh`
- Strict gate:
  - `make test-go-strict` (80%)
- Custom threshold:
  - `MIN_GO_COVERAGE=75 bash scripts/test-go.sh`

## 4) Test Design Rules

- Prefer table-driven tests for branch-heavy logic.
- Use explicit mocks in unit tests; avoid real network/DB calls.
- One test checks one behavior (clear given/when/then).
- Validate both success and failure paths.
- Keep tests deterministic (no sleep/time race conditions).
- Use meaningful names: `Test<Function>_<Condition>_<Expected>`.

## 5) Recommended Coverage Expansion Plan

### Phase 1 (must-have)
- `auth-service/service` (done, extend with `RefreshToken`, `ChangePassword`, negative claims cases).
- `diary-service/service` (done, extend with MQ queueing and consumer callbacks).
- `api-gateway/handler` core endpoints and auth middleware paths.

### Phase 2 (high value)
- `storage-service/service` (URL generation, download/upload/delete error branches via client abstraction).
- `search-service/service` (query composition, highlight/snippet logic).
- `notification-service/service` and websocket handlers.

### Phase 3 (integration)
- Repositories with real Postgres/Redis.
- Queue flows with RabbitMQ.
- End-to-end pipeline tests in docker-compose test environment.

## 6) CI/CD Recommendation

On each PR:
1. `go test` with race detector for unit layer:
   - `go test -race ./...`
2. coverage gate:
   - `make test`
3. strict nightly/merge gate:
   - `make test-go-strict`

Artifacts:
- `backend-coverage.out` stored in CI for reports/trends.

