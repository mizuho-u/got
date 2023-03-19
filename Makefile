OUTPUT_DIR := $(shell pwd)/bin/got

.PHONY: build
build: $(shell find . -type f -name '*.go' -print)
	go build -o $(OUTPUT_DIR) github.com/mizuho-u/got

.PHONY:	test
test:
	go test `go list ./... | grep -v github.com/mizuho-u/got/test/e2e` -cover

.PHONY:	e2etest
e2etest: build
	go test github.com/mizuho-u/got/test/e2e -args -build $(OUTPUT_DIR)

.PHONY: fulltest
fulltest: test e2etest