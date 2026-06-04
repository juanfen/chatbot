# Go 接口与 HTTP 规范

## 1. 统一响应 Schema（强制）

- 所有接口统一返回：
  - `code`：业务状态码；
  - `message`：提示信息；
  - `data`：业务数据（可为 `null`）。
- 成功/失败都必须走统一结构，便于前端统一处理。

## 2. HTTP Method 限制

- 业务接口只允许使用 GET、POST。
- 禁止 PUT、DELETE、PATCH 等方法进入业务处理链路。
- OPTIONS 仅用于 CORS 预检，不承载业务逻辑。

## 3. 鉴权

- 除登录、健康检查等白名单接口外，所有请求必须通过 JWT 校验。
- 推荐 Header 方案：`Authorization: Bearer <token>`。
- 鉴权失败返回 401，且响应结构保持统一。

## 4. 接口设计建议

- 读操作优先 GET；写操作统一 POST。
- 路径统一前缀（如 `/api/v1`），便于版本演进。
- GET 参数使用 query/path；POST 使用 JSON body + DTO 校验。

## 5. 中间件约束

- 必备中间件：请求 ID、恢复 panic、JWT 鉴权、方法限制、访问日志。
- panic 必须被恢复并转换为统一错误响应，避免泄露内部堆栈。
