# VAD 服务化接入说明

本文档说明如何将 `xiaozhi-server` 原本本地执行的 Silero VAD 推理，改造成调用独立 HTTP 服务的形式。

当前示例默认部署信息如下：

- VAD 服务部署机：`192.168.99.103`
- VAD 服务端口：`8766`
- 服务地址：`http://192.168.99.103:8766`

如果你的部署地址不同，只需要替换配置中的 `base_url` 和端口即可。

## 1. 改造目标

本次改造完成了三部分：

- `xiaozhi-server` 新增 `remote_http` 类型的 VAD Provider，后端不再直接加载本地 VAD 模型，而是调用独立 VAD 服务
- `manager-api` 新增一个可选的 VAD 模型配置模板 `SileroVAD 服务版`
- 新增独立项目 `main/xiaozhi-vad-service`，用于单独部署 VAD 推理服务，并支持 Docker 容器化

这样做的好处：

- 主服务与 VAD 推理解耦
- VAD 可以独立部署、重启、迁移
- 后续更换部署机器或扩展服务更方便
- `xiaozhi-server` 宿主机不必再直接持有 Silero VAD 运行环境

## 2. 整体架构

改造后的链路如下：

```text
设备音频(opus)
  -> xiaozhi-server
  -> remote_http VAD Provider
  -> POST /v1/vad/infer
  -> xiaozhi-vad-service
  -> 返回 have_voice / client_voice_stop / last_is_voice
  -> xiaozhi-server 继续执行 ASR / LLM / TTS
```

连接结束或音频状态重置时：

```text
xiaozhi-server
  -> DELETE /v1/vad/sessions/{session_id}
  -> 远端 VAD 服务释放该 session 的状态
```

## 3. 代码改动位置

### 3.1 `xiaozhi-server`

- `main/xiaozhi-server/core/providers/vad/remote_http.py`
  - 新增远程 VAD Provider
  - 调用 `GET /health`
  - 调用 `POST /v1/vad/infer`
  - 调用 `DELETE /v1/vad/sessions/{session_id}`

- `main/xiaozhi-server/core/providers/vad/base.py`
  - 增加 `reset_conn_resources`
  - 增加 `release_conn_resources`

- `main/xiaozhi-server/core/connection.py`
  - 增加 `reset_vad_states()`
  - 在连接重置时同步清理远端 session 状态

- `main/xiaozhi-server/config.yaml`
- `main/xiaozhi-server/data/.config.yaml`
  - 增加 `SileroVADService` 示例配置

### 3.2 `manager-api`

- `main/manager-api/src/main/resources/db/changelog/202604231130.sql`
  - 增加 `remote_http` 类型的 VAD 提供器
  - 增加 `VAD_SileroVADService` 默认模型配置

- `main/manager-api/src/main/resources/db/changelog/db.changelog-master.yaml`
  - 注册新的 changelog

### 3.3 独立 VAD 服务项目

新增目录：

```text
main/xiaozhi-vad-service/
├─ app.py
├─ requirements.txt
├─ Dockerfile
├─ docker-compose.yml
├─ .dockerignore
└─ README.md
```

## 4. VAD 服务接口说明

### 4.1 健康检查

```http
GET /health
```

示例：

```bash
curl http://192.168.99.103:8766/health
```

正常返回示例：

```json
{
  "status": "healthy",
  "model_dir": "/app/models/snakers4_silero-vad",
  "session_count": 0
}
```

### 4.2 单包推理

```http
POST /v1/vad/infer
Content-Type: application/octet-stream
```

请求参数使用 query string：

- `session_id`
- `threshold`
- `threshold_low`
- `min_silence_duration_ms`
- `frame_window_threshold`

请求体：

- 原始 `opus` 音频包二进制内容

返回示例：

```json
{
  "have_voice": true,
  "client_have_voice": true,
  "client_voice_stop": false,
  "last_is_voice": true
}
```

### 4.3 删除会话

```http
DELETE /v1/vad/sessions/{session_id}
```

用于释放远端 session 的 VAD 状态，避免连接关闭后状态残留。

## 5. 在 Linux 服务器部署 VAD 服务

以下示例假设你在 Linux 服务器上操作，部署路径为：

```bash
~/work/chatbot/main/xiaozhi-vad-service
```

### 5.1 前置条件

- 已安装 Docker
- 已安装 Docker Compose
- 宿主机可访问模型目录
- 防火墙允许 `8766` 端口入站访问

默认挂载的模型目录为：

```bash
../xiaozhi-server/models/snakers4_silero-vad
```

也就是相对于 `main/xiaozhi-vad-service`，会去挂载：

```bash
main/xiaozhi-server/models/snakers4_silero-vad
```

### 5.2 启动服务

进入目录后执行：

```bash
cd ~/work/chatbot/main/xiaozhi-vad-service
docker compose up --build -d
```

如果你修改了依赖文件，建议强制重建：

```bash
docker compose down
docker compose build --no-cache
docker compose up -d
```

### 5.3 查看容器日志

查看所有 compose 日志：

```bash
docker compose logs -f
```

仅查看当前服务：

```bash
docker compose logs -f xiaozhi-vad-service
```

按容器名查看：

```bash
docker logs -f xiaozhi-vad-service
```

### 5.4 检查服务是否正常

先在服务器本机检查：

```bash
curl http://127.0.0.1:8766/health
```

再使用局域网地址检查：

```bash
curl http://192.168.99.103:8766/health
```

