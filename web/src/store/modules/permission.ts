import { defineStore } from 'pinia';
import type { RouteRecordRaw } from 'vue-router';
import cloneDeep from 'lodash/cloneDeep';

import router, { fixedRouterList, homepageRouterList } from '@/router';
import { store } from '@/store';

function hasPermission(route: RouteRecordRaw, permissions: string[]) {
  const requiredPermission = route.meta?.permission as string | undefined;
  if (!requiredPermission) return true;
  if (!permissions || !permissions.length) return false;
  return permissions.includes('all') || permissions.includes(requiredPermission);
}

function filterRoutesByPermission(routes: RouteRecordRaw[], permissions: string[]) {
  const routesCopy = cloneDeep(routes);

  const loop = (list: RouteRecordRaw[]): RouteRecordRaw[] => {
    return list
      .map((route) => {
        const hasOriginChildren = Array.isArray(route.children) && route.children.length > 0;
        const children: RouteRecordRaw[] = hasOriginChildren ? loop(route.children) : [];
        const current: RouteRecordRaw = { ...route, children };
        const canVisitCurrent = hasPermission(current, permissions);

        // 父级菜单如果没有任何可访问子路由，则不显示
        if (hasOriginChildren && children.length === 0) {
          return null;
        }

        if (children.length > 0) {
          return current;
        }

        return canVisitCurrent ? current : null;
      })
      .filter(Boolean) as RouteRecordRaw[];
  };

  return loop(routesCopy);
}

export const usePermissionStore = defineStore('permission', {
  state: () => ({
    whiteListRouters: ['/login'],
    routers: [],
    removeRoutes: [],
    asyncRoutes: [],
    isRoutesReady: false,
  }),
  actions: {
    async initRoutes(permissions: string[] = []) {
      const filterBaseRoutes = [...homepageRouterList, ...fixedRouterList];
      const accessedRouters = permissions.length ? filterRoutesByPermission(filterBaseRoutes, permissions) : filterBaseRoutes;

      // 在菜单展示全部路由
      this.routers = accessedRouters;
      // 在菜单只展示动态路由和首页
      // this.routers = [...homepageRouterList, ...accessedRouters];
      // 在菜单只展示动态路由
      // this.routers = [...accessedRouters];
    },
    async buildAsyncRoutes(permissions: string[] = []) {
      // 当前项目使用固定路由，不依赖后端菜单接口
      this.asyncRoutes = [];
      await this.initRoutes(permissions);
      this.isRoutesReady = true;
      return this.asyncRoutes;
    },
    async restoreRoutes() {
      // 不需要在此额外调用initRoutes更新侧边导肮内容，在登录后asyncRoutes为空会调用
      this.asyncRoutes.forEach((item: RouteRecordRaw) => {
        if (item.name) {
          router.removeRoute(item.name);
        }
      });
      this.asyncRoutes = [];
      this.routers = [];
      this.isRoutesReady = false;
    },
  },
});

export function getPermissionStore() {
  return usePermissionStore(store);
}
