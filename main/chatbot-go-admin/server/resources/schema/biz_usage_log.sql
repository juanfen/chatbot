-- biz_usage_log: 用户调用明细（用于首页图表与明细表）

CREATE TABLE IF NOT EXISTS biz_usage_log (
  id VARCHAR(64) PRIMARY KEY,
  user_id VARCHAR(64) NOT NULL,
  channel VARCHAR(64) NOT NULL DEFAULT 'experience',
  ref_text VARCHAR(512) NOT NULL DEFAULT '',
  create_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE biz_usage_log IS '调用流水';
CREATE INDEX IF NOT EXISTS idx_biz_usage_log_user_time ON biz_usage_log(user_id, create_time DESC);
