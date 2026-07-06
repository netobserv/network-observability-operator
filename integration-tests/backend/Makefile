GOLANGCI_LINT_VERSION = v2.12.2

# Detect OS
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
	GOOS = linux
endif
ifeq ($(UNAME_S),Darwin)
	GOOS = darwin
endif

# Detect architecture
UNAME_M := $(shell uname -m)
ifeq ($(UNAME_M),x86_64)
	GOARCH = amd64
endif
ifeq ($(UNAME_M),arm64)
	GOARCH = arm64
endif
ifeq ($(UNAME_M),aarch64)
	GOARCH = arm64
endif

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

##@ Build
.PHONY: install-ginkgo
install-ginkgo: ## Install ginkgo CLI
	@echo "### Installing ginkgo CLI"
	go install github.com/onsi/ginkgo/v2/ginkgo@latest

.PHONY: build-test-binary
build-test-binary: install-ginkgo ## Build test binary for current platform with ginkgo CLI
	@echo "### Building e2e test binary for $(GOOS)/$(GOARCH)"
	GOOS=$(GOOS) GOARCH=$(GOARCH) ginkgo build --ldflags="-s -w" -o e2e-tests.test
