.PHONY: build run test

build:
	go build -o bin/oapic ./cmd/oapic

run: build
	go run ./cmd/oapic list --source demo/task-board-demo/demo1.go

test:
	go test ./...