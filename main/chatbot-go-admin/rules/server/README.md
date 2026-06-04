# Go 后端开发规范索引

本目录为 Go 版本后端（go 模块）的强制开发规范。目标是与 `rules/server` 的约束保持一致，只是将实现语境映射为 Go 常用工程实践。新开发需按顺序阅读 01～06 后再开始编码。

---

## 开发前必读（强制）

| 顺序 | 文档 | 说明 |
|------|------|------|
| 1 | [01-architecture.md](01-architecture.md) | 分层、包结构、技术栈、接口约定 |
| 2 | [02-coding-style.md](02-coding-style.md) | 命名、注释、格式、错误处理与日志 |
| 3 | [03-api-and-http.md](03-api-and-http.md) | 统一响应、仅 GET/POST、JWT、接口设计 |
| - | [HTTP_CONTRACT.md](HTTP_CONTRACT.md) | 前后端统一 HTTP/状态码/Envelope 约定（强制） |
| 4 | [04-database-and-gorm.md](04-database-and-gorm.md) | 表结构规范、Schema SQL、Repository/GORM 约束 |
| 5 | [05-utils-and-logging.md](05-utils-and-logging.md) | 工具包设计与日志规范 |
| 6 | [06-schema-version-check.md](06-schema-version-check.md) | 启动时 Schema 校验、自动建表与差异建议 |

---

## 新功能涉及新表：完整步骤（强制）

1. 按 [04](04-database-and-gorm.md) 设计表结构（小写下划线表名、`id VARCHAR(64)`、`create_time`/`update_time`/`deleted`、COMMENT）。
2. 在 `server/resources/schema/` 新增 `{表名}.sql`（完整 DDL + 基础数据，若无基础数据需注释说明）。
3. 新增 Model/Repository/Service/Handler，保持分层职责清晰。
4. 启动应用执行 [06](06-schema-version-check.md) 校验：
   - 表不存在：执行对应 schema SQL 建表并插入基础数据；
   - 表存在且一致：正常启动；
   - 表存在但不一致：启动失败并输出建议 SQL，人工处理后重启。

禁止绕过 schema 文件直接在数据库手工建表或改表。