如果本机可通、局域网不可通，通常是防火墙或安全策略问题。

## 6. 配置 `xiaozhi-server` 调用远程 VAD

### 6.1 直接修改 YAML 配置

如果你是最简部署，可以修改：

- `main/xiaozhi-server/config.yaml`
- 或 `main/xiaozhi-server/data/.config.yaml`

配置示例如下：

```yaml
selected_module:
  VAD: SileroVADService

VAD:
  SileroVADService:
    type: remote_http
    base_url: http://192.168.99.103:8766
    timeout_ms: 1500
    threshold: 0.5
    threshold_low: 0.3
    min_silence_duration_ms: 200
    frame_window_threshold: 3
    api_key: ""
```

字段说明：

- `type`
  - 固定写 `remote_http`

- `base_url`
  - VAD 服务地址
  - 当前示例为 `http://192.168.99.103:8766`

- `timeout_ms`
  - 每个音频包调用远端服务的超时时间

- `threshold`
  - 高阈值，语音概率大于等于该值时判为语音

- `threshold_low`
  - 低阈值，语音概率小于等于该值时判为非语音

- `min_silence_duration_ms`
  - 达到该静音时长后，认为本轮说话结束

- `frame_window_threshold`
  - 滑动窗口中至少多少帧为 `True` 才算检测到人声

- `api_key`
  - 可选，若 VAD 服务配置了鉴权，则这里填写对应值

### 6.2 使用 `manager-api` 配置

如果你的部署方式使用了 `manager-api` 和智控台：

1. 确保新的 migration 已执行
2. 在模型配置中可以看到新的 VAD 模型：
   - `SileroVAD 服务版`
3. 选择该模型，并将 `base_url` 配置为：

```text
http://192.168.99.103:8766
```

修改完成后，重启相关服务，或按现有配置热更新流程生效。

## 7. 启动顺序建议

推荐顺序如下：

1. 先启动 `xiaozhi-vad-service`
2. 用 `curl` 验证 `/health`
3. 再启动 `xiaozhi-server`
4. 最后进行设备侧语音测试

不建议先启动 `xiaozhi-server` 再去补启 VAD 服务，因为当前 `remote_http` Provider 初始化时会先做健康检查。

## 8. 常见问题排查

### 8.1 `curl` 提示连接被拒绝

典型现象：

```bash
curl: (7) Failed to connect to 192.168.99.103 port 8766: Connection refused
```

排查步骤：

```bash
docker ps -a | grep xiaozhi-vad-service
docker compose logs --tail 200
ss -lntp | grep 8766
```

一般有两类原因：

- 容器进程启动失败，服务根本没有监听 `8766`
- 服务已经监听本机端口，但局域网访问被防火墙拦截

### 8.2 日志提示 `No module named 'torchaudio'`

典型日志：

```text
ModuleNotFoundError: No module named 'torchaudio'
```

说明镜像内缺少 `torchaudio`，需要重新构建镜像：

```bash
docker compose down
docker compose build --no-cache
docker compose up -d
```

### 8.3 本机能访问，局域网 IP 不能访问

先确认本机正常：

```bash
curl http://127.0.0.1:8766/health
```

再确认局域网地址：

```bash
curl http://192.168.99.103:8766/health
```

如果只有第二个失败，请检查：

```bash
sudo ufw status
sudo firewall-cmd --list-ports
ss -lntp | grep 8766
```

必要时放行端口：

```bash
sudo ufw allow 8766/tcp
```

或：

```bash
sudo firewall-cmd --add-port=8766/tcp --permanent
sudo firewall-cmd --reload
```

### 8.4 `xiaozhi-server` 启动时报 VAD 服务不可用

说明 `remote_http` Provider 在初始化健康检查时访问失败。请依次检查：

1. `xiaozhi-vad-service` 是否已启动
2. `base_url` 是否写对
3. `8766` 是否通
4. 目标服务器是否允许跨机访问

### 8.5 模型目录挂载错误

如果容器日志里出现模型加载路径错误，请检查：

```bash
../xiaozhi-server/models/snakers4_silero-vad
```

这个路径在你的 Linux 服务器上是否真实存在。

可用以下命令确认：

```bash
ls -lah ../xiaozhi-server/models/snakers4_silero-vad
```

## 9. 推荐验证流程

建议按照下面顺序完整验证：

### 9.1 服务层验证

```bash
docker compose up --build -d
docker compose logs -f
curl http://127.0.0.1:8766/health
curl http://192.168.99.103:8766/health
```

### 9.2 主服务层验证

将 `xiaozhi-server` 切换到：

```yaml
selected_module:
  VAD: SileroVADService
```

启动 `xiaozhi-server` 后，观察其日志是否能正常初始化 VAD。

### 9.3 业务链路验证

使用设备或测试页发起语音输入，确认：

- 语音开始时能被检测到
- 语音结束后能触发 `client_voice_stop`
- ASR 能继续识别
- 整条链路不受影响

## 10. 补充说明

当前的 VAD 服务设计为“单实例内存会话状态”模式：

- 每个 `session_id` 在服务端维护一份 VAD 状态
- 适合单实例部署
- 如果未来要做多实例负载均衡，需要额外考虑会话粘滞或集中式状态存储

如果后续需要继续补：

- Swagger/OpenAPI 接口说明
- Nginx 反向代理配置
- Systemd 启动方式
- 多实例部署建议

可以继续在此文档基础上扩展。
