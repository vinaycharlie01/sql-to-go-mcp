.PHONY: build run test test-v test-cover lint clean

BINARY := sql-repository-mcp
BUILD_DIR := bin

build:
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/server

run: build
	./$(BUILD_DIR)/$(BINARY)

test:
	go test ./...

test-v:
	go test -v ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-bench:
	go test -bench=. -benchmem -count=3 ./pkg/...

lint:
	golangci-lint run ./...

clean:
	rm -rf $(BUILD_DIR) coverage.out coverage.html

tidy:
	go mod tidy

vet:
	go vet ./...
