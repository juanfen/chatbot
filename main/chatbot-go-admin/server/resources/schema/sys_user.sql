-- sys_user
-- base data: default admin user (username=admin, password=admin123). Please change password immediately.

CREATE TABLE IF NOT EXISTS sys_user (
  id VARCHAR(64) PRIMARY KEY,
  username VARCHAR(64) NOT NULL,
  password_hash VARCHAR(128) NOT NULL,
  nickname VARCHAR(128) NOT NULL DEFAULT '',
  status INT NOT NULL DEFAULT 1,
  is_admin INT NOT NULL DEFAULT 0,
  create_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted INT NOT NULL DEFAULT 0
);

COMMENT ON TABLE sys_user IS '系统用户';
COMMENT ON COLUMN sys_user.id IS '主键';
COMMENT ON COLUMN sys_user.username IS '用户名（唯一）';
COMMENT ON COLUMN sys_user.password_hash IS '密码哈希（bcrypt）';
COMMENT ON COLUMN sys_user.nickname IS '昵称';
COMMENT ON COLUMN sys_user.status IS '状态：1启用 0禁用';
COMMENT ON COLUMN sys_user.is_admin IS '是否管理员：1是 0否';
COMMENT ON COLUMN sys_user.create_time IS '创建时间';
COMMENT ON COLUMN sys_user.update_time IS '更新时间';
COMMENT ON COLUMN sys_user.deleted IS '逻辑删除：0否 1是';

CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_user_username ON sys_user(username) WHERE deleted = 0;

INSERT INTO sys_user (id, username, password_hash, nickname, status, is_admin, deleted)
VALUES ('00000000-0000-0000-0000-000000000001', 'admin', '$2a$10$rs2JIgjJfo4Cm30K4WZ5dOohnTMv4B3/y9oZT.s/4s3g7mb/kkh.i', '管理员', 1, 1, 0)
ON CONFLICT DO NOTHING;

