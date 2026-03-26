import { defineStore } from 'pinia';

import { getProfileApi, loginApi } from '@/api/auth';
import { usePermissionStore } from '@/store';
import type { UserInfo } from '@/types/interface';

const InitUserInfo: UserInfo = {
  name: '', // 用户名，用于展示在页面右上角头像处
  roles: [], // 前端权限模型使用 如果使用请配置modules/permission-fe.ts使用
  permissions: [],
};

export const useUserStore = defineStore('user', {
  state: () => ({
    token: '',
    userInfo: { ...InitUserInfo },
  }),
  getters: {
    roles: (state) => {
      return state.userInfo?.roles;
    },
    permissions: (state) => {
      return state.userInfo?.permissions || [];
    },
  },
  actions: {
    async login(userInfo: Record<string, unknown>) {
      const account = String(userInfo.account || '');
      const password = String(userInfo.password || '');

      const res = await loginApi({
        username: account,
        password,
      });

      this.token = res.token;
      this.userInfo = {
        name: res.user.nickname || res.user.username,
        roles: res.roles || [],
        permissions: res.permissions || [],
      };
    },
    async getUserInfo() {
      const res = await getProfileApi();
      this.userInfo = {
        name: res.user.nickname || res.user.username,
        roles: res.roles || [],
        permissions: res.permissions || [],
      };
    },
    async logout() {
      this.token = '';
      this.userInfo = { ...InitUserInfo };
    },
  },
  persist: {
    afterRestore: () => {
      const permissionStore = usePermissionStore();
      permissionStore.initRoutes();
    },
    key: 'user',
    paths: ['token'],
  },
});
