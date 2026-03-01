BINARY_NAME=devforge
VERSION?=dev
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build build-all checksums clean test

build:
	go build $(LDFLAGS) -o $(BINARY_NAME)

build-all:
	mkdir -p dist
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe

checksums:
	cd dist && shasum -a 256 * > checksums.txt

clean:
	rm -rf dist $(BINARY_NAME)

test:
	go test ./...
