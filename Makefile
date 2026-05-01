.PHONY: build test lint clean

build:
	go build -o volumeleaders-agent ./cmd/volumeleaders-agent

test:
	go test -v ./...

lint:
	golangci-lint run ./...

clean:
	go clean
	rm -f volumeleaders-agent
	rm -rf dist/
