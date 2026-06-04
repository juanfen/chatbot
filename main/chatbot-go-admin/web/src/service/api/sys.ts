import { request } from '../request';

export function fetchSysUserList(params: { page?: number; pageSize?: number; keyword?: string }) {
  return request<Api.Sys.UserList>(
    {
    url: '/sys/user/list',
    method: 'get',
    params
    },
    { total: 0, list: [] }
  );
}

export function fetchSysUserCreate(data: {
  username: string;
  password: string;
  nickname?: string;
  isAdmin?: number;
  roleIds?: string[];
  trialQuota?: number | null;
  defaultPackageId?: string;
}) {
  return request<null>({ url: '/sys/user/create', method: 'post', data }, null);
}

export function fetchSysUserUpdate(data: {
  id: string;
  nickname?: string;
  password?: string;
  status?: number;
  isAdmin?: number;
  roleIds?: string[];
  trialQuota?: number | null;
  defaultPackageId?: string;
}) {
  return request<null>({ url: '/sys/user/update', method: 'post', data }, null);
}

export function fetchSysUserDelete(id: string) {
  return request<null>({ url: '/sys/user/delete', method: 'post', data: { id } }, null);
}

export function fetchSysRoleList() {
  return request<Api.Sys.RoleList>({ url: '/sys/role/list', method: 'get' }, { list: [] });
}

export function fetchSysRoleCreate(data: { code: string; name: string }) {
  return request<null>({ url: '/sys/role/create', method: 'post', data }, null);
}

export function fetchSysRoleUpdate(data: { id: string; name?: string; status: number }) {
  return request<null>({ url: '/sys/role/update', method: 'post', data }, null);
}

export function fetchSysRoleDelete(id: string) {
  return request<null>({ url: '/sys/role/delete', method: 'post', data: { id } }, null);
}

export function fetchSysRoleSetRoutes(data: { roleId: string; routeIds: string[] }) {
  return request<null>({ url: '/sys/role/setRoutes', method: 'post', data }, null);
}

export function fetchSysRouteList() {
  return request<Api.Sys.RouteList>({ url: '/sys/route/list', method: 'get' }, { list: [] });
}

export function fetchSysRouteCreate(data: {
  parentId?: string;
  name: string;
  path: string;
  component: string;
  title?: string;
  i18nKey?: string;
  icon?: string;
  orderNo?: number;
  hideInMenu?: number;
  keepAlive?: number;
  status?: number | null;
}) {
  return request<null>({ url: '/sys/route/create', method: 'post', data }, null);
}

export function fetchSysRouteUpdate(data: {
  id: string;
  parentId?: string;
  path?: string;
  component?: string;
  title?: string;
  i18nKey?: string;
  icon?: string;
  orderNo?: number;
  hideInMenu?: number;
  keepAlive?: number;
  status?: number;
}) {
  return request<null>({ url: '/sys/route/update', method: 'post', data }, null);
}

export function fetchSysRouteDelete(id: string) {
  return request<null>({ url: '/sys/route/delete', method: 'post', data: { id } }, null);
}

/**
 * 获取算法运行时配置（仅 8 个允许的顶层 key）。
 * admin 后端会反向代理到 xiaozhi-go-core 的 `/xiaozhi/v1/runtime/config`。
 */
export function fetchSysRuntimeConfig() {
  return request<Record<string, any>>({ url: '/sys/runtime-config', method: 'get' }, {});
}

/**
 * 提交补丁；后端深度合并、数组整体替换；只需提交想改动的字段。
 */
export function fetchSysRuntimeConfigPatch(patch: Record<string, any>) {
  return request<null>({ url: '/sys/runtime-config', method: 'post', data: patch }, null);
}
