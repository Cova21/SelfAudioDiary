.PHONY: test test-go test-go-strict test-go-full-report

test: test-go

test-go:
	bash scripts/test-go.sh

test-go-strict:
	MIN_GO_COVERAGE=80 bash scripts/test-go.sh

test-go-full-report:
	cd backend && go test -covermode=atomic -coverprofile=../backend-coverage-full.out ./...
