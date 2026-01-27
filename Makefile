.PHONY: build test lint clean help build-arm64

BINARY_NAME=sadp-rpi

build:
	go build -o $(BINARY_NAME) ./cmd/sadp

build-arm64:
	GOARCH=arm64 GOOS=linux go build -o $(BINARY_NAME)-arm64 ./cmd/sadp

test:
	go test -v ./...

lint:
	golangci-lint run

clean:
	rm -f $(BINARY_NAME)
	go clean

help:
	@echo "Makefile targets:"
	@echo "  build   - Build the binary"
	@echo "  test    - Run all tests"
	@echo "  lint    - Run golangci-lint"
	@echo "  clean   - Remove binary and clean go cache"
