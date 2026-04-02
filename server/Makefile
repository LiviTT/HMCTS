SHELL=bash
TESTABLE=$$(go list ./...)

lint: ## lints with golangci-lint
	@ golangci-lint run -c .golangci.yml
.PHONY: lint

fmt:
	@ goimports -w -local github.com/LiviTT/HMCTS $$(find . -type f -name '*.go' -not -path "./vendor/*")
.PHONY: fmt
