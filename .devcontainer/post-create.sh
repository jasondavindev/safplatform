#!/usr/bin/env bash
set -euo pipefail

echo "Go: $(go version)"
echo "GOROOT: $(go env GOROOT)"
echo "GOPATH: $(go env GOPATH)"
echo "gopls:  $(gopls version | head -n1)"

if [ -f go.mod ]; then
  go mod download all || true
fi
