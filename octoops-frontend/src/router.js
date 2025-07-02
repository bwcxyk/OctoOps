import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/overview',
    component: () => import('./views/Overview.vue'),
    name: 'Overview'
  },
  {
    path: '/',
    redirect: { name: 'Overview' }
  },
  {
    path: '/batchtask',
    component: () => import('./views/seatunnel/batch/BatchTask.vue'),
    name: 'BatchTask'
  },
  {
    path: '/batchtask/new',
    component: () => import('./views/seatunnel/batch/BatchTaskEdit.vue'),
    name: 'BatchTaskNew'
  },
  {
    path: '/batchtask/edit/:id',
    component: () => import('./views/seatunnel/batch/BatchTaskEdit.vue'),
    name: 'BatchTaskEdit'
  },
  {
    path: '/streamtask',
    name: 'StreamTask',
    component: () => import('./views/seatunnel/stream/StreamTask.vue')
  },
  {
    path: '/streamtask/new',
    name: 'StreamTaskNew',
    component: () => import('./views/seatunnel/stream/StreamTaskEdit.vue')
  },
  {
    path: '/streamtask/edit/:id',
    name: 'StreamTaskEdit',
    component: () => import('./views/seatunnel/stream/StreamTaskEdit.vue')
  },
  {
    path: '/scheduler',
    component: () => import('./views/Scheduler.vue'),
    name: 'Scheduler'
  },
  {
    path: '/tasklog',
    component: () => import('./views/TaskLog.vue'),
    name: 'TaskLog'
  },
  {
    path: '/ecs-security-group',
    component: () => import('./views/EcsSecurityGroup.vue'),
    name: 'EcsSecurityGroup'
  },
  {
    path: '/alert-email',
    component: () => import('./views/alert/AlertEmail.vue'),
    name: 'AlertEmail'
  },
  {
    path: '/alert-robot',
    component: () => import('./views/alert/AlertRobot.vue'),
    name: 'AlertRobot'
  },
  {
    path: '/task/timer',
    component: () => import('./views/TaskTimer.vue'),
    name: 'TaskTimer'
  },
  {
    path: '/alert-group',
    component: () => import('./views/alert/AlertGroup.vue'),
    name: 'AlertGroup'
  },
  {
    path: '/alert-template',
    component: () => import('./views/alert/AlertTemplate.vue'),
    name: 'AlertTemplate'
  },
  {
    path: '/alert-channel',
    component: () => import('./views/alert/AlertChannel.vue'),
    name: 'AlertChannel'
  },
  // ... 其他路由 ...
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router 