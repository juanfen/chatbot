# xiaozhi-go-core

Go 核心服务。当前已可直接承接设备 WebSocket、OTA HTTP、VAD/ASR、LLM、TTS 主链路。

## 目录结构

```
cmd/server/main.go              # 入口：配置加载、指标启动、优雅关停
internal/
  adapter/llm_client.go         # OpenAI-compatible LLM SSE 流式调用
                                #   - tool_calls 增量检测
                                #   - 指数退避重试
  config/config.go              # JSON 配置加载与默认值
  domain/message.go             # 领域模型：Message/ToolCall/FunctionCall/ChatChunk
  observability/
    logger.go                   # 终端友好的单行日志，可切回 JSON
    metrics.go                  # 内存指标收集：请求数/错误数/首包延迟/P99
  gateway/
    ws_server.go                # 设备 WebSocket /xiaozhi/v1/
    http_server.go              # OTA/HTTP 服务
    ota_handler.go              # OTA响应与固件下载
  service/chat_service.go       # StreamChat 核心编排
                                #   - tool_calls 检测 → 事件推送
                                #   - max_output_size 字符截断
                                #   - 并发限制 (atomic counter)
                                #   - 超时控制
config.json                     # 默认配置
```

## 核心功能

| 功能 | 说明 |
|------|------|
| 流式对话 | SSE 解析 → WebSocket 事件推送 |
| 设备 WebSocket | 直接提供 `/xiaozhi/v1/`，处理 hello/listen/abort/ping 和二进制 Opus 音频 |
| OTA HTTP | 直接提供 `/xiaozhi/ota/` 和 `/xiaozhi/ota/download/{filename}` |
| 上行音频管线 | Go 负责 VAD、FunASR、声纹、LLM、TTS 编排；由 WebSocket 网关直接驱动 |
| tool_calls | 检测 LLM 的 function_call/tool_calls；WebSocket 主链路第一阶段仅记录占位 |
| Metrics HTTP | 暴露 `/metrics`（Prometheus 文本格式）与 `/status`（JSON 快照） |
| 输出限制 | 根据 `max_output_size` 按 UTF-8 字符截断 |
| 并发控制 | `max_concurrency` 超限返回 `CHAT_OVERLOADED` |
| 重试 | LLM 调用失败指数退避重试（可配置次数） |
| 指标 | 内存采集 + 60s 周期日志输出（请求数/错误率/首包P99） |
| 日志 | 默认单行文本输出，每个请求带 session_id/request_id/device_id；可用环境变量切回 JSON |

## 日志

默认日志为终端友好的单行文本，便于直接排查 ASR/VAD/LLM/TTS 链路。例如：

```text
04-28 08:52:30.551 INFO  starting xiaozhi-go-core config_file=config.json
04-28 08:54:01.126 ERROR asr_connect_failed session=xxx uri="wss://www.kemande.cn:10096" elapsed=1203 error="..."
```

可选环境变量：

- `XIAOZHI_GO_LOG_LEVEL=debug|info|warn|error`
- `XIAOZHI_GO_LOG_FORMAT=json`：切回 JSON 日志格式

## 快速启动

```bash
cd main/xiaozhi-go-core
GOPROXY=https://goproxy.cn,direct go build -o bin/xiaozhi-go-core ./cmd/server/

# 修改 config.json 中 llm.api_key
./bin/xiaozhi-go-core -config config.json -metrics_addr :9090
```

## 通信协议

设备入口：

- WebSocket: `/xiaozhi/v1/`
- OTA: `/xiaozhi/ota/`
- 固件下载: `/xiaozhi/ota/download/{filename}`

### 事件流

```
Python ─StreamChatRequest──▶ Go
Go ◀── CONTENT_CHUNK (n次) ──┤
Go ◀── TOOL_CALL (0~n次) ───┤
Go ◀── TTS_START (0~1次) ───┤
Go ◀── AUDIO_CHUNK (0~n次) ─┤
Go ◀── TTS_END (0~1次) ─────┤
Go ◀── FINISH ──────────────┘
       finish_reason: "stop" | "tool_calls"
```

当 `finish_reason=tool_calls` 时，Python 本地执行工具后重新发起 StreamChat 请求（附带工具结果），形成递归对话循环。

