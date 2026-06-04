-- sys_user_role
-- base data: bind admin user with admin role

CREATE TABLE IF NOT EXISTS sys_user_role (
  id VARCHAR(64) PRIMARY KEY,
  user_id VARCHAR(64) NOT NULL,
  role_id VARCHAR(64) NOT NULL,
  create_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted INT NOT NULL DEFAULT 0
);

COMMENT ON TABLE sys_user_role IS '用户角色关联';
COMMENT ON COLUMN sys_user_role.id IS '主键';
COMMENT ON COLUMN sys_user_role.user_id IS '用户ID';
COMMENT ON COLUMN sys_user_role.role_id IS '角色ID';
COMMENT ON COLUMN sys_user_role.create_time IS '创建时间';
COMMENT ON COLUMN sys_user_role.update_time IS '更新时间';
COMMENT ON COLUMN sys_user_role.deleted IS '逻辑删除：0否 1是';

CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_user_role_pair ON sys_user_role(user_id, role_id) WHERE deleted = 0;

INSERT INTO sys_user_role (id, user_id, role_id, deleted)
VALUES ('00000000-0000-0000-0000-000000000020', '00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000002', 0)
ON CONFLICT DO NOTHING;

