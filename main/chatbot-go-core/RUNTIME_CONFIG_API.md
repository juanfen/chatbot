# xiaozhi-go-core 运行时配置接口（热更新/查询）

> 说明：该接口用于**远程修改并即时生效** `config.json` 中的部分配置项，但**不会修改 `config.json` 文件**，仅更新进程内存中的生效配置快照。

## 基本信息

- **服务**：Gateway HTTP（监听地址来自 `gateway.http_listen_addr`）
- **数据格式**：请求/响应均为 JSON（错误时可能为纯文本）
- **生效方式**：POST 成功后，进程内配置快照原子替换，后续读取到新配置

## 接口列表

### 1) 获取当前生效配置（可热改部分）

- **Method**：`GET`
- **Path**：`/xiaozhi/v1/runtime/config`
- **鉴权**：无（已按你的要求绕开鉴权）
- **成功响应**：`200 OK`
- **响应 Content-Type**：`application/json; charset=utf-8`
- **响应体**：仅包含以下 8 个顶层键：
  - `gateway`
  - `ota`
  - `llm`
  - `tts`
  - `vad`
  - `asr`
  - `voiceprint`
  - `chatbot_admin`
  - `audio_pipeline`

示例：

```bash
curl -sS "http://127.0.0.1:8003/xiaozhi/v1/runtime/config"
```

---

### 2) 通知并热更新配置（深度合并补丁）

- **Method**：`POST`
- **Path**：`/xiaozhi/v1/runtime/config`
- **鉴权**：无（已按你的要求绕开鉴权）
- **请求 Content-Type**：`application/json`
- **请求体**：JSON Patch（补丁），仅允许以下 8 个顶层键：
  - `gateway`、`ota`、`llm`、`tts`、`vad`、`asr`、`voiceprint`、`chatbot_admin`、`audio_pipeline`
- **合并规则**：
  - **对象（object）**：递归深度合并（只覆盖你提供的字段）
  - **数组（array）**：整体替换（例如 `tts.services`、`gateway.wakeup_words`）
  - **未提供的字段**：保持原值不变
- **Body 大小限制**：约 4MB
- **成功响应**：`200 OK`
  - 响应体：`{"status":"ok"}`

示例：只改 LLM 模型

```bash
curl -sS -H "Content-Type: application/json" \
  -d '{"llm":{"model_name":"glm-4-air"}}' \
  "http://127.0.0.1:8003/xiaozhi/v1/runtime/config"
```

示例：改 TTS services（数组整体替换）

```bash
curl -sS -H "Content-Type: application/json" \
  -d '{
    "tts": {
      "enabled": true,
      "max_parallel": 4,
      "services": [
        {
          "name": "route-1",
          "enabled": true,
          "api_url": "http://127.0.0.1:5000/tts",
          "voice": "xiaozhi",
          "timeout_ms": 15000,
          "sample_rate_hz": 24000,
          "channels": 1,
          "frame_duration_ms": 60,
          "headers": {"X-From":"runtime-patch"}
        }
      ]
    }
  }' \
  "http://127.0.0.1:8003/xiaozhi/v1/runtime/config"
```

## 错误响应

### `400 Bad Request`

常见触发条件（POST）：

- JSON 解析失败
- Body 为空：`empty patch`
- 顶层 key 不在允许列表：`unknown or disallowed key "..."`
- 分区合并失败（类型不匹配等），错误会带前缀：`gateway: ...` / `llm: ...` 等

### `405 Method Not Allowed`

对该路径使用了非 GET/POST 方法。

## 数据结构（响应/补丁字段）

以下字段名与 `config.json` 的 JSON key 完全一致（即结构体 tag），你可以按需只提交部分字段。

### `gateway`

