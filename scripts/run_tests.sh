#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

echo "[1/2] unit tests"
go test -v $(go list ./... | rg -v '/tests$')

echo "[2/2] integrated suite"
go test -v ./tests
