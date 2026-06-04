#!/usr/bin/env bash
# 单机启动多个 xiaozhi-go-core 进程（不同 WS / HTTP / metrics 端口）。
# 用法（在 main/xiaozhi-go-core 目录下执行）:
#   ./run_multi_go_core_instances.sh <实例数> [WS起始端口] [HTTP起始端口] [metrics起始端口]
#
# 默认端口规划:
# - WS:      8000 起
# - HTTP:    WS_BASE + 实例数 + HTTP_GAP，默认 HTTP_GAP=100
# - metrics: 9090 起
#
# 可用环境变量:
# - BIN: 指定 xiaozhi-go-core 二进制路径，默认优先 ./xiaozhi-go-core，其次 ./bin/xiaozhi-go-core
# - BASE_CONFIG: 指定基础配置文件，默认 ./config.json
# - HTTP_GAP: 自动 HTTP 起始端口间隔，默认 100
# - TTS_ROUTE_MODE: tts 路由模式，dedicated(默认)/shared
#   - dedicated: 每个实例只保留一个专属 tts service（按实例号轮询）
#   - shared: 每个实例保留基础配置中的全部 tts services
# - TTS_ROUTE_OFFSET: dedicated 模式下的起始偏移，默认 0

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# 部署模式下脚本位于 .../go-core/current/，希望 PID/日志固定在 .../go-core 下，避免目录切换后丢失状态。
if [[ "$(basename "$SCRIPT_DIR")" == "current" ]]; then
  DEPLOY_ROOT="$(dirname "$SCRIPT_DIR")"
else
  DEPLOY_ROOT="$SCRIPT_DIR"
fi

WORK_ROOT="${GO_CORE_WORK_ROOT:-$DEPLOY_ROOT}"
cd "$SCRIPT_DIR"
export LD_LIBRARY_PATH="$SCRIPT_DIR/lib:${LD_LIBRARY_PATH:-}"

COUNT="${1:?实例数}"
WS_BASE="${2:-8000}"
HTTP_GAP="${HTTP_GAP:-100}"
METRICS_BASE="${4:-9090}"
BASE_CONFIG="${BASE_CONFIG:-${SCRIPT_DIR}/config.json}"
TTS_ROUTE_MODE="${TTS_ROUTE_MODE:-dedicated}"
TTS_ROUTE_OFFSET="${TTS_ROUTE_OFFSET:-0}"
PID_FILE="${WORK_ROOT}/.multi_go_core_pids"
RUNTIME_DIR="${WORK_ROOT}/multi_instance_runtime"
ALT_WORK_ROOT="$SCRIPT_DIR"
ALT_PID_FILE="${ALT_WORK_ROOT}/.multi_go_core_pids"
ALT_RUNTIME_DIR="${ALT_WORK_ROOT}/multi_instance_runtime"

ensure_writable_runtime() {
  local preferred_pid_file="$1"
  local preferred_runtime_dir="$2"
  local preferred_root
  preferred_root="$(dirname "$preferred_pid_file")"

  mkdir -p "$preferred_root" 2>/dev/null || true
  mkdir -p "$preferred_runtime_dir" 2>/dev/null || true

  if [[ -w "$preferred_root" && -w "$preferred_runtime_dir" ]]; then
    PID_FILE="$preferred_pid_file"
    RUNTIME_DIR="$preferred_runtime_dir"
    return 0
  fi

  echo "警告: 工作目录不可写，切换到脚本目录存放运行时文件。" >&2
  echo "  不可写目录: $preferred_root / $preferred_runtime_dir" >&2
  echo "  回退目录: $ALT_WORK_ROOT / $ALT_RUNTIME_DIR" >&2

  mkdir -p "$ALT_WORK_ROOT"
  mkdir -p "$ALT_RUNTIME_DIR"
  if [[ ! -w "$ALT_WORK_ROOT" || ! -w "$ALT_RUNTIME_DIR" ]]; then
    echo "错误: 回退目录也不可写，请修复目录权限后重试。" >&2
    exit 1
  fi
  PID_FILE="$ALT_PID_FILE"
  RUNTIME_DIR="$ALT_RUNTIME_DIR"
}

