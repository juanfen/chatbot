# Go 后端架构规范

## 1. 分层结构（强制）

- **Handler（或 Controller）**：仅处理路由绑定、参数解析校验、调用 Service、返回统一响应；禁止写业务逻辑和数据库访问。
- **Service**：编排业务规则、事务边界、跨领域调用；通过接口依赖 Repository。
- **Repository（或 DAO）**：数据访问层，封装 GORM/SQL 查询；禁止承载业务规则。
- **Model（Entity）**：与数据库表一一对应，字段与 schema 保持一致。
- **DTO**：请求/响应对象，入参与出参分离；禁止直接把 Model 作为 API 入参。

## 2. 推荐包结构

```text
server/
├── cmd/
│   └── api/                 # 程序入口
├── internal/
│   ├── config/              # 配置加载（env/yaml）
│   ├── handler/             # HTTP 处理层
│   ├── service/             # 业务层
│   ├── repository/          # 数据访问层
│   ├── model/               # 数据模型
│   ├── dto/                 # 请求/响应 DTO
│   ├── middleware/          # JWT、方法限制、恢复等
│   ├── common/              # 统一响应、常量、错误码
│   └── util/                # 工具函数
└── resources/
    └── schema/              # 一表一 SQL 文件
```

## 3. 技术栈建议

- **Go**：1.22+（统一以 `go.mod` 为准）
- **HTTP 框架**：Gin（或项目已确定框架）
- **ORM**：GORM（需要原生 SQL 时可补充 `database/sql`）
- **数据库**：PostgreSQL 15.1
- **鉴权**：JWT（登录等白名单接口除外）
- **HTTP Method**：仅 GET/POST

## 4. 配置与部署

- 数据库和密钥等配置通过环境变量或配置文件注入，禁止硬编码。
- 按环境区分配置（如 `dev`/`prod`），并确保默认值安全可控。

## 5. 接口约定

- 所有接口必须返回统一响应结构（`code/message/data`）。
- 新增表必须同步新增 `resources/schema/{table}.sql`，并通过启动时 schema 校验。

以上为 Go 后端强制规范，所有 go 版本 server 代码必须遵守。
