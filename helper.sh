#!/usr/bin/env bash

GOLANGCI_LINT_VERSION="v2.10.1"
GOFUMPT_VERSION="v0.9.2"
MOCKERY_VERSION="v3.6.4"
DEADCODE_VERSION="v0.42.0"

prerequisites() {
  if [[ "$(golangci-lint --version 2>&1)" != *"$GOLANGCI_LINT_VERSION"* ]]; then
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@"${GOLANGCI_LINT_VERSION}"
  fi
  if [[ "$(gofumpt --version 2>&1)" != *"$GOFUMPT_VERSION"* ]]; then
    go install mvdan.cc/gofumpt@"${GOFUMPT_VERSION}"
  fi
  if [[ "$(mockery --version 2>&1)" != *"$MOCKERY_VERSION"* ]]; then
    go install github.com/vektra/mockery/v3@"${MOCKERY_VERSION}"
  fi
  if [[ "$(deadcode --version 2>&1)" != *"$DEADCODE_VERSION"* ]]; then
    go install golang.org/x/tools/cmd/deadcode@"${DEADCODE_VERSION}"
  fi
}

fmt() {
    gofumpt -l -w .
}

check_deadcode() {
    deadcode ./...
}

lint() {
  fmt
  golangci-lint run
  check_deadcode
}

generate() {
  mockery
}

prerequisites

"$@"
