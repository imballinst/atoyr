GIT_HASH := $(shell git rev-parse HEAD)

.PHONY: build run test clean

build:
	go build -ldflags "-X main.GitHash=$(GIT_HASH)" -o bin/server ./cmd/service

run:
	go run ./cmd/service

test:
	go test -v ./...

clean:
	rm -rf bin
