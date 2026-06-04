BIN     := apple-compose
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build test test-integration lint vet fmt coverage coverage-html install clean release-dry

build:
	go build -ldflags "$(LDFLAGS)" -o $(BIN) .

test:
	go test -race ./...

test-integration:
	go test -tags integration -v -timeout 20m ./integration/

vet:
	go vet ./...

fmt:
	gofmt -w . cmd/ internal/ integration/

lint:
	golangci-lint run --timeout=5m

coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

coverage-html: coverage
	go tool cover -html=coverage.out -o coverage.html
	open coverage.html

install: build
	mv $(BIN) /usr/local/bin/$(BIN)

clean:
	rm -f $(BIN) coverage.out coverage.html

release-dry:
	goreleaser release --snapshot --clean
