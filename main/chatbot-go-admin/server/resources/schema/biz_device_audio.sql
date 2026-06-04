-- biz_device_audio: 设备上报的 WAV 音频（二进制存库，按 expire/delete 时间生命周期管理）

CREATE TABLE IF NOT EXISTS biz_device_audio (
  id VARCHAR(64) PRIMARY KEY,
  mac_id VARCHAR(128) NOT NULL,
  audio_blob BYTEA NOT NULL,
  audio_len VARCHAR(32) NOT NULL,
  context TEXT NOT NULL DEFAULT '暂无内容',
  create_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  expire_time TIMESTAMP NOT NULL,
  delete_time TIMESTAMP NOT NULL,
  speaker_id VARCHAR(32) DEFAULT NULL,
  bound_user_id VARCHAR(64) DEFAULT NULL
);

-- 已有表补列（幂等）
ALTER TABLE biz_device_audio ADD COLUMN IF NOT EXISTS context TEXT NOT NULL DEFAULT '暂无内容';
ALTER TABLE biz_device_audio ADD COLUMN IF NOT EXISTS speaker_id VARCHAR(32) DEFAULT NULL;
ALTER TABLE biz_device_audio ADD COLUMN IF NOT EXISTS bound_user_id VARCHAR(64) DEFAULT NULL;

COMMENT ON TABLE biz_device_audio IS '设备上报音频';
COMMENT ON COLUMN biz_device_audio.mac_id IS '设备唯一标识';
COMMENT ON COLUMN biz_device_audio.audio_blob IS 'WAV 音频二进制';
COMMENT ON COLUMN biz_device_audio.audio_len IS '音频时长描述，如 5s';
COMMENT ON COLUMN biz_device_audio.context IS '音频对应文本内容';
COMMENT ON COLUMN biz_device_audio.expire_time IS '创建后 7 天，到期后查询不可见';
COMMENT ON COLUMN biz_device_audio.delete_time IS '创建后 14 天，到期后物理删除；绑定音色后为无限期';
COMMENT ON COLUMN biz_device_audio.speaker_id IS '绑定的音色 speaker_id';
COMMENT ON COLUMN biz_device_audio.bound_user_id IS '执行绑定的用户 ID';

CREATE INDEX IF NOT EXISTS idx_biz_device_audio_mac ON biz_device_audio(mac_id);
CREATE INDEX IF NOT EXISTS idx_biz_device_audio_speaker ON biz_device_audio(speaker_id);
CREATE INDEX IF NOT EXISTS idx_biz_device_audio_expire ON biz_device_audio(expire_time);
CREATE INDEX IF NOT EXISTS idx_biz_device_audio_delete ON biz_device_audio(delete_time);
