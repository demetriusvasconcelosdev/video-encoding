server:
	go run ./cmd/encoder

test:
	go test -cover ./...

.PHONY: server test