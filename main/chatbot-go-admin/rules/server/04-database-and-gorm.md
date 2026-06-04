# Go 数据库与 GORM 规范

## 1. 数据库

- **数据库类型**：PostgreSQL 15.1
- 连接信息通过配置注入，禁止硬编码。
- 写操作必须在 Service 层管理事务边界。

## 2. 表结构规范（强制）

- 表名：小写下划线，如 `sys_user`、`biz_order_item`。
- 主键：`id VARCHAR(64)`，由应用层生成（UUID/雪花 ID 等）。
- 审计字段（每表必备）：
  - `create_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP`
  - `update_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP`
  - `deleted INT NOT NULL DEFAULT 0`
- 表与关键字段必须有 COMMENT。
- 表结构必须与 `resources/schema/{表名}.sql` 中 DDL 一致。

## 3. Schema SQL 与基础数据（强制）

- 一表一文件：每张业务表对应 `resources/schema/{表名}.sql`。
- 文件必须包含：
  - 完整 `CREATE TABLE IF NOT EXISTS ...`；
  - 必要基础数据 `INSERT INTO ...`；
  - 若无基础数据，需注释说明。
- 禁止只改数据库不改 schema 文件。

## 4. Model 规范

- Model 字段命名使用驼峰，数据库列名使用下划线并通过标签映射。
- 审计字段需与 schema 保持同名同类型。
- 逻辑删除字段统一使用 `deleted`（0/1），避免混用布尔与时间戳语义。

## 5. Repository 规范

- Repository 仅负责数据访问，不承载业务规则。
- 查询条件优先使用参数化方式，禁止字符串拼接 SQL。
- 复杂查询可使用原生 SQL，但必须参数化并保留可读性。

## 6. 分页与查询

- 统一分页入参（如 `page`、`pageSize`），并设置上限防止全量拖垮。
- 条件查询要有默认排序（通常按 `create_time DESC`）。

## 7. 事务

- 多表写操作使用显式事务。
- 事务失败必须回滚并返回可观测错误。
