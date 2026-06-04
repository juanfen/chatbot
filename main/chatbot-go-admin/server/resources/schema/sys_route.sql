-- sys_route: 动态菜单/路由元数据（与前端 Elegant 路由 name、component 约定一致）

CREATE TABLE IF NOT EXISTS sys_route (
  id VARCHAR(64) PRIMARY KEY,
  parent_id VARCHAR(64) NOT NULL DEFAULT '',
  name VARCHAR(128) NOT NULL,
  path VARCHAR(512) NOT NULL,
  component VARCHAR(512) NOT NULL,
  title VARCHAR(256) NOT NULL DEFAULT '',
  i18n_key VARCHAR(256) NOT NULL DEFAULT '',
  icon VARCHAR(256) NOT NULL DEFAULT '',
  order_no INT NOT NULL DEFAULT 0,
  hide_in_menu INT NOT NULL DEFAULT 0,
  keep_alive INT NOT NULL DEFAULT 0,
  status INT NOT NULL DEFAULT 1,
  create_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted INT NOT NULL DEFAULT 0
);

COMMENT ON TABLE sys_route IS '系统菜单/路由';
COMMENT ON COLUMN sys_route.id IS '主键';
COMMENT ON COLUMN sys_route.parent_id IS '父级 id，空串表示一级';
COMMENT ON COLUMN sys_route.name IS 'Elegant 路由名，如 home、sys-user';
COMMENT ON COLUMN sys_route.path IS '路径，如 /home';
COMMENT ON COLUMN sys_route.component IS '如 layout.base$view.home';
COMMENT ON COLUMN sys_route.status IS '1 启用 0 禁用';

CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_route_name ON sys_route(name) WHERE deleted = 0;

-- seed routes (idempotent with partial unique index on name)
INSERT INTO sys_route (id, parent_id, name, path, component, title, i18n_key, icon, order_no, hide_in_menu, keep_alive, status, deleted)
SELECT '00000000-0000-0000-0000-000000000010', '', 'home', '/home', 'layout.base$view.home', 'home', 'route.home', 'mdi:monitor-dashboard', 1, 0, 0, 1, 0
WHERE NOT EXISTS (SELECT 1 FROM sys_route WHERE name = 'home' AND deleted = 0);

INSERT INTO sys_route (id, parent_id, name, path, component, title, i18n_key, icon, order_no, hide_in_menu, keep_alive, status, deleted)
SELECT '00000000-0000-0000-0000-000000000011', '', 'sys-user', '/sys-user', 'layout.base$view.sys-user', '用户管理', 'route.sys-user', 'mdi:account-group', 20, 0, 0, 1, 0
WHERE NOT EXISTS (SELECT 1 FROM sys_route WHERE name = 'sys-user' AND deleted = 0);

INSERT INTO sys_route (id, parent_id, name, path, component, title, i18n_key, icon, order_no, hide_in_menu, keep_alive, status, deleted)
SELECT '00000000-0000-0000-0000-000000000012', '', 'sys-role', '/sys-role', 'layout.base$view.sys-role', '角色管理', 'route.sys-role', 'mdi:shield-account', 30, 0, 0, 1, 0
WHERE NOT EXISTS (SELECT 1 FROM sys_route WHERE name = 'sys-role' AND deleted = 0);

INSERT INTO sys_route (id, parent_id, name, path, component, title, i18n_key, icon, order_no, hide_in_menu, keep_alive, status, deleted)
SELECT '00000000-0000-0000-0000-000000000013', '', 'sys-route', '/sys-route', 'layout.base$view.sys-route', '菜单路由', 'route.sys-route', 'mdi:family-tree', 40, 0, 0, 1, 0
WHERE NOT EXISTS (SELECT 1 FROM sys_route WHERE name = 'sys-route' AND deleted = 0);

INSERT INTO sys_route (id, parent_id, name, path, component, title, i18n_key, icon, order_no, hide_in_menu, keep_alive, status, deleted)
SELECT '00000000-0000-0000-0000-000000000014', '', 'package-mall', '/package-mall', 'layout.base$view.package-mall', '套餐购买', 'route.package-mall', 'mdi:shopping-outline', 8, 0, 0, 1, 0
WHERE NOT EXISTS (SELECT 1 FROM sys_route WHERE name = 'package-mall' AND deleted = 0);

INSERT INTO sys_route (id, parent_id, name, path, component, title, i18n_key, icon, order_no, hide_in_menu, keep_alive, status, deleted)
SELECT '00000000-0000-0000-0000-000000000015', '', 'sys-package', '/sys-package', 'layout.base$view.sys-package', '套餐配置', 'route.sys-package', 'mdi:package-variant', 45, 0, 0, 1, 0
WHERE NOT EXISTS (SELECT 1 FROM sys_route WHERE name = 'sys-package' AND deleted = 0);

INSERT INTO sys_route (id, parent_id, name, path, component, title, i18n_key, icon, order_no, hide_in_menu, keep_alive, status, deleted)
SELECT '00000000-0000-0000-0000-000000000016', '', 'sys-announcement', '/sys-announcement', 'layout.base$view.sys-announcement', '公告管理', 'route.sys-announcement', 'mdi:bullhorn-outline', 46, 0, 0, 1, 0
WHERE NOT EXISTS (SELECT 1 FROM sys_route WHERE name = 'sys-announcement' AND deleted = 0);

INSERT INTO sys_route (id, parent_id, name, path, component, title, i18n_key, icon, order_no, hide_in_menu, keep_alive, status, deleted)
SELECT '00000000-0000-0000-0000-000000000017', '', 'sys-runtime-config', '/sys-runtime-config', 'layout.base$view.sys-runtime-config', '算法运行时配置', 'route.sys-runtime-config', 'mdi:cog-sync-outline', 47, 0, 0, 1, 0
WHERE NOT EXISTS (SELECT 1 FROM sys_route WHERE name = 'sys-runtime-config' AND deleted = 0);

INSERT INTO sys_route (id, parent_id, name, path, component, title, i18n_key, icon, order_no, hide_in_menu, keep_alive, status, deleted)
SELECT '00000000-0000-0000-0000-000000000018', '', 'voice-recognition', '/voice-recognition', 'layout.base$view.voice-recognition', '音色识别', 'route.voice-recognition', 'mdi:account-voice', 48, 0, 0, 1, 0
WHERE NOT EXISTS (SELECT 1 FROM sys_route WHERE name = 'voice-recognition' AND deleted = 0);
