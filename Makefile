# Build metadata injected into package main via -ldflags. On a tagged release
# GoReleaser sets these; this lets local and CI builds carry the same identity
# instead of the dev/none/unknown defaults. Override by passing e.g. VERSION=...
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT  ?= $(shell git rev-parse --short HEAD)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

BINARY := bin/payment-gateway

.PHONY: build vet test e2e clean

## build: compile the stamped binary into ./bin
build:
	@mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

## vet: run go vet across all packages
vet:
	go vet ./...

## test: run the unit test suite with the race detector
test:
	go test -race -count=1 ./...

## e2e: run the end-to-end tests (requires the bank simulator on :8080)
e2e:
	go test -tags=e2e -count=1 ./...

## clean: remove build artifacts
clean:
	rm -rf bin
