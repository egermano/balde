.PHONY: test test-short build lint vet fmt tidy run

test:
	go test ./... -v -count=1

test-short:
	go test ./... -short -count=1

build:
	go build -o bin/balde ./cli

vet:
	go vet ./...

fmt:
	gofmt -l -s -w .

lint: vet fmt

tidy:
	go mod tidy

run: build
	./bin/balde
