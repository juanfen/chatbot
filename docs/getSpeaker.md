# 按 MAC 查询已注册 Speaker ID 接口文档

## 概述

根据设备 `mac_id` 查询该设备下**已完成音色绑定**（已在管理端分配音色并成功注册到 SV）的 `speaker_id` 列表。仅返回 ID 数组，不含音色名称等其它字段。

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

---

## 查询已注册 Speaker ID

### 基本信息

| 项目 | 值 |
|------|-----|
| **URL** | `GET /api/v1/device/audio/speaker-ids` |
| **兼容 URL** | `GET /device/audio/speaker-ids` |
| **鉴权** | 不需要 |

### 查询参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `mac_id` | string | 是 | 设备唯一标识 |

### 数据来源说明

- 查询表 `biz_device_audio`
- 条件：`mac_id` 匹配，且 `speaker_id` 非空（表示该条音频已在「音色识别」界面完成绑定）
- 对 `speaker_id` **去重**后升序返回
- 绑定后的记录为永久保留，不会因 7 天过期规则从本接口消失

### 成功响应

**HTTP 200**，`code = 0`

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "macId": "AA:BB:CC:DD:EE:FF",
    "speakerIds": [
      "17804681263690001000000780468126",
      "55a93296274ec2b46f086ec03556e685"
    ]
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `macId` | string | 请求中的设备 MAC |
| `speakerIds` | string[] | 已注册的 32 位 `speaker_id` 列表；无数据时为 `[]` |

### 失败响应

**缺少 mac_id（HTTP 400）**

```json
{
  "code": 40000,
  "message": "mac_id is required",
  "data": null
}
```

---

## 请求示例

```bash
curl "http://127.0.0.1:8088/api/v1/device/audio/speaker-ids?mac_id=AA:BB:CC:DD:EE:FF"
```

**仅关心 ID 数组时**，读取响应中的 `data.speakerIds` 即可。

---

## 与其它接口的关系

| 接口 | 说明 |
|------|------|
| `POST /device/audio/upload` | 设备上报音频（尚未绑定 speaker） |
| `GET /device/audio/list` | 按 mac 查询未过期音频元数据（不含 speaker 绑定信息） |
| `GET /device/audio/speaker-ids` | **本接口**：按 mac 查询已绑定 speaker_id 列表 |
| `POST /sys/device-audio/bind` | 管理端分配音色（需登录），绑定后才会出现在本接口结果中 |

---

## 相关文档

- 按 Speaker ID 查音色名称：[getSpeakerName.md](./getSpeakerName.md)
- 设备音频上报：[audio_record.md](./audio_record.md)
