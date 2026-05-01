.PHONY: test lint clean

test:
	go test -v ./...

lint:
	golangci-lint run ./...

clean:
	go clean
	rm -f volumeleaders-agent
	rm -rf dist/
