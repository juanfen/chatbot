import { request } from '../request';

export function fetchBizQuotaSummary() {
  return request<Api.Biz.QuotaSummary>(
    { url: '/biz/quota/summary', method: 'get' },
    {
      isUnlimited: false,
      packageId: '',
      packageTitle: '',
      totalQuota: 0,
      usedQuota: 0,
      remainingQuota: 0
    }
  );
}

export function fetchBizUsageTrend(days = 14) {
  return request<Api.Biz.UsageTrendPoint[]>(
    { url: '/biz/usage/trend', method: 'get', params: { days } },
    () => []
  );
}

export function fetchBizUsageList(params: { page?: number; pageSize?: number }) {
  return request<Api.Biz.UsageList>(
    { url: '/biz/usage/list', method: 'get', params },
    { total: 0, list: [] }
  );
}

export function fetchBizAnnouncements() {
  return request<Api.Biz.Announcement[]>({ url: '/biz/announcements', method: 'get' }, () => []);
}

export function fetchBizPackageCatalog() {
  return request<Api.Biz.PackageCatalogItem[]>({ url: '/biz/packages/catalog', method: 'get' }, () => []);
}

export function fetchBizPurchaseSimulate(data: { packageId: string }) {
  return request<Api.Biz.PurchaseResult>(
    { url: '/biz/purchase/simulate', method: 'post', data },
    {
      ok: false,
      packageTitle: '',
      totalQuota: 0,
      usedQuota: 0,
      remainingQuota: 0,
      message: ''
    }
  );
}

export function fetchBizSimulateCall() {
  return request<null>({ url: '/biz/usage/simulateCall', method: 'post' }, null);
}

/** 管理端：套餐列表 */
export function fetchSysPackageList() {
  return request<Api.Biz.SysPackage[]>({ url: '/sys/package/list', method: 'get' }, () => []);
}

export function fetchSysPackageCreate(data: {
  code: string;
  name: string;
  description?: string;
  quotaCalls: number;
  sortNo?: number;
  status?: number;
}) {
  return request<null>({ url: '/sys/package/create', method: 'post', data }, null);
}

export function fetchSysPackageUpdate(data: {
  id: string;
  code?: string;
  name?: string;
  description?: string;
  quotaCalls?: number;
  sortNo?: number;
  status?: number;
}) {
  return request<null>({ url: '/sys/package/update', method: 'post', data }, null);
}

export function fetchSysPackageDelete(id: string) {
  return request<null>({ url: '/sys/package/delete', method: 'post', data: { id } }, null);
}

export function fetchSysAnnouncementList() {
  return request<Api.Biz.SysAnnouncement[]>({ url: '/sys/announcement/list', method: 'get' }, () => []);
}

export function fetchSysAnnouncementCreate(data: {
  title: string;
  content: string;
  sortNo?: number;
  status?: number;
  publishTime?: string | null;
}) {
  return request<null>({ url: '/sys/announcement/create', method: 'post', data }, null);
}

export function fetchSysAnnouncementUpdate(data: {
  id: string;
  title?: string;
  content?: string;
  sortNo?: number;
  status?: number;
  publishTime?: string | null;
}) {
  return request<null>({ url: '/sys/announcement/update', method: 'post', data }, null);
}

export function fetchSysAnnouncementDelete(id: string) {
  return request<null>({ url: '/sys/announcement/delete', method: 'post', data: { id } }, null);
}
