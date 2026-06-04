## Kemande 自研 go-core（主链路）

本仓库当前**实时语音主链路**由 Kemande 自研 `xiaozhi-go-core` 承接：设备直接连接 Go 网关，Go 侧完成 **VAD → ASR → LLM → TTS** 编排，并将音频流按协议回传到设备。

### 目录说明

- `main/xiaozhi-go-core/`: go-core 主服务（WebSocket/OTA/音频管线/指标）
- `main/xiaozhi-server/`: 测试与兼容工具（例如压测脚本、测试页面等）
- `docs/`: 方案与集成文档（例如 VAD/TTS 相关说明）
- `ARCHITECTURE.md`: 架构与数据流说明

### 主链路数据流

```text
ESP32 / 客户端
  │
  ├─ WebSocket: /xiaozhi/v1/     （设备会话 + Opus 音频帧）
  └─ OTA HTTP:   /xiaozhi/ota/    （固件/入口信息）
          │
          ▼
xiaozhi-go-core
  ├─ Gateway: hello / listen / abort / ping
  ├─ VoicePipeline: VAD → ASR → LLM → TTS
  └─ Observability: /metrics、/status、日志（session_id/request_id/device_id）
```

### 服务入口（默认）

默认配置：`main/xiaozhi-go-core/config.json`

| 服务 | 默认入口 |
|:---:|:---|
| WebSocket | `ws://<HOST>:8000/xiaozhi/v1/` |
| OTA | `http://<HOST>:8003/xiaozhi/ota/` |
| Metrics | `http://<HOST>:9090/metrics` |
| Status | `http://<HOST>:9090/status` |

### 快速启动（本地编译）

```bash
cd main/xiaozhi-go-core
GOPROXY=https://goproxy.cn,direct go build -o bin/xiaozhi-go-core ./cmd/server/
./bin/xiaozhi-go-core -config config.json -metrics_addr :9090
```

### 多实例（单机横向扩展）

适用于压测或单机多进程部署，脚本会为每个实例自动生成独立配置并规划端口。

```bash
cd main/xiaozhi-go-core
./run_multi_go_core_instances.sh <实例数> [WS起始端口] [HTTP起始端口] [metrics起始端口]
./stop_multi_go_core_instances.sh
```

### 配置要点（简述）

- **网关监听**：`gateway.ws_listen_addr`、`gateway.http_listen_addr`
- **设备回连地址**：`gateway.websocket_url`（用于 OTA/配置下发时给设备的 WebSocket 地址）
- **LLM**：`llm.base_url`、`llm.api_key`、`llm.model_name`
- **TTS**：`tts.enabled=true` 且至少一个 `tts.services[].enabled=true`，可配置多路并发路由
- **VAD/ASR/声纹**：分别配置 `vad`、`asr`、`voiceprint`（按当前产品线实际接入启用）

更完整说明请查看：

- `main/xiaozhi-go-core/README.md`
- `main/xiaozhi-go-core/exec.md`
- `ARCHITECTURE.md`
- `docs/vad-service-integration.md`
- `docs/TTS 修改方案.md`
