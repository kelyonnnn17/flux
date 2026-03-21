.PHONY: build setup

build:
	go build -o bin/flux main.go

setup:
	@./scripts/setup.sh
