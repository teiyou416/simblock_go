.PHONY: all build run test test-unit test-suite test-align lint clean

APP_NAME := simblock_go
MAIN_FILE := ./cmd/simblock/main.go

all: lint build test

build:
	@echo "==> Building $(APP_NAME)..."
	@go build -o bin/$(APP_NAME) $(MAIN_FILE)

run: build
	@echo "==> Running $(APP_NAME)..."
	@./bin/$(APP_NAME)

# Default test target: unit + integrated suite
test: test-unit test-suite

test-unit:
	@echo "==> Running unit tests..."
	@go test -v $$(go list ./... | rg -v '/tests$$')

test-suite:
	@echo "==> Running integrated suite..."
	@go test -v ./tests

test-align:
	@echo "==> Running Java/Go alignment check..."
	@./scripts/alignment.sh

lint:
	@echo "==> Running golangci-lint..."
	@golangci-lint run ./...

clean:
	@echo "==> Cleaning up..."
	@rm -rf bin/
	@go clean
