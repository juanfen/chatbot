# Web HTTP 请求/响应约定（强制）

本文件定义 Web 前端与后端之间 **统一** 的 HTTP 请求与响应协议。任何新增接口/页面都必须遵守。

## 1. 统一响应 Envelope（强制）

后端统一返回：

```ts
type Envelope<T> = {
  code: number;
  message: string;
  data: T | null;
};
```

- 成功：`code === 0`，`message === "ok"`
- 失败：`code !== 0`，`message` 为可展示错误信息，`data === null`

前端只认 `message` 字段，禁止使用/依赖 `msg`。

## 2. 前端请求工具层职责（强制）

所有 HTTP 行为必须集中在 `src/service/request/index.ts` 完成：

- 自动注入 `Authorization: Bearer <token>`
- 解包 `Envelope<T>`（页面拿到的必须是 `T`，而不是 FlatResponse/Envelope）
- 对异常进行统一提示与处理：
  - 网络错误提示
  - 业务失败提示（基于 `code/message`）
  - 需要登出的 code：直接 resetStore 并跳转登录
  - token 过期 code：自动刷新 token 并重放请求

**页面 / Store / 组件禁止进行 HTTP 异常处理**：

- 不要 try/catch 去弹错误提示
- 不要自己判断 401/403/500 去做登出/提示
- 只处理业务态：例如空列表、无数据等

## 3. API 层返回值约束（强制）

`src/service/api/*` 暴露的函数必须返回 **已解包的业务数据**，并在工具层提供 fallback，避免页面出现 `undefined.map` 一类错误。

示例（列表接口）：

```ts
return request<Api.Sys.UserList>({ url: '/sys/user/list', method: 'get', params }, { total: 0, list: [] });
```

## 4. 状态码语义（强制）

- HTTP Status 用于表达协议/鉴权/资源语义（401/403/404/405/500...）
- 业务成功/失败由 `code` 决定（0 为成功）

