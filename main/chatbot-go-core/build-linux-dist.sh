#!/usr/bin/env bash
# 构建 Linux 可执行文件，并生成可拷贝部署的目录（二进制 + config + start.sh）。
# chatbot-go-core 的上行音频管线依赖 Opus cgo 解码，macOS 上不能再用 CGO_ENABLED=0 直接交叉编译。
# 默认通过 Docker 在 Linux 环境内构建；如果你已经在 Linux 且安装了 libopus/libopusfile，可设置 USE_DOCKER=0。
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

GOOS=linux
GOARCH="${GOARCH:-amd64}"
USE_DOCKER="${USE_DOCKER:-auto}"
OUTPUT_ROOT="${OUTPUT_ROOT:-"$SCRIPT_DIR/dist"}"
OUT="$OUTPUT_ROOT"

mkdir -p "$OUT"
find "$OUT" -mindepth 1 -maxdepth 1 -exec rm -rf {} +

echo "==> GOOS=${GOOS} GOARCH=${GOARCH}"
echo "==> 输出目录: ${OUT}"

build_local() {
  export CGO_ENABLED=1
  export GOOS
  export GOARCH
  # 依赖下载走本机 Go 配置；国内环境可先: export GOPROXY=https://goproxy.cn,direct
  go build -trimpath -ldflags="-s -w" -o "$OUT/chatbot-go-core" ./cmd/server
  if command -v ldd >/dev/null 2>&1; then
    mkdir -p "$OUT/lib"
    ldd "$OUT/chatbot-go-core" \
      | awk '/=> \// && ($1 ~ /libopus|libopusfile|libogg/) {print $3}' \
      | while read -r lib; do cp -L "$lib" "$OUT/lib/"; done
  fi
}

build_docker() {
  if ! command -v docker >/dev/null 2>&1; then
    echo "错误: 当前构建需要 Docker。你也可以在 Linux 上安装 pkg-config/libopus-dev/libopusfile-dev 后执行 USE_DOCKER=0 ./build-linux-dist.sh" >&2
    exit 1
  fi
  # 将宿主机 OUT 映射为容器内 /src 下的相对路径（默认 OUT 为项目下 dist/）
  local out_in_container="/src${OUT#$SCRIPT_DIR}/chatbot-go-core"
  local docker_platform="linux/$GOARCH"
  docker run --rm \
    --platform "$docker_platform" \
    -v "$SCRIPT_DIR:/src" \
    -w /src \
    -e "GOOS=$GOOS" \
    -e "GOARCH=$GOARCH" \
    -e CGO_ENABLED=1 \
    -e "GOPROXY=${GOPROXY:-https://goproxy.cn,direct}" \
    -e "OUT_BIN=$out_in_container" \
    golang:1.24-bookworm \
    bash -c 'export PATH="/usr/local/go/bin:$PATH" && apt-get update && apt-get install -y --no-install-recommends pkg-config libopus-dev libopusfile-dev ca-certificates && go build -trimpath -ldflags="-s -w" -o "$OUT_BIN" ./cmd/server && out_dir="$(dirname "$OUT_BIN")" && mkdir -p "$out_dir/lib" && ldd "$OUT_BIN" | awk '\''/=> \// && ($1 ~ /libopus|libopusfile|libogg/) {print $3}'\'' | while read -r lib; do cp -L "$lib" "$out_dir/lib/"; done'
}

if [[ "$USE_DOCKER" == "1" || "$USE_DOCKER" == "true" ]]; then
  build_docker
elif [[ "$USE_DOCKER" == "0" || "$USE_DOCKER" == "false" ]]; then
  build_local
elif [[ "$(uname -s)" == "Linux" ]]; then
  build_local
else
  build_docker
fi

chmod +x "$OUT/chatbot-go-core"

if [[ ! -f config.json ]]; then
  echo "错误: 未找到 ${SCRIPT_DIR}/config.json" >&2
  exit 1
fi
cp -f config.json "$OUT/config.json"

cat >"$OUT/start.sh" <<'EOS'
#!/usr/bin/env bash
# 在 Linux 上于本目录启动 chatbot-go-core。
# 可选环境变量: CONFIG（默认 config.json）、METRICS_ADDR（默认 :9090）
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT"
CONFIG="${CONFIG:-config.json}"
METRICS_ADDR="${METRICS_ADDR:-:9090}"
export LD_LIBRARY_PATH="$ROOT/lib:${LD_LIBRARY_PATH:-}"
ulimit -n 1048576 && exec "./chatbot-go-core" -config "$CONFIG" -metrics_addr "$METRICS_ADDR"
EOS
chmod +x "$OUT/start.sh"

cat >"$OUT/README.deploy.md" <<'EOS'
# chatbot-go-core Linux 部署说明

打包目录默认已包含 `lib/` 下的 Opus 运行库，`start.sh` 会自动设置 `LD_LIBRARY_PATH`。

如果你不使用 `start.sh`，或目标机器需要使用系统库，也可以安装：

```bash
sudo apt-get update
sudo apt-get install -y libopus0 libopusfile0
```

启动：

```bash
./start.sh
```
EOS

echo "==> 完成。部署到 Linux 后在该目录执行: ./start.sh"
ls -la "$OUT"
