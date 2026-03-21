.PHONY: build setup test test-integration coverage

build:
	go build -o bin/flux main.go

setup:
	@./scripts/setup.sh

test:
	go test -v ./...

test-integration:
	go test -tags=integration -v ./tests/integration/...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
