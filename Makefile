.PHONY: build setup install uninstall reinstall test test-integration coverage

BINARY := flux
BUILD_OUT := bin/$(BINARY)
PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
INSTALL_PATH ?= $(BINDIR)/$(BINARY)

build:
	go build -o $(BUILD_OUT) main.go

setup:
	@./scripts/setup.sh

install: build
	mkdir -p "$(BINDIR)"
	install -m 0755 "$(BUILD_OUT)" "$(INSTALL_PATH)"
	@echo "OK Installed $(INSTALL_PATH)"

uninstall:
	rm -f "$(INSTALL_PATH)"
	@echo "OK Removed $(INSTALL_PATH)"

reinstall: uninstall install

test:
	go test -v ./...

test-integration:
	go test -tags=integration -v ./tests/integration/...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
