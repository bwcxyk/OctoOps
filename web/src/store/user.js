import { defineStore } from 'pinia'
import { loginApi, getProfileApi, getMenusApi } from '@/api/user'
import axios from 'axios'

export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    user: null,
    roles: [],
    permissions: [],
    menus: []
  }),
  actions: {
    async login({ username, password }) {
      try {
        const res = await loginApi({ username, password })
        if (res.code === 200) {
          this.token = res.data.token
          this.user = res.data.user
          this.roles = res.data.roles
          this.permissions = res.data.permissions
          localStorage.setItem('token', this.token)
          // 设置 axios 默认 header
          axios.defaults.headers.common['Authorization'] = 'Bearer ' + this.token
          await this.fetchMenus()
        } else {
          throw new Error(res.message)
        }
      } catch (err) {
        if (err.response && err.response.status === 401) {
          throw new Error('用户名或密码错误')
        } else {
          throw new Error('登录失败，请稍后重试')
        }
      }
    },
    async fetchUserInfo() {
      const res = await getProfileApi()
      if (res.code === 200) {
        this.user = res.data.user
        this.roles = res.data.roles
        this.permissions = res.data.permissions
        await this.fetchMenus()
      }
    },
    async fetchMenus() {
      const res = await getMenusApi()
      if (res.data && res.data.code === 200) {
        this.menus = res.data.data
      }
    },
    logout() {
      this.token = ''
      this.user = null
      this.roles = []
      this.permissions = []
      this.menus = []
      localStorage.removeItem('token')
      delete axios.defaults.headers.common['Authorization']
    }
  }
})
