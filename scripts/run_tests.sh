#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

echo "[1/3] unit tests"
go test -v $(go list ./... | rg -v '/tests$')

echo "[2/3] integrated suite"
go test -v ./tests

if [[ "${1:-}" == "--with-align" ]]; then
  echo "[3/3] java/go alignment check"
  ./scripts/compare_with_java.sh
else
  echo "[3/3] alignment skipped (pass --with-align to enable)"
fi
