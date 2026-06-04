-- biz_package: 可售/可分配套餐（管理员配置）

CREATE TABLE IF NOT EXISTS biz_package (
  id VARCHAR(64) PRIMARY KEY,
  code VARCHAR(64) NOT NULL,
  name VARCHAR(256) NOT NULL,
  description VARCHAR(2000) NOT NULL DEFAULT '',
  quota_calls BIGINT NOT NULL DEFAULT 0,
  sort_no INT NOT NULL DEFAULT 0,
  status INT NOT NULL DEFAULT 1,
  create_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted INT NOT NULL DEFAULT 0
);

COMMENT ON TABLE biz_package IS '套餐定义';
COMMENT ON COLUMN biz_package.quota_calls IS '套餐包含的可调用次数';
CREATE UNIQUE INDEX IF NOT EXISTS uk_biz_package_code ON biz_package(code) WHERE deleted = 0;

INSERT INTO biz_package (id, code, name, description, quota_calls, sort_no, status, deleted)
SELECT '00000000-0000-0000-0000-000000000020', 'trial', '体验包', '新用户体验链路，少量调用次数', 20, 10, 1, 0
WHERE NOT EXISTS (SELECT 1 FROM biz_package WHERE code = 'trial' AND deleted = 0);

INSERT INTO biz_package (id, code, name, description, quota_calls, sort_no, status, deleted)
SELECT '00000000-0000-0000-0000-000000000021', 'standard', '标准包', '适合日常试用与轻量业务', 500, 20, 1, 0
WHERE NOT EXISTS (SELECT 1 FROM biz_package WHERE code = 'standard' AND deleted = 0);

INSERT INTO biz_package (id, code, name, description, quota_calls, sort_no, status, deleted)
SELECT '00000000-0000-0000-0000-000000000022', 'pro', '专业包', '更高调用额度', 5000, 30, 1, 0
WHERE NOT EXISTS (SELECT 1 FROM biz_package WHERE code = 'pro' AND deleted = 0);
