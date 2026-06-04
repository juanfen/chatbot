#!/usr/bin/env bash
# 在任意平台执行：将 chatbot-go-admin API 交叉编译为 Linux x64，并打包 schema、.env 与运行脚本。
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

OUT="${SCRIPT_DIR}/dist/linux-amd64"
BIN_NAME="chatbot-admin-api"

echo "==> output: ${OUT}"
rm -rf "${OUT}"
mkdir -p "${OUT}/resources/schema"

echo "==> go build GOOS=linux GOARCH=amd64 CGO_ENABLED=0 ..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "${OUT}/${BIN_NAME}" ./cmd/api

echo "==> copy resources/schema ..."
if compgen -G "${SCRIPT_DIR}/resources/schema/*.sql" > /dev/null; then
  cp -f "${SCRIPT_DIR}/resources/schema/"*.sql "${OUT}/resources/schema/"
else
  echo "WARN: no SQL files under resources/schema" >&2
fi

echo "==> copy .env ..."
if [[ -f "${SCRIPT_DIR}/.env" ]]; then
  cp -f "${SCRIPT_DIR}/.env" "${OUT}/.env"
  chmod 600 "${OUT}/.env" 2>/dev/null || true
  echo "    (已打入本机 server/.env，含密钥请勿公开分发该目录或压缩包)"
elif [[ -f "${SCRIPT_DIR}/.env.example" ]]; then
  cp -f "${SCRIPT_DIR}/.env.example" "${OUT}/.env"
  chmod 600 "${OUT}/.env" 2>/dev/null || true
  echo "    WARN: 未找到 .env，已用 .env.example 写入包内 .env，部署前请修改密钥与数据库等配置" >&2
else
  echo "WARN: 无 .env 也无 .env.example，未打包环境文件" >&2
fi

echo "==> write run.sh ..."
cat > "${OUT}/run.sh" << 'EOS'
#!/usr/bin/env bash
# 与二进制同目录部署；在 Linux x64 上直接: ./run.sh
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT"

BIN_NAME="chatbot-admin-api"
if [[ ! -x "${ROOT}/${BIN_NAME}" ]]; then
  echo "error: ${BIN_NAME} not found or not executable" >&2
  exit 1
fi

# 数据目录（语音样本等）；与 .env 中 ADMIN_AUDIO_TARGET_PATH=data/audio 时一致
mkdir -p "${ROOT}/data/audio"

# 默认使用包内 schema；可用环境变量覆盖
export ADMIN_SCHEMA_DIR="${ADMIN_SCHEMA_DIR:-${ROOT}/resources/schema}"

# 若与本目录下放 .env，由进程内 godotenv 加载；也可用 ADMIN_DOTENV_PATH 指向其它路径
export ADMIN_DOTENV_PATH="${ADMIN_DOTENV_PATH:-${ROOT}/.env}"

exec "${ROOT}/${BIN_NAME}" "$@"
EOS

chmod +x "${OUT}/run.sh" "${OUT}/${BIN_NAME}"

echo "==> done."
echo "    tree (flat):"
ls -la "${OUT}"
echo ""
echo "    部署到 Linux：将目录同步到目标机后执行（包内已含 .env 时可直接改目标机上的 .env）:"
echo "      cd \"${OUT}\" && ./run.sh"
