#!/usr/bin/env bash

GOLANGCI_LINT_VERSION="v1.60.2"
GOFUMPT_VERSION="v0.7.0"
MOCKERY_VERSION="v2.50.0"

prerequisites() {
  if [[ "$(golangci-lint --version 2>&1)" != *"$GOLANGCI_LINT_VERSION"* ]]; then
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@"${GOLANGCI_LINT_VERSION}"
  fi
  if [[ "$(gofumpt --version 2>&1)" != *"$GOFUMPT_VERSION"* ]]; then
    go install mvdan.cc/gofumpt@"${GOFUMPT_VERSION}"
  fi
  if [[ "$(mockery --version 2>&1)" != *"$MOCKERY_VERSION"* ]]; then
    go install github.com/vektra/mockery/v2@"${MOCKERY_VERSION}"
  fi
}

fmt() {
    gofumpt -l -w .
}

lint() {
  fmt
  golangci-lint run
}

generate() {
  mockery
}

prerequisites

"$@"
