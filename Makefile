.PHONY: all build test clean sec lint bench

all: lint sec test build

build:
	go build -v ./...

test:
	go test -v -race ./...

bench:
	go run benchmark/main.go

sec:
	gosec -tags=gosec -exclude=G115 ./...

lint:
	golangci-lint run ./...

clean:
	rm -f cpu.prof mem.prof vp8_benchmark_results.csv coverage.out
	go clean
