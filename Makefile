.PHONY: test test-short build lint vet fmt tidy run

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -ldflags "-s -w -X github.com/egermano/balde/version.Version=$(VERSION) -X github.com/egermano/balde/version.Commit=$(COMMIT) -X github.com/egermano/balde/version.BuildDate=$(DATE)"

test:
	go test ./... -v -count=1

test-short:
	go test ./... -short -count=1

build:
	go build $(LDFLAGS) -o bin/balde .

vet:
	go vet ./...

fmt:
	gofmt -l -s -w .

lint: vet fmt

tidy:
	go mod tidy

run: build
	./bin/balde
