# Set up GOBIN so that our binaries are installed to ./bin instead of $GOPATH/bin.
PROJECT_ROOT = $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
export GOBIN = $(PROJECT_ROOT)/bin

# Check if golangci-lint exists and get its version
GOLANGCI_LINT_EXISTS := $(shell test -f $(GOBIN)/golangci-lint && echo "yes" || echo "no")
GOLANGCI_LINT_VERSION := $(shell if [ "$(GOLANGCI_LINT_EXISTS)" = "yes" ]; then $(GOBIN)/golangci-lint version --short 2>/dev/null; else echo "not-installed"; fi)
REQUIRED_GOLANGCI_LINT_VERSION := $(shell cat .golangci.version)

# Directories containing independent Go modules.
MODULE_DIRS = .

.PHONY: all
all: lint test

.PHONY: clean
clean:
	@rm -rf $(GOBIN)

.PHONY: test
test:
	@$(foreach mod,$(MODULE_DIRS),(cd $(mod) && go test -race ./...) &&) true

.PHONY: lint
lint: golangci-lint tidy-lint

# Install golangci-lint with the required version in GOBIN if it is not already installed.
.PHONY: install-golangci-lint
install-golangci-lint:
	@if [ "$(GOLANGCI_LINT_EXISTS)" = "no" ]; then \
		echo "[lint] golangci-lint not found in $(GOBIN), installing v$(REQUIRED_GOLANGCI_LINT_VERSION)"; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) v$(REQUIRED_GOLANGCI_LINT_VERSION); \
	elif [ "$(GOLANGCI_LINT_VERSION)" != "$(REQUIRED_GOLANGCI_LINT_VERSION)" ]; then \
		echo "[lint] updating golangci-lint from v$(GOLANGCI_LINT_VERSION) to v$(REQUIRED_GOLANGCI_LINT_VERSION)"; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) v$(REQUIRED_GOLANGCI_LINT_VERSION); \
	else \
		echo "[lint] golangci-lint v$(GOLANGCI_LINT_VERSION) is already installed and up to date"; \
	fi

.PHONY: golangci-lint
golangci-lint: install-golangci-lint
	@echo "[lint] $(shell $(GOBIN)/golangci-lint version)"
	@$(foreach mod,$(MODULE_DIRS), \
		(cd $(mod) && \
		echo "[lint] golangci-lint: $(mod)" && \
		$(GOBIN)/golangci-lint run --path-prefix $(mod)) &&) true

.PHONY: tidy-lint
tidy-lint:
	@$(foreach mod,$(MODULE_DIRS), \
		(cd $(mod) && \
		echo "[lint] mod tidy: $(mod)" && \
		go mod tidy && \
		git diff --exit-code -- go.mod go.sum) &&) true
