#!/usr/bin/env bash
# 在 server 目录下启动 chatbot-go-admin API（默认监听 :8088）
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT"

export ADMIN_HTTP_ADDR="${ADMIN_HTTP_ADDR:-:8088}"
export ADMIN_SCHEMA_DIR="${ADMIN_SCHEMA_DIR:-${ROOT}/resources/schema}"

if [[ -f "${ROOT}/.env" ]]; then
  export ADMIN_DOTENV_PATH="${ADMIN_DOTENV_PATH:-${ROOT}/.env}"
elif [[ -f "${ROOT}/.env.example" ]]; then
  echo "WARN: 未找到 .env，将使用 .env.example；建议 cp .env.example .env 后按需修改" >&2
  export ADMIN_DOTENV_PATH="${ADMIN_DOTENV_PATH:-${ROOT}/.env.example}"
fi

# 若 .env 使用相对路径 data/audio，需保证目录存在
mkdir -p "${ROOT}/data/audio"

echo "==> starting chatbot-go-admin on ${ADMIN_HTTP_ADDR}"
echo "    workdir: ${ROOT}"
if [[ -n "${ADMIN_DOTENV_PATH:-}" ]]; then
  echo "    dotenv:  ${ADMIN_DOTENV_PATH}"
fi

exec go run ./cmd/api "$@"
