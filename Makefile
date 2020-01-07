.PHONY: lint
lint:
	golangci-lint run --enable gofmt ./...