if [[ $# -ge 3 ]]; then
  HTTP_BASE="$3"
else
  HTTP_BASE=$((WS_BASE + COUNT + HTTP_GAP))
fi

if ! [[ "$COUNT" =~ ^[0-9]+$ ]] || [[ "$COUNT" -lt 1 ]]; then
  echo "实例数须为正整数" >&2
  exit 1
fi

if [[ ! -f "$BASE_CONFIG" ]]; then
  echo "基础配置不存在: $BASE_CONFIG" >&2
  exit 1
fi

ensure_writable_runtime "$PID_FILE" "$RUNTIME_DIR"

if [[ -n "${BIN:-}" ]]; then
  GO_CORE_BIN="$BIN"
elif [[ -x "${SCRIPT_DIR}/xiaozhi-go-core" ]]; then
  GO_CORE_BIN="${SCRIPT_DIR}/xiaozhi-go-core"
elif [[ -x "${SCRIPT_DIR}/bin/xiaozhi-go-core" ]]; then
  GO_CORE_BIN="${SCRIPT_DIR}/bin/xiaozhi-go-core"
else
  echo "未找到 xiaozhi-go-core 二进制。请先构建，或用 BIN=/path/to/xiaozhi-go-core 指定。" >&2
  exit 1
fi

WS_END=$((WS_BASE + COUNT - 1))
HTTP_END=$((HTTP_BASE + COUNT - 1))
METRICS_END=$((METRICS_BASE + COUNT - 1))

mkdir -p "$RUNTIME_DIR"
: >"$PID_FILE"

echo "脚本目录: $SCRIPT_DIR"
echo "工作目录: $WORK_ROOT"
echo "二进制: $GO_CORE_BIN"
echo "基础配置: $BASE_CONFIG"
echo "TTS 路由模式: $TTS_ROUTE_MODE (offset=${TTS_ROUTE_OFFSET})"
echo "PID 文件: $PID_FILE"
echo ""
echo "========== 端口规划（共 ${COUNT} 个实例）=========="
echo "WebSocket 端口区间: ${WS_BASE} – ${WS_END}"
echo "HTTP/OTA  端口区间: ${HTTP_BASE} – ${HTTP_END}"
echo "metrics   端口区间: ${METRICS_BASE} – ${METRICS_END}"
echo ""
printf "%-6s %-8s %-8s %-8s %-52s %s\n" "实例" "WS" "HTTP" "metrics" "WebSocket URL" "OTA URL"
echo "------------------------------------------------------------------------------------------------------"
for ((i = 0; i < COUNT; i++)); do
  WS=$((WS_BASE + i))
  HTTP=$((HTTP_BASE + i))
  METRICS=$((METRICS_BASE + i))
  printf "%-6s %-8s %-8s %-8s %-52s %s\n" "#${i}" "$WS" "$HTTP" "$METRICS" "ws://<HOST>:${WS}/xiaozhi/v1/" "http://<HOST>:${HTTP}/xiaozhi/ota/"
done
echo "------------------------------------------------------------------------------------------------------"
echo "（将 <HOST> 换为你的公网 IP 或域名）"
echo ""

for ((i = 0; i < COUNT; i++)); do
  WS=$((WS_BASE + i))
  HTTP=$((HTTP_BASE + i))
  METRICS=$((METRICS_BASE + i))
  INSTANCE_DIR="${RUNTIME_DIR}/instance_${i}"
  CONFIG_FILE="${INSTANCE_DIR}/config.json"
  LOG_FILE="${INSTANCE_DIR}/go_core_${i}.log"
  mkdir -p "$INSTANCE_DIR"

  python3 - "$BASE_CONFIG" "$CONFIG_FILE" "$WS" "$HTTP" "$i" "$COUNT" "$TTS_ROUTE_MODE" "$TTS_ROUTE_OFFSET" <<'PY'
import json
import sys

src, dst, ws, http, idx, count, route_mode, route_offset = sys.argv[1:9]
idx = int(idx)
count = int(count)
route_offset = int(route_offset)
with open(src, "r", encoding="utf-8") as f:
    cfg = json.load(f)

gateway = cfg.setdefault("gateway", {})
gateway["enabled"] = True
gateway["ws_listen_addr"] = f":{ws}"
gateway["http_listen_addr"] = f":{http}"
if gateway.get("websocket_url", "").startswith("你的") or gateway.get("websocket_url", "") == "":
    gateway["websocket_url"] = "你的websocket地址"

tts = cfg.setdefault("tts", {})
services = tts.get("services") or []
selected_name = ""
selected_url = ""
if route_mode == "dedicated" and len(services) > 0:
    selected_idx = (route_offset + idx) % len(services)
    selected = services[selected_idx]
    tts["services"] = [selected]
    selected_name = str(selected.get("name", ""))
    selected_url = str(selected.get("api_url", selected.get("url", "")))
elif route_mode == "shared":
    pass
else:
    raise ValueError(f"unsupported TTS_ROUTE_MODE={route_mode}, expected dedicated/shared")

with open(dst, "w", encoding="utf-8") as f:
    json.dump(cfg, f, ensure_ascii=False, indent=2)

if route_mode == "dedicated" and len(services) > 0:
    print(f"TTS_SELECTED[{idx}] index={selected_idx} name={selected_name} url={selected_url}")
elif len(services) == 0:
    print(f"TTS_SELECTED[{idx}] none(no tts.services configured)")
else:
    print(f"TTS_SELECTED[{idx}] shared(all services, count={len(services)})")
PY

  echo "启动 #$i: WS=$WS HTTP=$HTTP metrics=$METRICS"
  nohup env LD_LIBRARY_PATH="$LD_LIBRARY_PATH" "$GO_CORE_BIN" -config "$CONFIG_FILE" -metrics_addr ":${METRICS}" >>"$LOG_FILE" 2>&1 &
  pid=$!
  echo "$pid" >>"$PID_FILE"
  echo "       -> PID=$pid  配置: $CONFIG_FILE  日志: $LOG_FILE"
done

echo ""
echo "========== 各实例完整 URL（替换 HOST 后使用）=========="
URL_LIST=""
for ((i = 0; i < COUNT; i++)); do
  WS=$((WS_BASE + i))
  HTTP=$((HTTP_BASE + i))
  sep=""
  [[ -n "$URL_LIST" ]] && sep=","
  URL_LIST="${URL_LIST}${sep}ws://HOST:${WS}/xiaozhi/v1/"
  echo "  [#${i}] WS      ws://HOST:${WS}/xiaozhi/v1/"
  echo "  [#${i}] OTA     http://HOST:${HTTP}/xiaozhi/ota/"
  echo "  [#${i}] metrics http://HOST:$((METRICS_BASE + i))/status"
  echo ""
done
echo "=================================================="
echo ""
echo "已全部后台启动。查看日志: tail -f multi_instance_runtime/instance_0/go_core_0.log"
echo "压测示例（把 HOST 换成你的 IP；--urls 须列出全部 WS 端口）:"
echo "  python3 ../xiaozhi-server/performance_tester/ws_connection_stress_test.py \\"
echo "    --urls \"${URL_LIST}\" \\"
echo "    --start 50 --step 50 --max 2000"
