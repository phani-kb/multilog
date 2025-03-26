#!/bin/bash

FMT_TOOLS=("goimports" "golines" "gofmt")
FMT_INSTALL_TOOLS=("goimports" "golines")
GOBIN=$(go env GOPATH)/bin
GOLANGCI_LINT_VERSION="latest"

install_fmt_tools() {
  for tool in "${FMT_INSTALL_TOOLS[@]}"; do
    if ! command -v "$GOBIN/$tool" &>/dev/null; then
      echo "Installing $tool..."
      if [ "$tool" = "golines" ]; then
        go install github.com/segmentio/golines@latest
      else
        go install golang.org/x/tools/cmd/"$tool"@latest
      fi
    fi
  done
  export PATH="$GOBIN:$PATH"
}

install_golangci_lint() {
  if ! command -v "$GOBIN/golangci-lint" &>/dev/null; then
    echo "Installing golangci-lint..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@$GOLANGCI_LINT_VERSION
  fi
  export PATH="$GOBIN:$PATH"
}

run_lint() {
  install_golangci_lint
  echo "Running golangci-lint..."
  "$GOBIN/golangci-lint" run --config .golangci.yml ./...
}

fmt() {
  install_fmt_tools
  for tool in "${FMT_TOOLS[@]}"; do
    echo "Running $tool..."
    if [ "$tool" = "golines" ]; then
      "$GOBIN/$tool" --max-len=120 -w .
    elif [ "$tool" = "gofmt" ]; then
      $tool -s -w .
    else
      "$GOBIN/$tool" -w .
    fi
  done
}

clean_tools() {
  echo "Removing golangci-lint..."
  rm -f "$GOBIN/golangci-lint"
}

case "$1" in
fmt)
  fmt
  ;;
lint)
  run_lint
  ;;
install-tools)
  install_golangci_lint
  ;;
clean-tools)
  clean_tools
  ;;
*)
  echo "Usage: $0 {fmt|lint|install-tools|clean-tools}"
  exit 1
  ;;
esac