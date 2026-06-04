-- biz_announcement: 官方公告（用户首页展示）

CREATE TABLE IF NOT EXISTS biz_announcement (
  id VARCHAR(64) PRIMARY KEY,
  title VARCHAR(512) NOT NULL,
  content VARCHAR(8000) NOT NULL DEFAULT '',
  sort_no INT NOT NULL DEFAULT 0,
  status INT NOT NULL DEFAULT 1,
  publish_time TIMESTAMP,
  create_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted INT NOT NULL DEFAULT 0
);

COMMENT ON TABLE biz_announcement IS '公告';
CREATE INDEX IF NOT EXISTS idx_biz_announcement_status ON biz_announcement(status, sort_no) WHERE deleted = 0;

INSERT INTO biz_announcement (id, title, content, sort_no, status, publish_time, deleted)
SELECT '00000000-0000-0000-0000-000000000040', '欢迎使用 ChatBot 管理后台', '您可在首页查看套餐与调用额度，前往「套餐购买」可体验模拟下单流程。', 1, 1, CURRENT_TIMESTAMP, 0
WHERE NOT EXISTS (SELECT 1 FROM biz_announcement WHERE id = '00000000-0000-0000-0000-000000000040' AND deleted = 0);
