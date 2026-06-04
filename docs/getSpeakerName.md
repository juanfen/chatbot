# 按 Speaker ID 查询音色名称接口文档

## 概述

根据 `speaker_id` 查询在管理端「音色识别」中登记过的**音色名称**（`speaker_name`）。数据来自表 `biz_user_speaker`。

| 项目 | 说明 |
|------|------|
| 基础地址 | `http://<host>:8088`（默认端口，以实际部署为准） |
| API 前缀 | `/api/v1`（兼容无前缀路径，行为一致） |
| 响应格式 | 统一 Envelope（见下文） |
| 鉴权 | **不需要** JWT |

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
| `message` | string | 提示信息 |
| `data` | object \| null | 业务数据 |

### 业务状态码

| code | HTTP Status | 含义 |
|------|-------------|------|
| `0` | 200 | 成功 |
| `40000` | 400 | 请求参数错误 |
| `40400` | 404 | 未找到对应 speaker |
| `50000` | 500 | 服务端内部错误 |

---

## 查询音色名称

### 基本信息

| 项目 | 值 |
|------|-----|
| **URL** | `GET /api/v1/device/audio/speaker-name` |
| **兼容 URL** | `GET /device/audio/speaker-name` |
| **鉴权** | 不需要 |

### 查询参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `speaker_id` | string | 是 | 音色 ID（绑定音色时生成的 32 位标识） |

### 成功响应

**HTTP 200**，`code = 0`

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "speakerId": "17804681263690001000000780468126",
    "speakerName": "张三"
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `speakerId` | string | 请求中的 speaker_id |
| `speakerName` | string | 管理端配置的音色显示名称 |

### 失败响应

**缺少 speaker_id（HTTP 400）**

```json
{
  "code": 40000,
  "message": "speaker_id is required",
  "data": null
}
```

**speaker 不存在（HTTP 404）**

```json
{
  "code": 40400,
  "message": "speaker not found",
  "data": null
}
```

---

## 请求示例

```bash
curl "http://127.0.0.1:8088/api/v1/device/audio/speaker-name?speaker_id=17804681263690001000000780468126"
```

仅需要名称时，读取 `data.speakerName` 即可。

---

## 与其它接口的关系

| 接口 | 说明 |
|------|------|
| `GET /device/audio/speaker-ids` | 按 `mac_id` 查询该设备已绑定的 speaker_id 列表 |
| `GET /device/audio/speaker-name` | **本接口**：按 `speaker_id` 查询音色名称 |
| `POST /sys/device-audio/bind` | 管理端分配音色（需登录），成功后才会在 `biz_user_speaker` 中有记录 |

典型流程：先用 [getSpeaker.md](./getSpeaker.md) 拿到 `speakerIds`，再对本接口逐个查询 `speakerName` 用于展示。

---

## 相关文档

- 按 MAC 查 Speaker ID：[getSpeaker.md](./getSpeaker.md)
- 设备音频上报：[audio_record.md](./audio_record.md)
