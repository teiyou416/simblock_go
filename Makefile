.PHONY: all build test lint clean run

APP_NAME = simblock_go
MAIN_FILE = cmd/simblock/main.go

all: lint build test

build:
		@echo "==> Building $(APP_NAME)..."
			@go build -o bin/$(APP_NAME) $(MAIN_FILE)

run: build
		@echo "==> Running $(APP_NAME)..."
			@./bin/$(APP_NAME)

test:
		@echo "==> Running tests..."
			@go test -v ./...

lint:
		@echo "==> Running golangci-lint..."
			@golangci-lint run ./...

clean:
		@echo "==> Cleaning up..."
			@rm -rf bin/
				@go clean
