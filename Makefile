BUILD_DATE   = $(shell date +%Y%m%d-%H:%M:%S)
BUILD_USER   = $(shell whoami)
GIT_BRANCH   = $(shell git rev-parse --abbrev-ref HEAD)
GIT_REVISION = $(shell git rev-parse HEAD)
VERSION      = $(shell git describe --tags $(git rev-list --tags --max-count=1))

.PHONY: all
all: vault-kv-mv

vault-kv-mv: *.go
	@go get -v .
	@go build -v ./...

.PHONY: test
test:
	@go test -v ./...

.PHONY: test-coverage
test-coverage:
	@go test -race -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY: clean
clean:
	@rm -f vault-kv-mv coverage.txt
