GOLANGCI_LINT_VERSION = v2.8.0

##@ Development
.PHONY: prereqs
prereqs:
	@echo "### Test if prerequisites are met, and installing missing dependencies"
	test -f ./bin/golangci-lint-${GOLANGCI_LINT_VERSION} || ( \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s ${GOLANGCI_LINT_VERSION} \
		&& mv ./bin/golangci-lint ./bin/golangci-lint-${GOLANGCI_LINT_VERSION})

fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: lint
lint: prereqs ## Run linter (golangci-lint)
	@echo "### Linting code"
	./bin/golangci-lint-${GOLANGCI_LINT_VERSION} --config .golangci.yml run --timeout 15m ./...
