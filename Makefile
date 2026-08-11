GOLANGCI_LINT ?= go tool -modfile=tools/go.mod golangci-lint

FUZZTIME ?= 30s
COVERPROFILE ?= coverage.out

.DEFAULT_GOAL := check

.PHONY: help
help: ## Show this help
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'

.PHONY: check
check: fmt lint test ## Format, lint and test

.PHONY: test
test: ## Run the tests with the race detector and coverage
	@go test ./... -race -cover

.PHONY: cover
cover: ## Write and open an HTML coverage report
	@go test ./... -coverprofile=$(COVERPROFILE) -covermode=atomic
	@go tool cover -html=$(COVERPROFILE)

.PHONY: bench
bench: ## Run the benchmarks
	@go test ./... -run '^$$' -bench . -benchmem

.PHONY: fuzz
fuzz: ## Run each fuzz target for FUZZTIME (default 30s)
	@for target in FuzzParse FuzzParseNeverPanics FuzzRoundTrip FuzzGregorianRoundTrip; do \
		echo "==> $$target"; \
		go test -run '^$$' -fuzz "^$$target$$" -fuzztime $(FUZZTIME) . || exit 1; \
	done

.PHONY: fmt
fmt: tools ## Format the source
	@$(GOLANGCI_LINT) fmt -c .golangci.yaml

.PHONY: lint
lint: tools ## Run the linter
	@$(GOLANGCI_LINT) run -c .golangci.yaml

.PHONY: vuln
vuln: ## Scan for known vulnerabilities
	@go run golang.org/x/vuln/cmd/govulncheck@latest ./...

.PHONY: tidy
tidy: ## Tidy the module files
	@go mod tidy
	@go mod tidy -modfile=tools/go.mod

.PHONY: tools
tools:
	@go get -modfile=tools/go.mod -tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint

.PHONY: clean
clean: ## Remove generated files
	@rm -rf $(COVERPROFILE) testdata/fuzz
