-- biz_user_speaker: 用户创建的音色（speaker_id 为 32 位雪花 ID）

CREATE TABLE IF NOT EXISTS biz_user_speaker (
  speaker_id VARCHAR(32) PRIMARY KEY,
  user_id VARCHAR(64) NOT NULL,
  speaker_name VARCHAR(128) NOT NULL,
  personal_lines TEXT NOT NULL DEFAULT '[]',
  create_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE biz_user_speaker IS '用户音色配置';
COMMENT ON COLUMN biz_user_speaker.speaker_id IS '32 位雪花 ID';
COMMENT ON COLUMN biz_user_speaker.speaker_name IS '音色名称';
COMMENT ON COLUMN biz_user_speaker.personal_lines IS '个性化内容列表（JSON 字符串数组，每项≤20字）';

CREATE UNIQUE INDEX IF NOT EXISTS uk_biz_user_speaker_user_name ON biz_user_speaker(user_id, speaker_name);
CREATE INDEX IF NOT EXISTS idx_biz_user_speaker_user ON biz_user_speaker(user_id);
