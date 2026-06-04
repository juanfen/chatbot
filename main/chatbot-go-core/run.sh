#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT"

BIN="${ROOT}/bin/xiaozhi-go-core"
echo "building ${BIN}..."
mkdir -p "${ROOT}/bin"
go build -o "$BIN" ./cmd/server

exec "$BIN" -config config.json -metrics_addr :9090
