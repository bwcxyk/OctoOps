import { createRouter, createWebHistory } from 'vue-router'
import MainLayout from '@/layouts/MainLayout.vue'
import BlankLayout from '@/layouts/BlankLayout.vue'
import { hasPermission } from '@/utils/permission'
import { useUserStore } from '@/store/user'

const routes = [
  {
    path: '/',
    component: MainLayout,
    children: [
      { path: '', redirect: { name: 'Overview' } },
      { path: 'overview', component: () => import('./views/Overview.vue'), name: 'Overview' },
      { path: 'batchtask', component: () => import('./views/seatunnel/batch/BatchTask.vue'), name: 'BatchTask', meta: { permission: 'etl:batch' } },
      { path: 'batchtask/new', component: () => import('./views/seatunnel/batch/BatchTaskEdit.vue'), name: 'BatchTaskNew', meta: { permission: 'etl:batch' } },
      { path: 'batchtask/edit/:id', component: () => import('./views/seatunnel/batch/BatchTaskEdit.vue'), name: 'BatchTaskEdit', meta: { permission: 'etl:batch' } },
      { path: 'streamtask', name: 'StreamTask', component: () => import('./views/seatunnel/stream/StreamTask.vue'), meta: { permission: 'etl:stream' } },
      { path: 'streamtask/new', name: 'StreamTaskNew', component: () => import('./views/seatunnel/stream/StreamTaskEdit.vue'), meta: { permission: 'etl:stream' } },
      { path: 'streamtask/edit/:id', name: 'StreamTaskEdit', component: () => import('./views/seatunnel/stream/StreamTaskEdit.vue'), meta: { permission: 'etl:stream' } },
      { path: 'scheduler', component: () => import('./views/Scheduler.vue'), name: 'Scheduler', meta: { permission: 'task:scheduler' } },
      { path: 'tasklog', component: () => import('./views/TaskLog.vue'), name: 'TaskLog', meta: { permission: 'tasklog' } },
      { path: 'ecs-security-group', component: () => import('./views/EcsSecurityGroup.vue'), name: 'EcsSecurityGroup', meta: { permission: 'aliyun:ecs_sg' } },
      { path: 'alert-email', component: () => import('./views/alert/AlertEmail.vue'), name: 'AlertEmail', meta: { permission: 'alert:read' } },
      { path: 'alert-robot', component: () => import('./views/alert/AlertRobot.vue'), name: 'AlertRobot', meta: { permission: 'alert:read' } },
      { path: 'task/timer', component: () => import('./views/CustomTask.vue'), name: 'CustomTask', meta: { permission: 'task:schedule' } },
      { path: 'alert-group', component: () => import('./views/alert/AlertGroup.vue'), name: 'AlertGroup', meta: { permission: 'notify:group' } },
      { path: 'alert-template', component: () => import('./views/alert/AlertTemplate.vue'), name: 'AlertTemplate', meta: { permission: 'notify:template' } },
      { path: 'alert-channel', component: () => import('./views/alert/AlertChannel.vue'), name: 'AlertChannel', meta: { permission: 'notify:channel' } },
      { path: 'rbac/user', component: () => import('./views/rbac/UserManage.vue'), name: 'UserManage', meta: { permission: 'rbac:user:read' } },
      { path: 'rbac/role', component: () => import('./views/rbac/RoleManage.vue'), name: 'RoleManage', meta: { permission: 'rbac:role:read' } },
      { path: 'rbac/permission', component: () => import('./views/rbac/PermissionManage.vue'), name: 'PermissionManage', meta: { permission: 'rbac:permission:read' } },
      { path: '403', component: () => import('./views/403.vue'), name: 'Forbidden' },
    ]
  },
  {
    path: '/login',
    component: BlankLayout,
    children: [
      { path: '', name: 'Login', component: () => import('./views/Login.vue') }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 路由守卫：RBAC 权限控制
router.beforeEach(async (to, from, next) => {
  const userStore = useUserStore()
  const required = to.meta.permission

  // 未登录，只允许访问 /login
  if (!userStore.token) {
    if (to.matched.some(r => r.name === 'Login')) {
      return next()
    } else {
      return next('/login')
    }
  }

  // 已登录，访问 /login 自动跳转首页
  if (userStore.token && to.matched.some(r => r.name === 'Login')) {
    return next('/')
  }

  // 关键：如果权限数据为空，自动拉取
  if (userStore.permissions.length === 0 || userStore.roles.length === 0) {
    try {
      await userStore.fetchUserInfo()
    } catch (e) {
      userStore.logout()
      return next('/login')
    }
  }

  // RBAC 权限判断
  if (required && !hasPermission(required)) {
    return next('/403')
  }

  next()
})

export default router 