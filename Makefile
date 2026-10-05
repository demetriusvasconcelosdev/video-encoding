.PHONY: build run test
build:
	go build -o bin/encoder ./cmd/encoder
run:
	go run ./cmd/encoder
test:
	go test ./...
