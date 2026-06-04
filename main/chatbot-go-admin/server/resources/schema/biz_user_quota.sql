-- biz_user_quota: 用户当前套餐与额度快照

CREATE TABLE IF NOT EXISTS biz_user_quota (
  user_id VARCHAR(64) PRIMARY KEY,
  package_id VARCHAR(64) NOT NULL DEFAULT '',
  package_title VARCHAR(256) NOT NULL DEFAULT '',
  total_quota BIGINT NOT NULL DEFAULT 0,
  used_quota BIGINT NOT NULL DEFAULT 0,
  update_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE biz_user_quota IS '用户额度';
COMMENT ON COLUMN biz_user_quota.package_title IS '展示用套餐名称快照';
