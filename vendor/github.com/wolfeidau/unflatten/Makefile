SOURCE_FILES?=$$(go list ./... | grep -v /vendor/)
TEST_PATTERN?=.
TEST_OPTIONS?=

GO ?= go

ci: setup lint test
.PHONY: ci

# Install all the build and lint dependencies
setup:
	@$(GO) get -u github.com/golang/dep/cmd/dep
	@$(GO) get -u github.com/pierrre/gotestcover
	@$(GO) get -u golang.org/x/tools/cmd/cover
	@$(GO) get -u honnef.co/go/tools/cmd/megacheck
	@dep ensure
.PHONY: setup

# Run all the tests
test:
	@gotestcover $(TEST_OPTIONS) -covermode=atomic -coverprofile=coverage.txt $(SOURCE_FILES) -run $(TEST_PATTERN) -timeout=2m
.PHONY: test

# Run all the tests and opens the coverage report
cover: test
	@$(GO) tool cover -html=coverage.txt
.PHONY: cover

# Run all the linters
lint:
	@megacheck
.PHONY: lint

# Run all the tests and code checks
ci: setup test lint
.PHONY: ci
