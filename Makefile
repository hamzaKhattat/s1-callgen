.PHONY: build run clean test

build:
	go build -o bin/callgen cmd/callgen/main.go

run: build
	./bin/callgen

clean:
	rm -rf bin/
	go clean -cache

test:
	go test -v ./...
