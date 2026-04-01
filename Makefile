.PHONY: build test lint coverage release clean

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

build:
	go build -ldflags "-X main.version=$(VERSION)" -o bin/cantool .

test:
	go test -race -count=1 ./...

coverage:
	go test -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out

lint:
	golangci-lint run ./...

release:
	goreleaser release --clean

clean:
	rm -rf bin/ dist/ coverage.out
