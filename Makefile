COMMAND_NAME  := bqtableschema
COMMIT_HASH   := $(shell git rev-parse HEAD)
ROOT_DIR      := $(shell git rev-parse --show-toplevel)
MAIN_DIR      := ${ROOT_DIR}
TEST_DIR      := ${ROOT_DIR}/_test
COVERAGE_FILE := ${TEST_DIR}/coverage.out
COVERAGE_HTML := ${TEST_DIR}/coverage.html
TEST_CMD      := go test -v -race -cover -coverprofile=${COVERAGE_FILE} ./...

OPEN_CMD := $(shell if command -v explorer.exe; then : "noop"; elif command -v open; then : "noop"; else echo "echo"; fi)

.PHONY: help
help:  ## display this documents
	@grep -E '^[0-9a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: init
init:  ## init
	@mkdir -p ${TEST_DIR}

.PHONY: run
run:  ## go run
	go run ${MAIN_DIR}

.PHONY: test
test: init ## go test
	${TEST_CMD}

.PHONY: cover
cover: init ## open coverage.html
	${TEST_CMD} || true
	go tool cover -html=${COVERAGE_FILE} -o ${COVERAGE_HTML}
	${OPEN_CMD} ${COVERAGE_HTML}
