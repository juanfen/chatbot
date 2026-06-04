-- 音色个性化文案：JSON 数组，每项不超过 20 字（由应用层校验）

ALTER TABLE biz_user_speaker
  ADD COLUMN IF NOT EXISTS personal_lines TEXT NOT NULL DEFAULT '[]';

COMMENT ON COLUMN biz_user_speaker.personal_lines IS '个性化内容列表（JSON 字符串数组，每项≤20字）';
