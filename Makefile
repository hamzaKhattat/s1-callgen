.PHONY: build run clean

build:
   go build -o bin/callgen cmd/callgen/main.go

run: build
   ./bin/callgen -csv configs/numbers.csv

clean:
   rm -rf bin/

test:
   go test -v ./...
