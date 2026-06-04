-- sys_role
-- base data: admin role

CREATE TABLE IF NOT EXISTS sys_role (
  id VARCHAR(64) PRIMARY KEY,
  code VARCHAR(64) NOT NULL,
  name VARCHAR(128) NOT NULL,
  status INT NOT NULL DEFAULT 1,
  create_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted INT NOT NULL DEFAULT 0
);

COMMENT ON TABLE sys_role IS '系统角色';
COMMENT ON COLUMN sys_role.id IS '主键';
COMMENT ON COLUMN sys_role.code IS '角色编码（唯一）';
COMMENT ON COLUMN sys_role.name IS '角色名称';
COMMENT ON COLUMN sys_role.status IS '状态：1启用 0禁用';
COMMENT ON COLUMN sys_role.create_time IS '创建时间';
COMMENT ON COLUMN sys_role.update_time IS '更新时间';
COMMENT ON COLUMN sys_role.deleted IS '逻辑删除：0否 1是';

CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_role_code ON sys_role(code) WHERE deleted = 0;

INSERT INTO sys_role (id, code, name, status, deleted)
VALUES ('00000000-0000-0000-0000-000000000002', 'admin', '管理员', 1, 0)
ON CONFLICT DO NOTHING;

INSERT INTO sys_role (id, code, name, status, deleted)
VALUES ('00000000-0000-0000-0000-000000000003', 'member', '普通用户', 1, 0)
ON CONFLICT DO NOTHING;

