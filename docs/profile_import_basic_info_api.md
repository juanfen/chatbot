# 用户基础画像导入接口

## 用途

前端让用户主动填写基础画像时，调用该接口把自然语言画像写入记忆服务。

接口不会要求前端提供 `category/key`。服务端会把 `user_basic_info_list`
交给现有长期记忆抽取链路，自动生成：

- `profile_facts`：结构化用户画像
- `memory_events`：可向量检索的长期记忆事件

## Endpoint

```http
POST /profile/import_basic_info
```

完整地址示例：

```text
http://192.168.99.103:8176/profile/import_basic_info
```

## Request JSON

```json
{
  "mac_id": "74:4E:9A:B2:D7:B1",
  "user_id": "17804681263690001000000780468126",
  "user_basic_info_list": [
    "我是一名程序员",
    "我今年28岁",
    "我的 MBTI 是 ENFP"
  ]
}
```

## Request Schema

```json
{
  "type": "object",
  "required": ["user_id", "mac_id", "user_basic_info_list"],
  "properties": {
    "user_id": {
      "type": "string",
      "description": "用户唯一标识，对应记忆系统 user_id"
    },
    "mac_id": {
      "type": "string",
      "description": "设备唯一标识；服务端会映射为记忆系统 device_id/run_id"
    },
    "device_id": {
      "type": "string",
      "description": "可选；如果传入 device_id，则优先使用 device_id，否则使用 mac_id"
    },
    "user_basic_info_list": {
      "type": "array",
      "minItems": 1,
      "maxItems": 200,
      "items": {
        "type": "string",
        "minLength": 1
      },
      "description": "用户主动填写的自然语言画像条目"
    },
    "source": {
      "type": "string",
      "default": "chatbot_user_basic_info",
      "description": "写入来源，通常不用传"
    },
    "update_profile": {
      "type": "boolean",
      "default": true,
      "description": "是否写入结构化用户画像"
    },
    "infer_memories": {
      "type": "boolean",
      "default": true,
      "description": "是否写入可检索长期记忆事件"
    },
    "metadata": {
      "type": "object",
      "default": {},
      "description": "业务侧附加信息"
    }
  }
}
```

## Response JSON

```json
{
  "scope": {
    "user_id": "17804681263690001000000780468126",
    "run_id": "74:4E:9A:B2:D7:B1"
  },
  "mac_id": "74:4E:9A:B2:D7:B1",
  "user_basic_info_count": 3,
  "result": {
    "scope": {
      "user_id": "17804681263690001000000780468126",
      "run_id": "74:4E:9A:B2:D7:B1"
    },
    "import_id": "generated-import-id",
    "message_count": 3,
    "auto_reflection": {
      "status": "pending",
      "reason": "auto_reflection_deferred"
    },
    "chunks_processed": 1,
    "target_messages": 3,
    "profile_facts_upserted": 3,
    "memories_written": 3,
    "memories_skipped_duplicate": 0,
    "memory_write_decisions": {
      "NEW": 3,
      "UPDATE": 0,
      "CONFLICT": 0,
      "NOOP": 0
    },
    "extraction_errors": 0,
    "profile": {}
  }
}
```

字段说明：

- `profile_facts_upserted`：本次新增或更新的用户画像数量
- `memories_written`：本次写入的长期记忆事件数量
- `memories_skipped_duplicate`：被去重跳过的长期记忆数量
- `memory_write_decisions`：长期记忆冲突检测结果统计
- `auto_reflection.status=pending`：只标记待反思，不会在该请求里阻塞调用反思大模型

## curl 示例

```bash
curl -X POST http://192.168.99.103:8176/profile/import_basic_info \
  -H 'Content-Type: application/json' \
  -d '{
    "mac_id": "74:4E:9A:B2:D7:B1",
    "user_id": "17804681263690001000000780468126",
    "user_basic_info_list": [
      "我是一名程序员",
      "我今年28岁",
      "我的 MBTI 是 ENFP"
    ]
  }'
```

## Python 示例

```python
import requests

MEMORY_BASE = "http://192.168.99.103:8176"

payload = {
    "mac_id": "74:4E:9A:B2:D7:B1",
    "user_id": "17804681263690001000000780468126",
    "user_basic_info_list": [
        "我是一名程序员",
        "我今年28岁",
        "我的 MBTI 是 ENFP",
    ],
}

resp = requests.post(
    f"{MEMORY_BASE}/profile/import_basic_info",
    json=payload,
    timeout=30,
)
resp.raise_for_status()
data = resp.json()

print(data["result"]["profile_facts_upserted"])
print(data["result"]["memories_written"])
```

## 设计说明

- `mac_id` 会映射到当前记忆系统的 `device_id/run_id`。
- 该接口适合用户主动填写资料，不适合每轮普通聊天调用。
- 该接口会调用大模型做画像抽取，因此建议在用户保存资料时调用，而不是在首字链路中调用。
- 如果项目组已经能提供结构化 `category/key/value`，可以改用 `/profile/upsert_facts`。
