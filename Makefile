OUTPUT_DIR := $(shell pwd)/bin/got

.PHONY: build
build: $(shell find . -type f -name '*.go' -print)
	go build -o $(OUTPUT_DIR) github.com/mizuho-u/got

.PHONY:	test
test:
	go test -v `go list ./... | grep -v github.com/mizuho-u/got/test/e2e` -cover

e2etest: build
	go test -v github.com/mizuho-u/got/test/e2e -args -build $(OUTPUT_DIR)