# AgentSpillCheck Makefile
.PHONY: build test vet lint e2e

build:
	go build ./...

test:
	go test ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

e2e:
	go test ./test/...
