.PHONY: build setup install install-global uninstall reinstall test test-integration coverage

BINARY := flux
BUILD_OUT := bin/$(BINARY)
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
ifeq ($(wildcard /opt/homebrew/bin),/opt/homebrew/bin)
PREFIX ?= /opt/homebrew
else
PREFIX ?= /usr/local
endif
else
PREFIX ?= /usr/local
endif
BINDIR ?= $(PREFIX)/bin
INSTALL_PATH ?= $(BINDIR)/$(BINARY)

build:
	go build -o $(BUILD_OUT) main.go

setup:
	@./scripts/setup.sh

install-global:
	@./scripts/install-go.sh

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