当上游启用音频事件时，`StreamChatResponse` 会带 `AUDIO_CHUNK`（含 `tts_stream_id/chunk_seq/codec/audio_chunk`），由 WebSocket 网关按序重排并下发到设备。

### 设备 WebSocket 音频管线

设备直接连接 go-core 的 `/xiaozhi/v1/` 后，二进制音频会进入 Go 内部 VoicePipeline：

```
Device ─hello/listen/audio──▶ Go Gateway
Go Gateway ─VAD/ASR/LLM/TTS──▶ Device
```

第一阶段 VAD 仍调用独立 `xiaozhi-vad-service`，ASR 对接 FunASR WebSocket offline 协议，声纹识别对接 `/voiceprint/identify` HTTP 接口。Go 侧 Opus 编解码依赖系统 `opus`、`opusfile`、`pkg-config`。

不再实现 vision 相关 HTTP 接口；当前产品线不需要 `/mcp/vision/explain`。

## ExecuteTool 白名单

当前 Go 侧已支持最小白名单工具（用于验证服务端执行链路）：

- `echo_text`：参数 `{"text":"..."}`，原样返回；
- `time_now`：参数 `{"format":"2006-01-02 15:04:05"}`，返回当前时间（为空则 RFC3339）；
- `math_add`：参数 `{"a":1,"b":2}`，返回两数和。

其他工具默认返回 `TOOL_ACTION_NOTFOUND`，继续由 Python 网关处理。

## 配置说明

```json
{
  "server": {
    "request_timeout_ms": 8000,
    "stream_idle_timeout_ms": 15000,
    "max_concurrency": 100
  },
  "gateway": {
    "enabled": true,
    "ws_listen_addr": ":8000",
    "http_listen_addr": ":8003",
    "websocket_url": "你的websocket地址",
    "auth_key": "",
    "auth": {
      "enabled": false,
      "allowed_devices": [],
      "expire_seconds": 2592000
    },
    "mqtt_gateway": ""
  },
  "ota": {
    "enabled": true,
    "bin_dir": "data/bin",
    "firmware_cache_ttl": 30
  },
  "xiaozhi": {
    "type": "hello",
    "version": 1,
    "transport": "websocket",
    "audio_params": {
      "format": "opus",
      "sample_rate": 16000,
      "channels": 1,
      "frame_duration": 60
    }
  },
  "llm": {
    "base_url": "https://open.bigmodel.cn/api/paas/v4/",
    "api_key": "YOUR_KEY",
    "model_name": "glm-4-flash",
    "timeout_ms": 6000
  },
  "tts": {
    "enabled": true,
    "max_parallel": 4,
    "segment_max_chars": 80,
    "services": [
      {
        "name": "tts-route-1",
        "enabled": true,
        "api_url": "http://YOUR_HOST:PORT/clone_voice",
        "voice": "晓伊",
        "timeout_ms": 15000,
        "sample_rate_hz": 24000,
        "channels": 1,
        "frame_duration_ms": 60,
        "headers": {}
      }
    ]
  },
  "vad": {
    "base_url": "http://127.0.0.1:8766",
    "timeout_ms": 1500,
    "threshold": 0.5,
    "threshold_low": 0.3,
    "min_silence_duration_ms": 200,
    "frame_window_threshold": 3,
    "api_key": ""
  },
  "asr": {
    "host": "127.0.0.1",
    "port": 10095,
    "is_ssl": false,
    "api_key": "none",
    "timeout_ms": 15000
  },
  "voiceprint": {
    "enabled": false,
    "url": "",
    "speakers": [],
    "similarity_threshold": 0.4,
    "timeout_ms": 10000
  },
  "audio_pipeline": {
    "enabled": false,
    "preroll_frames": 10,
    "min_utterance_frames": 15,
    "audio_frame_buffer_size": 128
  }
}
```

说明：
- TTS 仅对接 Python 侧 **IndexStreamTTS**（`index_stream.py`）：`POST` JSON `{"text","character"}`，流式 PCM；`AUDIO_CHUNK` 为 `pcm_s16le`，由 xiaozhi-server 编码为 Opus 后下发。
- `tts.enabled=true` 且至少一个 `services[].enabled=true`，且每条需 `api_url`（或 `url`）与 **`voice`（对应 character）** 非空时，go-core 才会调用上游。
- 可配置多路服务（1~N），go-core 会并发调用并通过 `tts_stream_id/chunk_seq` 输出事件。
