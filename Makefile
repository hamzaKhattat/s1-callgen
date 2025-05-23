.PHONY: build run test clean import-test-data load-test

build:
	go build -o bin/callgen cmd/callgen/main.go
	go build -o bin/loadtest cmd/loadtest/main.go
	go build -o bin/generate_test_data testdata/generate_test_data.go

run: build
	./bin/callgen -config configs/config.json

test: build
	./bin/callgen -test -calls 100 -concurrent 10 -duration 60

generate-test-data: build
	./bin/generate_test_data
	@echo "Test data generated in testdata/test_numbers.csv"

import-test-data: generate-test-data
	./bin/callgen -import testdata/test_numbers.csv

load-test: build
	./bin/loadtest -duration 300 -rampup 60 -max 500 -cps 20

stats:
	./bin/callgen -stats

clean:
	rm -rf bin/
	rm -f testdata/test_numbers.csv
