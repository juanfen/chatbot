#!/usr/bin/env bash
# 停止由 run_multi_go_core_instances.sh 启动的 go-core 进程。
# 先按 PID 文件停止，再按端口兜底清理，避免残留进程占用端口。
#
# 用法:
#   ./stop_multi_go_core_instances.sh [实例数] [WS起始端口] [HTTP起始端口] [metrics起始端口]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
if [[ "$(basename "$SCRIPT_DIR")" == "current" ]]; then
  DEPLOY_ROOT="$(dirname "$SCRIPT_DIR")"
else
  DEPLOY_ROOT="$SCRIPT_DIR"
fi

WORK_ROOT="${GO_CORE_WORK_ROOT:-$DEPLOY_ROOT}"
HTTP_GAP="${HTTP_GAP:-100}"
COUNT="${1:-4}"
WS_BASE="${2:-8000}"
if [[ $# -ge 3 ]]; then
  HTTP_BASE="$3"
else
  HTTP_BASE=$((WS_BASE + COUNT + HTTP_GAP))
fi
METRICS_BASE="${4:-9090}"

PID_FILE_PRIMARY="${WORK_ROOT}/.multi_go_core_pids"
PID_FILE_FALLBACK="${SCRIPT_DIR}/.multi_go_core_pids"

declare -A seen_pids=()
declare -a target_pids=()

add_pid() {
  local pid="$1"
  [[ -z "$pid" ]] && return 0
  [[ ! "$pid" =~ ^[0-9]+$ ]] && return 0
  if [[ -z "${seen_pids[$pid]+x}" ]]; then
    seen_pids["$pid"]=1
    target_pids+=("$pid")
  fi
}

read_pid_file() {
  local file="$1"
  if [[ -f "$file" ]]; then
    while read -r pid; do
      add_pid "$pid"
    done <"$file"
  fi
}

find_pids_by_port() {
  local port="$1"
  if command -v ss >/dev/null 2>&1; then
    while read -r pid; do
      add_pid "$pid"
    done < <(ss -ltnp 2>/dev/null | awk -v p=":${port}" '$4 ~ p { if (match($0,/pid=[0-9]+/)) { print substr($0,RSTART+4,RLENGTH-4) } }')
  elif command -v lsof >/dev/null 2>&1; then
    while read -r pid; do
      add_pid "$pid"
    done < <(lsof -ti "tcp:${port}" 2>/dev/null || true)
  fi
}

wait_ports_release() {
  local max_wait_s="${1:-8}"
  local elapsed=0
  while (( elapsed < max_wait_s )); do
    local busy=0
    for ((i = 0; i < COUNT; i++)); do
      local ws_port=$((WS_BASE + i))
      local http_port=$((HTTP_BASE + i))
      local metrics_port=$((METRICS_BASE + i))
      if command -v ss >/dev/null 2>&1; then
        ss -ltn "( sport = :${ws_port} or sport = :${http_port} or sport = :${metrics_port} )" 2>/dev/null | awk 'NR>1{exit 1}' || busy=1
      elif command -v lsof >/dev/null 2>&1; then
        lsof -ti "tcp:${ws_port}" "tcp:${http_port}" "tcp:${metrics_port}" >/dev/null 2>&1 && busy=1 || true
      fi
    done
    if (( busy == 0 )); then
      return 0
    fi
    sleep 1
    elapsed=$((elapsed + 1))
  done
  return 1
}

read_pid_file "$PID_FILE_PRIMARY"
read_pid_file "$PID_FILE_FALLBACK"

for ((i = 0; i < COUNT; i++)); do
  find_pids_by_port "$((WS_BASE + i))"
  find_pids_by_port "$((HTTP_BASE + i))"
  find_pids_by_port "$((METRICS_BASE + i))"
done

if [[ ${#target_pids[@]} -eq 0 ]]; then
  echo "未发现可停止的进程（PID 文件与端口扫描均为空）。"
  rm -f "$PID_FILE_PRIMARY" "$PID_FILE_FALLBACK"
  exit 0
fi

echo "准备停止进程数: ${#target_pids[@]}"

for pid in "${target_pids[@]}"; do
  if kill -0 "$pid" 2>/dev/null; then
    echo "停止 PID=$pid (SIGTERM)"
    kill "$pid" 2>/dev/null || true
  else
    echo "PID=$pid 已不存在，跳过"
  fi
done

sleep 2

for pid in "${target_pids[@]}"; do
  if kill -0 "$pid" 2>/dev/null; then
    echo "强制停止 PID=$pid (SIGKILL)"
    kill -9 "$pid" 2>/dev/null || true
  fi
done

if wait_ports_release 10; then
  echo "端口释放完成。"
else
  echo "警告: 仍有目标端口占用，请手工检查。"
fi

rm -f "$PID_FILE_PRIMARY" "$PID_FILE_FALLBACK"
echo "完成。"
