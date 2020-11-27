COMMIT_HASH  := $(shell git rev-parse HEAD)
COMMAND_NAME := bqtableschema
MAIN_DIR     := .

OPEN_CMD := $(shell if command -v explorer.exe; then : "noop"; elif command -v open; then : "noop"; else echo "echo"; fi)

.PHONY: help
help:  ## display this documents
	@grep -E '^[0-9a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

clean:  ## clean up
	-rm -rf bqtableschema/

run:  ## go run
	go run main.go

generate: run ## generate

.PHONY: test
test:  ## go test
	go test -v -race -cover -coverprofile=coverage.out ./...

test2:  ## open coverage.html
	go test -v -race -cover -coverprofile=coverage.out ./... || true
	go tool cover -html=coverage.out -o coverage.html
	${OPEN_CMD} coverage.html
