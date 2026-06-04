# xiaozhi-go-core 启动说明

## 本地编译与启动

```bash
cd "main/xiaozhi-go-core"
GOPROXY=https://goproxy.cn,direct go build -o bin/xiaozhi-go-core ./cmd/server/
./bin/xiaozhi-go-core -config config.json -metrics_addr :9090
```

## 服务入口

- WebSocket: `ws://<HOST>:8000/xiaozhi/v1/`
- OTA: `http://<HOST>:8003/xiaozhi/ota/`
- Metrics: `http://<HOST>:9090/metrics`
- Status: `http://<HOST>:9090/status`

## 多实例

```bash
./run_multi_go_core_instances.sh <实例数> [WS起始端口] [HTTP起始端口] [metrics起始端口]
./stop_multi_go_core_instances.sh
```

## Linux 部署包启动

```bash
./start.sh
```

`start.sh` 与多实例脚本都会自动注入：

```bash
LD_LIBRARY_PATH=./lib:${LD_LIBRARY_PATH}
```

用于加载 `libopus.so.0`、`libopusfile.so.0`、`libogg.so.0`。
