-- biz_user_preference: 移动端「标准文本」等用户偏好

CREATE TABLE IF NOT EXISTS biz_user_preference (
  user_id VARCHAR(64) PRIMARY KEY,
  standard_text VARCHAR(8000) NOT NULL DEFAULT '',
  create_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE biz_user_preference IS '用户偏好（标准朗读文本等）';
COMMENT ON COLUMN biz_user_preference.user_id IS '用户 ID';
COMMENT ON COLUMN biz_user_preference.standard_text IS '录音时展示的标准文本';
