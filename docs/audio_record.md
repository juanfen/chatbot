# 设备音频上报接口文档

## 概述

设备端向 chatbot-go-admin 服务端上报 WAV 音频及元数据。接口**无需登录**（无 JWT）。

| 项目 | 说明 |
|------|------|
| 基础地址 | `http://<host>:8088`（默认端口，以实际部署为准） |
| API 前缀 | `/api/v1`（兼容无前缀路径，行为一致） |
| 响应格式 | 统一 Envelope（见下文） |
| 允许方法 | `GET` / `POST` |

---

## 统一响应结构

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `code` | number | 业务状态码，`0` 表示成功 |
| `message` | string | 提示信息，失败时可展示给用户 |
| `data` | object \| null | 业务数据，失败时为 `null` |

### 业务状态码

| code | HTTP Status | 含义 |
|------|-------------|------|
| `0` | 200 | 成功 |
| `40000` | 400 | 请求参数/格式错误 |
| `50000` | 500 | 服务端内部错误 |

---

## 1. 上传设备音频

上报一条音频记录（二进制写入数据库）。

### 基本信息

| 项目 | 值 |
|------|-----|
| **URL** | `POST /api/v1/device/audio/upload` |
| **兼容 URL** | `POST /device/audio/upload` |
| **Content-Type** | `multipart/form-data` |
| **鉴权** | 不需要 |
| **单文件大小上限** | 32 MiB |

### 请求参数（form-data）

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `mac_id` | string | 是 | 设备唯一标识 |
| `audio_len` | string | 是 | 音频时长描述，如 `5s`、`12.5s` |
| `context` | string | 否 | 音频对应文本内容；未传、空串或仅空白时，服务端写入 **`暂无内容`** |
| `audio_blob` | file | 是 | WAV 音频文件；需为 `.wav` 或含 RIFF/WAVE 文件头 |

### 数据生命周期（服务端自动计算）

| 字段 | 规则 |
|------|------|
| `create_time` | 上传时刻 |
| `expire_time` | `create_time + 7 天`；到期后**列表查询不可见** |
| `delete_time` | `create_time + 14 天`；到期后**从数据库物理删除**（后台每小时清理） |

### 成功响应

**HTTP 200**，`code = 0`

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "macId": "AA:BB:CC:DD:EE:FF",
    "audioLen": "5s",
    "context": "你好，这是测试语音",
    "createTime": "2026-06-03T10:00:00+08:00",
    "expireTime": "2026-06-10T10:00:00+08:00",
    "deleteTime": "2026-06-17T10:00:00+08:00"
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | string | 记录 UUID |
| `macId` | string | 设备标识 |
| `audioLen` | string | 时长描述 |
| `context` | string | 文本内容（空值时已替换为 `暂无内容`） |
| `createTime` | string (ISO8601) | 创建时间 |
| `expireTime` | string (ISO8601) | 过期时间（查询截止） |
| `deleteTime` | string (ISO8601) | 物理删除时间 |

### 失败响应示例

**参数错误（HTTP 400，`code: 40000`）**

```json
{
  "code": 40000,
  "message": "mac_id is required",
  "data": null
}
```

| message | 触发条件 |
|---------|----------|
| `invalid multipart form` | 表单解析失败 |
| `audio_blob is required` | 未上传文件 |
| `mac_id is required` | `mac_id` 为空 |
| `audio_len is required` | `audio_len` 为空 |
| `audio_blob is empty` | 文件内容为空 |
| `audio_blob exceeds size limit` | 超过 32 MiB |
| `audio_blob must be wav` | 非 WAV 格式 |

**服务器错误（HTTP 500，`code: 50000`）**

```json
{
  "code": 50000,
  "message": "internal error",
  "data": null
}
```

### 请求示例

**cURL**

```bash
curl -X POST "http://127.0.0.1:8088/api/v1/device/audio/upload" \
  -F "mac_id=AA:BB:CC:DD:EE:FF" \
  -F "audio_len=5s" \
  -F "context=你好，这是测试语音" \
  -F "audio_blob=@/path/to/sample.wav"
```

**不传 context（使用默认值）**

```bash
curl -X POST "http://127.0.0.1:8088/api/v1/device/audio/upload" \
  -F "mac_id=AA:BB:CC:DD:EE:FF" \
  -F "audio_len=5s" \
  -F "audio_blob=@/path/to/sample.wav"
```

响应中 `context` 为 `"暂无内容"`。

---

## 2. 查询设备音频列表

按设备查询**未过期**记录（不含音频二进制，仅元数据）。

### 基本信息

| 项目 | 值 |
|------|-----|
| **URL** | `GET /api/v1/device/audio/list` |
| **兼容 URL** | `GET /device/audio/list` |
| **鉴权** | 不需要 |

### 查询参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `mac_id` | string | 是 | 设备唯一标识 |

### 筛选规则

- 仅返回 `expire_time > 当前时间` 的记录
- 按 `create_time` 降序
- 响应中**不包含** `audio_blob`

### 成功响应

**HTTP 200**，`code = 0`

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "list": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "macId": "AA:BB:CC:DD:EE:FF",
        "audioLen": "5s",
        "context": "暂无内容",
        "createTime": "2026-06-03T10:00:00+08:00",
        "expireTime": "2026-06-10T10:00:00+08:00"
      }
    ]
  }
}
```

无数据时 `list` 为空数组 `[]`。

### 失败响应

```json
{
  "code": 40000,
  "message": "mac_id is required",
  "data": null
}
```

### 请求示例

```bash
curl "http://127.0.0.1:8088/api/v1/device/audio/list?mac_id=AA:BB:CC:DD:EE:FF"
```

---

## 数据库表（参考）

表名：`biz_device_audio`

| 列名 | 类型 | 说明 |
|------|------|------|
| `id` | VARCHAR(64) | UUID 主键 |
| `mac_id` | VARCHAR(128) | 设备标识 |
| `audio_blob` | BYTEA | WAV 二进制 |
| `audio_len` | VARCHAR(32) | 时长描述 |
| `context` | TEXT | 文本内容，默认 `暂无内容` |
| `create_time` | TIMESTAMP | 创建时间 |
| `expire_time` | TIMESTAMP | 创建后 7 天 |
| `delete_time` | TIMESTAMP | 创建后 14 天 |

---

## 调用流程示意

```
设备 ──POST /device/audio/upload──► 服务端入库
                                      ├─ expire_time = now + 7d
                                      └─ delete_time = now + 14d

设备/业务 ──GET /device/audio/list──► 仅返回未过期记录

后台任务（每小时）──► DELETE WHERE delete_time <= now()
```
