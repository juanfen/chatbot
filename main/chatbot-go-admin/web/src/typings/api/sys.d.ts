declare namespace Api {
  namespace Sys {
    interface User {
      id: string;
      username: string;
      nickname: string;
      status: number;
      isAdmin: number;
      roleIds: string[];
      totalQuota?: number;
      usedQuota?: number;
      packageTitle?: string;
    }

    interface UserList {
      total: number;
      list: User[];
    }

    interface Role {
      id: string;
      code: string;
      name: string;
      status: number;
      routeIds: string[];
    }

    interface RoleList {
      list: Role[];
    }

    interface RouteRow {
      id: string;
      parentId: string;
      name: string;
      path: string;
      component: string;
      title: string;
      i18nKey: string;
      icon: string;
      orderNo: number;
      hideInMenu: number;
      keepAlive: number;
      status: number;
    }

    interface RouteList {
      list: RouteRow[];
    }
  }
}
