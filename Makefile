.PHONY: run server worker test

server:
	go run ./cmd/server

worker:
	go run ./cmd/worker

run: server

test:
	go test ./...
