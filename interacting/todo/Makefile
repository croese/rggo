.DEFAULT_GOAL := build

fmt:
	go fmt ./...
.PHONY:fmt

lint: fmt
	golint ./...
.PHONY:lint

vet: fmt
	go vet ./...
.PHONY:vet

test: vet
	go test -v ./...
.PHONY:test

build: test
	go build -o todo cmd/todo/main.go
.PHONY:build
