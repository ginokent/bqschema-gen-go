COMMAND_NAME  := bqtableschema
COMMIT_HASH   := $(shell git rev-parse HEAD)
ROOT_DIR      := $(shell git rev-parse --show-toplevel)
MAIN_DIR      := ${ROOT_DIR}
COVERAGE_FILE := ${ROOT_DIR}/coverage.txt
COVERAGE_HTML := ${ROOT_DIR}/coverage.html
TEST_CMD      := GOTEST=true OUTPUT_FILE=/dev/null go test -v -race -cover -coverprofile=${COVERAGE_FILE} ./...

OPEN_CMD := $(shell if command -v explorer.exe 1>/dev/null; then echo "explorer.exe"; elif uname -s | grep -q Darwin; then echo "open"; else echo "echo"; fi)

.PHONY: help
help:  ## display this documents
	@grep -E '^[0-9a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: lint
lint:  ## go fmt and go vet
	# tidy
	go mod tidy
	git diff --exit-code ${ROOT_DIR}/go.mod
	git diff --exit-code ${ROOT_DIR}/go.sum
	# fmt
	go fmt ./...
	# vet
	go vet ./...

.PHONY: run
run:  ## go run
	# run
	go run ${MAIN_DIR}

.PHONY: test
test:  ## go test
	# test
	${TEST_CMD}

.PHONY: cover
cover:  ## open coverage.html
	# test
	${TEST_CMD} || true
	# cover
	go tool cover -html=${COVERAGE_FILE} -o ${COVERAGE_HTML}
	${OPEN_CMD} ${COVERAGE_HTML}

.PHONY: ci
ci: lint cover ## for CI

