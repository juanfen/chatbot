# HTTP 协议与响应约定（强制）

本文件定义 **前后端统一** 的 HTTP 请求/响应协议与状态码语义。任何新增/修改接口必须遵守。

## 1. 统一响应 Envelope（强制）

后端所有接口都必须返回如下包体结构（成功/失败都一致）：

```json
{
  \"code\": 0,
  \"message\": \"ok\",
  \"data\": {}
}
```

- **code**：业务状态码（数字）
- **message**：面向用户的提示信息（字符串，前端可直接展示）
- **data**：业务数据（可以为 `null`）

禁止使用 `msg` 字段；前端只认 `message`。

## 2. HTTP Status 与业务 code 的分工（强制）

- **HTTP Status**：表达“协议/鉴权/资源”语义
  - 200：请求成功到达业务层（无论业务成功/失败，包体都存在）
  - 400：请求参数/格式错误
  - 401：未登录/Token 无效
  - 403：已登录但无权限
  - 404：资源不存在
  - 405：Method 不允许（仅允许 GET/POST，OPTIONS 仅用于预检）
  - 500：服务端内部错误

- **业务 code**：表达“业务是否成功”
  - `0`：成功
  - `40000`：BadRequest（与 400 对应）
  - `40100`：Unauthorized（与 401 对应）
  - `40300`：Forbidden（与 403 对应）
  - `40400`：NotFound（与 404 对应）
  - `50000`：Internal（与 500 对应）

业务失败时，后端依旧返回 Envelope，`data=null`，并给出可展示的 `message`。

## 3. Method 约束（强制）

- 业务接口仅允许：`GET` / `POST`
- `OPTIONS`：仅用于 CORS 预检
- 禁止：`PUT` / `PATCH` / `DELETE` 等进入业务链路

## 4. 鉴权（强制）

- 除登录、健康检查等白名单接口外，所有接口必须 JWT 鉴权
- Header：`Authorization: Bearer <token>`

## 5. 代码落地位置（强制）

- 统一响应封装：`server/internal/common/response.go`
- 方法限制中间件：`server/internal/middleware/methodlimit.go`
- JWT 中间件：`server/internal/middleware/auth.go`