```json
{
  "enabled": true,
  "ws_listen_addr": ":8000",
  "http_listen_addr": ":8003",
  "ws_ping_interval_seconds": 15,
  "ws_pong_timeout_seconds": 30,
  "websocket_url": "xxx",
  "auth_key": "xxx",
  "auth": {
    "enabled": false,
    "allowed_devices": [],
    "expire_seconds": 2592000
  },
  "mqtt_gateway": "",
  "wakeup_words": ["你好小智"],
  "enable_greeting": true,
  "close_connection_no_voice_time": 120,
  "end_prompt": "...",
  "end_prompt_enabled": true,
  "system_prompt": "...",
  "enable_stop_tts_notify": false,
  "stop_tts_notify_voice": "config/assets/tts_notify.mp3",
  "tts_audio_send_delay_ms": -1,
  "enable_wakeup_words_response_cache": false,
  "wakeup_words_config_file": "data/.wakeup_words.yaml",
  "wakeup_words_fallback_file": "config/assets/wakeup_words_short.wav"
}
```

`tts_audio_send_delay_ms`：`-1` 按每包真实音频时长（`xiaozhi.audio_params.frame_duration`）相对墙钟节流；`>0` 仅作每包时长下限。TTS PCM 在网关按 24k/60ms 切帧后逐帧 Opus，一帧一个 WS 二进制包。

### `ota`

```json
{
  "enabled": true,
  "bin_dir": "data/bin",
  "firmware_cache_ttl": 30
}
```

### `llm`

```json
{
  "base_url": "https://...",
  "api_key": "xxx",
  "model_name": "xxx",
  "timeout_ms": 6000,
  "use_external": false,
  "external": {
    "base_url": "https://...",
    "api_key": "xxx",
    "model_name": "xxx",
    "timeout_ms": 8000
  }
}
```

### `tts`

```json
{
  "enabled": false,
  "max_parallel": 4,
  "segment_max_chars": 80,
  "flush_min_chars": 24,
  "flush_interval_ms": 400,
  "services": [
    {
      "name": "route-1",
      "enabled": true,
      "url": "http://...",
      "api_url": "http://...",
      "voice": "xxx",
      "api_key": "xxx",
      "timeout_ms": 15000,
      "sample_rate_hz": 24000,
      "channels": 1,
      "frame_duration_ms": 60,
      "headers": {"k":"v"}
    }
  ]
}
```

### `vad`

```json
{
  "base_url": "http://127.0.0.1:8766",
  "timeout_ms": 1500,
  "threshold": 0.5,
  "threshold_low": 0.3,
  "min_silence_duration_ms": 200,
  "frame_window_threshold": 3,
  "api_key": "xxx"
}
```

### `asr`

```json
{
  "host": "127.0.0.1",
  "port": 10095,
  "is_ssl": false,
  "api_key": "none",
  "timeout_ms": 15000
}
```

### `voiceprint`

```json
{
  "enabled": false,
  "url": "https://host/path?key=YOUR_KEY",
  "speakers": ["id,name", "id2,name2"],
  "similarity_threshold": 0.4,
  "timeout_ms": 10000
}
```

### `chatbot_admin`

长音频（≥ `record_audio_len_sec`）在 ASR 后、进入 LLM 前同步上报至 chatbot-go-admin `POST /api/v1/device/audio/upload`。`gateway` 侧 `VoiceprintEnabled` 语义对应本段 `enabled`。

```json
{
  "enabled": true,
  "ip": "127.0.0.1",
  "port": 8088,
  "record_audio_len_sec": 5,
  "timeout_ms": 15000
}
```

### `audio_pipeline`

```json
{
  "enabled": false,
  "preroll_frames": 10,
  "min_utterance_frames": 15,
  "audio_frame_buffer_size": 128
}
```

## 生效边界与注意事项

- **会即时生效的典型项**：`llm`、`tts`、`vad`、`asr`、`voiceprint`、`gateway.system_prompt`、`gateway.wakeup_words`、`ota.*` 等（后续调用会读取新快照）。
- **监听端口类**：`gateway.ws_listen_addr` / `gateway.http_listen_addr` 改了只会更新内存值，**不会让已监听端口自动迁移**；要换端口仍需重启进程。

