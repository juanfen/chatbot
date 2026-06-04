-- biz_user_voice: 用户上传的语音样本（磁盘文件 + 元数据）

CREATE TABLE IF NOT EXISTS biz_user_voice (
  id VARCHAR(64) PRIMARY KEY,
  user_id VARCHAR(64) NOT NULL,
  file_name VARCHAR(256) NOT NULL,
  upload_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  duration_seconds DOUBLE PRECISION NOT NULL DEFAULT 0,
  is_active INT NOT NULL DEFAULT 0,
  create_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted INT NOT NULL DEFAULT 0
);

COMMENT ON TABLE biz_user_voice IS '用户语音样本';
COMMENT ON COLUMN biz_user_voice.file_name IS '相对于 audio_target_path 的文件名，如 {uuid}.wav';
COMMENT ON COLUMN biz_user_voice.duration_seconds IS '音频时长（秒）';
COMMENT ON COLUMN biz_user_voice.is_active IS '是否当前激活：1 是 0 否';

CREATE INDEX IF NOT EXISTS idx_biz_user_voice_user ON biz_user_voice(user_id) WHERE deleted = 0;
