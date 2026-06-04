#!/usr/bin/env python3
"""
Analyze per-request go-core trace logs under logs/requests.

Usage:
  python tools/analyze_request_logs.py
  python tools/analyze_request_logs.py --log-dir logs/requests --top 20
"""

from __future__ import annotations

import argparse
import math
from dataclasses import dataclass
from pathlib import Path
from typing import Dict, List, Optional


@dataclass
class RequestMetrics:
    request_id: str
    finish_reason: str = ""
    total_ms: Optional[float] = None
    llm_first_chunk_ms: Optional[float] = None
    tts_call_start_ms: Optional[float] = None
    tts_call_end_ms: Optional[float] = None
    chat_response_out_ms: Optional[float] = None
    tts_route_hits: int = 0

    @property
    def tts_phase_ms(self) -> Optional[float]:
        if self.tts_call_start_ms is None or self.tts_call_end_ms is None:
            return None
        return self.tts_call_end_ms - self.tts_call_start_ms


def parse_kv_line(line: str) -> Dict[str, str]:
    # Format: "YYYY-mm-dd HH:MM:SS stage=xxx key=val ..."
    out: Dict[str, str] = {}
    tokens = line.strip().split()
    for tok in tokens:
        if "=" not in tok:
            continue
        k, v = tok.split("=", 1)
        out[k.strip()] = v.strip()
    return out


def to_float(val: Optional[str]) -> Optional[float]:
    if val is None:
        return None
    try:
        return float(val)
    except ValueError:
        return None


def load_one(path: Path) -> RequestMetrics:
    req_id = path.stem
    m = RequestMetrics(request_id=req_id)
    for raw in path.read_text(encoding="utf-8", errors="replace").splitlines():
        kv = parse_kv_line(raw)
        stage = kv.get("stage", "")
        if not stage:
            continue

        if stage == "chat_response_out":
            m.finish_reason = kv.get("finish_reason", m.finish_reason)
            m.total_ms = to_float(kv.get("total_ms")) or m.total_ms
            m.chat_response_out_ms = m.total_ms
        elif stage == "llm_first_chunk":
            m.llm_first_chunk_ms = to_float(kv.get("first_chunk_ms"))
        elif stage == "tts_call_start":
            # use first tts start in this request
            if m.tts_call_start_ms is None:
                m.tts_call_start_ms = 0.0
        elif stage == "tts_call_end":
            # sum or keep last; here keep last and estimate from total durations
            end_mark = to_float(kv.get("audio_chunks_emitted"))
            if end_mark is not None:
                # placeholder mark indicating a tts phase happened
                m.tts_call_end_ms = m.tts_call_end_ms or 0.0
        elif stage == "tts_route_hit":
            m.tts_route_hits += 1

    # if only marker-style tts values were found, tts_phase_ms is not derivable directly
    # keep them as None unless we have explicit numbers.
    if m.tts_call_start_ms == 0.0 and m.tts_call_end_ms == 0.0:
        m.tts_call_start_ms = None
        m.tts_call_end_ms = None

    return m


def percentile(values: List[float], p: float) -> Optional[float]:
    if not values:
        return None
    values = sorted(values)
    if len(values) == 1:
        return values[0]
    rank = (p / 100.0) * (len(values) - 1)
    low = math.floor(rank)
    high = math.ceil(rank)
    if low == high:
        return values[low]
    frac = rank - low
    return values[low] * (1 - frac) + values[high] * frac


def fmt_ms(v: Optional[float]) -> str:
    if v is None:
        return "-"
    return f"{v:.0f}ms"


def main() -> None:
    parser = argparse.ArgumentParser(description="Analyze go-core request trace logs")
    parser.add_argument("--log-dir", default="logs/requests", help="request log directory")
    parser.add_argument("--top", type=int, default=10, help="top N slow requests")
    args = parser.parse_args()

    log_dir = Path(args.log_dir)
    if not log_dir.exists():
        print(f"[ERROR] log dir not found: {log_dir}")
        return

    files = sorted(log_dir.glob("*.log"))
    if not files:
        print(f"[ERROR] no .log files found under: {log_dir}")
        return

    metrics = [load_one(p) for p in files]
    total = len(metrics)
    success = sum(1 for m in metrics if m.finish_reason == "stop")
    tool_calls = sum(1 for m in metrics if m.finish_reason == "tool_calls")
    failed = total - success - tool_calls

    total_ms_vals = [m.total_ms for m in metrics if m.total_ms is not None]
    first_chunk_vals = [m.llm_first_chunk_ms for m in metrics if m.llm_first_chunk_ms is not None]

    print("=== Request Log Summary ===")
    print(f"log_dir: {log_dir}")
    print(f"total_requests: {total}")
    print(f"finish_stop: {success}")
    print(f"finish_tool_calls: {tool_calls}")
    print(f"other_or_failed: {failed}")
    print("")

    print("=== Latency Percentiles ===")
    print(
        f"total_ms    P50={fmt_ms(percentile(total_ms_vals, 50))} "
        f"P95={fmt_ms(percentile(total_ms_vals, 95))} "
        f"P99={fmt_ms(percentile(total_ms_vals, 99))}"
    )
    print(
        f"first_chunk P50={fmt_ms(percentile(first_chunk_vals, 50))} "
        f"P95={fmt_ms(percentile(first_chunk_vals, 95))} "
        f"P99={fmt_ms(percentile(first_chunk_vals, 99))}"
    )
    print("")

    print(f"=== Top {args.top} Slow Requests (by total_ms) ===")
    slow = sorted(
        [m for m in metrics if m.total_ms is not None],
        key=lambda x: x.total_ms or -1,
        reverse=True,
    )[: args.top]
    for i, m in enumerate(slow, start=1):
        print(
            f"{i:02d}. request_id={m.request_id} "
            f"finish={m.finish_reason or '-'} "
            f"total={fmt_ms(m.total_ms)} "
            f"first_chunk={fmt_ms(m.llm_first_chunk_ms)} "
            f"tts_route_hits={m.tts_route_hits}"
        )


if __name__ == "__main__":
    main()
